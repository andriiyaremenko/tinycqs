# tinycqs

This package is intended to help user separate ones workflow on commands and queries (CQS part of CQRS).  
It also possible to chain commands into single workflow.  
Each command is executed in separate gorutine, and each query is executed in same gorutine it was called.

## Commands
`import "github.com/andriiyaremenko/tinycqs/command"`

### CONSTANTS

```go
const (
	DoneEventType          string = "DONE"
	CatchAllErrorEventType string = "Error#*"
)
```
`Event` returned as a result of successful processing `Event` or chain of `Event`s
```go
const Done doneEv = doneEv(DoneEventType)
```
### VARIABLES

```go
var ErrorEventType = func(eventType string) string { return fmt.Sprintf("Error#%s", eventType) }
```
```go
var MoreThanOneCatchAllErrorHandler = fmt.Errorf(`you can use only one handler for "%s" event`, CatchAllErrorEventType)
```
```go
var WorkerStopped = errors.New("command worker is stopped")
```

### TYPES

`*CommandHandler` implements `Handler`
```go
type CommandHandler struct {
	EType      string
	HandleFunc func(ctx context.Context, w EventWriter, e Event)
}
```

Returns `EType`
```go
func (ch *CommandHandler) EventType() string
```

Runs `HandleFunc` in separate gorutine
```go
func (ch *CommandHandler) Handle(ctx context.Context, w EventWriter, event Event)
```

Returns `Handler` with `EventType` equals `eventType` and `Handle` 
based on `handle` running in separate gorutine
```go
func CommandHandlerFunc(eventType string, handle func(context.Context, []byte) error) Handler
```

Sealed interface to handle `Event`s
```go 
type Commands interface {
	Handle(ctx context.Context, event Event) Event
	HandleOnly(ctx context.Context, event Event, only ...string) Event
}
```

Returns new `Commands` with Concurrency Limit equals to `0` or `error`
Concurrency Limit is amount of `Event`s that can be processed concurrently
```go
func NewCommands(handlers ...Handler) (Commands, error)
```

Returns new `Commands` with Concurrency Limit equals to `limit` or `error`
Concurrency Limit is amount of `Event`s that can be processed concurrently
```go
func NewCommandsWithConcurrencyLimit(limit int, handlers ...Handler) (Commands, error)
```

Sealed interface of worker that is able to handle `Event`s regardless of
their type Supposed to be bound to single context for all `Handle` calls
```go
type CommandsWorker interface {
	IsRunning() bool
	Handle(event Event) error
}
```

Returns `CommandWorker` based on `Commands`.  
`eventSink` is used to channel all unhandled `error`s in form of `Event`
```go
func NewWorker(ctx context.Context, eventSink func(Event), commands Commands) CommandsWorker
```

`E` implements `Event`
```go
type E struct {
	EType    string
	EPayload []byte
}
```

Event carries information needed to execute `Commands.Handle` and `Commands.HandleOnly`
```go
type Event interface {
	EventType() string
	Payload() []byte
	Err() error
}
```

Serves to pass `Event`s to `Handler`s
```go
type EventReader interface {
	Read() <-chan Event
	Close()
	GetWriter() EventWriter
}
```

Servers to write `Event`s in `Handle.Handle` to chain `Event`s  
**Do not forget to call `Done()` when finished writing**
```go
type EventWriter interface {
	Write(e Event)
	Done()
}
```

Handles single type of `Event`
```go
type Handler interface {
	EventType() string
	Handle(ctx context.Context, w EventWriter, event Event)
}
```

