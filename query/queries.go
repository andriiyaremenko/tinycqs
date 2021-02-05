package query

import (
	"context"
	"encoding/json"
)

// Returns new Queries or error.
func NewQueries(handlers ...Handler) (Queries, error) {
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

func (q *queries) Handle(ctx context.Context,
	query string, payload []byte) ([]byte, error) {
	for _, h := range q.handlers {
		if h.QueryName() == query {
			return h.Handle(ctx, payload)
		}
	}

	return nil, NewErrQueryHandlerNotFound(query)
}

func (q *queries) HandleJSONEncoded(ctx context.Context,
	query string, v interface{}, payload []byte) error {
	b, err := q.Handle(ctx, query, payload)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(b, v); err != nil {
		return err
	}

	return nil
}

func (q *queries) MarshalJSON() ([]byte, error) {
	events := make([]string, 0, 1)
	for _, q := range q.handlers {
		events = append(events, q.QueryName())
	}

	return json.Marshal(events)
}
