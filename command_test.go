package tinycqs

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/andriiyaremenko/tinycqs/command"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestImplementations(t *testing.T) {
	t.Run("Ev should implement Event interface", testEvImplementsEvent)
}

func testEvImplementsEvent(t *testing.T) {
	assert := assert.New(t)

	assert.Implementsf((*command.Event)(nil), command.E{}, "no error should be returned")
	assert.Implementsf((*command.Event)(nil), &command.E{}, "no error should be returned")
}

func TestCommand(t *testing.T) {
	t.Run("Should be able to create command and handle command", testCanCreateCommand)
	t.Run("Command should error if ho handlers exists matching command",
		testCommandShouldErrIfNoHandlersMatch)
	t.Run("Command should be able to handle commands in only list",
		testCommandCanHandleOnlyListOfCommands)
	t.Run("Command should be able to ignore commands absent in only list",
		testCommandHandleOnlyIgnoresCommandsAbsentInList)
	t.Run("Command should error if no handlers match command and command is in list",
		testCommandHandleOnlyShouldErrIfNoHandlersMatch)
	t.Run("HandleOnly should ignore events chaining", testCommandHandleOnlyShouldNotChainEvents)
	t.Run("Command should be able to chain events", testCommandHandleChainEvents)
	t.Run("Command should error if no handler found while chaining events",
		testCommandHandleChainEventsShouldExhaustOrErr)
	t.Run("Command should be able to chain events and handle several events from one handler",
		testCommandHandleChainEventsSeveralEvents)
	t.Run("Command Handle should error if context was cancelled",
		testCommandHandleShouldRespectContext)
	t.Run("Command HandleOnly should error if context was cancelled",
		testCommandHandleOnlyShouldRespectContext)
	t.Run("Command should be able to chain events and use registered error handlers",
		testCommandHandleChainEventsShouldUseErrorHandlers)
	t.Run("Command should be able to chain events and use registered global error handler",
		testCommandHandleChainEventsShouldUseGlobalErrHandler)
	t.Run("Command should restrict only one global error handler",
		testCommandYouCanUseOnlyOneUseGlobalErrHandler)
	t.Run("Correlation IDs should be correctly resolved in Metadata",
		testMetadata)
}

func testCanCreateCommand(t *testing.T) {
	ctx := context.TODO()
	ctx, cancel := context.WithCancel(ctx)

	defer cancel()

	assert := assert.New(t)
	handler := func(ctx context.Context, _ []byte) error {
		return nil
	}
	c, _ := command.NewCommands(
		command.CommandHandlerFunc("test_1", handler),
	)

	assert.NoError(c.Handle(ctx, command.E{EType: "test_1"}).Err(), "no error should be returned")
}

func testCommandShouldErrIfNoHandlersMatch(t *testing.T) {
	ctx := context.TODO()
	ctx, cancel := context.WithCancel(ctx)

	defer cancel()

	assert := assert.New(t)
	handler := func(ctx context.Context, _ []byte) error {
		return nil
	}
	c, _ := command.NewCommands(
		command.CommandHandlerFunc("test_1", handler),
	)
	err := c.Handle(ctx, command.E{EType: "test_2"})
	assert.EqualError(err.Err(), "failed to process event test_2: handler not found for command test_2", "error should be returned")
	assert.IsType(&command.ErrEvent{}, err, "error should be of type *command.ErrEvent")
	assert.IsType(&command.ErrCommandHandlerNotFound{}, (err.(*command.ErrEvent)).Unwrap(), "underlying error should be of type *command.ErrCommandHandlerNotFound")
}

func testCommandCanHandleOnlyListOfCommands(t *testing.T) {
	ctx := context.TODO()
	ctx, cancel := context.WithCancel(ctx)

	defer cancel()

	assert := assert.New(t)
	handler := func(ctx context.Context, _ []byte) error {
		return nil
	}
	c, _ := command.NewCommands(
		command.CommandHandlerFunc("test_1", handler),
	)

	assert.NoError(c.HandleOnly(ctx, command.E{EType: "test_1"}, "test_1").Err(), "no error should be returned")
}

