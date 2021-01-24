package jsonrpc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"regexp"
	"sync"

	"github.com/andriiyaremenko/tinycqs/command"
	"github.com/google/uuid"
)

var (
	regexpRequestID     = regexp.MustCompile("^(?i)id|requestid|request_id$")
	regexpCausationID   = regexp.MustCompile("^(?i)causationid|causation_id$")
	regexpCorrelationID = regexp.MustCompile("^(?i)correlationid|correlation_id$")
)

type keyValue struct {
	key   string
	value string
}

type cSlice struct {
	mu    sync.Mutex
	items []interface{}
}

func (cs *cSlice) append(item interface{}) {
	cs.mu.Lock()
	cs.items = append(cs.items, item)
	cs.mu.Unlock()
}

func (cs *cSlice) value() []interface{} {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	return cs.items
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

func getMetadata(req *http.Request) (command.Metadata, string, string, string, bool) {
	var id *keyValue
	var causationID *keyValue
	var correlationID *keyValue

	for key, v := range req.Header {
		switch {
		case regexpRequestID.MatchString(key):
			id = &keyValue{key: key, value: v[0]}
		case regexpCausationID.MatchString(key):
			causationID = &keyValue{key: key, value: v[0]}
		case regexpCorrelationID.MatchString(key):
			correlationID = &keyValue{key: key, value: v[0]}
		default:
		}
	}

	hasMetadata := id != nil && causationID != nil && correlationID != nil
	if !hasMetadata {
		return nil, "", "", "", false
	}

	metadata := command.M{
		EID:            id.value,
		ECausationID:   causationID.value,
		ECorrelationID: correlationID.value}

	return metadata, id.key, causationID.key, correlationID.key, true
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

func addMetadata(w http.ResponseWriter, req *http.Request) command.Metadata {
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

	return metadata
}
