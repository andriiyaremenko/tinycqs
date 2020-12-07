package query

import (
	"context"
)

type Handler interface {
	QueryName() string
	Handle(ctx context.Context, payload []byte) ([]byte, error)
}

type Queries interface {
	Handle(ctx context.Context, query string, payload []byte) ([]byte, error)
	HandleJSONEncoded(ctx context.Context, query string, v interface{}, payload []byte) error
}
