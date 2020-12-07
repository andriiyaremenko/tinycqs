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
	handlers   []Handler
	channelCap int
}

func (hf *commands) HandleOnly(ctx context.Context, event Event, only ...string) Event {
	exists := false
	for _, c := range only {
		if exists = c == event.EventType(); exists {
			break
		}
	}

	if !exists {
		return Done
	}

	h, ok := hf.getHandle(event)
	if !ok {
		return NewErrEvent(event, &ErrCommandHandlerNotFound{event.EventType()})
	}

	evChan := h.Handle(ctx, event.Payload())

	for {
		select {
		case <-ctx.Done():
			return NewErrEvent(event, ctx.Err())
		case ev := <-evChan:
			if ev == nil {
				return Done
			}
			return ev
		}
	}
}

func (hf *commands) Handle(ctx context.Context, event Event) Event {
	ctx, cancel := context.WithCancel(ctx)

	defer cancel()

	h, ok := hf.getHandle(event)
	if !ok {
		return NewErrEvent(event, &ErrCommandHandlerNotFound{event.EventType()})
	}

	events := h.Handle(ctx, event.Payload())

	if err := hf.handleNext(ctx, events); err != nil {
		return NewErrEvent(event, err)
	}

	return Done
}

func (hf *commands) sealed() {}

func (hf *commands) handleNext(ctx context.Context, events <-chan Event) error {
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

			if err := event.Err(); err != nil {
				return err
			}

			h, ok := hf.getHandle(event)
			if !ok {
				return &ErrCommandHandlerNotFound{event.EventType()}
			}

			events = hf.mergeEvents(events, h.Handle(ctx, event.Payload()))
		}
	}
}

func (hf *commands) getHandle(event Event) (Handler, bool) {
	for _, h := range hf.handlers {
		if h.EventType() == event.EventType() {
			return h, true
		}
	}
	return nil, false
}

func (hf *commands) mergeEvents(cs ...<-chan Event) <-chan Event {
	out := make(chan Event, hf.channelCap)
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
