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
	q, _ := query.New(
		query.HandlerFunc("test_1", handler),
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
	q, _ := query.New(
		query.HandlerFunc("test_1", handler),
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
	q, _ := query.New(
		query.HandlerFunc("test_1", handler),
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
	handler := &query.BaseHandler{
		Name: "test_1",
		HandleFunc: func(ctx context.Context, w query.ResultWriter, _ []byte) {
			defer w.Done()

			w.Write(query.Q{Name: "test_1", B: []byte("success 1")})
			w.Write(query.Q{Name: "test_1", B: []byte("success 2")})
			w.Write(query.Q{Name: "test_1", B: []byte("success 3")})
		}}
	q, _ := query.New(
		handler,
	)

	compareList := []string{"success 1", "success 2", "success 3"}
	for qResult := range q.Handle(ctx, "test_1", nil) {
		v, err := qResult.Body(), qResult.Err()
		vString := string(v)
		idx := -1
		for i, compare := range compareList {
			if compare == vString {
				idx = i
				break
			}
		}
		assert.NoError(err, "no error should be returned")
		assert.Equal(compareList[idx], vString)
		compareList = append(compareList[:idx], compareList[idx+1:]...)
	}

	if len(compareList) != 0 {
		assert.Failf("not all results have been received", "%v", compareList)
	}
}
