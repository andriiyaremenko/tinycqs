package command

import (
	"context"
)

func CommandHandlerFunc(eventType string, handle func(context.Context, []byte) error) Handler {
	return &commandHandler{eventType, handle}
}

type commandHandler struct {
	eventType string
	handle    func(context.Context, []byte) error
}

func (ch *commandHandler) EventType() string {
	return ch.eventType
}

func (ch *commandHandler) Handle(ctx context.Context, payload []byte) error {
	return ch.handle(ctx, payload)
}
