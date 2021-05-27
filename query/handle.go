package query

import (
	"context"
)

type BaseHandler struct {
	Name       string
	HandleFunc func(ctx context.Context, w ResultWriter, payload []byte)
}

func (qh *BaseHandler) QueryName() string {
	return qh.Name
}

func (qh *BaseHandler) Handle(ctx context.Context, w ResultWriter, payload []byte) <-chan Result {
	r := w.GetReader()

	go qh.HandleFunc(ctx, w, payload)

	return r.Read()
}

// Returns Handler with QueryName equals queryName.
// and Handle based on handle.
func HandlerFunc(queryName string, handle func(context.Context, []byte) ([]byte, error)) Handler {
	return &queryHandler{queryName, handle}
}

type queryHandler struct {
	queryName string
	handle    func(context.Context, []byte) ([]byte, error)
}

func (ch *queryHandler) QueryName() string {
	return ch.queryName
}

func (ch *queryHandler) Handle(ctx context.Context, w ResultWriter, payload []byte) <-chan Result {
	r := w.GetReader()

	go func() {
		payload, err := ch.handle(ctx, payload)

		w.Write(Q{
			Name:  ch.queryName,
			B:     payload,
			Error: err})

		w.Done()
	}()

	return r.Read()
}
