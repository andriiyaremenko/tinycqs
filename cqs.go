package tinycqs

import (
	"context"

	"github.com/andriiyaremenko/tinycqs/command"
	"github.com/andriiyaremenko/tinycqs/internal"
	"github.com/andriiyaremenko/tinycqs/query"
)

func NewCommandDemultiplexer(handlers ...command.Handler) command.Demultiplexer {
	return internal.NewCommandDemultiplexer(handlers...)
}

func NewQueryDemultiplexer(handlers ...query.Handler) query.Demultiplexer {
	return internal.NewQueryDemultiplexer(handlers...)
}

func CommandHandlerFunc(commandName string, handle func(context.Context, []byte) error) command.Handler {
	return internal.CommandHandlerFunc(commandName, handle)
}

func QueryHandlerFunc(queryName string, handle func(context.Context, []byte) ([]byte, error)) query.Handler {
	return internal.QueryHandlerFunc(queryName, handle)
}