func testCommandHandleOnlyIgnoresCommandsAbsentInList(t *testing.T) {
	ctx := context.TODO()
	ctx, cancel := context.WithCancel(ctx)

	defer cancel()

	assert := assert.New(t)
	handler := func(ctx context.Context, _ []byte) error {
		return nil
	}
	c, _ := command.NewCommands(
		command.CommandHandlerFunc("test_1", handler),
	)

	assert.NoError(c.HandleOnly(ctx, command.E{EType: "test_2"}, "test_1").Err(), "no error should be returned")
}

func testCommandHandleOnlyShouldErrIfNoHandlersMatch(t *testing.T) {
	ctx := context.TODO()
	ctx, cancel := context.WithCancel(ctx)

	defer cancel()

	assert := assert.New(t)
	handler := func(ctx context.Context, _ []byte) error {
		return nil
	}
	c, _ := command.NewCommands(
		command.CommandHandlerFunc("test_1", handler),
	)

	err := c.HandleOnly(ctx, command.E{EType: "test_2"}, "test_1", "test_2")
	assert.EqualError(err.Err(), "failed to process event test_2: handler not found for command test_2", "error should be returned")
	assert.IsType(&command.ErrEvent{}, err, "error should be of type *command.ErrEvent")
	assert.IsType(&command.ErrCommandHandlerNotFound{}, (err.(*command.ErrEvent)).Unwrap(), "underlying error should be of type *command.ErrCommandHandlerNotFound")
}

func testCommandHandleOnlyShouldNotChainEvents(t *testing.T) {
	ctx := context.TODO()
	ctx, cancel := context.WithCancel(ctx)

	defer cancel()

	assert := assert.New(t)
	handler1 := &command.CommandHandler{
		EType: "test_1",
		HandleFunc: func(ctx context.Context, r command.EventWriter, _ command.Event) {
			defer r.Done()

			r.Write(command.E{EType: "test_2"})
		}}
	handler2WasCalled := false
	handler2 := &command.CommandHandler{
		EType: "test_2",
		HandleFunc: func(ctx context.Context, r command.EventWriter, _ command.Event) {
			defer r.Done()

			handler2WasCalled = true
			r.Write(command.E{EType: "test_3"})
		}}

	c, _ := command.NewCommands(
		handler1,
		handler2,
	)

	err := c.HandleOnly(ctx, command.E{EType: "test_1"})
	assert.NoError(err.Err(), "no error should be returned")
	assert.False(handler2WasCalled, "second handler should not have been called")
}

func testCommandHandleChainEvents(t *testing.T) {
	ctx := context.TODO()
	ctx, cancel := context.WithCancel(ctx)

	defer cancel()

	assert := assert.New(t)
	handler1 := &command.CommandHandler{
		EType: "test_1",
		HandleFunc: func(ctx context.Context, r command.EventWriter, _ command.Event) {
			defer r.Done()

			r.Write(command.E{EType: "test_2"})
		}}
	handler2WasCalled := false
	handler2 := &command.CommandHandler{
		EType: "test_2",
		HandleFunc: func(ctx context.Context, r command.EventWriter, _ command.Event) {
			defer r.Done()

			handler2WasCalled = true
		}}

	c, _ := command.NewCommands(
		handler1,
		handler2,
	)

	err := c.Handle(ctx, command.E{EType: "test_1"})
	assert.NoError(err.Err(), "no error should be returned")
	assert.True(handler2WasCalled, "second handler should have been called")
}

func testCommandHandleChainEventsShouldExhaustOrErr(t *testing.T) {
	ctx := context.TODO()
	ctx, cancel := context.WithCancel(ctx)

	defer cancel()

	assert := assert.New(t)
	handler1 := &command.CommandHandler{
		EType: "test_1",
		HandleFunc: func(ctx context.Context, r command.EventWriter, _ command.Event) {
			defer r.Done()

			r.Write(command.E{EType: "test_2"})
		}}
	handler2 := &command.CommandHandler{
		EType: "test_2",
		HandleFunc: func(ctx context.Context, r command.EventWriter, _ command.Event) {
			defer r.Done()

			r.Write(command.E{EType: "test_3"})
		}}

	c, _ := command.NewCommands(
		handler1,
		handler2,
	)

	err := c.Handle(ctx, command.E{EType: "test_1"})
	assert.EqualError(err.Err(), "failed to process event test_1: aggregated error occurred: [\n\thandler not found for command test_3\n]", "error should be returned")
	assert.IsType(&command.ErrAggregatedEvent{}, err, "error should be of type *command.ErrEvent")
}

