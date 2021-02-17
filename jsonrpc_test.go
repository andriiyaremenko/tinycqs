package tinycqs

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/andriiyaremenko/tinycqs/command"
	"github.com/andriiyaremenko/tinycqs/jsonrpc"
	"github.com/andriiyaremenko/tinycqs/query"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

const (
	requestBody             = `{"jsonrpc": "2.0", "id": 1, "method": "test", "params": {"test": "test"}}`
	notificationRequestBody = `{"jsonrpc": "2.0", "method": "test", "params": {"test": "test"}}`
	batchBody               = `[
	{"jsonrpc": "2.0", "id": 1, "method": "test", "params": {"test": "test"}},
	{"jsonrpc": "2.0", "method": "test_1", "params": {"test": "test"}}
	]`
)

func TestJSONRPC(t *testing.T) {
	t.Run("Queries", TestQueries)
	t.Run("Commands", TestCommands)
	t.Run("Worker", TestWorker)
	t.Run("Handler", TestHandler)
}

func TestQueries(t *testing.T) {
	t.Run("Should return 404 on non POST requests", testQueriesShouldReturn404)
	t.Run("Should return 400 on request with invalid format", testQueriesShouldReturn400InvalidFormat)
	t.Run("Should return 400 on execution error", testQueriesShouldReturn400ExecutionError)
	t.Run("Should return result on successful execution", testQueriesShouldReturn200)
}

func TestCommands(t *testing.T) {
	t.Run("Should return 404 on non POST requests", testCommandsShouldReturn404)
	t.Run("Should return 400 on request with invalid format", testCommandsShouldReturn400InvalidFormat)
	t.Run("Should return 400 on execution error", testCommandsShouldReturn400ExecutionError)
	t.Run("Should return result on successful execution", testCommandsShouldReturn200)
}

func TestWorker(t *testing.T) {
	t.Run("Should return 404 on non POST requests", testWorkerShouldReturn404)
	t.Run("Should return 400 on request with invalid format", testWorkerShouldReturn400InvalidFormat)
	t.Run("Should return 400 on execution error", testWorkerShouldReturn400ExecutionError)
	t.Run("Should return result on successful execution", testWorkerShouldReturn200)
}

func TestHandler(t *testing.T) {
	t.Run("Should handle batch requests", testHandlerShouldHandleBatchRequests)
	t.Run("Should be able to marshal handler to JSON", testShouldMarshalHandlerToJSON)
}

func testShouldReturn404(assert *assert.Assertions, handler http.Handler) {
	ts := httptest.NewServer(handler)

	defer ts.Close()

	cl := http.Client{}

	u, err := url.Parse(ts.URL)
	if err != nil {
		assert.FailNow(err.Error())
	}

	req := &http.Request{
		Method: http.MethodGet,
		URL:    u,
		Header: http.Header{"Content-Type": []string{"application/json"}}}

	resp, err := cl.Do(req)

	assert.Equal(http.StatusNotFound, resp.StatusCode, "should return 404")
	assert.NoError(err, "no error should be returned")
	resp.Body.Close()
}

func testQueriesShouldReturn404(t *testing.T) {
	assert := assert.New(t)

	q, err := query.NewQueries()
	if err != nil {
		assert.FailNow(err.Error())
	}

	testShouldReturn404(assert, jsonrpc.Queries(q))
}

func testCommandsShouldReturn404(t *testing.T) {
	assert := assert.New(t)

	c, err := command.NewCommands()
	if err != nil {
		assert.FailNow(err.Error())
	}

	testShouldReturn404(assert, jsonrpc.Commands(c))
}

func testWorkerShouldReturn404(t *testing.T) {
	ctx := context.TODO()
	assert := assert.New(t)

	c, err := command.NewCommands()
	if err != nil {
		assert.FailNow(err.Error())
	}

	w := command.NewWorker(ctx, func(e command.Event) {}, c, 200)

	testShouldReturn404(assert, jsonrpc.CommandsWorker(w))
}

