package command

import (
	"context"
	"sync"
)

func NewCommandsWithConcurrencyLimit(limit int, handlers ...Handler) (Commands, error) {
	for _, h := range handlers {
		if h.EventType() == "" {
			return nil, &ErrIncorrectHandler{h}
		}
	}

	return &commands{handlers, limit}, nil
}

func NewCommands(handlers ...Handler) (Commands, error) {
	for _, h := range handlers {
		if h.EventType() == "" {
			return nil, &ErrIncorrectHandler{h}
		}
	}

	return &commands{handlers, 0}, nil
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

	h, ok := c.getHandle(event)
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

	h, ok := c.getHandle(event)
	if !ok {
		return NewErrEvent(event, &ErrCommandHandlerNotFound{event.EventType()})
	}

	events := h.Handle(ctx, event)

	if err := c.handleNext(ctx, events); err != nil {
		return NewErrEvent(event, err)
	}

	return Done
}

func (c *commands) sealed() {}

func (c *commands) handleNext(ctx context.Context, events <-chan Event) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case event, ok := <-events:
			if !ok {
				return nil
			}

			if event == nil || event == Done {
				continue
			}

			h, ok := c.getHandle(event)
			err := event.Err()
			isUnhandledEvent := !ok && err == nil
			isUnhandledError := !ok && err != nil

			if isUnhandledEvent {
				return &ErrCommandHandlerNotFound{event.EventType()}
			}

			if isUnhandledError {
				return err
			}

			events = c.mergeEvents(events, h.Handle(ctx, event))
		}
	}
}

func (c *commands) getHandle(event Event) (Handler, bool) {
	for _, h := range c.handlers {
		if h.EventType() == event.EventType() {
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
