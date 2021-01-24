package jsonrpc

import (
	"net/http"

	"github.com/andriiyaremenko/tinycqs/query"
)

// Turns query.Queries into http.Handler.
// Every query.Handler handles Request with corresponding Method.
// Request.Props are passed to query.Queries.Handle as Event.Payload.
// Does not supports only JSON RPC Notifications.
func Queries(queries query.Queries) http.Handler {
	return &Handler{Queries: queries}
}
