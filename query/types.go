package query

import (
	"context"
)

type QueryWriterWithCancel func() (write func(QueryResult), done func())

// Result returned by query.Handler.
type QueryResult interface {
	// Query name.
	QueryName() string
	// Information returned by executing query.
	Body() []byte
	// Error caused by executing query.
	Err() error
	// Unmarshals body if it is correct json string.
	// Returns error otherwise.
	UnmarshalJSONBody(v interface{}) error
}

// Handles single type of query.
type Handler interface {
	// Query name.
	QueryName() string
	// Method to handle query.
	// Returns result of query execution.
	Handle(ctx context.Context, payload []byte) <-chan QueryResult
}

// Interface to handle queries.
type Queries interface {
	// Handles query regardless of its type.
	// Returns result of this query execution.
	Handle(ctx context.Context, query string, payload []byte) <-chan QueryResult
}
