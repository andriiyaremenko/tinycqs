package tinycqs

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/andriiyaremenko/tinycqs/query"
	"github.com/stretchr/testify/assert"
)

func TestQuery(t *testing.T) {
	t.Run("Should be able to create query and handle query", testCanCreateQuery)
	t.Run("Should be able to create query, handle query and decode result from JSON", testCanCreateQueryAndHandleJSONEncoded)
	t.Run("Query should error if ho handlers exists matching query", testQueryShouldErrIfNoHandlersMatch)
	t.Run("Test heavy query", testHeavyQuery)
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
	qResult := <-q.Handle(ctx, "test_1", nil)
	v, err := qResult.Body(), qResult.Err()
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
	qResult := <-q.Handle(ctx, "test_1", nil)
	err := qResult.UnmarshalJSONBody(&str)

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
	qResult := <-q.Handle(ctx, "test_2", nil)
	v, err := qResult.Body(), qResult.Err()

	assert.Nil(v, "value should be nil")
	assert.EqualError(err, "handler not found for query test_2", "error should be returned")
	assert.IsType(&query.ErrQueryHandlerNotFound{}, err, "error should be of type *query.ErrQueryHandlerNotFound")
}

func testHeavyQuery(t *testing.T) {
	ctx := context.TODO()
	ctx, cancel := context.WithCancel(ctx)

	defer cancel()

	assert := assert.New(t)
	handler := &query.QueryHandler{
		QName: "test_1",
		HandleFunc: func(ctx context.Context, getWriter query.QueryWriterWithCancel, _ []byte) {
			write, done := getWriter()

			defer done()

			write(query.Q{Name: "test_1", B: []byte("success 1")})
			write(query.Q{Name: "test_1", B: []byte("success 2")})
			write(query.Q{Name: "test_1", B: []byte("success 3")})
		}}
	q, _ := query.NewQueries(
		handler,
	)

	i := 1
	for qResult := range q.Handle(ctx, "test_1", nil) {
		v, err := qResult.Body(), qResult.Err()
		assert.NoError(err, "no error should be returned")
		assert.Equal(fmt.Sprintf("success %d", i), string(v))
		i++
	}
}
