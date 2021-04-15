package query

import (
	"context"
)

type QueryHandler struct {
	QName      string
	HandleFunc func(ctx context.Context, getWriter QueryWriterWithCancel, payload []byte)
}

func (qh *QueryHandler) QueryName() string {
	return qh.QName
}

func (qh *QueryHandler) Handle(ctx context.Context, payload []byte) <-chan QueryResult {
	resultCh := make(chan QueryResult)
	getWriter := func() (func(QueryResult), func()) {
		return func(queryResult QueryResult) { resultCh <- queryResult }, func() { close(resultCh) }
	}

	go qh.HandleFunc(ctx, getWriter, payload)

	return resultCh
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

func (ch *queryHandler) Handle(ctx context.Context, payload []byte) <-chan QueryResult {
	result := make(chan QueryResult)

	go func() {
		payload, err := ch.handle(ctx, payload)

		result <- Q{
			Name:  ch.queryName,
			B:     payload,
			Error: err}

		close(result)
	}()

	return result
}
