package main

import (
	"log"
	"net/http"
	"testing"

	"github.com/gorilla/mux"
	"github.com/kelseyhightower/envconfig"
	"go.uber.org/zap"

	"github.com/alitari/mockgo-server/mockgo/starter"
	"github.com/alitari/mockgo-server/mockgo/testutil"
	"github.com/stretchr/testify/assert"
)

var apiPassword = testutil.RandString(10)

func setupMain(t *testing.T) {
	env := map[string]string{
		"LOGLEVEL_API":  "DEBUG",
		"LOGLEVEL_MOCK": "DEBUG",
		"MOCK_DIR":      "../../../test/main",
		"API_PASSWORD":  apiPassword,
	}
	for key, value := range env {
		t.Setenv(key, value)
	}
	starter.BasicConfig = &starter.BasicConfiguration{}
	if err := envconfig.Process("", starter.BasicConfig); err != nil {
		log.Fatal("can't create configuration", zap.Error(err))
	}

	serve = func(router *mux.Router) {
		testutil.StartServing(router)
	}
	main()
}

func TestMain_health(t *testing.T) {
	setupMain(t)
	assert.NoError(t, testutil.AssertResponseStatusOfRequestCall(t,
		testutil.CreateOutgoingRequest(t, http.MethodGet, "/__/health", testutil.CreateHeader(), ""), http.StatusOK))
	testutil.StopServing()
}

func TestMain_basicAuth(t *testing.T) {
	setupMain(t)
	assert.NoError(t, testutil.AssertResponseStatusOfRequestCall(t,
		testutil.CreateOutgoingRequest(t, http.MethodGet, "/__/kvstore/some", testutil.CreateHeader().WithJSONAccept(), ""), http.StatusUnauthorized))
	assert.NoError(t, testutil.AssertResponseStatusOfRequestCall(t,
		testutil.CreateOutgoingRequest(t, http.MethodGet, "/__/matches/helloId", testutil.CreateHeader().WithJSONAccept(), ""), http.StatusUnauthorized))
	testutil.StopServing()
}

func TestMain_metrics(t *testing.T) {
	setupMain(t)
	matchRequest := testutil.CreateOutgoingRequest(t, http.MethodGet, "/hello", testutil.CreateHeader().WithAuth("mockgo", apiPassword), "")
	assertFunc := func(response *http.Response, responseBody string) {
		assert.Equal(t, http.StatusOK, response.StatusCode)
		assert.Equal(t, "{\n    \"hello\": \"from Mockgo!\"\n}", responseBody)
	}
	assert.NoError(t, testutil.AssertResponseOfRequestCall(t, matchRequest, assertFunc))
	assert.NoError(t, testutil.AssertResponseOfRequestCall(t, matchRequest, assertFunc))

	mismatchRequest := testutil.CreateOutgoingRequest(t, http.MethodGet, "/mismatch", testutil.CreateHeader().WithAuth("mockgo", apiPassword), "")
	assert.NoError(t, testutil.AssertResponseStatusOfRequestCall(t, mismatchRequest, http.StatusNotFound))

	metricsRequest := testutil.CreateOutgoingRequest(t, http.MethodGet, "/__/metrics", testutil.CreateHeader().WithAuth("mockgo", apiPassword), "")
	assert.NoError(t, testutil.AssertResponseOfRequestCall(t, metricsRequest, func(response *http.Response, responseBody string) {
		assert.Equal(t, http.StatusOK, response.StatusCode)
		assert.Contains(t, responseBody, "# HELP matches Number of matches of an endpoint")
		assert.Contains(t, responseBody, "# TYPE matches counter")
		assert.Contains(t, responseBody, "matches{endpoint=\"helloId\"}")
		assert.Contains(t, responseBody, "# HELP mismatches Number of mismatches")
		assert.Contains(t, responseBody, "# TYPE mismatches counter")
		assert.Contains(t, responseBody, "mismatches")
	}))
	testutil.StopServing()
}

