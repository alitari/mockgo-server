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
	// set kvstore with a template func
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
	// get kvstore with a template func
	util.RequestCall(t, httpClient, http.MethodGet, urlPrefix+"/getkvstore/maintest", nil, "", 200, func(responseBody string) {
		expectedResponseBody := `{
    "message": "get kvstore successfully",
    "key": "maintest",
    "value": "{\"key\":\"value\"}"
}`
		assert.Equal(t, expectedResponseBody, responseBody)
	})
}

// 	// get kvstore with the config api
// 	requestToAllNodes(t, true, http.MethodGet, "/kvstore/maintest", map[string][]string{headers.Accept: {"application/json"}, headers.Authorization: {config.BasicAuth("mockgo", configPassword)}}, "", http.StatusOK, func(responseBody string) {
// 		assert.Equal(t, `{"mainTest1":"mainTest1Value"}`, responseBody)
// 	})
// 	// get kvstore with a template func
// 	requestToAllNodes(t, false, http.MethodGet, "/getkvstore/maintest", nil, "", 200, func(responseBody string) {
// 		expectedResponseBody := `{
//     "message": "get kvstore successfully",
//     "key": "maintest",
//     "value": "{\"mainTest1\":\"mainTest1Value\"}"
// }`
// 		assert.Equal(t, expectedResponseBody, responseBody)
// 	})

// func TestMain_getMatches(t *testing.T) {
// 	// get matches
// 	requestToNode(t, 0, true, http.MethodGet, "/matches", map[string][]string{headers.Accept: {"application/json"}, headers.Authorization: {config.BasicAuth("mockgo", configPassword)}}, "", http.StatusOK, func(responseBody string) {
// 		assert.Equal(t, "{}", responseBody)
// 	})
// 	// mock request
// 	requestToNode(t, 0, false, http.MethodGet, "/hello", map[string][]string{headers.Accept: {"application/json"}}, "", http.StatusOK, func(responseBody string) {
// 		assert.Equal(t, "{\n    \"hello\": \"from Mockgo!\"\n}", responseBody)
// 	})
// 	var assertMatchesResponsesFunc func(responseBody string)
// 	if os.Getenv("MATCHES_COUNT_ONLY") == "false" {
// 		assertMatchesResponsesFunc = func(responseBody string) {
// 			var matchData map[string][]*model.Match
// 			err := json.Unmarshal([]byte(responseBody), &matchData)
// 			assert.NoError(t, err)
// 			matches := matchData["helloId"]
// 			assert.NotNil(t, matches)
// 			assert.Len(t, matches, 1)
// 			match := matches[0]
// 			assert.Equal(t, "helloId", match.EndpointId)
// 			assert.Greater(t, time.Now(), match.Timestamp)
// 			assert.Equal(t, http.MethodGet, match.ActualRequest.Method)
// 			assert.Equal(t, "localhost:8081", match.ActualRequest.Host)
// 			assert.Equal(t, "/hello", match.ActualRequest.URL)
// 			assert.Equal(t, map[string][]string{"Accept": {"application/json"}, "Accept-Encoding": {"gzip"}, "User-Agent": {"Go-http-client/1.1"}}, match.ActualRequest.Header)
// 			assert.Equal(t, http.StatusOK, match.ActualResponse.StatusCode)
// 			assert.Empty(t, match.ActualResponse.Header)
// 		}
// 	} else {
// 		assertMatchesResponsesFunc = func(responseBody string) {
// 			var matchesCountData map[string]int64
// 			err := json.Unmarshal([]byte(responseBody), &matchesCountData)
// 			assert.NoError(t, err)
// 			assert.NotNil(t, matchesCountData)
// 			matchesCount := matchesCountData["helloId"]
// 			assert.Equal(t, int64(1), matchesCount)
// 		}
// 	}
// 	requestToAllNodes(t, true, http.MethodGet, "/matches", map[string][]string{headers.Accept: {"application/json"}, headers.Authorization: {config.BasicAuth("mockgo", configPassword)}}, "", http.StatusOK, assertMatchesResponsesFunc)

// 	requestToNode(t, 0, true, http.MethodDelete, "/matches", map[string][]string{headers.Authorization: {config.BasicAuth("mockgo", configPassword)}}, "", http.StatusOK, func(responseBody string) {
// 		assert.Empty(t, responseBody)
// 	})

// 	requestToAllNodes(t, true, http.MethodGet, "/matches", map[string][]string{headers.Accept: {"application/json"}, headers.Authorization: {config.BasicAuth("mockgo", configPassword)}}, "", http.StatusOK, func(responseBody string) {
// 		assert.Equal(t, "{}", responseBody)
// 	})

// }

