package tinycqs

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/andriiyaremenko/tinycqs/command"
	"github.com/andriiyaremenko/tinycqs/query"
	"github.com/stretchr/testify/assert"
)

func TestCanCreateCommandDemultiplexer(t *testing.T) {
	t.Log("Should be able to create command demultiplexer and handle command")

	ctx := context.TODO()
	assert := assert.New(t)
	handler := func(ctx context.Context, _ []byte) error {
		return nil
	}
	c := NewCommandDemultiplexer(
		CommandHandlerFunc("test_1", handler),
	)

	assert.NoError(c.Handle(ctx, "test_1", nil), "no error should be returned")
}

func TestCanCreateQueryDemultiplexer(t *testing.T) {
	t.Log("Should be able to create query demultiplexer and handle query")

	ctx := context.TODO()
	assert := assert.New(t)
	handler := func(ctx context.Context, _ []byte) ([]byte, error) {
		return []byte("works"), nil
	}
	c := NewQueryDemultiplexer(
		QueryHandlerFunc("test_1", handler),
	)
	v, err := c.Handle(ctx, "test_1", nil)
	assert.NoError(err, "no error should be returned")
	assert.Equal("works", string(v))
}

func TestCanCreateQueryDemultiplexerAndHandleJSONEncoded(t *testing.T) {
	t.Log("Should be able to create query demultiplexer and handle query")

	ctx := context.TODO()
	assert := assert.New(t)
	handler := func(ctx context.Context, _ []byte) ([]byte, error) {
		return json.Marshal("works")
	}
	c := NewQueryDemultiplexer(
		QueryHandlerFunc("test_1", handler),
	)

	var str string
	err := c.HandleJSONEncoded(ctx, "test_1", &str, nil)

	assert.NoError(err, "no error should be returned")
	assert.Equal("works", str)
}

func TestCommandDemultiplexerShouldErrIfNoHandlersMatch(t *testing.T) {
	t.Log("Command demultiplexer should error if ho handlers exists matching command")

	ctx := context.TODO()
	assert := assert.New(t)
	handler := func(ctx context.Context, _ []byte) error {
		return nil
	}
	c := NewCommandDemultiplexer(
		CommandHandlerFunc("test_1", handler),
	)
	err := c.Handle(ctx, "test_2", nil)
	assert.EqualError(err, "handler not found for command test_2", "error should be returned")
	assert.IsType(&command.ErrCommandHandlerNotFound{}, err, "error should be of type *command.ErrCommandHandlerNotFound")
}

func TestQueryDemultiplexerShouldErrIfNoHandlersMatch(t *testing.T) {
	t.Log("Query demultiplexer should error if ho handlers exists matching query")

	ctx := context.TODO()
	assert := assert.New(t)
	handler := func(ctx context.Context, _ []byte) ([]byte, error) {
		return []byte("works"), nil
	}
	c := NewQueryDemultiplexer(
		QueryHandlerFunc("test_1", handler),
	)
	v, err := c.Handle(ctx, "test_2", nil)
	assert.Nil(v, "value should be nil")
	assert.EqualError(err, "handler not found for query test_2", "error should be returned")
	assert.IsType(&query.ErrQueryHandlerNotFound{}, err, "error should be of type *query.ErrQueryHandlerNotFound")
}

func TestCommandDemultiplexerCanHandleOnlyListOfCommands(t *testing.T) {
	t.Log("Command demultiplexer should be able to handle commands in only list")

	ctx := context.TODO()
	assert := assert.New(t)
	handler := func(ctx context.Context, _ []byte) error {
		return nil
	}
	c := NewCommandDemultiplexer(
		CommandHandlerFunc("test_1", handler),
	)

	assert.NoError(c.HandleOnly(ctx, "test_1", nil, "test_1"), "no error should be returned")
}

func TestCommandDemultiplexerHandleOnlyIgnoresCommandsAbsentInList(t *testing.T) {
	t.Log("Command demultiplexer should be able to ignore commands absent in only list")

	ctx := context.TODO()
	assert := assert.New(t)
	handler := func(ctx context.Context, _ []byte) error {
		return nil
	}
	c := NewCommandDemultiplexer(
		CommandHandlerFunc("test_1", handler),
	)

	assert.NoError(c.HandleOnly(ctx, "test_2", nil, "test_1"), "no error should be returned")
}

func TestCommandDemultiplexerHandleOnlyShouldErrIfNoHandlersMatch(t *testing.T) {
	t.Log("Command demultiplexer should error if no handlers match command and command is in list")

	ctx := context.TODO()
	assert := assert.New(t)
	handler := func(ctx context.Context, _ []byte) error {
		return nil
	}
	c := NewCommandDemultiplexer(
		CommandHandlerFunc("test_1", handler),
	)

	err := c.HandleOnly(ctx, "test_2", nil, "test_1", "test_2")
	assert.EqualError(err, "handler not found for command test_2", "error should be returned")
	assert.IsType(&command.ErrCommandHandlerNotFound{}, err, "error should be of type *command.ErrCommandHandlerNotFound")
}
