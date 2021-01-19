package httprpc

import (
	"net/http"

	"github.com/andriiyaremenko/tinycqs/command"
	"github.com/andriiyaremenko/tinycqs/query"
	"github.com/google/uuid"
)

// Turns query.Queries into http.Handler.
// Every query.Handler handles Request with corresponding Method.
// Request.Props are passed to query.Queries.Handle as Event.Payload
func Queries(queries query.Queries) http.Handler {
	return &queriesHandler{queries}
}

type queriesHandler struct {
	queries query.Queries
}

func (h *queriesHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.NotFound(w, req)
		return
	}

	method, payload, err := getRequest(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	metadata, idKey, causationIDKey, correlationIDKey, hasMetadata := getMetadata(req)
	if hasMetadata {
		metadata = metadata.New(uuid.New().String())
	}

	if !hasMetadata {
		idKey = "RequestID"
		causationIDKey = "CausationID"
		correlationIDKey = "CorrelationID"
		id := uuid.New().String()
		metadata = command.M{EID: id, ECausationID: id, ECorrelationID: id}
	}

	w.Header().Add(idKey, metadata.ID())
	w.Header().Add(causationIDKey, metadata.CausationID())
	w.Header().Add(correlationIDKey, metadata.CorrelationID())

	resp, err := h.queries.Handle(req.Context(), method, payload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.Write(resp)
}