func testCommandHandleChainEventsSeveralEvents(t *testing.T) {
	ctx := context.TODO()
	ctx, cancel := context.WithCancel(ctx)

	defer cancel()

	assert := assert.New(t)
	handler1 := &command.CommandHandler{
		EType: "test_1",
		HandleFunc: func(ctx context.Context, r command.EventWriter, _ command.Event) {
			defer r.Done()

			r.Write(command.E{EType: "test_2"})
			r.Write(command.E{EType: "test_2"})
			r.Write(command.E{EType: "test_2"})
			r.Write(command.E{EType: "test_3"})
		}}
	handler2WasCalled := &wasCalledCounter{}
	handler2 := &command.CommandHandler{
		EType: "test_2",
		HandleFunc: func(ctx context.Context, r command.EventWriter, _ command.Event) {
			defer r.Done()

			handler2WasCalled.increase()
			r.Write(command.E{EType: "test_3"})
		}}

	handler3WasCalled := &wasCalledCounter{}
	handlerFunc3 := func(ctx context.Context, _ []byte) error {
		handler3WasCalled.increase()
		return nil
	}
	c, _ := command.NewCommandsWithConcurrencyLimit(
		20,
		handler1,
		handler2,
		command.CommandHandlerFunc("test_3", handlerFunc3),
	)

	err := c.Handle(ctx, command.E{EType: "test_1"})
	assert.NoError(err.Err(), "no error should be returned")
	assert.Equal(3, handler2WasCalled.count, "second handler should have been called three times")
	assert.Equal(4, handler3WasCalled.count, "third handler should have been called four times")
}

func testCommandHandleShouldRespectContext(t *testing.T) {
	ctx := context.TODO()
	ctx, cancel := context.WithTimeout(ctx, time.Millisecond*200)

	defer cancel()

	assert := assert.New(t)

	handler1 := &command.CommandHandler{
		EType: "test_1",
		HandleFunc: func(ctx context.Context, r command.EventWriter, _ command.Event) {
			defer r.Done()

			r.Write(command.E{EType: "test_2"})
			r.Write(command.E{EType: "test_2"})
			r.Write(command.E{EType: "test_2"})
			r.Write(command.E{EType: "test_3"})
		}}
	handler2 := &command.CommandHandler{
		EType: "test_2",
		HandleFunc: func(ctx context.Context, r command.EventWriter, _ command.Event) {
			defer r.Done()

			r.Write(command.E{EType: "test_3"})
		}}

	handlerFunc3 := func(ctx context.Context, _ []byte) error {
		time.Sleep(time.Millisecond * 400)
		return nil
	}
	c, _ := command.NewCommandsWithConcurrencyLimit(
		20,
		handler1,
		handler2,
		command.CommandHandlerFunc("test_3", handlerFunc3),
	)

	err := c.Handle(ctx, command.E{EType: "test_1"})
	assert.EqualError(err.Err(), "failed to process event test_1: context deadline exceeded", "error should be returned")
	assert.IsType(&command.ErrEvent{}, err, "error should be of type *command.ErrEvent")
}

func testCommandHandleOnlyShouldRespectContext(t *testing.T) {
	ctx := context.TODO()
	ctx, cancel := context.WithTimeout(ctx, time.Millisecond*200)

	defer cancel()

	assert := assert.New(t)

	handler := &command.CommandHandler{
		EType: "test_1",
		HandleFunc: func(ctx context.Context, r command.EventWriter, _ command.Event) {
			defer r.Done()

			time.Sleep(time.Millisecond * 400)
			r.Write(command.E{EType: "test_2"})
		}}
	c, _ := command.NewCommands(handler)

	err := c.HandleOnly(ctx, command.E{EType: "test_1"}, "test_1")
	assert.EqualError(err.Err(), "failed to process event test_1: context deadline exceeded", "error should be returned")
	assert.IsType(&command.ErrEvent{}, err, "error should be of type *command.ErrEvent")
}

