package tinycqs

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/andriiyaremenko/tinycqs/command"
	"github.com/andriiyaremenko/tinycqs/query"
	"github.com/stretchr/testify/assert"
)

func TestEvInplementsEvent(t *testing.T) {
	t.Log("Ev should implement Event interface")

	assert := assert.New(t)

	assert.Implementsf((*command.Event)(nil), command.Ev{}, "no error should be returned")
	assert.Implementsf((*command.Event)(nil), &command.Ev{}, "no error should be returned")
}

func TestCanCreateCommand(t *testing.T) {
	t.Log("Should be able to create command and handle command")

	ctx := context.TODO()
	assert := assert.New(t)
	handler := func(ctx context.Context, _ []byte) error {
		return nil
	}
	c := command.NewCommand(
		command.CommandHandlerFunc("test_1", handler),
	)

	assert.NoError(c.Handle(ctx, command.Ev{Type: "test_1"}), "no error should be returned")
}

func TestCanCreateQuery(t *testing.T) {
	t.Log("Should be able to create query and handle query")

	ctx := context.TODO()
	assert := assert.New(t)
	handler := func(ctx context.Context, _ []byte) ([]byte, error) {
		return []byte("works"), nil
	}
	c := query.NewQuery(
		query.QueryHandlerFunc("test_1", handler),
	)
	v, err := c.Handle(ctx, "test_1", nil)
	assert.NoError(err, "no error should be returned")
	assert.Equal("works", string(v))
}

func TestCanCreateQueryAndHandleJSONEncoded(t *testing.T) {
	t.Log("Should be able to create query and handle query")

	ctx := context.TODO()
	assert := assert.New(t)
	handler := func(ctx context.Context, _ []byte) ([]byte, error) {
		return json.Marshal("works")
	}
	c := query.NewQuery(
		query.QueryHandlerFunc("test_1", handler),
	)

	var str string
	err := c.HandleJSONEncoded(ctx, "test_1", &str, nil)

	assert.NoError(err, "no error should be returned")
	assert.Equal("works", str)
}

func TestCommandShouldErrIfNoHandlersMatch(t *testing.T) {
	t.Log("Command should error if ho handlers exists matching command")

	ctx := context.TODO()
	assert := assert.New(t)
	handler := func(ctx context.Context, _ []byte) error {
		return nil
	}
	c := command.NewCommand(
		command.CommandHandlerFunc("test_1", handler),
	)
	err := c.Handle(ctx, command.Ev{Type: "test_2"})
	assert.EqualError(err, "handler not found for command test_2", "error should be returned")
	assert.IsType(&command.ErrCommandHandlerNotFound{}, err, "error should be of type *command.ErrCommandHandlerNotFound")
}

func TestQueryShouldErrIfNoHandlersMatch(t *testing.T) {
	t.Log("Query should error if ho handlers exists matching query")

	ctx := context.TODO()
	assert := assert.New(t)
	handler := func(ctx context.Context, _ []byte) ([]byte, error) {
		return []byte("works"), nil
	}
	c := query.NewQuery(
		query.QueryHandlerFunc("test_1", handler),
	)
	v, err := c.Handle(ctx, "test_2", nil)
	assert.Nil(v, "value should be nil")
	assert.EqualError(err, "handler not found for query test_2", "error should be returned")
	assert.IsType(&query.ErrQueryHandlerNotFound{}, err, "error should be of type *query.ErrQueryHandlerNotFound")
}

func TestCommandCanHandleOnlyListOfCommands(t *testing.T) {
	t.Log("Command should be able to handle commands in only list")

	ctx := context.TODO()
	assert := assert.New(t)
	handler := func(ctx context.Context, _ []byte) error {
		return nil
	}
	c := command.NewCommand(
		command.CommandHandlerFunc("test_1", handler),
	)

	assert.NoError(c.HandleOnly(ctx, command.Ev{Type: "test_1"}, "test_1"), "no error should be returned")
}

func TestCommandHandleOnlyIgnoresCommandsAbsentInList(t *testing.T) {
	t.Log("Command should be able to ignore commands absent in only list")

	ctx := context.TODO()
	assert := assert.New(t)
	handler := func(ctx context.Context, _ []byte) error {
		return nil
	}
	c := command.NewCommand(
		command.CommandHandlerFunc("test_1", handler),
	)

	assert.NoError(c.HandleOnly(ctx, command.Ev{Type: "test_2"}, "test_1"), "no error should be returned")
}

func TestCommandHandleOnlyShouldErrIfNoHandlersMatch(t *testing.T) {
	t.Log("Command should error if no handlers match command and command is in list")

	ctx := context.TODO()
	assert := assert.New(t)
	handler := func(ctx context.Context, _ []byte) error {
		return nil
	}
	c := command.NewCommand(
		command.CommandHandlerFunc("test_1", handler),
	)

	err := c.HandleOnly(ctx, command.Ev{Type: "test_2"}, "test_1", "test_2")
	assert.EqualError(err, "handler not found for command test_2", "error should be returned")
	assert.IsType(&command.ErrCommandHandlerNotFound{}, err, "error should be of type *command.ErrCommandHandlerNotFound")
}
