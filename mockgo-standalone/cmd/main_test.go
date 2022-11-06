package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/alitari/mockgo/util"
	"github.com/go-http-utils/headers"
	"github.com/stretchr/testify/assert"
)

const ()

var httpClient = http.Client{Timeout: time.Duration(1) * time.Second}
var port = 8081
var urlPrefix = fmt.Sprintf("http://localhost:%d", port)
var apiPassword = util.RandString(10)

func TestMain(m *testing.M) {
	os.Setenv("LOGLEVEL_API", "2")
	os.Setenv("LOGLEVEL_MOCK", "2")
	os.Setenv("MOCK_PORT", strconv.Itoa(port))
	os.Setenv("API_PASSWORD", apiPassword)
	os.Setenv("MOCK_DIR", "../../test/main")
	os.Setenv("MATCHES_RECORD_REQUESTS", "true")
	os.Setenv("MISMATCHES_RECORD_REQUESTS", "true")
	go main()
	time.Sleep(200 * time.Millisecond)
	code := util.RunAndCheckCoverage("main", m, 0.65)
	os.Exit(code)
}

func TestMain_health(t *testing.T) {
	util.RequestCall(t, httpClient, http.MethodGet, urlPrefix+"/__/health", map[string][]string{}, "", http.StatusOK, nil)
}

func TestMain_metrics(t *testing.T) {
	// 2 matches  1 mismatch
	util.RequestCall(t, httpClient, http.MethodGet, urlPrefix+"/hello", map[string][]string{}, "", http.StatusOK, func(responseBody string, header map[string][]string) {
		assert.Equal(t, "{\n    \"hello\": \"from Mockgo!\"\n}", responseBody)
	})
	util.RequestCall(t, httpClient, http.MethodGet, urlPrefix+"/hello", map[string][]string{}, "", http.StatusOK, func(responseBody string, header map[string][]string) {
		assert.Equal(t, "{\n    \"hello\": \"from Mockgo!\"\n}", responseBody)
	})

	util.RequestCall(t, httpClient, http.MethodGet, urlPrefix+"/mismtach", map[string][]string{}, "", http.StatusNotFound, nil)

	util.RequestCall(t, httpClient, http.MethodGet, urlPrefix+"/__/metrics", map[string][]string{}, "", http.StatusOK, func(responseBody string, header map[string][]string) {
		assert.Contains(t, responseBody, "# HELP matches Number of matches of an endpoint")
		assert.Contains(t, responseBody, "# TYPE matches counter")
		assert.Contains(t, responseBody, "matches{endpoint=\"helloId\"} 2")
		assert.Contains(t, responseBody, "# HELP mismatches Number of mismatches")
		assert.Contains(t, responseBody, "# TYPE mismatches counter")
		assert.Contains(t, responseBody, "mismatches 1")
	})
}

func TestMain_templateFunctionsPutKVStore(t *testing.T) {
	util.RequestCall(t, httpClient, http.MethodPut, urlPrefix+"/setkvstore/maintest", nil, `{ "mainTest1": "mainTest1Value" }`, 200, func(responseBody string, header map[string][]string) {
		expectedResponseBody := `{
    "message": "set kvstore successfully",
    "key": "maintest",
    "value": "{ \"mainTest1\": \"mainTest1Value\" }"
}`
		assert.Equal(t, expectedResponseBody, responseBody)
	})
	util.RequestCall(t, httpClient, http.MethodGet, urlPrefix+"/__/kvstore/maintest", map[string][]string{headers.Accept: {"application/json"}, headers.Authorization: {util.BasicAuth("mockgo", apiPassword)}},
		"", http.StatusOK, func(responseBody string, header map[string][]string) {
			assert.Equal(t, `{ "mainTest1": "mainTest1Value" }`, responseBody)
		})
}

func TestMain_templateFunctionsGetKVStore(t *testing.T) {
	util.RequestCall(t, httpClient, http.MethodPut, urlPrefix+"/__/kvstore/maintest", map[string][]string{headers.ContentType: {"application/json"}, headers.Authorization: {util.BasicAuth("mockgo", apiPassword)}},
		`{ "key": "value" }`, http.StatusNoContent, func(responseBody string, header map[string][]string) {
		})
	util.RequestCall(t, httpClient, http.MethodGet, urlPrefix+"/getkvstore/maintest", nil, "", 200, func(responseBody string, header map[string][]string) {
		expectedResponseBody := `{
    "message": "get kvstore successfully",
    "key": "maintest",
    "value": "{\"key\":\"value\"}"
}`
		assert.Equal(t, expectedResponseBody, responseBody)
	})
}

func TestMain_templateFunctionsGetKVStoreInline(t *testing.T) {
	util.RequestCall(t, httpClient, http.MethodPut, urlPrefix+"/__/kvstore/maintest", map[string][]string{headers.ContentType: {"application/json"}, headers.Authorization: {util.BasicAuth("mockgo", apiPassword)}},
		`{ "key": "valueInline" }`, http.StatusNoContent, func(responseBody string, header map[string][]string) {
		})
	util.RequestCall(t, httpClient, http.MethodGet, urlPrefix+"/getkvstoreInline/maintest", nil, "", 200, func(responseBody string, header map[string][]string) {
		expectedResponseBody := `{
    "message": "get kvstore successfully",
    "key": "maintest",
    "value": "{\"key\":\"valueInline\"}"
}`
		assert.Equal(t, expectedResponseBody, responseBody)
	})
}