func testShouldReturn400InvalidFormat(assert *assert.Assertions, handler http.Handler) {
	ts := httptest.NewServer(handler)

	defer ts.Close()

	var b bytes.Buffer
	b.WriteString(`{"someWrongArguments": 1}`)

	cl := http.Client{}

	u, err := url.Parse(ts.URL)
	if err != nil {
		assert.FailNow(err.Error())
	}

	req := &http.Request{
		Method: http.MethodPost,
		URL:    u,
		Body:   ioutil.NopCloser(&b),
		Header: http.Header{"Content-Type": []string{"application/json"}}}

	resp, err := cl.Do(req)
	if err != nil {
		assert.FailNow(err.Error())
	}

	resp.Body.Close()
	assert.Equal(http.StatusBadRequest, resp.StatusCode, "should return 400")
	assert.NoError(err, "no error should be returned")
}

func testQueriesShouldReturn400InvalidFormat(t *testing.T) {
	assert := assert.New(t)
	q, err := query.NewQueries()
	if err != nil {
		assert.FailNow(err.Error())
	}

	testShouldReturn400InvalidFormat(assert, jsonrpc.Queries(q))
}

func testCommandsShouldReturn400InvalidFormat(t *testing.T) {
	assert := assert.New(t)
	c, err := command.NewCommands()
	if err != nil {
		assert.FailNow(err.Error())
	}

	testShouldReturn400InvalidFormat(assert, jsonrpc.Commands(c))
}

func testWorkerShouldReturn400InvalidFormat(t *testing.T) {
	ctx := context.TODO()
	assert := assert.New(t)

	c, err := command.NewCommands()
	if err != nil {
		assert.FailNow(err.Error())
	}

	w := command.NewWorker(ctx, func(e command.Event) {}, c, 200)

	testShouldReturn400InvalidFormat(assert, jsonrpc.CommandsWorker(w))
}

func testShouldReturn400ExecutionError(assert *assert.Assertions, handler http.Handler, body string, code int) {
	ts := httptest.NewServer(handler)

	defer ts.Close()

	var b bytes.Buffer
	b.WriteString(body)

	id := uuid.New().String()
	correlationID := uuid.New().String()
	causationID := uuid.New().String()
	cl := http.Client{}

	u, err := url.Parse(ts.URL)
	if err != nil {
		assert.FailNow(err.Error())
	}

	req := &http.Request{
		Method: http.MethodPost,
		URL:    u,
		Body:   ioutil.NopCloser(&b),
		Header: http.Header{
			"Content-Type":   []string{"application/json"},
			"Request_id":     []string{id},
			"Correlation_id": []string{correlationID},
			"Causation_id":   []string{causationID}}}
	resp, err := cl.Do(req)

	if code == http.StatusBadRequest {
		respBody, err := ioutil.ReadAll(resp.Body)

		if err != nil {
			assert.FailNow(err.Error())
		}

		response := new(jsonrpc.ErrorResponse)
		if err := json.Unmarshal(respBody, response); err != nil {
			assert.FailNowf("failed to read response", "%s: %s", err.Error(), string(body))
		}

		assert.EqualValues(1, response.ID, "json rpc request id should equal 1")
		assert.Equal("2.0", response.Version, `json rpc request version should equal "2.0"`)
	} else {
		resp.Body.Close()
	}

	assert.Equalf(code, resp.StatusCode, "should return %d", code)
	assert.NotEmpty(resp.Header["Request_id"], "request id should not be empty")
	assert.NotEqual([]string{id}, resp.Header["Request_id"], "request id should not equal incoming request id")
	assert.Equal([]string{correlationID}, resp.Header["Correlation_id"], "correlation id should equal incoming request correlation id")
	assert.Equal([]string{id}, resp.Header["Causation_id"], "causation id should equal incoming request id")
}

func testQueriesShouldReturn400ExecutionError(t *testing.T) {
	assert := assert.New(t)
	fn := func(ctx context.Context, _ []byte) ([]byte, error) {
		return nil, errors.New("fail")
	}
	q, err := query.NewQueries(query.QueryHandlerFunc("test", fn))

	if err != nil {
		assert.FailNow(err.Error())
	}

	testShouldReturn400ExecutionError(assert, jsonrpc.Queries(q), requestBody, http.StatusBadRequest)
}

func testCommandsShouldReturn400ExecutionError(t *testing.T) {
	assert := assert.New(t)
	fn := func(ctx context.Context, _ []byte) error {
		return errors.New("fail")
	}
	c, err := command.NewCommands(command.CommandHandlerFunc("test", fn))

	if err != nil {
		assert.FailNow(err.Error())
	}

	testShouldReturn400ExecutionError(assert, jsonrpc.Commands(c), requestBody, http.StatusBadRequest)
}

