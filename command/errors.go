package command

import (
	"fmt"
	"strings"
	"sync"
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

func NewErrAggregatedEvent(initialEvent Event) *ErrAggregatedEvent {
	return &ErrAggregatedEvent{initialEvent: initialEvent}
}

type ErrAggregatedEvent struct {
	mu sync.RWMutex

	initialEvent Event
	errors       []error
}

func (err *ErrAggregatedEvent) EventType() string {
	return ErrorEventType(err.initialEvent.EventType())
}

func (err *ErrAggregatedEvent) Payload() []byte {
	return err.initialEvent.Payload()
}

func (err *ErrAggregatedEvent) Err() error {
	err.mu.RLock()
	defer err.mu.RUnlock()

	if len(err.errors) == 0 {
		return nil
	}

	var sb strings.Builder

	sb.WriteByte('\n')

	for _, e := range err.errors {
		sb.WriteByte('\t')
		sb.WriteString(e.Error())
		sb.WriteByte('\n')
	}

	return fmt.Errorf(
		"failed to process event %s: aggregated error occurred: [%s]",
		err.initialEvent.EventType(), sb.String())
}

func (err *ErrAggregatedEvent) Error() string {
	return err.Err().Error()
}

func (err *ErrAggregatedEvent) Event() Event {
	return err.initialEvent
}

func (err *ErrAggregatedEvent) Inner() []error {
	err.mu.RLock()
	defer err.mu.RUnlock()
	return err.errors
}

func (err *ErrAggregatedEvent) Append(errs ...error) {
	err.mu.Lock()
	defer err.mu.Unlock()
	err.errors = append(err.errors, errs...)
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

const NilEvent ErrNilEvent = "NilEvent"

type ErrNilEvent string

func (err ErrNilEvent) EventType() string {
	return ErrorEventType(string(err))
}

func (err ErrNilEvent) Payload() []byte {
	return nil
}

func (err ErrNilEvent) Err() error {
	return err
}

func (err ErrNilEvent) Error() string {
	return "got event with value of nil"
}
