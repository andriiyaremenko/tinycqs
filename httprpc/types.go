package httprpc

import _ "encoding/json"

// RPC request model
type Request struct {
	// RPC method to call.
	Method string `json:"method"`
	// RPC method parameters to pass to method
	Params map[string]interface{} `json:"params"`
}
