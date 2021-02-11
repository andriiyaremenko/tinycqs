package command

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/andriiyaremenko/tinycqs/tracing"
)

const doneWriting doneEv = doneEv("EVENT_WRITER_DONE_WRITING")

// returns wrapped Event by calling .Event() method if event has it or nil otherwise.
func Unwrap(event Event) Event {
	withE, ok := event.(interface{ Event() Event })
	if !ok {
		return nil
	}

	return withE.Event()
}

// returns Done event type for eventType event.
func DoneEventType(eventType string) string { return fmt.Sprintf("DONE#%s", eventType) }

// Event returned as a result of successful processing Event or chain of Events.
// If *DoneEvent is passed to EventWriter.Write it will appear in final result.
func Done(event Event) Event { return &DoneEvent{E: event} }

// returns true if event is *DoneEvent and .EventType() equals DoneEventType(eventType).
func IsDone(event Event, eventType string) bool {
	done, ok := event.(*DoneEvent)
	if !ok {
		if event := Unwrap(event); event != nil {
			return IsDone(event, eventType)
		}

		return false
	}

	if done.EventType() != DoneEventType(eventType) {
		return false
	}

	return true
}

// returns true if event is *DoneEvent.
func AsDoneEvent(event Event) *DoneEvent {
	done, ok := event.(*DoneEvent)
	if !ok {
		if event := Unwrap(event); event != nil {
			return AsDoneEvent(event)
		}

		return nil
	}

	return done
}

// unwraps event if event is *DoneEvent and returns it.
// returns original event otherwise.
func UnwrapDoneEvent(event Event) Event {
	if done := AsDoneEvent(event); done != nil {
		return done.Event()
	}

	return event
}

// *DoneEvent is returned by Commands.Handle if Event was handled successfully.
// If *DoneEvent is passed to EventWriter.Write it will appear in final result.
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

// E implements Event.
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

// Adds Metadata to event.
func WithMetadata(event Event, metadata tracing.Metadata) EventWithMetadata {
	return &eventWithMetadata{event, metadata}
}

// Returns EventWithMetadata if event can be converted to it. nil otherwise.
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
	metadata tracing.Metadata
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

func (e *eventWithMetadata) Metadata() tracing.Metadata {
	return e.metadata
}

func (e *eventWithMetadata) Event() Event {
	return e.event
}

type EventMessage struct {
	ID            string `json:"id"`
	CausationID   string `json:"causationId"`
	CorrelationID string `json:"correlationId"`

	EventType string          `json:"type"`
	Payload   json.RawMessage `json:"payload"`
}

func newResult(event EventWithMetadata) *result {
	return &result{event: event}
}

type result struct {
	mu sync.Mutex

	event   EventWithMetadata
	results []json.RawMessage
	errors  []error
}

func (r *result) Append(done *DoneEvent, metadata tracing.Metadata) {
	r.mu.Lock()
	defer r.mu.Unlock()

	ev := done.Event()
	message := EventMessage{
		ID:            metadata.ID(),
		CausationID:   metadata.CausationID(),
		CorrelationID: metadata.CorrelationID(),
		EventType:     ev.EventType(),
		Payload:       ev.Payload()}

	b, err := json.Marshal(message)
	if err != nil {
		jsonString := fmt.Sprintf(`"%s"`, string(ev.Payload()))
		message.Payload = []byte(jsonString)
		b, err = json.Marshal(message)
	}

	if err != nil {
		r.errors = append(r.errors, err)
		return
	}

	r.results = append(r.results, b)
}

func (r *result) EventType() string {
	return r.event.EventType()
}

func (r *result) Payload() []byte {
	r.mu.Lock()
	defer r.mu.Unlock()

	results, err := json.Marshal(r.results)
	if err != nil {
		return nil
	}

	metadata := r.event.Metadata()
	payload := EventMessage{
		ID:            metadata.ID(),
		CausationID:   metadata.CausationID(),
		CorrelationID: metadata.CorrelationID(),
		EventType:     DoneEventType(r.event.EventType()),
		Payload:       results}
	b, err := json.Marshal(payload)

	if err != nil {
		return nil
	}

	return b
}

func (r *result) Err() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if len(r.errors) == 0 {
		return nil
	}

	var sb strings.Builder

	sb.WriteByte('\n')

	for _, e := range r.errors {
		sb.WriteByte('\t')
		sb.WriteString(e.Error())
		sb.WriteByte('\n')
	}

	return fmt.Errorf(
		"failed to process event %s: failed to marshal results: [%s]",
		r.event.EventType(), sb.String())
}

func (r *result) Event() Event {
	return r.event
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

func (dine doneEv) Metadata() tracing.Metadata {
	return nil
}
