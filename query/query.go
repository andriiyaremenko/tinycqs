package query

import (
	"context"
	"encoding/json"
)

func NewQuery(handlers ...Handler) Query {
	return &query{handlers}
}

type query struct {
	handlers []Handler
}

func (hf *query) Handle(ctx context.Context,
	query string, payload []byte) ([]byte, error) {
	for _, h := range hf.handlers {
		if h.QueryName() == query {
			return h.Handle(ctx, payload)
		}
	}

	return nil, NewErrQueryHandlerNotFound(query)
}

func (hf *query) HandleJSONEncoded(ctx context.Context,
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