### Errors
Returns new `*ErrAggregatedEvent` caused by `event` or `Event`s dispatched
while processing `event` `*ErrAggregatedEvent` implements `error` and `Event`
```go
func NewErrAggregatedEvent(initialEvent Event) *ErrAggregatedEvent
```
Appends `errors` to `*ErrAggregatedEvent` `error` list
```go
func (err *ErrAggregatedEvent) Append(errors ...error)
```
Returns `*ErrAggregatedEvent` as an `error`
```go
func (err *ErrAggregatedEvent) Err() error
```
Implementation of `error`
```go
func (err *ErrAggregatedEvent) Error() string
```
Returns underlying (initial) `Event`
```go
func (err *ErrAggregatedEvent) Event() Event
```
Returns error event type
```go
func (err *ErrAggregatedEvent) EventType() string
```
Returns list of all `error`s caused by processing initial 'Event' or
`Event`s dispatched while processing this `Event`
```go
func (err *ErrAggregatedEvent) Inner() []error
```
Returns payload of initial event
```go
func (err *ErrAggregatedEvent) Payload() []byte
```
`error` type returned if no `Handler`s was found for particular `Event`
```go
type ErrCommandHandlerNotFound struct {
	// Has unexported fields.
}
```
Implementation of `error`
```go
func (err *ErrCommandHandlerNotFound) Error() string
```
Returns new `*ErrEvent` caused by `event`  
`*ErrEvent` implements `error` and `Event`
```go
func NewErrEvent(event Event, err error) *ErrEvent
```
Returns `*ErrEvent` as an `error`
```go
func (err *ErrEvent) Err() error
```
Implementation of `error`
```go
func (err *ErrEvent) Error() string
```
Returns underlying `Event`
```go
func (err *ErrEvent) Event() Event
```
Returns error event type
```go
func (err *ErrEvent) EventType() string
```
Returns payload of event caused the error
```go
func (err *ErrEvent) Payload() []byte
```
Returns underlying `error`
```go
func (err *ErrEvent) Unwrap() error
```
`error` type returned if incorrect `Handler` was passed to `Commands`
```go
type ErrIncorrectHandler struct {
	// Has unexported fields.
}
```
Implementation of `error`
```go
func (err *ErrIncorrectHandler) Error() string
```
`error` type returned if `Event` equals `nil`  
`ErrNilEvent` implements `error` and `Event`
```go
type ErrNilEvent string
```
`ErrNilEvent` instance
```go
const NilEvent ErrNilEvent = "NilEvent"
```
Returns `ErrNilEvent` as an `error`
```go
func (err ErrNilEvent) Err() error
```
Implementation of `error`
```go
func (err ErrNilEvent) Error() string
```
Returns error event type
```go
func (err ErrNilEvent) EventType() string
```
Returns `nil`
```go
func (err ErrNilEvent) Payload() []byte
```

## Queries
`import "github.com/andriiyaremenko/tinycqs/query"`


### TYPES
Handles single type of query
```go
type Handler interface {
	QueryName() string
	Handle(ctx context.Context, payload []byte) ([]byte, error)
}
```

Returns `Handler` with `QueryName` equals `queryName` and `Handle` based on `handle`
```go
func QueryHandlerFunc(queryName string, handle func(context.Context, []byte) ([]byte, error)) Handler
```

Interface to handle queries
```go
type Queries interface {
	Handle(ctx context.Context, query string, payload []byte) ([]byte, error)
	HandleJSONEncoded(ctx context.Context, query string, v interface{}, payload []byte) error
}
```

Returns new `Queries` or `error`
```go
func NewQueries(handlers ...Handler) (Queries, error)
```

### Errors

`error` type returned if incorrect `Handler` was passed to `Queries`
```go
type ErrIncorrectHandler struct {
	// Has unexported fields.
}
```

Implementation of `error`
```go
func (err *ErrIncorrectHandler) Error() string
```

`error` type returned if no `Handler`s was found for particular query
```go
type ErrQueryHandlerNotFound struct {
	// Has unexported fields.
}
```

returns `*ErrQueryHandlerNotFound` of `queryName`
```go
func NewErrQueryHandlerNotFound(queryName string) *ErrQueryHandlerNotFound
```

Implementation of `error`
```go
func (err *ErrQueryHandlerNotFound) Error() string
```
