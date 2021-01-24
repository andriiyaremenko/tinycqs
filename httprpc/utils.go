package httprpc

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"regexp"

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

func getRequest(req *http.Request) (*Request, []byte, int, error) {
	b, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return nil, nil, ParseError, fmt.Errorf("failed to read request body: %s", err)
	}

	reqModel := new(Request)
	if err := json.Unmarshal(b, reqModel); err != nil {
		return nil, nil, InvalidRequest, fmt.Errorf("invalid request format: %s", err)
	}

	if reqModel.Method == "" {
		return nil, nil, InvalidRequest, fmt.Errorf("invalid request format: %s", string(b))
	}

	if reqModel.Version != ProtocolVersion {
		return nil, nil, InvalidRequest, fmt.Errorf("invalid request format: %s", string(b))
	}

	isNil := reqModel.ID == nil
	_, isString := reqModel.ID.(string)
	_, isInt := reqModel.ID.(int)

	if fl, ok := reqModel.ID.(float64); ok {
		isInt = math.Mod(fl, 1) == 0
	}

	if !isString && !isInt && !isNil {
		return nil, nil, InvalidRequest, fmt.Errorf("invalid request format: %s", string(b))
	}

	payload, err := json.Marshal(reqModel.Params)
	if err != nil {
		return nil, nil, InvalidRequest, fmt.Errorf("invalid request format: %s", err)
	}

	return reqModel, payload, 0, err
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
