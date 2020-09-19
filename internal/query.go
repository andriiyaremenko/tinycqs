package internal

import (
	"context"

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
