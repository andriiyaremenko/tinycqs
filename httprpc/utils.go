package httprpc

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"

	"github.com/andriiyaremenko/tinycqs/command"
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

func getRequest(req *http.Request) (string, []byte, error) {
	b, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return "", nil, fmt.Errorf("failed to read request body: %s", err)
	}

	reqModel := new(Request)
	if err := json.Unmarshal(b, reqModel); err != nil {
		return "", nil, fmt.Errorf("invalid request format: %s", err)
	}

	if reqModel.Method == "" {
		return "", nil, fmt.Errorf("invalid request format: %s", string(b))
	}

	payload, err := json.Marshal(reqModel.Params)
	if err != nil {
		return "", nil, fmt.Errorf("invalid request format: %s", err)
	}

	return reqModel.Method, payload, err
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
