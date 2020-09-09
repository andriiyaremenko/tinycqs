package query

import (
	"context"
)

type Query struct {
	Name string
	Body []byte
}

type Handler interface {
	QueryName() string
	Handle(ctx context.Context, query *Query) ([]byte, error)
}

type Demultiplexer interface {
	Handle(ctx context.Context, query *Query) ([]byte, error)
}
