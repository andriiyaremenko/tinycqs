package jsonrpc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
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
	b, err := ioutil.ReadAll(req.Body)
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
	metadata, ok := tracing.GetMetadataFromHeaders(req)
	idKey, causationIDKey, correlationIDKey := tracing.GetTracingHeaderNames(req)

	// means it is not new
	if ok {
		metadata = metadata.New(uuid.New().String())
	}

	w.Header().Add(idKey, metadata.ID())
	w.Header().Add(causationIDKey, metadata.CausationID())
	w.Header().Add(correlationIDKey, metadata.CorrelationID())

	return metadata
}
