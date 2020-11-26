package internal

import (
	"context"
	"encoding/json"

	q "github.com/andriiyaremenko/tinycqs/query"
)

type QueryDemultiplexer struct {
	handlers []q.Handler
}

func (hf *QueryDemultiplexer) Handle(ctx context.Context,
	query string, payload []byte) ([]byte, error) {
	for _, h := range hf.handlers {
		if h.QueryName() == query {
			return h.Handle(ctx, payload)
		}
	}

	return nil, q.NewErrQueryHandlerNotFound(query)
}

func (hf *QueryDemultiplexer) HandleJSONEncoded(ctx context.Context,
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

func NewQueryDemultiplexer(handlers ...q.Handler) q.Demultiplexer {
	return &QueryDemultiplexer{handlers}
}

type QueryHandler struct {
	queryName string
	handle    func(context.Context, []byte) ([]byte, error)
}

func (ch *QueryHandler) QueryName() string {
	return ch.queryName
}

func (ch *QueryHandler) Handle(ctx context.Context, payload []byte) ([]byte, error) {
	return ch.handle(ctx, payload)
}

func QueryHandlerFunc(queryName string,
	handle func(context.Context, []byte) ([]byte, error)) q.Handler {
	return &QueryHandler{queryName, handle}
}
