package query

import (
	"context"
)

type QueryHandler struct {
	QName      string
	HandleFunc func(ctx context.Context, w QueryResultWriter, payload []byte)
}

func (qh *QueryHandler) QueryName() string {
	return qh.QName
}

func (qh *QueryHandler) Handle(ctx context.Context, w QueryResultWriter, payload []byte) <-chan QueryResult {
	r := w.GetReader()

	go qh.HandleFunc(ctx, w, payload)

	return r.Read()
}

// Returns Handler with QueryName equals queryName.
// and Handle based on handle.
func QueryHandlerFunc(queryName string,
	handle func(context.Context, []byte) ([]byte, error)) Handler {
	return &queryHandler{queryName, handle}
}

type queryHandler struct {
	queryName string
	handle    func(context.Context, []byte) ([]byte, error)
}

func (ch *queryHandler) QueryName() string {
	return ch.queryName
}

func (ch *queryHandler) Handle(ctx context.Context, w QueryResultWriter, payload []byte) <-chan QueryResult {
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
