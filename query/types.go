package query

import (
	"context"
)

// Handles single type of query
type Handler interface {
	// Query name
	QueryName() string
	// Method to handle query
	// Returns result of `query` execution
	Handle(ctx context.Context, payload []byte) ([]byte, error)
}

// Interface to handle queries
type Queries interface {
	// Handles `query` regardless of its type
	// Returns result of this `query` execution
	Handle(ctx context.Context, query string, payload []byte) ([]byte, error)
	// Handles `query` regardless of its type
	// Returns result of this `query` execution by unmarshalling it from JSON into `v`
	HandleJSONEncoded(ctx context.Context, query string, v interface{}, payload []byte) error
}
