package jsonrpc

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/andriiyaremenko/tinycqs/command"
	"github.com/andriiyaremenko/tinycqs/query"
	"github.com/andriiyaremenko/tinycqs/tracing"
)

// *Handler implements http.Handler.
// Handler uses query.Queries, command.Commands and command.CommandsWorker to process requests
type Handler struct {
	Queries  query.Queries
	Commands command.Commands
	Worker   command.CommandsWorker
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.NotFound(w, req)
		return
	}

	metadata := addMetadata(w, req)
	w.Header().Add("Content-Type", "application/json")

	reqModels, isBatch, errCode, err := getRequests(req)

	if err != nil {
		errResponse := new(ErrorResponse)
		errResponse.Version = ProtocolVersion
		errResponse.Error = Error{Code: errCode, Message: err.Error()}

		writeErrorResponse(w, errResponse)

		return
	}

	ctx := req.Context()
	responses := make([]interface{}, 0, 1)

	for _, reqModel := range reqModels {
		payload, err := json.Marshal(reqModel.Params)
		if err != nil {
			writeErrorResponse(w,
				reqModel.NewErrorResponse(InvalidRequest, fmt.Sprintf("request format: %s", err), nil))

			return
		}

		if reqModel.ID != nil {
			successResp, errResp := h.handleQueries(ctx, reqModel, metadata, payload)
			if errResp != nil && errResp.Error.Code == MethodNotFound {
				successResp, errResp = h.handleCommand(ctx, reqModel, metadata, payload)
			}

			if errResp != nil {
				responses = append(responses, errResp)
				continue
			}

			responses = append(responses, successResp)
			continue
		}

		_, errResp := h.handleCommand(ctx, reqModel, metadata, payload)
		if errResp != nil && errResp.Error.Code == MethodNotFound {
			errResp = h.workerHandleCommand(reqModel, metadata, payload)
		}

		if errResp != nil {
			responses = append(responses, errResp)
		}
	}

	switch {
	case len(responses) == 0:
		w.WriteHeader(http.StatusNoContent)
	case isBatch:
		b, err := json.Marshal(responses)
		if err == nil {
			w.Write(b)

			return
		}

		errResponse := new(ErrorResponse)
		errResponse.Version = ProtocolVersion
		errResponse.Error = Error{Code: InternalApplicationError, Message: err.Error()}
		writeErrorResponse(w, errResponse)
	default:
		resp := responses[0]
		errResponse, ok := resp.(*ErrorResponse)

		if ok {
			writeErrorResponse(w, errResponse)

			return
		}

		b, err := json.Marshal(resp)
		if err != nil {
			writeErrorResponse(w, reqModels[0].NewErrorResponse(InternalApplicationError, err.Error(), nil))

			return
		}

		w.Write(b)
		return
	}
}

func (h *Handler) MarshalJSON() ([]byte, error) {
	var sb bytes.Buffer

	sb.WriteString("{")

	if h.Queries != nil {
		b, err := json.Marshal(h.Queries)
		if err != nil {
			return nil, err
		}

		sb.WriteString(`"queries":`)
		sb.Write(b)
	}

	if h.Commands != nil {
		sb.WriteString(", ")
		b, err := json.Marshal(h.Commands)
		if err != nil {
			return nil, err
		}

		sb.WriteString(`"commands":`)
		sb.Write(b)
	}

	if h.Queries != nil {
		sb.WriteString(", ")
		b, err := json.Marshal(h.Queries)
		if err != nil {
			return nil, err
		}

		sb.WriteString(`"worker":`)
		sb.Write(b)
	}

	sb.WriteString("}")

	return sb.Bytes(), nil
}

