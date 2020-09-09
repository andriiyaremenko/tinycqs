package command

import (
	"context"
)

type Command struct {
	Name string
	Body []byte
}

type Handler interface {
	CommandName() string
	Handle(ctx context.Context, command *Command) error
}

type Demultiplexer interface {
	Handle(ctx context.Context, command *Command) error
}
