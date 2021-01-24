package jsonrpc

import (
	"net/http"

	"github.com/andriiyaremenko/tinycqs/command"
)

// Turns command.Commands into http.Handler.
// Every command.Handler handles Request with corresponding Method.
// Request.Props are passed to command.Commands.Handle as Event.Payload.
func Commands(commands command.Commands) http.Handler {
	return &Handler{Commands: commands}
}
