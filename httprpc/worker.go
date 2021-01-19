package httprpc

import (
	"net/http"

	"github.com/andriiyaremenko/tinycqs/command"
	"github.com/google/uuid"
)

// Turns command.CommandsWorker into http.Handler.
// Every command.Handler handles Request with corresponding Method.
// Request.Props are passed to command.CommandsWorker.Handle as Event.Payload
func CommandsWorker(worker command.CommandsWorker) http.Handler {
	return &commandsWorkerHandler{worker}
}

type commandsWorkerHandler struct {
	worker command.CommandsWorker
}

func (h *commandsWorkerHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.NotFound(w, req)
		return
	}

	method, payload, err := getRequest(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var ev command.Event = command.E{EType: method, EPayload: payload}
	metadata, idKey, causationIDKey, correlationIDKey, hasMetadata := getMetadata(req)

	if hasMetadata {
		ev = command.WithMetadata(ev, metadata)
		metadata = metadata.New(uuid.New().String())
	}

	if !hasMetadata {
		idKey = "RequestID"
		causationIDKey = "CausationID"
		correlationIDKey = "CorrelationID"
		id := uuid.New().String()
		metadata = command.M{EID: id, ECausationID: id, ECorrelationID: id}
		ev = command.WithMetadata(ev, metadata)
	}

	w.Header().Add(idKey, metadata.ID())
	w.Header().Add(causationIDKey, metadata.CausationID())
	w.Header().Add(correlationIDKey, metadata.CorrelationID())

	if err := h.worker.Handle(ev); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)

	return
}
