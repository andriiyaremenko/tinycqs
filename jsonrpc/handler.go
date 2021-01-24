package jsonrpc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"

	"github.com/andriiyaremenko/tinycqs/command"
	"github.com/andriiyaremenko/tinycqs/query"
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
	responses := new(cSlice)
	var wg sync.WaitGroup

	for _, reqModel := range reqModels {
		wg.Add(1)
		go func(reqModel Request) {
			defer wg.Done()

			payload, err := json.Marshal(reqModel.Params)
			if err != nil {
				writeErrorResponse(w, reqModel.NewErrorResponse(InvalidRequest, fmt.Sprintf("request format: %s", err), nil))

				return
			}

			if reqModel.ID != nil {
				successResp, errResp := h.handleQueries(ctx, reqModel, metadata, payload)
				if errResp != nil && errResp.Error.Code == MethodNotFound {
					successResp, errResp = h.handleCommand(ctx, reqModel, metadata, payload)
				}

				if errResp != nil {
					responses.append(errResp)
					return
				}

				responses.append(successResp)
				return
			}

			errResp := h.workerHandleCommand(reqModel, metadata, payload)
			if errResp != nil && errResp.Error.Code == MethodNotFound {
				_, errResp = h.handleCommand(ctx, reqModel, metadata, payload)
			}

			if errResp != nil {
				responses.append(errResp)
			}
		}(reqModel)
	}

	wg.Wait()
	items := responses.value()

	switch {
	case len(items) == 0:
		w.WriteHeader(http.StatusNoContent)
	case isBatch:
		b, err := json.Marshal(items)
		if err == nil {
			w.Write(b)

			return
		}

		errResponse := new(ErrorResponse)
		errResponse.Version = ProtocolVersion
		errResponse.Error = Error{Code: InternalApplicationError, Message: err.Error()}
		writeErrorResponse(w, errResponse)
	default:
		resp := items[0]
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

func (h *Handler) handleCommand(ctx context.Context, reqModel Request, metadata command.Metadata, payload []byte) (*SuccessResponse, *ErrorResponse) {
	if h.Commands == nil {
		return nil, reqModel.NewErrorResponse(MethodNotFound,
			fmt.Sprintf("handler not found for command %s", reqModel.Method), nil)
	}

	var ev command.Event = command.E{EType: reqModel.Method, EPayload: payload}
	switch reqModel.ID {
	case nil:
		ev = h.Commands.Handle(ctx, command.WithMetadata(ev, metadata))
	default:
		ev = h.Commands.HandleOnly(ctx, command.WithMetadata(ev, metadata), reqModel.Method)
	}

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
		result := make(map[string]interface{})
		result["message"] = ev.EventType()
		result["params"] = reqModel.Params

		return reqModel.NewResponse(result), nil
	}

	return nil, nil
}

func (h *Handler) handleQueries(ctx context.Context, reqModel Request, metadata command.Metadata, payload []byte) (*SuccessResponse, *ErrorResponse) {
	if h.Queries == nil {
		return nil, reqModel.NewErrorResponse(MethodNotFound,
			fmt.Sprintf("handler not found for query %s", reqModel.Method), nil)
	}

	if reqModel.ID == nil {
		return nil, reqModel.NewErrorResponse(InternalError, "Queries does not support JSON-RPC Notifications", nil)
	}

	var errResponse *ErrorResponse
	result := make(map[string]interface{})
	err := h.Queries.HandleJSONEncoded(ctx, reqModel.Method, &result, payload)
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

func (h *Handler) workerHandleCommand(reqModel Request, metadata command.Metadata, payload []byte) *ErrorResponse {
	if h.Worker == nil {
		return reqModel.NewErrorResponse(MethodNotFound,
			fmt.Sprintf("handler not found for command %s", reqModel.Method), nil)
	}

	if reqModel.ID != nil {
		return reqModel.NewErrorResponse(InternalError, "CommandsWorker supports only JSON-RPC Notifications", nil)
	}

	var ev command.Event = command.E{EType: reqModel.Method, EPayload: payload}

	var errResponse *ErrorResponse
	err := h.Worker.Handle(command.WithMetadata(ev, metadata))
	methodNotSupported := new(command.ErrCommandHandlerNotFound)

	if errors.As(err, &methodNotSupported) {
		errResponse = reqModel.NewErrorResponse(MethodNotFound, err.Error(), nil)
	}

	if errResponse == nil && err != nil {
		errResponse = reqModel.NewErrorResponse(InternalError, err.Error(), nil)
	}

	if errResponse != nil {
		return errResponse
	}

	return nil
}
