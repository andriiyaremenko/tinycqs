package command

import (
	"context"
)

type CommandHandler struct {
	EType      string
	HandleFunc func(ctx context.Context, event Event) <-chan Event
}

func (ch *CommandHandler) EventType() string {
	return ch.EType
}

func (ch *CommandHandler) Handle(ctx context.Context, event Event) <-chan Event {
	return ch.HandleFunc(ctx, event)
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

func (ch *commandHandler) Handle(ctx context.Context, event Event) <-chan Event {
	events := make(chan Event)

	go func() {
		defer close(events)
		for {
			select {
			case <-ctx.Done():
				events <- NewErrEvent(event, ctx.Err())

				return
			default:
				if err := ch.handle(ctx, event.Payload()); err != nil {
					events <- NewErrEvent(event, err)

					return
				}

				events <- Done
				return
			}
		}
	}()

	return events
}
