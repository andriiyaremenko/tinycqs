package jsonrpc

import (
	"net/http"

	"github.com/andriiyaremenko/tinycqs/command"
)

// Turns command.CommandsWorker into http.Handler.
// Every command.Handler handles Request with corresponding Method.
// Request.Props are passed to command.CommandsWorker.Handle as Event.Payload.
// Supports only JSON RPC Notifications.
func CommandsWorker(worker command.CommandsWorker) http.Handler {
	return &Handler{Worker: worker}
}
