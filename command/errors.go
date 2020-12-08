package command

import (
	"fmt"
)

var MoreThanOneCatchAllErrorHandler = fmt.Errorf(`you can use only one handler for "%s" event`, CatchAllErrorEventType)

const CatchAllErrorEventType = "Error#*"

var ErrorEventType = func(eventType string) string { return fmt.Sprintf("Error#%s", eventType) }

func NewErrEvent(event Event, err error) *ErrEvent {
	return &ErrEvent{event, err}
}

type ErrEvent struct {
	event Event
	cause error
}

func (err *ErrEvent) EventType() string {
	return ErrorEventType(err.event.EventType())
}

func (err *ErrEvent) Payload() []byte {
	return err.event.Payload()
}

func (err *ErrEvent) Err() error {
	return err
}

func (err *ErrEvent) Error() string {
	return fmt.Sprintf("failed to process event %s: %s", err.event.EventType(), err.Unwrap())
}

func (err *ErrEvent) Unwrap() error {
	return err.cause
}

func (err *ErrEvent) Event() Event {
	return err.event
}

type ErrDone struct {
	Cause error
}

func (err *ErrDone) EventType() string {
	return DoneEventType
}

func (err *ErrDone) Payload() []byte {
	return nil
}

func (err *ErrDone) Err() error {
	return err.Cause
}

func (err *ErrDone) Error() string {
	return fmt.Sprintf("failed to process event %s: %s", err.EventType(), err.Unwrap())
}

func (err *ErrDone) Unwrap() error {
	return err.Cause
}

type ErrIncorrectHandler struct {
	handler Handler
}

func (err *ErrIncorrectHandler) Error() string {
	return fmt.Sprintf("command handler %#v has incorrect format", err.handler)
}

type ErrCommandHandlerNotFound struct {
	commandName string
}

func (err *ErrCommandHandlerNotFound) Error() string {
	return fmt.Sprintf("handler not found for command %s", err.commandName)
}
