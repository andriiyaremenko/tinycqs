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

func (hf *queries) Handle(ctx context.Context,
	query string, payload []byte) ([]byte, error) {
	for _, h := range hf.handlers {
		if h.QueryName() == query {
			return h.Handle(ctx, payload)
		}
	}

	return nil, NewErrQueryHandlerNotFound(query)
}

func (hf *queries) HandleJSONEncoded(ctx context.Context,
	query string, v interface{}, payload []byte) error {
	b, err := hf.Handle(ctx, query, payload)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(b, v); err != nil {
		return err
	}

	return nil
}
