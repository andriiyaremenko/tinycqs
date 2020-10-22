package internal

import (
	"context"

	c "github.com/andriiyaremenko/tinycqs/command"
)

type CommandDemultiplexer struct {
	handlers []c.Handler
}

func (hf *CommandDemultiplexer) Handle(ctx context.Context, command string, payload []byte) error {
	for _, h := range hf.handlers {
		if h.CommandName() == command {
			return h.Handle(ctx, payload)
		}
	}

	return c.NewErrCommandHandlerNotFound(command)
}

func (hf *CommandDemultiplexer) HandleOnly(ctx context.Context, command string, payload []byte, only ...string) error {
	exists := false
	for _, c := range only {
		if exists = c == command; exists {
			break
		}
	}

	if !exists {
		return nil
	}

	for _, h := range hf.handlers {
		if h.CommandName() == command {
			return h.Handle(ctx, payload)
		}
	}

	return c.NewErrCommandHandlerNotFound(command)
}

func NewCommandDemultiplexer(handlers ...c.Handler) c.Demultiplexer {
	return &CommandDemultiplexer{handlers}
}

type CommandHandler struct {
	commandName string
	handle      func(context.Context, []byte) error
}

func (ch *CommandHandler) CommandName() string {
	return ch.commandName
}

func (ch *CommandHandler) Handle(ctx context.Context, payload []byte) error {
	return ch.handle(ctx, payload)
}

func CommandHandlerFunc(commandName string, handle func(context.Context, []byte) error) c.Handler {
	return &CommandHandler{commandName, handle}
}
