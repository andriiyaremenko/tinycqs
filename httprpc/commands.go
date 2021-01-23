package httprpc

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/andriiyaremenko/tinycqs/command"
)

// Turns command.Commands into http.Handler.
// Every command.Handler handles Request with corresponding Method.
// Request.Props are passed to command.Commands.Handle as Event.Payload.
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

	var ev command.Event = command.E{EType: reqModel.Method, EPayload: payload}
	ev = h.commands.Handle(req.Context(), command.WithMetadata(ev, metadata))

	err = ev.Err()
	methodNotSupported := new(command.ErrCommandHandlerNotFound)

	if errors.As(err, &methodNotSupported) {
		errResponse = reqModel.NewErrorResponse(MethodNotFound, err.Error(), nil)
	}

	if errResponse == nil && err != nil {
		errResponse = reqModel.NewErrorResponse(InternalApplicationError, err.Error(), nil)
	}

	if errResponse != nil {
		writeErrorResponse(w, errResponse)

		return
	}

	if reqModel.ID != nil {
		result := make(map[string]interface{})
		result["message"] = ev.EventType()
		result["params"] = reqModel.Params

		b, err := json.Marshal(reqModel.NewResponse(result))
		if err != nil {
			writeErrorResponse(w, reqModel.NewErrorResponse(InternalApplicationError, err.Error(), nil))

			return
		}

		w.Write(b)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
