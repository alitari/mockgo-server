package main

import (
	"log"
	"net/http"
	"os"
	"testing"

	"github.com/alitari/mockgo-server/mockgo/testutil"
	"github.com/stretchr/testify/assert"
)

var apiPassword = testutil.RandString(10)

func TestMain(m *testing.M) {
	os.Setenv("LOGLEVEL_API", "2")
	os.Setenv("LOGLEVEL_MOCK", "2")
	os.Setenv("API_PASSWORD", apiPassword)
	os.Setenv("MOCK_DIR", "../../test/main")
	os.Setenv("MATCHES_RECORD_REQUESTS", "true")
	os.Setenv("MISMATCHES_RECORD_REQUESTS", "true")
	router, _, err := setupRouter()
	if err != nil {
		log.Fatalf("can't setup router : %v", err)
	}
	testutil.StartServing(router)
	code := testutil.RunAndCheckCoverage("main_test", m, 0.60)
	testutil.StopServing()
	os.Exit(code)
}

func TestMain_health(t *testing.T) {
	assert.NoError(t, testutil.AssertResponseStatusOfRequestCall(t,
		testutil.CreateOutgoingRequest(t, http.MethodGet, "/__/health", testutil.CreateHeader(), ""), http.StatusOK))
}

func TestMain_metrics(t *testing.T) {
	matchRequest := testutil.CreateOutgoingRequest(t, http.MethodGet, "/hello", testutil.CreateHeader(), "")
	assertFunc := func(response *http.Response, responseBody string) {
		assert.Equal(t, http.StatusOK, response.StatusCode)
		assert.Equal(t, "{\n    \"hello\": \"from Mockgo!\"\n}", responseBody)
	}
	assert.NoError(t, testutil.AssertResponseOfRequestCall(t, matchRequest, assertFunc))
	assert.NoError(t, testutil.AssertResponseOfRequestCall(t, matchRequest, assertFunc))

	mismatchRequest := testutil.CreateOutgoingRequest(t, http.MethodGet, "/mismatch", testutil.CreateHeader(), "")
	assert.NoError(t, testutil.AssertResponseStatusOfRequestCall(t, mismatchRequest, http.StatusNotFound))

	metricsRequest := testutil.CreateOutgoingRequest(t, http.MethodGet, "/__/metrics", testutil.CreateHeader(), "")
	assert.NoError(t, testutil.AssertResponseOfRequestCall(t, metricsRequest, func(response *http.Response, responseBody string) {
		assert.Equal(t, http.StatusOK, response.StatusCode)
		assert.Contains(t, responseBody, "# HELP matches Number of matches of an endpoint")
		assert.Contains(t, responseBody, "# TYPE matches counter")
		assert.Contains(t, responseBody, "matches{endpoint=\"helloId\"}")
		assert.Contains(t, responseBody, "# HELP mismatches Number of mismatches")
		assert.Contains(t, responseBody, "# TYPE mismatches counter")
		assert.Contains(t, responseBody, "mismatches")
	}))
}

func TestMain_templateFunctionsPutKVStore(t *testing.T) {
	putKVStoreRequest := testutil.CreateOutgoingRequest(t, http.MethodPut, "/setkvstore/maintest", testutil.CreateHeader(), `{ "mainTest1": "mainTest1Value" }`)
	assert.NoError(t, testutil.AssertResponseOfRequestCall(t, putKVStoreRequest, func(response *http.Response, responseBody string) {
		assert.Equal(t, http.StatusOK, response.StatusCode)
		expectedResponseBody := `{
    "message": "set kvstore successfully",
    "key": "maintest",
    "value": "{ \"mainTest1\": \"mainTest1Value\" }"
}`
		assert.Equal(t, expectedResponseBody, responseBody)
	}))

	getKVStoreRequest := testutil.CreateOutgoingRequest(t, http.MethodGet, "/__/kvstore/maintest", testutil.CreateHeader().WithJSONAccept().WithAuth("mockgo", apiPassword), "")
	assert.NoError(t, testutil.AssertResponseOfRequestCall(t, getKVStoreRequest, func(response *http.Response, responseBody string) {
		assert.Equal(t, http.StatusOK, response.StatusCode)
		assert.Equal(t, `{ "mainTest1": "mainTest1Value" }`, responseBody)
	}))

}

func TestMain_templateFunctionsGetKVStore(t *testing.T) {
	putKVStoreRequest := testutil.CreateOutgoingRequest(t, http.MethodPut, "/__/kvstore/maintest", testutil.CreateHeader().WithJSONContentType().WithAuth("mockgo", apiPassword), `{ "key": "value" }`)
	assert.NoError(t, testutil.AssertResponseStatusOfRequestCall(t, putKVStoreRequest, http.StatusNoContent))

	getKVStoreRequest := testutil.CreateOutgoingRequest(t, http.MethodGet, "/getkvstore/maintest", testutil.CreateHeader(), "")
	assert.NoError(t, testutil.AssertResponseOfRequestCall(t, getKVStoreRequest, func(response *http.Response, responseBody string) {
		assert.Equal(t, http.StatusOK, response.StatusCode)
		expectedResponseBody := `{
    "message": "get kvstore successfully",
    "key": "maintest",
    "value": "{\"key\":\"value\"}"
}`
		assert.Equal(t, expectedResponseBody, responseBody)
	}))
}

