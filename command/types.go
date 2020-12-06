package command

import (
	"context"
)

type Event interface {
	EventType() string
	Payload() []byte
}

type Handler interface {
	EventType() string
	Handle(ctx context.Context, payload []byte) error
}

type Command interface {
	Handle(ctx context.Context, event Event) error
	HandleOnly(ctx context.Context, event Event, only ...string) error
}
