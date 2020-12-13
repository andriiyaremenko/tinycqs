package tinycqs

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/andriiyaremenko/tinycqs/query"
	"github.com/stretchr/testify/assert"
)

func TestQuery(t *testing.T) {
	t.Run("Should be able to create query and handle query", testCanCreateQuery)
	t.Run("Should be able to create query, handle query and decode result from JSON", testCanCreateQueryAndHandleJSONEncoded)
	t.Run("Query should error if ho handlers exists matching query", testQueryShouldErrIfNoHandlersMatch)
}

func testCanCreateQuery(t *testing.T) {
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

func testCanCreateQueryAndHandleJSONEncoded(t *testing.T) {
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

func testQueryShouldErrIfNoHandlersMatch(t *testing.T) {
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
