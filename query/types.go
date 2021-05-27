package query

import (
	"context"
)

// Result returned by query.Handler.
type Result interface {
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

// Serves to read query results.
type ResultReader interface {
	// Reads query results.
	Read() <-chan Result
}

// Serves to write query results from handlers
// Do NOT forget to call Done() when finished writing.
type ResultWriter interface {
	// Writes a query result
	Write(Result)
	// Signals that handler is done writing results.
	Done()

	// Returns QueryResultReader based on this writer.
	GetReader() ResultReader
}

// Handles single type of query.
type Handler interface {
	// Query name.
	QueryName() string
	// Method to handle query.
	// Returns result of query execution.
	Handle(ctx context.Context, w ResultWriter, payload []byte) <-chan Result
}

// Interface to handle queries.
type Queries interface {
	// Handles query regardless of its type.
	// Returns result of this query execution.
	Handle(ctx context.Context, query string, payload []byte) <-chan Result
}
