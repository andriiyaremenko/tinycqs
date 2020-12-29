package command

import (
	"context"
	"sync"
)

// Returns new `Commands` with Concurrency Limit equals to `limit` or `error`
// Concurrency Limit is amount of `Event`s that can be processed concurrently
func NewCommandsWithConcurrencyLimit(limit int, handlers ...Handler) (Commands, error) {
	globalErrHandlersN := 0
	for _, h := range handlers {
		if h.EventType() == CatchAllErrorEventType {
			globalErrHandlersN++
		}

		if h.EventType() == "" {
			return nil, &ErrIncorrectHandler{h}
		}
	}

	if globalErrHandlersN > 1 {
		return nil, MoreThanOneCatchAllErrorHandler
	}

	return &commands{handlers: handlers, cLimit: limit}, nil
}

// Returns new `Commands` with Concurrency Limit equals to `0` or `error`
// Concurrency Limit is amount of `Event`s that can be processed concurrently
func NewCommands(handlers ...Handler) (Commands, error) {
	return NewCommandsWithConcurrencyLimit(0, handlers...)
}

type commands struct {
	handlers []Handler
	cLimit   int
}

func (c *commands) HandleOnly(ctx context.Context, event Event, only ...string) Event {
	exists := false
	for _, c := range only {
		if exists = c == event.EventType(); exists {
			break
		}
	}

	if !exists {
		return Done(event)
	}

	h, ok := c.getHandle(event.EventType())
	if !ok {
		return NewErrEvent(event, &ErrCommandHandlerNotFound{event.EventType()})
	}

	rw := newEventRW(ctx, c.cLimit)

	defer rw.Close()

	h.Handle(ctx, rw.GetWriter(), event)

	for {
		select {
		case <-ctx.Done():
			return NewErrEvent(event, ctx.Err())
		case ev := <-rw.Read():
			if ev == nil {
				return Done(event)
			}

			return ev
		}
	}
}

func (c *commands) Handle(ctx context.Context, event Event) Event {
	rw := newEventRW(ctx, c.cLimit)

	defer rw.Close()

	return c.handleNext(ctx, event, rw)
}

func (c *commands) sealed() {}

func (c *commands) handleNext(ctx context.Context, initialEvent Event, rw EventReader) Event {
	h, ok := c.getHandle(initialEvent.EventType())
	if !ok {
		return NewErrEvent(initialEvent, &ErrCommandHandlerNotFound{initialEvent.EventType()})
	}

	var wg sync.WaitGroup

	wg.Add(1)
	h.Handle(ctx, rw.GetWriter(), initialEvent)

	errAggregated := NewErrAggregatedEvent(initialEvent)
	resultCh := make(chan Event)

	go func() {
		wg.Wait()

		resultCh <- errAggregated

		close(resultCh)
	}()

	for {
		select {
		case <-ctx.Done():
			return NewErrEvent(initialEvent, ctx.Err())
		case event := <-rw.Read():
			if event == doneWriting {
				wg.Done()

				continue
			}

			if event == nil {
				return NilEvent
			}

			h, ok := c.getHandle(event.EventType())
			err := event.Err()
			isUnhandledEvent := !ok && err == nil
			isUnhandledError := !ok && err != nil

			if isUnhandledEvent {
				errAggregated.Append(&ErrCommandHandlerNotFound{event.EventType()})

				continue
			}

			if isUnhandledError {
				h, ok = c.getHandle(CatchAllErrorEventType)
				if !ok {
					errAggregated.Append(err)

					continue
				}
			}

			select {
			case <-ctx.Done():
				continue
			default:
				wg.Add(1)
				h.Handle(ctx, rw.GetWriter(), event)
			}
		case r := <-resultCh:
			if r.Err() == nil {
				return Done(initialEvent)
			}

			return r
		}
	}
}

func (c *commands) getHandle(eventType string) (Handler, bool) {
	for _, h := range c.handlers {
		if h.EventType() == eventType {
			return h, true
		}
	}
	return nil, false
}
