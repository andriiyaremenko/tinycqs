package tracing

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