func TestMain_templateFunctionsAddKVStore(t *testing.T) {
	util.RequestCall(t, httpClient, http.MethodPut, urlPrefix+"/__/kvstore/maintest", map[string][]string{headers.ContentType: {"application/json"}, headers.Authorization: {util.BasicAuth("mockgo", apiPassword)}},
		`{ "key": "value" }`, http.StatusNoContent, func(responseBody string, header map[string][]string) {
		})
	util.RequestCall(t, httpClient, http.MethodPost, urlPrefix+"/addkvstore/maintest", nil,
		`{ "path": "/key2", "value": "value2" }`, 200, func(responseBody string, header map[string][]string) {
			expectedResponseBody := `{
    "message": "add kvstore successfully",
    "key": "maintest",
    "body": "{ \"path\": \"/key2\", \"value\": \"value2\" }",
    "path": "/key2",
    "value": "value2"
}`
			assert.Equal(t, expectedResponseBody, responseBody)
		})
	util.RequestCall(t, httpClient, http.MethodGet, urlPrefix+"/__/kvstore/maintest", map[string][]string{headers.Accept: {"application/json"}, headers.Authorization: {util.BasicAuth("mockgo", apiPassword)}},
		"", http.StatusOK, func(responseBody string, header map[string][]string) {
			assert.Equal(t, `{"key":"value","key2":"value2"}`, responseBody)
		})
}

func TestMain_templateFunctionsRemoveKVStore(t *testing.T) {
	util.RequestCall(t, httpClient, http.MethodPut, urlPrefix+"/__/kvstore/maintest", map[string][]string{headers.ContentType: {"application/json"}, headers.Authorization: {util.BasicAuth("mockgo", apiPassword)}},
		`{ "key": "value" }`, http.StatusNoContent, func(responseBody string, header map[string][]string) {
		})
	util.RequestCall(t, httpClient, http.MethodDelete, urlPrefix+"/removekvstore/maintest", nil,
		`{ "path": "/key" }`, 200, func(responseBody string, header map[string][]string) {
			expectedResponseBody := `{
    "message": "remove kvstore successfully",
    "key": "maintest",
    "body": "{ \"path\": \"/key\" }",
    "path": "/key"
}`
			assert.Equal(t, expectedResponseBody, responseBody)
		})
	util.RequestCall(t, httpClient, http.MethodGet, urlPrefix+"/__/kvstore/maintest", map[string][]string{headers.Accept: {"application/json"}, headers.Authorization: {util.BasicAuth("mockgo", apiPassword)}},
		"", http.StatusOK, func(responseBody string, header map[string][]string) {
			assert.Equal(t, `{}`, responseBody)
		})
}

func TestMain_templateFunctionsLookupKVStore(t *testing.T) {
	util.RequestCall(t, httpClient, http.MethodPut, urlPrefix+"/__/kvstore/maintest", map[string][]string{headers.ContentType: {"application/json"}, headers.Authorization: {util.BasicAuth("mockgo", apiPassword)}},
		`{ "key": "value" }`, http.StatusNoContent, func(responseBody string, header map[string][]string) {
		})
	util.RequestCall(t, httpClient, http.MethodGet, urlPrefix+"/lookupkvstore/maintest", nil,
		`{ "jsonPath": "$.key" }`, 200, func(responseBody string, header map[string][]string) {
			expectedResponseBody := `{
    "message": "lookup kvstore successfully",
    "key": "maintest",
    "body": "{ \"jsonPath\": \"$.key\" }",
    "jsonPath": "$.key",
    "value": "value"
}`
			assert.Equal(t, expectedResponseBody, responseBody)
		})
}

func TestMain_templateFunctionsResponseCode(t *testing.T) {
	util.RequestCall(t, httpClient, http.MethodGet, urlPrefix+"/statusCode/201", nil, "Alex", 201, func(responseBody string, header map[string][]string) {
		expectedResponseBody := `Alex`
		assert.Equal(t, expectedResponseBody, responseBody)
		assert.Equal(t, []string{"/statusCode/201"}, header["Header1"])
	})
}

func TestMain_templateFunctionsPeople(t *testing.T) {
	util.RequestCall(t, httpClient, http.MethodPut, urlPrefix+"/__/kvstore/people", map[string][]string{headers.ContentType: {"application/json"}, headers.Authorization: {util.BasicAuth("mockgo", apiPassword)}},
		`{ "adults": [], "childs": [] }`, http.StatusNoContent, func(responseBody string, header map[string][]string) {
		})

	util.RequestCall(t, httpClient, http.MethodPut, urlPrefix+"/storePeople", nil, `{ "wrongkey": "Alex", "age": 55 }`, 400, func(responseBody string, header map[string][]string) {
	})
	util.RequestCall(t, httpClient, http.MethodPut, urlPrefix+"/storePeople", nil, `{ "name": "Alex", "age": 55 }`, 200, func(responseBody string, header map[string][]string) {
	})
	util.RequestCall(t, httpClient, http.MethodPut, urlPrefix+"/storePeople", nil, `{ "name": "Dani", "age": 45 }`, 200, func(responseBody string, header map[string][]string) {
	})
	util.RequestCall(t, httpClient, http.MethodPut, urlPrefix+"/storePeople", nil, `{ "name": "Klara", "age": 16 }`, 200, func(responseBody string, header map[string][]string) {
	})

	util.RequestCall(t, httpClient, http.MethodGet, urlPrefix+"/getPeople/childs", nil, "", 200,
		func(responseBody string, header map[string][]string) {
			assert.Equal(t, "[\n  {\n    \"age\": 16,\n    \"name\": \"Klara\"\n  }\n]", responseBody)
		})

	util.RequestCall(t, httpClient, http.MethodGet, urlPrefix+"/getPeople/adults", nil, "", 200,
		func(responseBody string, header map[string][]string) {
			assert.Equal(t, "[\n  {\n    \"age\": 55,\n    \"name\": \"Alex\"\n  },\n  {\n    \"age\": 45,\n    \"name\": \"Dani\"\n  }\n]", responseBody)
		})
}
