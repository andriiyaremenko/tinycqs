package query

import (
	"context"
	"encoding/json"
)

// Returns new Queries or error.
func New(handlers ...Handler) (Queries, error) {
	for _, h := range handlers {
		if h.QueryName() == "" {
			return nil, &ErrIncorrectHandler{h}
		}
	}

	return &queries{handlers}, nil
}

type queries struct {
	handlers []Handler
}

func (q *queries) Handle(ctx context.Context, query string, payload []byte) <-chan Result {
	w := NewQueryResultWriter()
	for _, h := range q.handlers {
		if h.QueryName() == query {
			return h.Handle(ctx, w, payload)
		}
	}

	r := w.GetReader()
	go func() {
		w.Write(Q{
			Name:  query,
			B:     nil,
			Error: NewErrQueryHandlerNotFound(query)})

		w.Done()
	}()

	return r.Read()
}

func (q *queries) MarshalJSON() ([]byte, error) {
	events := make([]string, 0, 1)
	for _, q := range q.handlers {
		events = append(events, q.QueryName())
	}

	return json.Marshal(events)
}