func testCommandHandleChainEventsShouldUseErrorHandlers(t *testing.T) {
	ctx := context.TODO()
	ctx, cancel := context.WithCancel(ctx)

	defer cancel()

	assert := assert.New(t)
	handler1 := &command.CommandHandler{
		EType: "test_1",
		HandleFunc: func(ctx context.Context, r command.EventWriter, e command.Event) {
			defer r.Done()

			r.Write(command.E{EType: "test_2"})
			r.Write(command.E{EType: "test_2"})
			r.Write(command.E{EType: "test_2"})
			r.Write(command.NewErrEvent(e, errors.New("some error")))
			r.Write(command.E{EType: "test_3"})
		}}
	handler2WasCalled := &wasCalledCounter{}
	handler2 := &command.CommandHandler{
		EType: "test_2",
		HandleFunc: func(ctx context.Context, r command.EventWriter, _ command.Event) {
			defer r.Done()

			handler2WasCalled.increase()
			r.Write(command.E{EType: "test_3"})
		}}

	handler3WasCalled := &wasCalledCounter{}
	handlerFunc3 := func(ctx context.Context, _ []byte) error {
		handler3WasCalled.increase()
		return nil
	}

	handlerErr := &command.CommandHandler{
		EType: command.ErrorEventType("test_1"),
		HandleFunc: func(ctx context.Context, r command.EventWriter, e command.Event) {
			r.Done()
		}}
	c, _ := command.NewCommandsWithConcurrencyLimit(
		20,
		handler1,
		handler2,
		handlerErr,
		command.CommandHandlerFunc("test_3", handlerFunc3),
	)

	err := c.Handle(ctx, command.E{EType: "test_1"})
	assert.NoError(err.Err(), "no error should be returned")
	assert.Equal(3, handler2WasCalled.count, "second handler should have been called three times")
	assert.Equal(4, handler3WasCalled.count, "third handler should have been called four times")
}

func testCommandHandleChainEventsShouldUseGlobalErrHandler(t *testing.T) {
	ctx := context.TODO()
	ctx, cancel := context.WithCancel(ctx)

	defer cancel()

	assert := assert.New(t)
	handler1 := &command.CommandHandler{
		EType: "test_1",
		HandleFunc: func(ctx context.Context, r command.EventWriter, e command.Event) {
			defer r.Done()

			r.Write(command.E{EType: "test_2"})
			r.Write(command.E{EType: "test_2"})
			r.Write(command.E{EType: "test_2"})
			r.Write(command.NewErrEvent(e, errors.New("some error")))
			r.Write(command.E{EType: "test_3"})
		}}
	handler2WasCalled := &wasCalledCounter{}
	handler2 := &command.CommandHandler{
		EType: "test_2",
		HandleFunc: func(ctx context.Context, r command.EventWriter, _ command.Event) {
			defer r.Done()

			handler2WasCalled.increase()
			r.Write(command.E{EType: "test_3"})
		}}

	handler3WasCalled := &wasCalledCounter{}
	handlerFunc3 := func(ctx context.Context, _ []byte) error {
		handler3WasCalled.increase()
		return nil
	}

	handlerErr := &command.CommandHandler{
		EType: command.CatchAllErrorEventType,
		HandleFunc: func(ctx context.Context, r command.EventWriter, e command.Event) {
			r.Done()
		}}
	c, _ := command.NewCommandsWithConcurrencyLimit(
		20,
		handler1,
		handler2,
		handlerErr,
		command.CommandHandlerFunc("test_3", handlerFunc3),
	)

	err := c.Handle(ctx, command.E{EType: "test_1"})
	assert.NoError(err.Err(), "no error should be returned")
	assert.Equal(3, handler2WasCalled.count, "second handler should have been called three times")
	assert.Equal(4, handler3WasCalled.count, "third handler should have been called four times")
}

