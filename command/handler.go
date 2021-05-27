package command

import (
	"context"
)

// *BaseHandler implements Handler.
type BaseHandler struct {
	Type       string
	HandleFunc func(ctx context.Context, w EventWriter, e Event)
}

// Returns EType.
func (ch *BaseHandler) EventType() string {
	return ch.Type
}

// Runs HandleFunc.
func (ch *BaseHandler) Handle(ctx context.Context, w EventWriter, event Event) {
	ch.HandleFunc(ctx, w, event)
}

// Returns Handler with EventType equals eventType.
// and Handle based on handle.
func HandlerFunc(eventType string, handle func(context.Context, []byte) error) Handler {
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
	defer w.Done()
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
