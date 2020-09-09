package internal

import (
	"context"

	c "github.com/andriiyaremenko/tinycqs/command"
)

type CommandDemultiplexer struct {
	handlers []c.Handler
}

func (hf *CommandDemultiplexer) Handle(ctx context.Context, command *c.Command) error {
	for _, h := range hf.handlers {
		if h.CommandName() == command.Name {
			return h.Handle(ctx, command)
		}
	}

	return c.NewErrCommandHandlerNotFound(command.Name)
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

func (ch *CommandHandler) Handle(ctx context.Context, command *c.Command) error {
	return ch.handle(ctx, command.Body)
}

func CommandHandlerFunc(commandName string, handle func(context.Context, []byte) error) c.Handler {
	return &CommandHandler{commandName, handle}
}
