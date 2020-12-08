package command

import (
	"context"
	"sync"
)

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

	return &commands{handlers, limit}, nil
}

func NewCommands(handlers ...Handler) (Commands, error) {
	return NewCommandsWithConcurrencyLimit(0, handlers...)
}

type commands struct {
	handlers         []Handler
	concurrencyLimit int
}

func (c *commands) ConcurrencyLimit() int {
	return c.concurrencyLimit
}
func (c *commands) SetConcurrencyLimit(limit int) {
	c.concurrencyLimit = limit
}

func (c *commands) HandleOnly(ctx context.Context, event Event, only ...string) Event {
	exists := false
	for _, c := range only {
		if exists = c == event.EventType(); exists {
			break
		}
	}

	if !exists {
		return Done
	}

	h, ok := c.getHandle(event.EventType())
	if !ok {
		return NewErrEvent(event, &ErrCommandHandlerNotFound{event.EventType()})
	}

	for {
		select {
		case <-ctx.Done():
			return NewErrEvent(event, ctx.Err())
		case ev := <-h.Handle(ctx, event):
			if ev == nil {
				return Done
			}

			return ev
		}
	}
}

func (c *commands) Handle(ctx context.Context, event Event) Event {
	ctx, cancel := context.WithCancel(ctx)

	defer cancel()

	h, ok := c.getHandle(event.EventType())
	if !ok {
		return NewErrEvent(event, &ErrCommandHandlerNotFound{event.EventType()})
	}

	events := h.Handle(ctx, event)

	return c.handleNext(ctx, event, events)
}

func (c *commands) sealed() {}

func (c *commands) handleNext(ctx context.Context, initialEvent Event, events <-chan Event) Event {
	for {
		select {
		case <-ctx.Done():
			return NewErrEvent(initialEvent, ctx.Err())
		case event, ok := <-events:
			if !ok {
				return Done
			}

			if event == nil || event == Done {
				continue
			}

			h, ok := c.getHandle(event.EventType())
			err := event.Err()
			isUnhandledEvent := !ok && err == nil
			isUnhandledError := !ok && err != nil

			if isUnhandledEvent {
				return NewErrEvent(event, &ErrCommandHandlerNotFound{event.EventType()})
			}

			if isUnhandledError {
				h, ok = c.getHandle(CatchAllErrorEventType)
				if !ok {
					return event
				}
			}

			events = c.mergeEvents(events, h.Handle(ctx, event))
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

func (c *commands) mergeEvents(cs ...<-chan Event) <-chan Event {
	out := make(chan Event, c.concurrencyLimit)
	var wg sync.WaitGroup

	wg.Add(len(cs))

	for _, c := range cs {
		go func(c <-chan Event) {
			for v := range c {
				out <- v
			}

			wg.Done()
		}(c)
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}