func TestMain_templateFunctionsPutKVStore(t *testing.T) {
	setupMain(t)
	putKVStoreRequest := testutil.CreateOutgoingRequest(t, http.MethodPut, "/setkvstore/mainstore/mykey", testutil.CreateHeader(), `{ "mainTest1": "mainTest1Value" }`)
	assert.NoError(t, testutil.AssertResponseOfRequestCall(t, putKVStoreRequest, func(response *http.Response, responseBody string) {
		assert.Equal(t, http.StatusOK, response.StatusCode)
		expectedResponseBody := `{
    "message": "set kvstore successfully",
    "store": "mainstore",
    "key": "mykey",
    "value": "{ \"mainTest1\": \"mainTest1Value\" }"
}`
		assert.Equal(t, expectedResponseBody, responseBody)
	}))

	getKVStoreRequest := testutil.CreateOutgoingRequest(t, http.MethodGet, "/__/kvstore/mainstore/mykey", testutil.CreateHeader().WithJSONAccept().WithAuth("mockgo", apiPassword), "")
	assert.NoError(t, testutil.AssertResponseOfRequestCall(t, getKVStoreRequest, func(response *http.Response, responseBody string) {
		assert.Equal(t, http.StatusOK, response.StatusCode)
		assert.Equal(t, `{ "mainTest1": "mainTest1Value" }`, responseBody)
	}))
	testutil.StopServing()

}

func TestMain_templateFunctionsGetKVStore(t *testing.T) {
	setupMain(t)
	putKVStoreRequest := testutil.CreateOutgoingRequest(t, http.MethodPut, "/__/kvstore/mainstore/mykey", testutil.CreateHeader().WithJSONContentType().WithAuth("mockgo", apiPassword), `{ "key": "value" }`)
	assert.NoError(t, testutil.AssertResponseStatusOfRequestCall(t, putKVStoreRequest, http.StatusNoContent))

	getKVStoreRequest := testutil.CreateOutgoingRequest(t, http.MethodGet, "/getkvstore/mainstore/mykey", testutil.CreateHeader(), "")
	assert.NoError(t, testutil.AssertResponseOfRequestCall(t, getKVStoreRequest, func(response *http.Response, responseBody string) {
		assert.Equal(t, http.StatusOK, response.StatusCode)
		expectedResponseBody := `{
    "message": "get kvstore successfully",
    "store": "mainstore",
    "key": "mykey",
    "value": "{\"key\":\"value\"}"
}`
		assert.Equal(t, expectedResponseBody, responseBody)
	}))
	testutil.StopServing()
}

func TestMain_templateFunctionsGetKVStoreInline(t *testing.T) {
	setupMain(t)
	putKVStoreRequest := testutil.CreateOutgoingRequest(t, http.MethodPut, "/__/kvstore/mainstoreInline/mykeyInline", testutil.CreateHeader().WithJSONContentType().WithAuth("mockgo", apiPassword), `{ "key": "valueInline" }`)
	assert.NoError(t, testutil.AssertResponseStatusOfRequestCall(t, putKVStoreRequest, http.StatusNoContent))

	getKVStoreRequest := testutil.CreateOutgoingRequest(t, http.MethodGet, "/getkvstoreInline/mainstoreInline/mykeyInline", testutil.CreateHeader(), "")
	assert.NoError(t, testutil.AssertResponseOfRequestCall(t, getKVStoreRequest, func(response *http.Response, responseBody string) {
		assert.Equal(t, http.StatusOK, response.StatusCode)
		expectedResponseBody := `{
    "message": "get kvstore successfully",
    "store": "mainstoreInline",
    "key": "mykeyInline",
    "value": "{\"key\":\"valueInline\"}"
}`
		assert.Equal(t, expectedResponseBody, responseBody)
	}))
	testutil.StopServing()

}

func TestMain_templateFunctionsRemoveKVStore(t *testing.T) {
	setupMain(t)
	putKVStoreRequest := testutil.CreateOutgoingRequest(t, http.MethodPut, "/__/kvstore/mainstore/anotherkey", testutil.CreateHeader().WithJSONContentType().WithAuth("mockgo", apiPassword), `{ "key": "value" }`)
	assert.NoError(t, testutil.AssertResponseStatusOfRequestCall(t, putKVStoreRequest, http.StatusNoContent))

	removeKVStoreRequest := testutil.CreateOutgoingRequest(t, http.MethodDelete, "/removekvstore/mainstore/anotherkey", testutil.CreateHeader(), "")
	assert.NoError(t, testutil.AssertResponseOfRequestCall(t, removeKVStoreRequest, func(response *http.Response, responseBody string) {
		assert.Equal(t, http.StatusOK, response.StatusCode)
		expectedResponseBody := `{
    "message": "remove kvstore successfully",
    "store": "mainstore",
    "key": "anotherkey",
}`
		assert.Equal(t, expectedResponseBody, responseBody)
	}))
	getKVStoreRequest := testutil.CreateOutgoingRequest(t, http.MethodGet, "/__/kvstore/mainstore/anotherkey", testutil.CreateHeader().WithJSONAccept().WithAuth("mockgo", apiPassword), "")
	assert.NoError(t, testutil.AssertResponseOfRequestCall(t, getKVStoreRequest, func(response *http.Response, responseBody string) {
		assert.Equal(t, http.StatusNotFound, response.StatusCode)
	}))
	testutil.StopServing()
}

