package tinycqs

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/andriiyaremenko/tinycqs/command"
	"github.com/stretchr/testify/assert"
)

func TestCommandWorker(t *testing.T) {
	t.Run("Command Worker should start and Handle commands", testWorkerShouldStartAndHandleCommands)
}

func testWorkerShouldStartAndHandleCommands(t *testing.T) {
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
	eventSink := func(e command.Event) {
		t.Log(e.EventType())
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
	time.Sleep(time.Millisecond * 200)
	assert.Equal(6, handler2WasCalled.getCount(), "second handler should have been called three times")
	assert.Equal(8, handler3WasCalled.getCount(), "third handler should have been called four times")
}
