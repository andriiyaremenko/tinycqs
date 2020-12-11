package command

import (
	"context"
)

type Event interface {
	EventType() string
	Payload() []byte
	Err() error
}

type EventReader interface {
	Read() <-chan Event
	Close()

	GetWriter() EventWriter
}

type EventWriter interface {
	Write(e Event)
	Done()
}

type Handler interface {
	EventType() string
	Handle(ctx context.Context, w EventWriter, event Event)
}

type CommandsWorker interface {
	sealed()

	IsRunning() bool
	Handle(event Event) error
}

type Commands interface {
	sealed()

	Handle(ctx context.Context, event Event) Event
	HandleOnly(ctx context.Context, event Event, only ...string) Event
}