func TestMain_templateFunctionsResponseCode(t *testing.T) {
	setupMain(t)
	getStatusCodeRequest := testutil.CreateOutgoingRequest(t, http.MethodGet, "/statusCode/201", testutil.CreateHeader(), "Alex")
	assert.NoError(t, testutil.AssertResponseOfRequestCall(t, getStatusCodeRequest, func(response *http.Response, responseBody string) {
		assert.Equal(t, http.StatusCreated, response.StatusCode)
		assert.Equal(t, "Alex", responseBody)
		assert.Equal(t, []string{"/statusCode/201"}, response.Header["Header1"])
	}))
	testutil.StopServing()
}

func TestMain_templateFunctionsQueryParams(t *testing.T) {
	setupMain(t)
	getQueryParamsRequest := testutil.CreateOutgoingRequest(t, http.MethodGet, "/queryParams?foo=bar&foo2=bar2", testutil.CreateHeader(), "")
	assert.NoError(t, testutil.AssertResponseOfRequestCall(t, getQueryParamsRequest, func(response *http.Response, responseBody string) {
		assert.Equal(t, http.StatusOK, response.StatusCode)
		expectedResponseBody := `{
    "foo" : "bar",
    "foo2" : "bar2",
}`
		assert.Equal(t, expectedResponseBody, responseBody)
	}))
	testutil.StopServing()
}

func TestMain_templateFunctionsPeople(t *testing.T) {
	setupMain(t)
	storeWrongAlexRequest := testutil.CreateOutgoingRequest(t, http.MethodPut, "/storePeople", testutil.CreateHeader(), `{ "wrongkey": "Alex", "age": 55 }`)
	assert.NoError(t, testutil.AssertResponseStatusOfRequestCall(t, storeWrongAlexRequest, http.StatusBadRequest))
	storeAlexRequest := testutil.CreateOutgoingRequest(t, http.MethodPut, "/storePeople", testutil.CreateHeader(), `{ "name": "Alex", "age": 55 }`)
	assert.NoError(t, testutil.AssertResponseStatusOfRequestCall(t, storeAlexRequest, http.StatusOK))
	storeDaniRequest := testutil.CreateOutgoingRequest(t, http.MethodPut, "/storePeople", testutil.CreateHeader(), `{ "name": "Dani", "age": 45 }`)
	assert.NoError(t, testutil.AssertResponseStatusOfRequestCall(t, storeDaniRequest, http.StatusOK))
	storeKlaraRequest := testutil.CreateOutgoingRequest(t, http.MethodPut, "/storePeople", testutil.CreateHeader(), `{ "name": "Klara", "age": 16 }`)
	assert.NoError(t, testutil.AssertResponseStatusOfRequestCall(t, storeKlaraRequest, http.StatusOK))
	getChildsRequest := testutil.CreateOutgoingRequest(t, http.MethodGet, "/getPeople/children", testutil.CreateHeader(), "")
	assert.NoError(t, testutil.AssertResponseOfRequestCall(t, getChildsRequest, func(response *http.Response, responseBody string) {
		assert.Equal(t, http.StatusOK, response.StatusCode)
		assert.Equal(t, `{
  "Klara": {
    "age": 16,
    "name": "Klara"
  }
}`, responseBody)
	}))
	getAdultsRequest := testutil.CreateOutgoingRequest(t, http.MethodGet, "/getPeople/adults", testutil.CreateHeader(), "")
	assert.NoError(t, testutil.AssertResponseOfRequestCall(t, getAdultsRequest, func(response *http.Response, responseBody string) {
		assert.Equal(t, http.StatusOK, response.StatusCode)
		assert.Equal(t, `{
  "Alex": {
    "age": 55,
    "name": "Alex"
  },
  "Dani": {
    "age": 45,
    "name": "Dani"
  }
}`, responseBody)
	}))
	testutil.StopServing()
}
