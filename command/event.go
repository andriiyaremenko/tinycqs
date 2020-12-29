package command

import (
	"fmt"
)

const (
	CatchAllErrorEventType string = "ERROR#*"
	doneWriting            doneEv = doneEv("EVENT_WRITER_DONE_WRITING")
)

var ErrorEventType = func(eventType string) string { return fmt.Sprintf("ERROR#%s", eventType) }
var DoneEventType = func(eventType string) string { return fmt.Sprintf("DONE#%s", eventType) }

// `Event` returned as a result of successful processing `Event` or chain of `Event`s
var Done = func(event Event) Event { return &DoneEvent{Event: event} }
var IsDone = func(event Event, eventType string) bool {
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
	Event Event
}

func (done *DoneEvent) EventType() string {
	return DoneEventType(done.Event.EventType())
}

func (done *DoneEvent) Payload() []byte {
	return done.Event.Payload()
}

func (done *DoneEvent) Err() error {
	return nil
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
