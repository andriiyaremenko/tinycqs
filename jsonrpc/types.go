package jsonrpc

import _ "encoding/json"

const (
	// Invalid JSON was received by the server.
	//An error occurred on the server while parsing the JSON text.
	ParseError int = -32700
	// The JSON sent is not a valid Request object.
	InvalidRequest int = -32600
	// The method does not exist / is not available.
	MethodNotFound int = -32601
	// Invalid method parameter(s).
	InvalidParams int = -32602
	// Internal JSON RPC error.
	InternalError int = -32603
	// Internal application error.
	InternalApplicationError int = -32000
)

const ProtocolVersion string = "2.0"

// JSON RPC request model.
type Request struct {
	// JSON RPC version. Must be exactly "2.0".
	Version string `json:"jsonrpc"`
	// JSON RPC request ID.
	ID interface{} `json:"id"`
	// JSON RPC method to call.
	Method string `json:"method"`
	// JSON RPC method parameters to pass to method.
	Params map[string]interface{} `json:"params"`
}

// Returns new JSON RPC Response based on request ID and Version.
func (r *Request) NewResponse(result map[string]interface{}) *SuccessResponse {
	return &SuccessResponse{
		Version: r.Version,
		ID:      r.ID,
		Result:  result}
}

// Returns new JSON RPC Error Response based on request ID and Version.
func (r *Request) NewErrorResponse(rpcCode int, message string, data interface{}) *ErrorResponse {
	return &ErrorResponse{
		Version: r.Version,
		ID:      r.ID,
		Error: Error{
			Code:    rpcCode,
			Message: message,
			Data:    data}}
}

// JSON RPC success response model.
type SuccessResponse struct {
	// JSON RPC version. Must be exactly "2.0".
	Version string `json:"jsonrpc"`
	// JSON RPC request ID.
	ID interface{} `json:"id"`
	// Result object.
	Result map[string]interface{} `json:"result"`
}

// JSON RPC error response model.
type ErrorResponse struct {
	// JSON RPC version. Must be exactly "2.0".
	Version string `json:"jsonrpc"`
	// JSON RPC request ID.
	ID interface{} `json:"id"`
	// Error object.
	Error Error `json:"error"`
}

// JSON RPC error model.
type Error struct {
	// JSON RPC error code.
	Code int `json:"code"`
	// JSON RPC error message.
	Message string `json:"message"`
	// A Primitive or Structured value that contains additional information about the error.
	Data interface{} `json:"data"`
}
