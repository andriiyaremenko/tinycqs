package query

import (
	"context"
)

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

func (ch *queryHandler) Handle(ctx context.Context, payload []byte) ([]byte, error) {
	return ch.handle(ctx, payload)
}