func testWorkerShouldReturn400ExecutionError(t *testing.T) {
	assert := assert.New(t)
	fn := func(ctx context.Context, _ []byte) error {
		return errors.New("fail")
	}
	c, err := command.NewCommands(command.CommandHandlerFunc("test", fn))

	if err != nil {
		assert.FailNow(err.Error())
	}

	w := command.NewWorker(context.TODO(), func(e command.Event) {}, c, 200)
	testShouldReturn400ExecutionError(assert, jsonrpc.CommandsWorker(w), notificationRequestBody, http.StatusNoContent)
}

func testShouldReturn200(assert *assert.Assertions, handler http.Handler, body string, code int) {
	ts := httptest.NewServer(handler)

	defer ts.Close()

	var b bytes.Buffer
	b.WriteString(body)

	id := uuid.New().String()
	correlationID := uuid.New().String()
	causationID := uuid.New().String()
	cl := http.Client{}

	u, err := url.Parse(ts.URL)
	if err != nil {
		assert.FailNow(err.Error())
	}

	req := &http.Request{
		Method: http.MethodPost,
		URL:    u,
		Body:   ioutil.NopCloser(&b),
		Header: http.Header{
			"Content-Type":   []string{"application/json"},
			"Request_id":     []string{id},
			"Correlation_id": []string{correlationID},
			"Causation_id":   []string{causationID}}}

	resp, err := cl.Do(req)
	if err != nil {
		assert.FailNow(err.Error())
	}

	assert.Equal(code, resp.StatusCode, fmt.Sprintf("should return %d", code))
	assert.NotEmpty(resp.Header["Request_id"], "request id should not be empty")
	assert.NotEqual([]string{id}, resp.Header["Request_id"], "request id should not equal incoming request id")
	assert.Equal([]string{correlationID}, resp.Header["Correlation_id"], "correlation id should equal incoming request correlation id")
	assert.Equal([]string{id}, resp.Header["Causation_id"], "causation id should equal incoming request id")

	if code == 200 {
		assert.Equal([]string{"application/json"}, resp.Header["Content-Type"], "Content-Type should be application/json")

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			assert.FailNow(err.Error())
		}

		response := new(jsonrpc.SuccessResponse)
		if err := json.Unmarshal(body, response); err != nil {
			assert.FailNowf("failed to read response", "%s: %s", err.Error(), string(body))
		}

		assert.Equal("success", response.Result["result"], `response.Result should contain "{"result": "sucess"}"`)
		assert.EqualValues(1, response.ID, "json rpc request id should equal 1")
		assert.Equal("2.0", response.Version, `json rpc request version should equal "2.0"`)
	} else {
		resp.Body.Close()
	}
}

func testQueriesShouldReturn200(t *testing.T) {
	assert := assert.New(t)
	fn := func(ctx context.Context, _ []byte) ([]byte, error) {
		return []byte(`{"result": "success"}`), nil
	}
	q, err := query.NewQueries(query.QueryHandlerFunc("test", fn))

	if err != nil {
		assert.FailNow(err.Error())
	}

	testShouldReturn200(assert, jsonrpc.Queries(q), requestBody, http.StatusOK)
}

func testCommandsShouldReturn200(t *testing.T) {
	assert := assert.New(t)
	fn := func(ctx context.Context, _ []byte) error {
		return nil
	}
	c, err := command.NewCommands(command.CommandHandlerFunc("test", fn))

	if err != nil {
		assert.FailNow(err.Error())
	}

	testShouldReturn200(assert, jsonrpc.Commands(c), notificationRequestBody, http.StatusNoContent)
}

func testWorkerShouldReturn200(t *testing.T) {
	assert := assert.New(t)
	fn := func(ctx context.Context, _ []byte) error {
		return nil
	}
	c, err := command.NewCommands(command.CommandHandlerFunc("test", fn))

	if err != nil {
		assert.FailNow(err.Error())
	}

	w := command.NewWorker(context.TODO(), func(e command.Event) {}, c, 200)
	testShouldReturn200(assert, jsonrpc.CommandsWorker(w), notificationRequestBody, http.StatusNoContent)
}

