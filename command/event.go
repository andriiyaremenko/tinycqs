package command

import (
	"fmt"
)

const doneWriting doneEv = doneEv("EVENT_WRITER_DONE_WRITING")

// returns wrapped `Event` by calling `.Event()` method if `event` has it or `nil` otherwise
func Unwrap(event Event) Event {
	withE, ok := event.(interface{ Event() Event })
	if !ok {
		return nil
	}

	return withE.Event()
}

// returns Done event type for `eventType` event
func DoneEventType(eventType string) string { return fmt.Sprintf("DONE#%s", eventType) }

// `Event` returned as a result of successful processing `Event` or chain of `Event`s
func Done(event Event) Event { return &DoneEvent{E: event} }

// returns `true` if `event` is `*DoneEvent` and `.EventType()` equals `DoneEventType(eventType)`
func IsDone(event Event, eventType string) bool {
	done, ok := event.(*DoneEvent)
	if !ok ||
		done.EventType() != DoneEventType(eventType) {
		return false
	}

	return true
}

type doneEv string

func (done doneEv) EventType() string {
	return string(done)
}

func (done doneEv) Payload() []byte {
	return nil
}

func (done doneEv) Err() error {
	return nil
}

// `DoneEvent` is returned if `Event`` was handled successfully
type DoneEvent struct {
	E Event
}

func (done *DoneEvent) EventType() string {
	return DoneEventType(done.E.EventType())
}

func (done *DoneEvent) Payload() []byte {
	return done.E.Payload()
}

func (done *DoneEvent) Err() error {
	return nil
}

func (done *DoneEvent) Event() Event {
	return done.E
}

// `E` implements `Event`
type E struct {
	EType    string
	EPayload []byte
}

func (e E) EventType() string {
	return e.EType
}

func (e E) Payload() []byte {
	return e.EPayload
}

func (e E) Err() error {
	return nil
}

// Adds `Metadata` to `event`
func WithMetadata(event Event, metadata Metadata) EventWithMetadata {
	return &eventWithMetadata{event, metadata}
}

// Returns `EventWithMetadata` if `event` can be converted to it. `nil` otherwise
func AsEventWithMetadata(event Event) EventWithMetadata {
	withM, ok := event.(EventWithMetadata)
	if ok {
		return withM
	}

	if e := Unwrap(event); e != nil {
		return AsEventWithMetadata(e)
	}

	return nil
}

type eventWithMetadata struct {
	event    Event
	metadata Metadata
}

func (e *eventWithMetadata) EventType() string {
	return e.event.EventType()
}

func (e *eventWithMetadata) Payload() []byte {
	return e.event.Payload()
}

func (e *eventWithMetadata) Err() error {
	return e.event.Err()
}

func (e *eventWithMetadata) Metadata() Metadata {
	return e.metadata
}

func (e *eventWithMetadata) Event() Event {
	return e.event
}

// `M` implements `Metadata`
type M struct {
	EID            string
	ECorrelationID string
	ECausationID   string
}

func (m M) New(id string) Metadata {
	return M{EID: id, ECorrelationID: m.ECorrelationID, ECausationID: m.EID}
}

func (m M) ID() string {
	return m.EID
}

func (m M) CorrelationID() string {
	return m.ECorrelationID
}

func (m M) CausationID() string {
	return m.ECausationID
}
