package command

import (
	"context"
)

type CommandHandler struct {
	EType      string
	HandleFunc func(ctx context.Context, payload []byte) <-chan Event
}

func (ch *CommandHandler) EventType() string {
	return ch.EType
}

func (ch *CommandHandler) Handle(ctx context.Context, payload []byte) <-chan Event {
	return ch.HandleFunc(ctx, payload)
}

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

func (ch *commandHandler) Handle(ctx context.Context, payload []byte) <-chan Event {
	events := make(chan Event)

	go func() {
		defer close(events)
		for {
			select {
			case <-ctx.Done():
				events <- NewErrEvent(&E{EType: ch.eventType, EPayload: payload}, ctx.Err())

				return
			default:
				if err := ch.handle(ctx, payload); err != nil {
					events <- NewErrEvent(&E{EType: ch.eventType, EPayload: payload}, err)

					return
				}

				events <- Done
				return
			}
		}
	}()

	return events
}