func TestMain_templateFunctionsGetKVStoreInline(t *testing.T) {
	putKVStoreRequest := testutil.CreateOutgoingRequest(t, http.MethodPut, "/__/kvstore/maintest", testutil.CreateHeader().WithJSONContentType().WithAuth("mockgo", apiPassword), `{ "key": "valueInline" }`)
	assert.NoError(t, testutil.AssertResponseStatusOfRequestCall(t, putKVStoreRequest, http.StatusNoContent))

	getKVStoreRequest := testutil.CreateOutgoingRequest(t, http.MethodGet, "/getkvstoreInline/maintest", testutil.CreateHeader(), "")
	assert.NoError(t, testutil.AssertResponseOfRequestCall(t, getKVStoreRequest, func(response *http.Response, responseBody string) {
		assert.Equal(t, http.StatusOK, response.StatusCode)
		expectedResponseBody := `{
    "message": "get kvstore successfully",
    "key": "maintest",
    "value": "{\"key\":\"valueInline\"}"
}`
		assert.Equal(t, expectedResponseBody, responseBody)
	}))

}

func TestMain_templateFunctionsAddKVStore(t *testing.T) {
	putKVStoreRequest := testutil.CreateOutgoingRequest(t, http.MethodPut, "/__/kvstore/maintest", testutil.CreateHeader().WithJSONContentType().WithAuth("mockgo", apiPassword), `{ "key": "value" }`)
	assert.NoError(t, testutil.AssertResponseStatusOfRequestCall(t, putKVStoreRequest, http.StatusNoContent))

	addKVStoreRequest := testutil.CreateOutgoingRequest(t, http.MethodPost, "/addkvstore/maintest", testutil.CreateHeader(), `{ "path": "/key2", "value": "value2" }`)
	assert.NoError(t, testutil.AssertResponseOfRequestCall(t, addKVStoreRequest, func(response *http.Response, responseBody string) {
		assert.Equal(t, http.StatusOK, response.StatusCode)
		expectedResponseBody := `{
    "message": "add kvstore successfully",
    "key": "maintest",
    "body": "{ \"path\": \"/key2\", \"value\": \"value2\" }",
    "path": "/key2",
    "value": "value2"
}`
		assert.Equal(t, expectedResponseBody, responseBody)
	}))

	getKVStoreRequest := testutil.CreateOutgoingRequest(t, http.MethodGet, "/__/kvstore/maintest", testutil.CreateHeader().WithJSONAccept().WithAuth("mockgo", apiPassword), "")
	assert.NoError(t, testutil.AssertResponseOfRequestCall(t, getKVStoreRequest, func(response *http.Response, responseBody string) {
		assert.Equal(t, http.StatusOK, response.StatusCode)
		assert.Equal(t, `{"key":"value","key2":"value2"}`, responseBody)
	}))
}

func TestMain_templateFunctionsRemoveKVStore(t *testing.T) {
	putKVStoreRequest := testutil.CreateOutgoingRequest(t, http.MethodPut, "/__/kvstore/maintest", testutil.CreateHeader().WithJSONContentType().WithAuth("mockgo", apiPassword), `{ "key": "value" }`)
	assert.NoError(t, testutil.AssertResponseStatusOfRequestCall(t, putKVStoreRequest, http.StatusNoContent))

	removeKVStoreRequest := testutil.CreateOutgoingRequest(t, http.MethodDelete, "/removekvstore/maintest", testutil.CreateHeader(), `{ "path": "/key" }`)
	assert.NoError(t, testutil.AssertResponseOfRequestCall(t, removeKVStoreRequest, func(response *http.Response, responseBody string) {
		assert.Equal(t, http.StatusOK, response.StatusCode)
		expectedResponseBody := `{
    "message": "remove kvstore successfully",
    "key": "maintest",
    "body": "{ \"path\": \"/key\" }",
    "path": "/key"
}`
		assert.Equal(t, expectedResponseBody, responseBody)
	}))
	getKVStoreRequest := testutil.CreateOutgoingRequest(t, http.MethodGet, "/__/kvstore/maintest", testutil.CreateHeader().WithJSONAccept().WithAuth("mockgo", apiPassword), "")
	assert.NoError(t, testutil.AssertResponseOfRequestCall(t, getKVStoreRequest, func(response *http.Response, responseBody string) {
		assert.Equal(t, http.StatusOK, response.StatusCode)
		assert.Equal(t, `{}`, responseBody)
	}))
}

