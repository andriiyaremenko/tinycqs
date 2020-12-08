package tinycqs

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/andriiyaremenko/tinycqs/command"
	"github.com/andriiyaremenko/tinycqs/query"
	"github.com/stretchr/testify/assert"
)

func TestEvImplementsEvent(t *testing.T) {
	t.Log("Ev should implement Event interface")

	assert := assert.New(t)

	assert.Implementsf((*command.Event)(nil), command.E{}, "no error should be returned")
	assert.Implementsf((*command.Event)(nil), &command.E{}, "no error should be returned")
}

func TestCanCreateCommand(t *testing.T) {
	t.Log("Should be able to create command and handle command")

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

func TestCanCreateQuery(t *testing.T) {
	t.Log("Should be able to create query and handle query")

	ctx := context.TODO()
	ctx, cancel := context.WithCancel(ctx)

	defer cancel()

	assert := assert.New(t)
	handler := func(ctx context.Context, _ []byte) ([]byte, error) {
		return []byte("works"), nil
	}
	q, _ := query.NewQueries(
		query.QueryHandlerFunc("test_1", handler),
	)
	v, err := q.Handle(ctx, "test_1", nil)
	assert.NoError(err, "no error should be returned")
	assert.Equal("works", string(v))
}

func TestCanCreateQueryAndHandleJSONEncoded(t *testing.T) {
	t.Log("Should be able to create query and handle query")

	ctx := context.TODO()
	ctx, cancel := context.WithCancel(ctx)

	defer cancel()

	assert := assert.New(t)
	handler := func(ctx context.Context, _ []byte) ([]byte, error) {
		return json.Marshal("works")
	}
	q, _ := query.NewQueries(
		query.QueryHandlerFunc("test_1", handler),
	)

	var str string
	err := q.HandleJSONEncoded(ctx, "test_1", &str, nil)

	assert.NoError(err, "no error should be returned")
	assert.Equal("works", str)
}

func TestCommandShouldErrIfNoHandlersMatch(t *testing.T) {
	t.Log("Command should error if ho handlers exists matching command")

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

func TestQueryShouldErrIfNoHandlersMatch(t *testing.T) {
	t.Log("Query should error if ho handlers exists matching query")

	ctx := context.TODO()
	ctx, cancel := context.WithCancel(ctx)

	defer cancel()

	assert := assert.New(t)
	handler := func(ctx context.Context, _ []byte) ([]byte, error) {
		return []byte("works"), nil
	}
	q, _ := query.NewQueries(
		query.QueryHandlerFunc("test_1", handler),
	)
	v, err := q.Handle(ctx, "test_2", nil)
	assert.Nil(v, "value should be nil")
	assert.EqualError(err, "handler not found for query test_2", "error should be returned")
	assert.IsType(&query.ErrQueryHandlerNotFound{}, err, "error should be of type *query.ErrQueryHandlerNotFound")
}

func TestCommandCanHandleOnlyListOfCommands(t *testing.T) {
	t.Log("Command should be able to handle commands in only list")

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

func TestCommandHandleOnlyIgnoresCommandsAbsentInList(t *testing.T) {
	t.Log("Command should be able to ignore commands absent in only list")

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

func TestCommandHandleOnlyShouldErrIfNoHandlersMatch(t *testing.T) {
	t.Log("Command should error if no handlers match command and command is in list")

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

func TestCommandHandleOnlyShouldNotChainEvents(t *testing.T) {
	t.Log("HandleOnly should ignore events chaining")

	ctx := context.TODO()
	ctx, cancel := context.WithCancel(ctx)

	defer cancel()

	assert := assert.New(t)
	handler1 := &command.CommandHandler{
		EType: "test_1",
		HandleFunc: func(ctx context.Context, _ command.Event) <-chan command.Event {
			respCh := make(chan command.Event)
			go func() {
				defer close(respCh)

				respCh <- command.E{EType: "test_2"}
			}()

			return respCh
		}}
	handler2WasCalled := false
	handler2 := &command.CommandHandler{
		EType: "test_2",
		HandleFunc: func(ctx context.Context, _ command.Event) <-chan command.Event {
			respCh := make(chan command.Event)
			go func() {
				defer close(respCh)

				handler2WasCalled = true
				respCh <- command.E{EType: "test_3"}
			}()

			return respCh
		}}

	c, _ := command.NewCommands(
		handler1,
		handler2,
	)

	err := c.HandleOnly(ctx, command.E{EType: "test_1"})
	assert.NoError(err.Err(), "no error should be returned")
	assert.False(handler2WasCalled, "second handler should not have been called")
}

func TestCommandHandleChainEvents(t *testing.T) {
	t.Log("Command should be able to chain events")

	ctx := context.TODO()
	ctx, cancel := context.WithCancel(ctx)

	defer cancel()

	assert := assert.New(t)
	handler1 := &command.CommandHandler{
		EType: "test_1",
		HandleFunc: func(ctx context.Context, _ command.Event) <-chan command.Event {
			respCh := make(chan command.Event)
			go func() {
				defer close(respCh)

				respCh <- command.E{EType: "test_2"}
			}()

			return respCh
		}}
	handler2WasCalled := false
	handler2 := &command.CommandHandler{
		EType: "test_2",
		HandleFunc: func(ctx context.Context, _ command.Event) <-chan command.Event {
			respCh := make(chan command.Event)
			go func() {
				defer close(respCh)

				handler2WasCalled = true
				respCh <- command.Done
			}()

			return respCh
		}}

	c, _ := command.NewCommands(
		handler1,
		handler2,
	)

	err := c.Handle(ctx, command.E{EType: "test_1"})
	assert.NoError(err.Err(), "no error should be returned")
	assert.True(handler2WasCalled, "second handler should have been called")
}

func TestCommandHandleChainEventsShouldExhaustOrErr(t *testing.T) {
	t.Log("Command should error if no handler found while chaining events")

	ctx := context.TODO()
	ctx, cancel := context.WithCancel(ctx)

	defer cancel()

	assert := assert.New(t)
	handler1 := &command.CommandHandler{
		EType: "test_1",
		HandleFunc: func(ctx context.Context, _ command.Event) <-chan command.Event {
			respCh := make(chan command.Event)
			go func() {
				defer close(respCh)

				respCh <- command.E{EType: "test_2"}
			}()

			return respCh
		}}
	handler2 := &command.CommandHandler{
		EType: "test_2",
		HandleFunc: func(ctx context.Context, _ command.Event) <-chan command.Event {
			respCh := make(chan command.Event)
			go func() {
				defer close(respCh)

				respCh <- command.E{EType: "test_3"}
			}()

			return respCh
		}}

	c, _ := command.NewCommands(
		handler1,
		handler2,
	)

	err := c.Handle(ctx, command.E{EType: "test_1"})
	assert.EqualError(err.Err(), "failed to process event test_3: handler not found for command test_3", "error should be returned")
	assert.IsType(&command.ErrEvent{}, err, "error should be of type *command.ErrEvent")
	assert.IsType(&command.ErrCommandHandlerNotFound{}, (err.(*command.ErrEvent)).Unwrap(), "underlying error should be of type *command.ErrCommandHandlerNotFound")
}

func TestCommandHandleChainEventsSeveralEvents(t *testing.T) {
	t.Log("Command should be able to chain events and handle several events from one handler")

	ctx := context.TODO()
	ctx, cancel := context.WithCancel(ctx)

	defer cancel()

	assert := assert.New(t)
	handler1 := &command.CommandHandler{
		EType: "test_1",
		HandleFunc: func(ctx context.Context, _ command.Event) <-chan command.Event {
			respCh := make(chan command.Event)
			go func() {
				defer close(respCh)

				respCh <- command.E{EType: "test_2"}
				respCh <- command.E{EType: "test_2"}
				respCh <- command.E{EType: "test_2"}
				respCh <- command.E{EType: "test_3"}
			}()

			return respCh
		}}
	handler2WasCalled := &wasCalledCounter{}
	handler2 := &command.CommandHandler{
		EType: "test_2",
		HandleFunc: func(ctx context.Context, _ command.Event) <-chan command.Event {
			respCh := make(chan command.Event)
			go func() {
				defer close(respCh)

				handler2WasCalled.increase()
				respCh <- command.E{EType: "test_3"}
			}()

			return respCh
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

func TestCommandHandleShouldRespectContext(t *testing.T) {
	t.Log("Command Handle should error if context was cancelled")

	ctx := context.TODO()
	ctx, cancel := context.WithTimeout(ctx, time.Second*2)

	defer cancel()

	assert := assert.New(t)

	handler1 := &command.CommandHandler{
		EType: "test_1",
		HandleFunc: func(ctx context.Context, _ command.Event) <-chan command.Event {
			respCh := make(chan command.Event)
			go func() {
				defer close(respCh)

				respCh <- command.E{EType: "test_2"}
				respCh <- command.E{EType: "test_2"}
				respCh <- command.E{EType: "test_2"}
				respCh <- command.E{EType: "test_3"}
			}()

			return respCh
		}}
	handler2 := &command.CommandHandler{
		EType: "test_2",
		HandleFunc: func(ctx context.Context, _ command.Event) <-chan command.Event {
			respCh := make(chan command.Event)
			go func() {
				defer close(respCh)

				respCh <- command.E{EType: "test_3"}
			}()

			return respCh
		}}

	handlerFunc3 := func(ctx context.Context, _ []byte) error {
		time.Sleep(time.Second * 4)
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

func TestCommandHandleOnlyShouldRespectContext(t *testing.T) {
	t.Log("Command HandleOnly should error if context was cancelled")

	ctx := context.TODO()
	ctx, cancel := context.WithTimeout(ctx, time.Second*2)

	defer cancel()

	assert := assert.New(t)

	handler := &command.CommandHandler{
		EType: "test_1",
		HandleFunc: func(ctx context.Context, _ command.Event) <-chan command.Event {
			respCh := make(chan command.Event)
			go func() {
				defer close(respCh)

				time.Sleep(time.Second * 4)
				respCh <- command.E{EType: "test_2"}
			}()

			return respCh
		}}
	c, _ := command.NewCommands(handler)

	err := c.HandleOnly(ctx, command.E{EType: "test_1"}, "test_1")
	assert.EqualError(err.Err(), "failed to process event test_1: context deadline exceeded", "error should be returned")
	assert.IsType(&command.ErrEvent{}, err, "error should be of type *command.ErrEvent")
}

func TestWorkerShouldStartAndHandleCommands(t *testing.T) {
	t.Log("Command Worker should start and Handle commands")

	ctx := context.TODO()
	ctx, cancel := context.WithCancel(ctx)

	defer cancel()

	assert := assert.New(t)
	handler1 := &command.CommandHandler{
		EType: "test_1",
		HandleFunc: func(ctx context.Context, _ command.Event) <-chan command.Event {
			respCh := make(chan command.Event)
			go func() {
				defer close(respCh)

				respCh <- command.E{EType: "test_2"}
				respCh <- command.E{EType: "test_2"}
				respCh <- command.E{EType: "test_2"}
				respCh <- command.E{EType: "test_3"}
			}()

			return respCh
		}}
	handler2WasCalled := &wasCalledCounter{}
	handler2 := &command.CommandHandler{
		EType: "test_2",
		HandleFunc: func(ctx context.Context, _ command.Event) <-chan command.Event {
			respCh := make(chan command.Event)
			go func() {
				defer close(respCh)

				handler2WasCalled.increase()
				respCh <- command.E{EType: "test_3"}
			}()

			return respCh
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
	eventSink := func(e command.Event) {
		t.Log(e)
	}

	w := command.NewWorker(ctx, eventSink, c)
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		err := w.Handle(command.E{EType: "test_1"})
		assert.NoError(err, "no error should be returned")
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		err := w.Handle(command.E{EType: "test_1"})
		assert.NoError(err, "no error should be returned")
	}()

	wg.Wait()
	time.Sleep(time.Second)
	assert.Equal(6, handler2WasCalled.getCount(), "second handler should have been called three times")
	assert.Equal(8, handler3WasCalled.getCount(), "third handler should have been called four times")
}

func TestCommandHandleChainEventsShouldUseErrorHandlers(t *testing.T) {
	t.Log("Command should be able to chain events and use registered error handlers")

	ctx := context.TODO()
	ctx, cancel := context.WithCancel(ctx)

	defer cancel()

	assert := assert.New(t)
	handler1 := &command.CommandHandler{
		EType: "test_1",
		HandleFunc: func(ctx context.Context, e command.Event) <-chan command.Event {
			respCh := make(chan command.Event)
			go func() {
				defer close(respCh)

				respCh <- command.E{EType: "test_2"}
				respCh <- command.E{EType: "test_2"}
				respCh <- command.E{EType: "test_2"}
				respCh <- command.NewErrEvent(e, errors.New("some error"))
				respCh <- command.E{EType: "test_3"}
			}()

			return respCh
		}}
	handler2WasCalled := &wasCalledCounter{}
	handler2 := &command.CommandHandler{
		EType: "test_2",
		HandleFunc: func(ctx context.Context, _ command.Event) <-chan command.Event {
			respCh := make(chan command.Event)
			go func() {
				defer close(respCh)

				handler2WasCalled.increase()
				respCh <- command.E{EType: "test_3"}
			}()

			return respCh
		}}

	handler3WasCalled := &wasCalledCounter{}
	handlerFunc3 := func(ctx context.Context, _ []byte) error {
		handler3WasCalled.increase()
		return nil
	}

	handlerErr := &command.CommandHandler{
		EType: command.ErrorEventType("test_1"),
		HandleFunc: func(ctx context.Context, e command.Event) <-chan command.Event {
			respCh := make(chan command.Event)
			go func() {
				defer close(respCh)
				respCh <- command.Done
			}()

			return respCh
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

func TestCommandHandleChainEventsShouldUseGlobalErrHandler(t *testing.T) {
	t.Log("Command should be able to chain events and use registered global error handler")

	ctx := context.TODO()
	ctx, cancel := context.WithCancel(ctx)

	defer cancel()

	assert := assert.New(t)
	handler1 := &command.CommandHandler{
		EType: "test_1",
		HandleFunc: func(ctx context.Context, e command.Event) <-chan command.Event {
			respCh := make(chan command.Event)
			go func() {
				defer close(respCh)

				respCh <- command.E{EType: "test_2"}
				respCh <- command.E{EType: "test_2"}
				respCh <- command.E{EType: "test_2"}
				respCh <- command.NewErrEvent(e, errors.New("some error"))
				respCh <- command.E{EType: "test_3"}
			}()

			return respCh
		}}
	handler2WasCalled := &wasCalledCounter{}
	handler2 := &command.CommandHandler{
		EType: "test_2",
		HandleFunc: func(ctx context.Context, _ command.Event) <-chan command.Event {
			respCh := make(chan command.Event)
			go func() {
				defer close(respCh)

				handler2WasCalled.increase()
				respCh <- command.E{EType: "test_3"}
			}()

			return respCh
		}}

	handler3WasCalled := &wasCalledCounter{}
	handlerFunc3 := func(ctx context.Context, _ []byte) error {
		handler3WasCalled.increase()
		return nil
	}

	handlerErr := &command.CommandHandler{
		EType: command.CatchAllErrorEventType,
		HandleFunc: func(ctx context.Context, e command.Event) <-chan command.Event {
			respCh := make(chan command.Event)
			go func() {
				defer close(respCh)
				respCh <- command.Done
			}()

			return respCh
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

func TestCommandYouCAnUseOnlyOneUseGlobalErrHandler(t *testing.T) {
	t.Log("Command should restrict only one global error handler")

	ctx := context.TODO()
	ctx, cancel := context.WithCancel(ctx)

	defer cancel()

	assert := assert.New(t)
	handlerErr1 := &command.CommandHandler{
		EType: command.CatchAllErrorEventType,
		HandleFunc: func(ctx context.Context, e command.Event) <-chan command.Event {
			respCh := make(chan command.Event)
			go func() {
				defer close(respCh)
				respCh <- command.Done
			}()

			return respCh
		}}
	handlerErr2 := &command.CommandHandler{
		EType: command.CatchAllErrorEventType,
		HandleFunc: func(ctx context.Context, e command.Event) <-chan command.Event {
			respCh := make(chan command.Event)
			go func() {
				defer close(respCh)
				respCh <- command.Done
			}()

			return respCh
		}}

	_, err := command.NewCommands(
		handlerErr1,
		handlerErr2,
	)

	assert.EqualError(err, command.MoreThanOneCatchAllErrorHandler.Error(), "error should be returned")
}

type wasCalledCounter struct {
	mu    sync.Mutex
	count int
}

func (cc *wasCalledCounter) increase() {
	cc.mu.Lock()
	cc.count++
	cc.mu.Unlock()
}

func (cc *wasCalledCounter) getCount() int {
	cc.mu.Lock()
	defer cc.mu.Unlock()
	return cc.count
}