func testCommandYouCanUseOnlyOneUseGlobalErrHandler(t *testing.T) {
	ctx := context.TODO()
	ctx, cancel := context.WithCancel(ctx)

	defer cancel()

	assert := assert.New(t)
	handlerErr1 := &command.CommandHandler{
		EType: command.CatchAllErrorEventType,
		HandleFunc: func(ctx context.Context, r command.EventWriter, e command.Event) {
			r.Done()
		}}
	handlerErr2 := &command.CommandHandler{
		EType: command.CatchAllErrorEventType,
		HandleFunc: func(ctx context.Context, r command.EventWriter, e command.Event) {
			r.Done()
		}}

	_, err := command.NewCommands(
		handlerErr1,
		handlerErr2,
	)

	assert.EqualError(err, command.MoreThanOneCatchAllErrorHandler.Error(), "error should be returned")
}

func testMetadata(t *testing.T) {
	ctx := context.TODO()
	ctx, cancel := context.WithCancel(ctx)

	defer cancel()

	assert := assert.New(t)
	id := uuid.New().String()
	correlationID := uuid.New().String()
	causationID := uuid.New().String()
	handler1 := &command.CommandHandler{
		EType: "test_1",
		HandleFunc: func(ctx context.Context, r command.EventWriter, e command.Event) {
			defer r.Done()

			withMetadata := command.AsEventWithMetadata(e)
			assert.NotNil(withMetadata, "metadata should be presented in first event")
			assert.Equal(id, withMetadata.Metadata().ID(), "ID should equal ID in first event")
			assert.Equal(correlationID, withMetadata.Metadata().CorrelationID(),
				"correlation ID should equal correlationID in first event")
			assert.Equal(causationID, withMetadata.Metadata().CausationID(),
				"causation ID should equal causationID in first event")

			r.Write(command.E{EType: "test_2"})
			r.Write(command.E{EType: "test_2"})
			r.Write(command.E{EType: "test_2"})
			r.Write(command.E{EType: "test_3"})
		}}
	handler2 := &command.CommandHandler{
		EType: "test_2",
		HandleFunc: func(ctx context.Context, r command.EventWriter, e command.Event) {
			defer r.Done()

			withMetadata := command.AsEventWithMetadata(e)
			assert.NotNil(withMetadata, "metadata should be presented in second event")
			assert.NotEqual(id, withMetadata.Metadata().ID(), "ID should not equal ID in second event")
			assert.Equal(correlationID, withMetadata.Metadata().CorrelationID(),
				"correlation ID should equal correlationID in second event")
			assert.Equal(id, withMetadata.Metadata().CausationID(),
				"causation ID should equal id in second event")

			r.Write(command.E{EType: "test_3"})
		}}
	handler3 := &command.CommandHandler{
		EType: "test_3",
		HandleFunc: func(ctx context.Context, r command.EventWriter, e command.Event) {
			defer r.Done()

			withMetadata := command.AsEventWithMetadata(e)
			assert.NotNil(withMetadata, "metadata should be presented in third event")
			assert.NotEqual(id, withMetadata.Metadata().ID(), "ID should not equal ID in third event")
			assert.Equal(correlationID, withMetadata.Metadata().CorrelationID(),
				"correlation ID should equal correlationID in third event")
		}}

	c, _ := command.NewCommandsWithConcurrencyLimit(
		20,
		handler1,
		handler2,
		handler3,
	)
	ev := c.Handle(ctx,
		command.WithMetadata(command.E{EType: "test_1"},
			command.M{EID: id, ECorrelationID: correlationID, ECausationID: causationID}))

	assert.NoError(ev.Err(), "no error should be returned")

	unwrapped := command.Unwrap(ev)
	withMetadata := command.AsEventWithMetadata(unwrapped)
	assert.NotNil(withMetadata, "metadata should be presented in unwrapped done event")

	assert.Equal(id, withMetadata.Metadata().ID(), "ID should equal ID in unwrapped done event")
	assert.Equal(correlationID, withMetadata.Metadata().CorrelationID(),
		"correlation ID should equal correlationID in unwrapped done event")
	assert.Equal(causationID, withMetadata.Metadata().CausationID(),
		"causation ID should equal causationID in unwrapped done event")
}
