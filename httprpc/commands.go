package httprpc

import (
	"net/http"

	"github.com/andriiyaremenko/tinycqs/command"
	"github.com/google/uuid"
)

// Turns command.Commands into http.Handler.
// Every command.Handler handles Request with corresponding Method.
// Request.Props are passed to command.Commands.Handle as Event.Payload
func Commands(commands command.Commands) http.Handler {
	return &commandsHandler{commands}
}

type commandsHandler struct {
	commands command.Commands
}

func (h *commandsHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
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
	}

	ev = h.commands.Handle(req.Context(), ev)
	if withMetadata := command.AsEventWithMetadata(ev); !hasMetadata && withMetadata != nil {
		metadata = withMetadata.Metadata()
	}

	w.Header().Add(idKey, metadata.ID())
	w.Header().Add(causationIDKey, metadata.CausationID())
	w.Header().Add(correlationIDKey, metadata.CorrelationID())

	if err := ev.Err(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if !command.IsDone(ev, method) {
		w.Header().Add("Content-Type", "application/json")
		w.Write(ev.Payload())

		return
	}

	w.WriteHeader(http.StatusNoContent)
}
