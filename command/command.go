package command

import (
	"context"
)

func NewCommand(handlers ...Handler) Command {
	return &command{handlers}
}

type command struct {
	handlers []Handler
}

func (hf *command) Handle(ctx context.Context, event Event) error {
	for _, h := range hf.handlers {
		if h.EventType() == event.EventType() {
			return h.Handle(ctx, event.Payload())
		}
	}

	return NewErrCommandHandlerNotFound(event.EventType())
}

func (hf *command) HandleOnly(ctx context.Context, event Event, only ...string) error {
	exists := false
	for _, c := range only {
		if exists = c == event.EventType(); exists {
			break
		}
	}

	if !exists {
		return nil
	}

	return hf.Handle(ctx, event)
}
