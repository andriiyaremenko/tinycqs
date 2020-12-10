package command

import (
	"context"
)

type CommandHandler struct {
	EType      string
	HandleFunc func(ctx context.Context, w EventWriter, e Event)
}

func (ch *CommandHandler) EventType() string {
	return ch.EType
}

func (ch *CommandHandler) Handle(ctx context.Context, w EventWriter, event Event) {
	ch.HandleFunc(ctx, w, event)
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

func (ch *commandHandler) Handle(ctx context.Context, w EventWriter, event Event) {
	go func() {
		defer w.Done()
		for {
			select {
			case <-ctx.Done():
				w.Write(NewErrEvent(event, ctx.Err()))

				return
			default:
				if err := ch.handle(ctx, event.Payload()); err != nil {
					w.Write(NewErrEvent(event, err))

					return
				}

				return
			}
		}
	}()
}