// func TestMain_getMismatches(t *testing.T) {
// 	// get mismatches
// 	requestToNode(t, 0, true, http.MethodGet, "/mismatches", map[string][]string{headers.Accept: {"application/json"}, headers.Authorization: {config.BasicAuth("mockgo", configPassword)}}, "", http.StatusOK, func(responseBody string) {
// 		if os.Getenv("MISMATCHES_COUNT_ONLY") == "false" {
// 			assert.Equal(t, "[]", responseBody)
// 		} else {
// 			assert.Equal(t, "0", responseBody)
// 		}
// 	})
// 	// create mismatch request
// 	requestToNode(t, 0, false, http.MethodGet, "/nohello", nil, "", http.StatusNotFound, func(responseBody string) {
// 		assert.Equal(t, "404 page not found\n", responseBody)
// 	})
// 	var assertMismatchesResponsesFunc func(responseBody string)
// 	if os.Getenv("MISMATCHES_COUNT_ONLY") == "false" {
// 		assertMismatchesResponsesFunc = func(responseBody string) {
// 			var mismatchData []*model.Mismatch
// 			err := json.Unmarshal([]byte(responseBody), &mismatchData)
// 			assert.NoError(t, err)
// 			assert.NotNil(t, mismatchData)
// 			assert.Len(t, mismatchData, 1)
// 			assert.Greater(t, time.Now(), mismatchData[0].Timestamp)
// 			assert.Equal(t, "path '/nohello' not matched, subpath which matched: ''", mismatchData[0].MismatchDetails)

// 			actualRequest := mismatchData[0].ActualRequest
// 			assert.Equal(t, http.MethodGet, actualRequest.Method)
// 			assert.Equal(t, "localhost:8081", actualRequest.Host)
// 			assert.Equal(t, "/nohello", actualRequest.URL)
// 			assert.Equal(t, map[string][]string{"Accept-Encoding": {"gzip"}, "User-Agent": {"Go-http-client/1.1"}}, actualRequest.Header)
// 		}
// 	} else {
// 		assertMismatchesResponsesFunc = func(responseBody string) {
// 			var mismatchesCountData int64
// 			err := json.Unmarshal([]byte(responseBody), &mismatchesCountData)
// 			assert.NoError(t, err)
// 			assert.Equal(t, int64(1), mismatchesCountData)
// 		}
// 	}
// 	requestToAllNodes(t, true, http.MethodGet, "/mismatches", map[string][]string{headers.Accept: {"application/json"}, headers.Authorization: {config.BasicAuth("mockgo", configPassword)}}, "", http.StatusOK, assertMismatchesResponsesFunc)

// 	requestToNode(t, 0, true, http.MethodDelete, "/mismatches", map[string][]string{headers.Authorization: {config.BasicAuth("mockgo", configPassword)}}, "", http.StatusOK, func(responseBody string) {
// 		assert.Empty(t, responseBody)
// 	})

// 	requestToAllNodes(t, true, http.MethodGet, "/mismatches", map[string][]string{headers.Accept: {"application/json"}, headers.Authorization: {config.BasicAuth("mockgo", configPassword)}}, "", http.StatusOK, func(responseBody string) {
// 		if os.Getenv("MISMATCHES_COUNT_ONLY") == "false" {
// 			assert.Equal(t, "[]", responseBody)
// 		} else {
// 			assert.Equal(t, "0", responseBody)
// 		}
// 	})

// }

// func TestMain_transferMatches(t *testing.T) {
// 	header := map[string][]string{config.NoAdvertiseHeader: {"true"}, headers.Accept: {"application/json"}, headers.Authorization: {config.BasicAuth("mockgo", configPassword)}}
// 	assertMatchesCount := func(node, expectedCount int) {
// 		requestToNode(t, node, true, http.MethodGet, "/matches", header, "", http.StatusOK, func(responseBody string) {
// 			if os.Getenv("MATCHES_COUNT_ONLY") == "false" {
// 				var matchData map[string][]*model.Match
// 				err := json.Unmarshal([]byte(responseBody), &matchData)
// 				assert.NoError(t, err)
// 				assert.Equal(t, expectedCount, len(matchData["helloId"]))
// 			} else {
// 				var matchCountData map[string]int64
// 				err := json.Unmarshal([]byte(responseBody), &matchCountData)
// 				assert.NoError(t, err)
// 				assert.Equal(t, int64(expectedCount), matchCountData["helloId"])
// 			}
// 		})
// 	}

// 	assertMismatchesCount := func(node, expectedCount int) {
// 		requestToNode(t, node, true, http.MethodGet, "/mismatches", header, "", http.StatusOK, func(responseBody string) {
// 			if os.Getenv("MISMATCHES_COUNT_ONLY") == "false" {
// 				var matchData []*model.Mismatch
// 				err := json.Unmarshal([]byte(responseBody), &matchData)
// 				assert.NoError(t, err)
// 				assert.Equal(t, expectedCount, len(matchData))
// 			} else {
// 				var matchCountData int64
// 				err := json.Unmarshal([]byte(responseBody), &matchCountData)
// 				assert.NoError(t, err)
// 				assert.Equal(t, int64(expectedCount), matchCountData)
// 			}
// 		})
// 	}

