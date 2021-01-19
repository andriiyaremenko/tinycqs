package command

import (
	"fmt"
	"strings"
	"sync"
)

const CatchAllErrorEventType string = "ERROR#*"

var MoreThanOneCatchAllErrorHandler = fmt.Errorf(`you can use only one handler for "%s" event`,
	CatchAllErrorEventType)

func ErrorEventType(eventType string) string { return fmt.Sprintf("ERROR#%s", eventType) }

func IsError(event Event, eventType string) bool {
	return event.EventType() == ErrorEventType(eventType)
}

// Returns new *ErrEvent caused by event.
// *ErrEvent implements error and Event.
func NewErrEvent(event Event, err error) *ErrEvent {
	return &ErrEvent{event, err}
}

type ErrEvent struct {
	event Event
	cause error
}

// Returns error event type.
func (err *ErrEvent) EventType() string {
	return ErrorEventType(err.event.EventType())
}

// Returns payload of event caused the error.
func (err *ErrEvent) Payload() []byte {
	return err.event.Payload()
}

// Returns *ErrEvent as an error.
func (err *ErrEvent) Err() error {
	return err
}

// Implementation of error.
func (err *ErrEvent) Error() string {
	return fmt.Sprintf("failed to process event %s: %s", err.event.EventType(), err.Unwrap())
}

// Returns underlying error.
func (err *ErrEvent) Unwrap() error {
	return err.cause
}

// Returns underlying Event.
func (err *ErrEvent) Event() Event {
	return err.event
}

// Returns new *ErrAggregatedEvent caused by event or Events dispatched while processing event.
// *ErrAggregatedEvent implements error and Event.
func NewErrAggregatedEvent(initialEvent Event) *ErrAggregatedEvent {
	return &ErrAggregatedEvent{initialEvent: initialEvent}
}

type ErrAggregatedEvent struct {
	mu sync.RWMutex

	initialEvent Event
	errors       []error
}

// Returns error event type.
func (err *ErrAggregatedEvent) EventType() string {
	return ErrorEventType(err.initialEvent.EventType())
}

// Returns payload of initial event.
func (err *ErrAggregatedEvent) Payload() []byte {
	return err.initialEvent.Payload()
}

// Returns *ErrAggregatedEvent as an error.
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

// Implementation of error.
func (err *ErrAggregatedEvent) Error() string {
	return err.Err().Error()
}

// Returns underlying (initial) Event.
func (err *ErrAggregatedEvent) Event() Event {
	return err.initialEvent
}

// Returns list of all errors caused by processing initial Event.
// or Events dispatched while processing this Event.
func (err *ErrAggregatedEvent) Inner() []error {
	err.mu.RLock()
	defer err.mu.RUnlock()
	return err.errors
}

// Appends errors to *ErrAggregatedEvent error list.
func (err *ErrAggregatedEvent) Append(errors ...error) {
	err.mu.Lock()
	defer err.mu.Unlock()
	err.errors = append(err.errors, errors...)
}

// error type returned if incorrect Handler was passed to Commands.
type ErrIncorrectHandler struct {
	handler Handler
}

// Implementation of error.
func (err *ErrIncorrectHandler) Error() string {
	return fmt.Sprintf("command handler %#v has incorrect format", err.handler)
}

// error type returned if no Handlers was found for particular Event.
type ErrCommandHandlerNotFound struct {
	commandName string
}

// Implementation of error.
func (err *ErrCommandHandlerNotFound) Error() string {
	return fmt.Sprintf("handler not found for command %s", err.commandName)
}

// ErrNilEvent instance.
const NilEvent ErrNilEvent = "NilEvent"

// error type returned if Event equals nil.
// ErrNilEvent implements error and Event.
type ErrNilEvent string

// Returns error event type.
func (err ErrNilEvent) EventType() string {
	return ErrorEventType(string(err))
}

// Returns nil.
func (err ErrNilEvent) Payload() []byte {
	return nil
}

// Returns ErrNilEvent as an error.
func (err ErrNilEvent) Err() error {
	return err
}

// Implementation of error.
func (err ErrNilEvent) Error() string {
	return "got event with value of nil"
}
