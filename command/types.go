package command

import (
	"context"
)

type Event interface {
	EventType() string
	Payload() []byte
	Err() error
}

type Handler interface {
	EventType() string
	Handle(ctx context.Context, payload []byte) <-chan Event
}

type CommandsWorker interface {
	sealed()

	Handle(event Event) Event
	HandleOnly(event Event, only ...string) Event
}

type Commands interface {
	sealed()

	Handle(ctx context.Context, event Event) Event
	HandleOnly(ctx context.Context, event Event, only ...string) Event
}
