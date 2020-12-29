package command

import (
	"fmt"
)

var ErrorEventType = func(eventType string) string { return fmt.Sprintf("ERROR#%s", eventType) }
var DoneEventType = func(eventType string) string { return fmt.Sprintf("DONE#%s", eventType) }

// `Event` returned as a result of successful processing `Event` or chain of `Event`s
var Done = func(event Event) Event { return &DoneEvent{Event: event} }

const (
	CatchAllErrorEventType string = "ERROR#*"

	doneWriting doneEv = doneEv("EVENT_WRITER_DONE_WRITING")
)

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

type DoneEvent struct {
	Event Event
}

func (done *DoneEvent) EventType() string {
	return DoneEventType(done.Event.EventType())
}

func (done *DoneEvent) Payload() []byte {
	return nil
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
