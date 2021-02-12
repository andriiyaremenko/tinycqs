package tracing

import (
	"net/http"
	"regexp"

	"github.com/google/uuid"
)

var (
	RegexpRequestID     = regexp.MustCompile("^(?i)id|requestid|request_id$")
	RegexpCausationID   = regexp.MustCompile("^(?i)causationid|causation_id$")
	RegexpCorrelationID = regexp.MustCompile("^(?i)correlationid|correlation_id$")
)

// Reads Metadata from req.Header
func GetMetadataFromHeaders(req *http.Request) (Metadata, bool) {
	hasID, hasCausationID, hasCorrelationID := false, false, false
	id := uuid.New().String()
	metadata := M{
		EID:            id,
		ECausationID:   id,
		ECorrelationID: id}

	for key, v := range req.Header {
		switch {
		case RegexpRequestID.MatchString(key):
			metadata.EID = v[0]
			hasID = true
		case RegexpCausationID.MatchString(key):
			metadata.ECausationID = v[0]
			hasCausationID = true
		case RegexpCorrelationID.MatchString(key):
			metadata.ECorrelationID = v[0]
			hasCorrelationID = true
		default:
		}
	}

	hasMetadata := hasID && hasCausationID && hasCorrelationID

	return metadata, hasMetadata
}

// Reads Tracing Header names from req.Header
func GetTracingHeaderNames(req *http.Request) (requestID, causationID, correlationID string) {
	requestID = "RequestID"
	causationID = "CausationID"
	correlationID = "CorrelationID"

	for key := range req.Header {
		switch {
		case RegexpRequestID.MatchString(key):
			requestID = key
		case RegexpCausationID.MatchString(key):
			causationID = key
		case RegexpCorrelationID.MatchString(key):
			correlationID = key
		default:
		}
	}

	return
}
