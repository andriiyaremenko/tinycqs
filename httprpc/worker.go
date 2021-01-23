package httprpc

import (
	"errors"
	"net/http"

	"github.com/andriiyaremenko/tinycqs/command"
)

// Turns command.CommandsWorker into http.Handler.
// Every command.Handler handles Request with corresponding Method.
// Request.Props are passed to command.CommandsWorker.Handle as Event.Payload.
// Supports only JSON RPC Notifications.
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

	metadata := addMetadata(w, req)

	w.Header().Add("Content-Type", "application/json")

	var errResponse *ErrorResponse
	reqModel, payload, errCode, err := getRequest(req)

	if err != nil {
		errResponse := new(ErrorResponse)
		errResponse.Version = ProtocolVersion
		errResponse.Error = Error{Code: errCode, Message: err.Error()}

		writeErrorResponse(w, errResponse)

		return
	}

	if reqModel.ID != nil {
		writeErrorResponse(w,
			reqModel.NewErrorResponse(InternalError, "CommandsWorker supports only JSON-RPC Notifications", nil))

		return
	}

	var ev command.Event = command.E{EType: reqModel.Method, EPayload: payload}

	err = h.worker.Handle(command.WithMetadata(ev, metadata))
	methodNotSupported := new(command.ErrCommandHandlerNotFound)

	if errors.As(err, &methodNotSupported) {
		errResponse = reqModel.NewErrorResponse(MethodNotFound, err.Error(), nil)
	}

	if errResponse == nil && err != nil {
		errResponse = reqModel.NewErrorResponse(InternalError, err.Error(), nil)
	}

	if errResponse != nil {
		writeErrorResponse(w, errResponse)

		return
	}

	w.WriteHeader(http.StatusNoContent)
}
