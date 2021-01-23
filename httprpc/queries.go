package httprpc

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/andriiyaremenko/tinycqs/query"
)

// Turns query.Queries into http.Handler.
// Every query.Handler handles Request with corresponding Method.
// Request.Props are passed to query.Queries.Handle as Event.Payload.
// Does not supports only JSON RPC Notifications.
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

	addMetadata(w, req)
	w.Header().Add("Content-Type", "application/json")

	var errResponse *ErrorResponse
	reqModel, payload, errCode, err := getRequest(req)

	if err != nil {
		errResponse = new(ErrorResponse)
		errResponse.Version = ProtocolVersion
		errResponse.Error = Error{Code: errCode, Message: err.Error()}

		writeErrorResponse(w, errResponse)

		return
	}

	if reqModel.ID == nil {
		writeErrorResponse(w,
			reqModel.NewErrorResponse(InternalError, "Queries does not support JSON-RPC Notifications", nil))

		return
	}

	result := make(map[string]interface{})
	err = h.queries.HandleJSONEncoded(req.Context(), reqModel.Method, &result, payload)
	methodNotSupported := new(query.ErrQueryHandlerNotFound)

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

	b, err := json.Marshal(reqModel.NewResponse(result))
	if err != nil {
		writeErrorResponse(w, reqModel.NewErrorResponse(InternalApplicationError, err.Error(), nil))

		return
	}

	w.Write(b)
}