// 	createMatchRequests := func(node, count int) {
// 		for i := 0; i < count; i++ {
// 			requestToNode(t, node, false, http.MethodGet, "/hello", map[string][]string{headers.Accept: {"application/json"}}, "", http.StatusOK, func(responseBody string) {
// 				assert.Equal(t, "{\n    \"hello\": \"from Mockgo!\"\n}", responseBody)
// 			})
// 		}
// 	}

// 	createMismatchRequests := func(node, count int) {
// 		for i := 0; i < count; i++ {
// 			requestToNode(t, node, false, http.MethodGet, "/nohello", nil, "", http.StatusNotFound, func(responseBody string) {
// 				assert.Equal(t, "404 page not found\n", responseBody)
// 			})
// 		}
// 	}

// 	assertMatchesCount(0, 0)    // 0 matches in node 0
// 	assertMatchesCount(1, 0)    // 0 matches in node 1
// 	assertMismatchesCount(0, 0) // 0 mismatches in node 0
// 	assertMismatchesCount(1, 0) // 0 mismatches in node 1

// 	createMismatchRequests(0, 3) // create 3 mismatching requests to node 0
// 	createMismatchRequests(1, 1) // create 1 mismatching requests to node 1
// 	createMatchRequests(0, 5)    // create 5 matching requests to node 0
// 	createMatchRequests(1, 2)    // create 2 matching requests to node 1

// 	assertMismatchesCount(0, 3) // 3 mismatches in node 0
// 	assertMismatchesCount(1, 1) // 1 mismatch in node 1
// 	assertMatchesCount(0, 5)    // 5 matches in node 0
// 	assertMatchesCount(1, 2)    // 2 matches in node 1

// 	//transfer matches and mismatches node 0 -> node 1
// 	requestToNode(t, 0, true, http.MethodGet, "/transfermatches", nil, "", http.StatusOK, func(responseBody string) {
// 		assert.Equal(t, "", responseBody)
// 	})

// 	assertMatchesCount(0, 0)    // 0 matches in node 0
// 	assertMismatchesCount(0, 0) // 0 mismatches in node 0
// 	assertMatchesCount(1, 7)    // 7 matches in node 1
// 	assertMismatchesCount(1, 4) // 4 mismatches in node 1
// }

// 	// add kvstore with a template func
// 	requestToNode(t, 0, false, http.MethodPost, "/addkvstore/maintest", nil, `{ "path": "/mainTest2", "value": "mainTest2Value" }`, 200, func(responseBody string) {
// 		expectedResponseBody := `{
//     "message": "add kvstore successfully",
//     "key": "maintest"
//     "body": "{ \"path\": \"/mainTest2\", \"value\": \"mainTest2Value\" }"
//     "path": "/mainTest2"
//     "value": "mainTest2Value"
// }`
// 		assert.Equal(t, expectedResponseBody, responseBody)
// 	})

// 	// get kvstore with the config api
// 	requestToAllNodes(t, true, http.MethodGet, "/kvstore/maintest", map[string][]string{headers.Accept: {"application/json"}, headers.Authorization: {config.BasicAuth("mockgo", configPassword)}}, "", http.StatusOK, func(responseBody string) {
// 		assert.Equal(t, `{"mainTest1":"mainTest1Value","mainTest2":"mainTest2Value"}`, responseBody)
// 	})

// 	// remove kvstore with a template func
// 	requestToNode(t, 0, false, http.MethodPost, "/removekvstore/maintest", nil, `{ "path": "/mainTest1" }`, 200, func(responseBody string) {
// 		expectedResponseBody := `{
//     "message": "remove kvstore successfully",
//     "key": "maintest"
//     "body": "{ \"path\": \"/mainTest1\" }"
//     "path": "/mainTest1"
// }`
// 		assert.Equal(t, expectedResponseBody, responseBody)
// 	})

// 	// get kvstore with the config api
// 	requestToAllNodes(t, true, http.MethodGet, "/kvstore/maintest", map[string][]string{headers.Accept: {"application/json"}, headers.Authorization: {config.BasicAuth("mockgo", configPassword)}}, "", http.StatusOK, func(responseBody string) {
// 		assert.Equal(t, `{"mainTest1":"","mainTest2":"mainTest2Value"}`, responseBody)
// 	})

// 	// lookup kvstore with jsonPath
// 	requestToNode(t, 0, false, http.MethodPost, "/lookupkvstore/maintest", nil, `{ "jsonPath": "$.mainTest2" }`, 200, func(responseBody string) {
// 		expectedResponseBody := `{
//     "message": "lookup kvstore successfully",
//     "key": "maintest",
//     "body": "{ \"jsonPath\": \"$.mainTest2\" }",
//     "jsonPath": "$.mainTest2",
//     "value": "mainTest2Value"
// }`
// 		assert.Equal(t, expectedResponseBody, responseBody)
// 	})

// }

// func requestToAllNodes(t *testing.T, config bool, method, path string, header map[string][]string, body string, expectedStatus int, assertBody func(responseBody string)) {
// 	for i := 0; i < len(mockRouters); i++ {
// 		requestToNode(t, i, config, method, path, header, body, expectedStatus, assertBody)
// 	}
// }
