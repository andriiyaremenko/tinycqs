package jsonrpc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"

	"github.com/andriiyaremenko/tinycqs/tracing"
	"github.com/google/uuid"
)

type keyValue struct {
	key   string
	value string
}

func getRequest(b []byte) ([]Request, bool, int, error) {
	var reqModel Request
	if err := json.Unmarshal(b, &reqModel); err != nil {
		return nil, false, InvalidRequest, fmt.Errorf("invalid request format: %s", err)
	}

	if !isValid(reqModel) {
		return nil, false, InvalidRequest, fmt.Errorf("invalid request format: %s", string(b))
	}

	return []Request{reqModel}, false, 0, nil
}

func getRequests(req *http.Request) ([]Request, bool, int, error) {
	b, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, false, ParseError, fmt.Errorf("failed to read request body: %s", err)
	}

	if !bytes.HasPrefix(b, []byte("[")) {
		return getRequest(b)
	}

	reqModels := make([]Request, 0, 1)
	if err := json.Unmarshal(b, &reqModels); err != nil {
		return nil, false, InvalidRequest, fmt.Errorf("invalid request format: %s", err)
	}

	for _, reqModel := range reqModels {
		if !isValid(reqModel) {
			return nil, false, InvalidRequest, fmt.Errorf("invalid request format: %s", string(b))
		}
	}

	return reqModels, true, 0, nil
}

func isValid(reqModel Request) bool {
	if reqModel.Method == "" {
		return false
	}

	if reqModel.Version != ProtocolVersion {
		return false
	}

	isNil := reqModel.ID == nil
	_, isString := reqModel.ID.(string)
	_, isInt := reqModel.ID.(int)

	if fl, ok := reqModel.ID.(float64); ok {
		isInt = math.Mod(fl, 1) == 0
	}

	if !isString && !isInt && !isNil {
		return false
	}

	return true
}

func writeErrorResponse(w http.ResponseWriter, errResponse *ErrorResponse) {
	b, err := json.Marshal(errResponse)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusBadRequest)
	w.Write(b)
}

func addMetadata(w http.ResponseWriter, req *http.Request) tracing.Metadata {
	// metadata is starting point for our execution
	// and we should base our command execution on it
	// respMetadata is what should be returned as a result of the whole execution pipeline
	//																	-> some other branch of execution (next metadata value) --> ...
	// incoming request (metadata) --> start of execution (metadata) -| --> response (next metadata value == respMetadata)
	//																	-> some other branch of execution (next metadata value) --> ...
	metadata, ok := tracing.GetMetadataFromHeaders(req)
	respMetadata := metadata
	idKey, causationIDKey, correlationIDKey := tracing.GetTracingHeaderNames(req)

	// means metadata is not new
	if ok {
		respMetadata = metadata.New(uuid.New().String())
	}

	w.Header().Add(idKey, respMetadata.ID())
	w.Header().Add(causationIDKey, respMetadata.CausationID())
	w.Header().Add(correlationIDKey, respMetadata.CorrelationID())

	return metadata
}
