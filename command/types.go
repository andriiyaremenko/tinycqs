package command

import (
	"context"
)

// Event carries information needed to execute `Commands.Handle` and `Commands.HandleOnly`
type Event interface {
	// Type of event. Is used to select `Handler`
	EventType() string
	// Information needed to process `Event`
	Payload() []byte
	// `error` caused by `Event`
	Err() error
}

// Serves to pass `Event`s to `Handler`s
type EventReader interface {
	// Returns `Event` channel
	Read() <-chan Event
	// Closes `Event` channel
	Close()

	// Returns `EventWriter` instance on which this `EventReader` is based
	GetWriter() EventWriter
}

// Servers to write `Event`s in `Handle.Handle` to chain `Event`s
// **Do not forget to call `Done()` when finished writing**
type EventWriter interface {
	// Writes `Event` to a channel
	Write(e Event)
	// Signals `Commands` that `Handler` is done writing
	Done()
}

// Handles single type of `Event`
type Handler interface {
	// Type of `Event`
	EventType() string
	// Method to handle `Event`s
	Handle(ctx context.Context, w EventWriter, event Event)
}

// Sealed interface to handle `Event`s
type Commands interface {
	sealed()

	// Handles `event` regardless of its type
	// Can chain `Event`s if any occurred as a result of processing this `event`
	Handle(ctx context.Context, event Event) Event
	// Handles `event` regardless of its type without event chaining
	HandleOnly(ctx context.Context, event Event, only ...string) Event
}

// Sealed interface of worker that is able to handle `Event`s regardless of their type
// Supposed to be bound to single context for all `Handle` calls
type CommandsWorker interface {
	sealed()

	// Indicates if `CommandWorker` is running and able to handle `Event`s
	IsRunning() bool
	// Handles `event` regardless of its type
	// Can chain `Event`s if any occurred as a result of processing this `event`
	Handle(event Event) error
}
