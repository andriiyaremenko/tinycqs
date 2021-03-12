package command

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/andriiyaremenko/tinycqs/tracing"
	"github.com/google/uuid"
)

// Returns new Commands with Concurrency Limit equals to limit or error.
// Concurrency Limit is amount of Events that can be processed concurrently per each handler.
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

	if limit < 1 {
		return nil, LimitLessThanOne
	}

	return &commands{handlers: handlers, cLimit: limit}, nil
}

// Returns new Commands with Concurrency Limit equals to 0 or error.
// Concurrency Limit is amount of Events that can be processed concurrently.
func NewCommands(handlers ...Handler) (Commands, error) {
	return NewCommandsWithConcurrencyLimit(1, handlers...)
}

type commands struct {
	handlers []Handler
	cLimit   int
}

func (c *commands) MarshalJSON() ([]byte, error) {
	events := make([]string, 0, 1)
	for _, h := range c.handlers {
		events = append(events, h.EventType())
	}

	return json.Marshal(events)
}

func (c *commands) HandleOnly(ctx context.Context, event Event, only ...string) Event {
	withMetadata := AsEventWithMetadata(event)
	if withMetadata == nil {
		id := uuid.New().String()
		withMetadata = WithMetadata(event, tracing.M{EID: id, ECorrelationID: id, ECausationID: id})
	}

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

	rw := newEventRW(ctx)

	defer rw.Close()

	h.Handle(ctx, rw.GetWriter(withMetadata.Metadata()), withMetadata)

	for {
		select {
		case <-ctx.Done():
			return NewErrEvent(event, ctx.Err())
		case ev := <-rw.Read():
			if event == doneWriting {
				return Done(event)
			}

			return ev
		}
	}
}

func (c *commands) Handle(ctx context.Context, event Event) Event {
	rw := newEventRW(ctx)

	defer rw.Close()

	withMetadata := AsEventWithMetadata(event)
	if withMetadata == nil {
		id := uuid.New().String()
		withMetadata = WithMetadata(event, tracing.M{EID: id, ECorrelationID: id, ECausationID: id})
	}

	result := newResult(withMetadata)
	c.startWorkers(ctx, rw, result)

	if result.Err() != nil {
		return result
	}

	return Done(result)
}

func (c *commands) startWorkers(ctx context.Context, rw EventReader, result *result) {
	channels := make(map[string]chan EventWithMetadata)
	for _, h := range c.handlers {
		events := make(chan EventWithMetadata)
		channels[h.EventType()] = events
		for i := 0; i < c.cLimit; i++ {
			go func(h Handler) {
				for event := range events {
					h.Handle(ctx, rw.GetWriter(event.Metadata()), event)
				}
			}(h)
		}
	}

	done := make(chan struct{})

	go func() {
		var wg sync.WaitGroup

		event := result.event
		handle, ok := channels[event.EventType()]

		if !ok {
			result.errors.Append(&ErrCommandHandlerNotFound{event.EventType()})
			close(done)

			return
		}

		select {
		case <-ctx.Done():
			result.errors.Append(ctx.Err())
			close(done)

			return
		default:
		}

		wg.Add(1)
		handle <- event

		go func() {
			wg.Wait()
			close(done)
		}()

		for event := range rw.Read() {
			if err := ctx.Err(); err != nil {
				result.errors.Append(err)
			}

			if event == doneWriting {
				wg.Done()

				continue
			}

			if err := ctx.Err(); err != nil {
				continue
			}

			if event == nil {
				result.errors.Append(NilEvent)

				continue
			}

			if done := AsDoneEvent(event); done != nil {
				result.Append(done, event.Metadata())

				continue
			}

			handle, ok := channels[event.EventType()]
			err := event.Err()
			isUnhandledEvent := !ok && err == nil
			isUnhandledError := !ok && err != nil

			if isUnhandledEvent {
				result.errors.Append(&ErrCommandHandlerNotFound{event.EventType()})

				continue
			}

			if isUnhandledError {
				handle, ok = channels[CatchAllErrorEventType]
				if !ok {
					result.errors.Append(err)

					continue
				}
			}

			wg.Add(1)
			handle <- event
		}
	}()

	<-done
}

func (c *commands) getHandle(eventType string) (Handler, bool) {
	for _, h := range c.handlers {
		if h.EventType() == eventType {
			return h, true
		}
	}
	return nil, false
}

func (c *commands) sealed() {}
