package command

import (
	"fmt"
)

type ErrCommandHandlerNotFound struct {
	commandName string
}

func (err *ErrCommandHandlerNotFound) Error() string {
	return fmt.Sprintf("handler not found for command %s", err.commandName)
}

func NewErrCommandHandlerNotFound(commandName string) *ErrCommandHandlerNotFound {
	return &ErrCommandHandlerNotFound{commandName}
}
