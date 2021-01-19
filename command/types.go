package command

import (
	"context"
)

// Metadata carries additional information not used in command execution.
type Metadata interface {
	// Unique event ID.
	ID() string
	// Root event ID that triggered execution of the program.
	CorrelationID() string
	// ID of event that caused execution of current event.
	CausationID() string

	// New metadata for next event in execution chain.
	New(id string) Metadata
}

// Event carries information needed to execute Commands.Handle and Commands.HandleOnly.
type Event interface {
	// Type of event. Is used to select Handler.
	EventType() string
	// Information needed to process Event.
	Payload() []byte
	// error caused by Event.
	Err() error
}

// Event with metadata.
type EventWithMetadata interface {
	Event
	// returns current event Metadata.
	Metadata() Metadata
}

// Serves to pass Events to Handlers.
type EventReader interface {
	// Returns Event channel.
	Read() <-chan Event
	// Closes Event channel.
	Close()

	// Returns EventWriter instance on which this EventReader is based.
	GetWriter(Metadata) EventWriter
}

// Servers to write Events in Handle.Handle to chain Events.
// **Do not forget to call Done() when finished writing**.
type EventWriter interface {
	// Writes Event to a channel.
	Write(e Event)
	// Signals Commands that Handler is done writing.
	Done()
}

// Handles single type of Event.
type Handler interface {
	// Type of Event.
	EventType() string
	// Method to handle Events.
	Handle(ctx context.Context, w EventWriter, event Event)
}

// Sealed interface to handle Events.
type Commands interface {
	sealed()

	// Handles event regardless of its type.
	// Can chain Events if any occurred as a result of processing this event.
	Handle(ctx context.Context, event Event) Event
	// Handles event regardless of its type without event chaining.
	HandleOnly(ctx context.Context, event Event, only ...string) Event
}

// Sealed interface of worker that is able to handle Events regardless of their type.
// Supposed to be bound to single context for all Handle calls.
type CommandsWorker interface {
	sealed()

	// Indicates if CommandWorker is running and able to handle Events.
	IsRunning() bool
	// Handles event regardless of its type.
	// Can chain Events if any occurred as a result of processing this event.
	Handle(event Event) error
}
