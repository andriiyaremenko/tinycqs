package query

import (
	"context"
)

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

func (ch *queryHandler) Handle(ctx context.Context, payload []byte) ([]byte, error) {
	return ch.handle(ctx, payload)
}
