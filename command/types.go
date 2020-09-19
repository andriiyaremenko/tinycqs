package command

import (
	"context"
)

type Handler interface {
	CommandName() string
	Handle(ctx context.Context, payload []byte) error
}

type Demultiplexer interface {
	Handle(ctx context.Context, command string, payload []byte) error
	HandleOnly(ctx context.Context, only []string, command string, payload []byte) error
}