func TestMain_templateFunctionsLookupKVStore(t *testing.T) {
	putKVStoreRequest := testutil.CreateOutgoingRequest(t, http.MethodPut, "/__/kvstore/maintest", testutil.CreateHeader().WithJSONContentType().WithAuth("mockgo", apiPassword), `{ "key": "value" }`)
	assert.NoError(t, testutil.AssertResponseStatusOfRequestCall(t, putKVStoreRequest, http.StatusNoContent))

	lookupKVStoreRequest := testutil.CreateOutgoingRequest(t, http.MethodGet, "/lookupkvstore/maintest", testutil.CreateHeader(), `{ "jsonPath": "$.key" }`)
	assert.NoError(t, testutil.AssertResponseOfRequestCall(t, lookupKVStoreRequest, func(response *http.Response, responseBody string) {
		assert.Equal(t, http.StatusOK, response.StatusCode)
		expectedResponseBody := `{
    "message": "lookup kvstore successfully",
    "key": "maintest",
    "body": "{ \"jsonPath\": \"$.key\" }",
    "jsonPath": "$.key",
    "value": "value"
}`
		assert.Equal(t, expectedResponseBody, responseBody)
	}))

}

func TestMain_templateFunctionsResponseCode(t *testing.T) {
	getStatusCodeRequest := testutil.CreateOutgoingRequest(t, http.MethodGet, "/statusCode/201", testutil.CreateHeader(), "Alex")
	assert.NoError(t, testutil.AssertResponseOfRequestCall(t, getStatusCodeRequest, func(response *http.Response, responseBody string) {
		assert.Equal(t, http.StatusCreated, response.StatusCode)
		assert.Equal(t, "Alex", responseBody)
		assert.Equal(t, []string{"/statusCode/201"}, response.Header["Header1"])
	}))
}

func TestMain_templateFunctionsQueryParams(t *testing.T) {
	getQueryParamsRequest := testutil.CreateOutgoingRequest(t, http.MethodGet, "/queryParams?foo=bar&foo2=bar2", testutil.CreateHeader(), "")
	assert.NoError(t, testutil.AssertResponseOfRequestCall(t, getQueryParamsRequest, func(response *http.Response, responseBody string) {
		assert.Equal(t, http.StatusOK, response.StatusCode)
		expectedResponseBody := `{
    "foo" : "bar",
    "foo2" : "bar2",
}`
		assert.Equal(t, expectedResponseBody, responseBody)
	}))
}

func TestMain_templateFunctionsPeople(t *testing.T) {
	setupStoreRequest := testutil.CreateOutgoingRequest(t, http.MethodPut, "/__/kvstore/people", testutil.CreateHeader().WithJSONContentType().WithAuth("mockgo", apiPassword), `{ "adults": [], "childs": [] }`)
	assert.NoError(t, testutil.AssertResponseStatusOfRequestCall(t, setupStoreRequest, http.StatusNoContent))
	storeWrongAlexRequest := testutil.CreateOutgoingRequest(t, http.MethodPut, "/storePeople", testutil.CreateHeader(), `{ "wrongkey": "Alex", "age": 55 }`)
	assert.NoError(t, testutil.AssertResponseStatusOfRequestCall(t, storeWrongAlexRequest, http.StatusBadRequest))
	storeAlexRequest := testutil.CreateOutgoingRequest(t, http.MethodPut, "/storePeople", testutil.CreateHeader(), `{ "name": "Alex", "age": 55 }`)
	assert.NoError(t, testutil.AssertResponseStatusOfRequestCall(t, storeAlexRequest, http.StatusOK))
	storeDaniRequest := testutil.CreateOutgoingRequest(t, http.MethodPut, "/storePeople", testutil.CreateHeader(), `{ "name": "Dani", "age": 45 }`)
	assert.NoError(t, testutil.AssertResponseStatusOfRequestCall(t, storeDaniRequest, http.StatusOK))
	storeKlaraRequest := testutil.CreateOutgoingRequest(t, http.MethodPut, "/storePeople", testutil.CreateHeader(), `{ "name": "Klara", "age": 16 }`)
	assert.NoError(t, testutil.AssertResponseStatusOfRequestCall(t, storeKlaraRequest, http.StatusOK))
	getChildsRequest := testutil.CreateOutgoingRequest(t, http.MethodGet, "/getPeople/childs", testutil.CreateHeader(), "")
	assert.NoError(t, testutil.AssertResponseOfRequestCall(t, getChildsRequest, func(response *http.Response, responseBody string) {
		assert.Equal(t, http.StatusOK, response.StatusCode)
		assert.Equal(t, "[\n  {\n    \"age\": 16,\n    \"name\": \"Klara\"\n  }\n]", responseBody)
	}))
	getAdultsRequest := testutil.CreateOutgoingRequest(t, http.MethodGet, "/getPeople/adults", testutil.CreateHeader(), "")
	assert.NoError(t, testutil.AssertResponseOfRequestCall(t, getAdultsRequest, func(response *http.Response, responseBody string) {
		assert.Equal(t, http.StatusOK, response.StatusCode)
		assert.Equal(t, "[\n  {\n    \"age\": 55,\n    \"name\": \"Alex\"\n  },\n  {\n    \"age\": 45,\n    \"name\": \"Dani\"\n  }\n]", responseBody)
	}))
}
