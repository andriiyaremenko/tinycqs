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
	Handle(ctx context.Context, event Event) <-chan Event
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
	ConcurrencyLimit() int
	SetConcurrencyLimit(limit int)
}