func (h *Handler) handleCommand(ctx context.Context, reqModel Request,
	metadata tracing.Metadata, payload []byte) (*SuccessResponse, *ErrorResponse) {
	if h.Commands == nil {
		return nil, reqModel.NewErrorResponse(MethodNotFound,
			fmt.Sprintf("handler not found for command %s", reqModel.Method), nil)
	}

	var ev command.Event = command.E{Type: reqModel.Method, P: payload}
	ev = h.Commands.Handle(ctx, command.WithMetadata(ev, metadata))

	var errResponse *ErrorResponse
	err := ev.Err()
	methodNotSupported := new(command.ErrCommandHandlerNotFound)

	if errors.As(err, &methodNotSupported) {
		errResponse = reqModel.NewErrorResponse(MethodNotFound, err.Error(), nil)
	}

	if errResponse == nil && err != nil {
		errResponse = reqModel.NewErrorResponse(InternalApplicationError, err.Error(), nil)
	}

	if errResponse != nil {
		return nil, errResponse
	}

	if reqModel.ID != nil {
		ev = command.UnwrapDoneEvent(ev)
		result := make(map[string]interface{})
		result["message"] = ev.EventType()
		result["params"] = json.RawMessage(ev.Payload())

		b, err := json.Marshal(result)
		if err != nil {
			return nil, reqModel.NewErrorResponse(InternalApplicationError, err.Error(), nil)
		}

		return reqModel.NewResponse(json.RawMessage(b)), nil
	}

	return nil, nil
}

func (h *Handler) handleQueries(ctx context.Context, reqModel Request,
	metadata tracing.Metadata, payload []byte) (*SuccessResponse, *ErrorResponse) {
	if h.Queries == nil {
		return nil, reqModel.NewErrorResponse(MethodNotFound,
			fmt.Sprintf("handler not found for query %s", reqModel.Method), nil)
	}

	if reqModel.ID == nil {
		return nil, reqModel.NewErrorResponse(InternalError,
			"Queries does not support JSON-RPC Notifications", nil)
	}

	var errResponse *ErrorResponse
	results := make([]query.Result, 0, 1)

	for qr := range h.Queries.Handle(ctx, reqModel.Method, payload) {
		results = append(results, qr)
	}

	var err error
	var result json.RawMessage

switchLabel:
	switch len(results) {
	case 0:
		err = errors.New("something went wrong...")
	case 1:
		if err = results[0].Err(); err == nil {
			result = json.RawMessage(results[0].Body())
		}
	default:
		qResults := make([]json.RawMessage, 0, 1)
		for _, qr := range results {
			if err = qr.Err(); err != nil {
				break switchLabel
			}

			qResult := json.RawMessage(qr.Body())
			qResults = append(qResults, qResult)
		}

		var b []byte

		b, err = json.Marshal(qResults)
		if err != nil {
			break switchLabel
		}

		result = json.RawMessage(b)
	}

	methodNotSupported := new(query.ErrQueryHandlerNotFound)

	if errors.As(err, &methodNotSupported) {
		errResponse = reqModel.NewErrorResponse(MethodNotFound, err.Error(), nil)
	}

	if errResponse == nil && err != nil {
		errResponse = reqModel.NewErrorResponse(InternalApplicationError, err.Error(), nil)
	}

	if errResponse != nil {
		return nil, errResponse
	}

	return reqModel.NewResponse(result), nil
}

func (h *Handler) workerHandleCommand(reqModel Request,
	metadata tracing.Metadata, payload []byte) *ErrorResponse {
	if h.Worker == nil {
		return reqModel.NewErrorResponse(MethodNotFound,
			fmt.Sprintf("handler not found for command %s", reqModel.Method), nil)
	}

	if reqModel.ID != nil {
		return reqModel.NewErrorResponse(InternalError,
			"CommandsWorker supports only JSON-RPC Notifications", nil)
	}

	var ev command.Event = command.E{Type: reqModel.Method, P: payload}
	var errResponse *ErrorResponse

	if err := h.Worker.Handle(command.WithMetadata(ev, metadata)); err != nil {
		errResponse = reqModel.NewErrorResponse(InternalError, err.Error(), nil)
	}

	if errResponse != nil {
		return errResponse
	}

	return nil
}
