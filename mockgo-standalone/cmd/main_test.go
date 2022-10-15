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

func TestMain_templateFunctionsPutKVStore(t *testing.T) {
	util.RequestCall(t, httpClient, http.MethodPut, urlPrefix+"/setkvstore/maintest", nil, `{ "mainTest1": "mainTest1Value" }`, 200, func(responseBody string) {
		expectedResponseBody := `{
    "message": "set kvstore successfully",
    "key": "maintest",
    "value": "{ \"mainTest1\": \"mainTest1Value\" }"
}`
		assert.Equal(t, expectedResponseBody, responseBody)
	})
	util.RequestCall(t, httpClient, http.MethodGet, urlPrefix+"/__/kvstore/maintest", map[string][]string{headers.Accept: {"application/json"}, headers.Authorization: {util.BasicAuth("mockgo", apiPassword)}},
		"", http.StatusOK, func(responseBody string) {
			assert.Equal(t, `{ "mainTest1": "mainTest1Value" }`, responseBody)
		})
}

func TestMain_templateFunctionsGetKVStore(t *testing.T) {
	util.RequestCall(t, httpClient, http.MethodPut, urlPrefix+"/__/kvstore/maintest", map[string][]string{headers.ContentType: {"application/json"}, headers.Authorization: {util.BasicAuth("mockgo", apiPassword)}},
		`{ "key": "value" }`, http.StatusNoContent, func(responseBody string) {
		})
	util.RequestCall(t, httpClient, http.MethodGet, urlPrefix+"/getkvstore/maintest", nil, "", 200, func(responseBody string) {
		expectedResponseBody := `{
    "message": "get kvstore successfully",
    "key": "maintest",
    "value": "{\"key\":\"value\"}"
}`
		assert.Equal(t, expectedResponseBody, responseBody)
	})
}

func TestMain_templateFunctionsAddKVStore(t *testing.T) {
	util.RequestCall(t, httpClient, http.MethodPut, urlPrefix+"/__/kvstore/maintest", map[string][]string{headers.ContentType: {"application/json"}, headers.Authorization: {util.BasicAuth("mockgo", apiPassword)}},
		`{ "key": "value" }`, http.StatusNoContent, func(responseBody string) {
		})
	util.RequestCall(t, httpClient, http.MethodPost, urlPrefix+"/addkvstore/maintest", nil,
		`{ "path": "/key2", "value": "value2" }`, 200, func(responseBody string) {
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
		"", http.StatusOK, func(responseBody string) {
			assert.Equal(t, `{"key":"value","key2":"value2"}`, responseBody)
		})
}

func TestMain_templateFunctionsRemoveKVStore(t *testing.T) {
	util.RequestCall(t, httpClient, http.MethodPut, urlPrefix+"/__/kvstore/maintest", map[string][]string{headers.ContentType: {"application/json"}, headers.Authorization: {util.BasicAuth("mockgo", apiPassword)}},
		`{ "key": "value" }`, http.StatusNoContent, func(responseBody string) {
		})
	util.RequestCall(t, httpClient, http.MethodDelete, urlPrefix+"/removekvstore/maintest", nil,
		`{ "path": "/key" }`, 200, func(responseBody string) {
			expectedResponseBody := `{
    "message": "remove kvstore successfully",
    "key": "maintest",
    "body": "{ \"path\": \"/key\" }",
    "path": "/key"
}`
			assert.Equal(t, expectedResponseBody, responseBody)
		})
	util.RequestCall(t, httpClient, http.MethodGet, urlPrefix+"/__/kvstore/maintest", map[string][]string{headers.Accept: {"application/json"}, headers.Authorization: {util.BasicAuth("mockgo", apiPassword)}},
		"", http.StatusOK, func(responseBody string) {
			assert.Equal(t, `{}`, responseBody)
		})
}

func TestMain_templateFunctionsLookupKVStore(t *testing.T) {
	util.RequestCall(t, httpClient, http.MethodPut, urlPrefix+"/__/kvstore/maintest", map[string][]string{headers.ContentType: {"application/json"}, headers.Authorization: {util.BasicAuth("mockgo", apiPassword)}},
		`{ "key": "value" }`, http.StatusNoContent, func(responseBody string) {
		})
	util.RequestCall(t, httpClient, http.MethodGet, urlPrefix+"/lookupkvstore/maintest", nil,
		`{ "jsonPath": "$.key" }`, 200, func(responseBody string) {
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