func testHandlerShouldHandleBatchRequests(t *testing.T) {
	assert := assert.New(t)
	cFn := func(ctx context.Context, _ []byte) error {
		return nil
	}
	c, err := command.NewCommands(command.CommandHandlerFunc("test_1", cFn))

	if err != nil {
		assert.FailNow(err.Error())
	}

	fn := func(ctx context.Context, _ []byte) ([]byte, error) {
		return []byte(`{"result": "success"}`), nil
	}

	q, err := query.NewQueries(query.QueryHandlerFunc("test", fn))
	if err != nil {
		assert.FailNow(err.Error())
	}

	ts := httptest.NewServer(&jsonrpc.Handler{Queries: q, Commands: c})

	defer ts.Close()

	var b bytes.Buffer
	b.WriteString(batchBody)

	id := uuid.New().String()
	correlationID := uuid.New().String()
	causationID := uuid.New().String()
	cl := http.Client{}

	u, err := url.Parse(ts.URL)
	if err != nil {
		assert.FailNow(err.Error())
	}

	req := &http.Request{
		Method: http.MethodPost,
		URL:    u,
		Body:   ioutil.NopCloser(&b),
		Header: http.Header{
			"Content-Type":   []string{"application/json"},
			"Request_id":     []string{id},
			"Correlation_id": []string{correlationID},
			"Causation_id":   []string{causationID}}}

	resp, err := cl.Do(req)
	if err != nil {
		assert.FailNow(err.Error())
	}

	assert.Equal(http.StatusOK, resp.StatusCode, "should return 200")
	assert.NotEmpty(resp.Header["Request_id"], "request id should not be empty")
	assert.NotEqual([]string{id}, resp.Header["Request_id"], "request id should not equal incoming request id")
	assert.Equal([]string{correlationID}, resp.Header["Correlation_id"], "correlation id should equal incoming request correlation id")
	assert.Equal([]string{id}, resp.Header["Causation_id"], "causation id should equal incoming request id")
	assert.Equal([]string{"application/json"}, resp.Header["Content-Type"], "Content-Type should be application/json")

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		assert.FailNow(err.Error())
	}

	var response []jsonrpc.SuccessResponse
	if err := json.Unmarshal(body, &response); err != nil {
		assert.FailNowf("failed to read response", "%s: %s", err.Error(), string(body))
	}

	assert.Equal("success", response[0].Result["result"], `response.Result should contain "{"result": "success"}"`)
	assert.EqualValues(1, response[0].ID, "json rpc request id should equal 1")
	assert.Equal("2.0", response[0].Version, `json rpc request version should equal "2.0"`)
}

func testShouldMarshalHandlerToJSON(t *testing.T) {
	assert := assert.New(t)
	cFn := func(ctx context.Context, _ []byte) error {
		return nil
	}
	c, err := command.NewCommands(command.CommandHandlerFunc("test_commands", cFn))

	if err != nil {
		assert.FailNow(err.Error())
	}

	c1, err := command.NewCommands(command.CommandHandlerFunc("test_worker", cFn))

	if err != nil {
		assert.FailNow(err.Error())
	}

	w := command.NewWorker(context.TODO(), func(e command.Event) {}, c1, 200)

	fn := func(ctx context.Context, _ []byte) ([]byte, error) {
		return []byte(`{"result": "success"}`), nil
	}

	q, err := query.NewQueries(query.QueryHandlerFunc("test_queries", fn))
	if err != nil {
		assert.FailNow(err.Error())
	}

	h := jsonrpc.Handler{Queries: q, Commands: c, Worker: w}

	b, err := json.Marshal(h)
	if err != nil {
		assert.FailNow(err.Error())
	}

	ad := new(testApiDefinition)
	if err := json.Unmarshal(b, ad); err != nil {
		assert.FailNow(err.Error())
	}

	assert.EqualValues([]string{"test_queries"}, ad.Queries, "should have correct query description")
	assert.EqualValues([]string{"test_commands"}, ad.Commands, "should have correct commands description")
	assert.EqualValues([]string{"test_worker"}, ad.Worker, "should have correct worker description")
}

type testApiDefinition struct {
	Queries  []string `json:"queries"`
	Commands []string `json:"commands"`
	Worker   []string `json:"worker"`
}
