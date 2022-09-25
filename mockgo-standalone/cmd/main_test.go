package main

import (
	"bytes"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var httpClient = http.Client{Timeout: time.Duration(1) * time.Second}
var port = 8081
var apiPassword = randString(10)

func TestMain(m *testing.M) {
	os.Setenv("LOGLEVEL_API", "2")
	os.Setenv("LOGLEVEL_MOCK", "2")
	os.Setenv("MOCK_PORT", strconv.Itoa(port))
	os.Setenv("API_PASSWORD", apiPassword)
	os.Setenv("MOCK_DIR", "../../test/main")
	os.Setenv("RESPONSE_DIR", "../../test/main/responses")
	os.Setenv("MATCHES_RECORD_REQUESTS", "true")
	os.Setenv("MISMATCHES_RECORD_REQUESTS", "true")
	go main()
	time.Sleep(200* time.Millisecond)
	code := runAndCheckCoverage("main", m, 0.65)
	os.Exit(code)
}


func request(t *testing.T, method, path string, header map[string][]string, body string, expectedStatus int, assertBody func(responseBody string)) {
	url := "http://localhost:" + strconv.Itoa(port) + path
	t.Logf("calling url: '%s' ...", url)

	var bodyReader io.Reader
	if len(body) > 0 {
		bodyReader = bytes.NewBufferString(body)
	}
	request, err := http.NewRequest(method, url, bodyReader)
	assert.NoError(t, err)
	request.Header = header
	response, err := httpClient.Do(request)
	assert.NoError(t, err)
	defer response.Body.Close()
	t.Logf("response status: '%s'", response.Status)
	assert.Equal(t, expectedStatus, response.StatusCode)
	respBytes, err := io.ReadAll(response.Body)
	respBody := string(respBytes)
	assert.NoError(t, err)
	t.Logf("response body:\n '%s'", respBody)
	if assertBody != nil {
		assertBody(respBody)
	}
}

func TestMain_health(t *testing.T) {
	request(t, http.MethodGet, "/health", map[string][]string{}, "", http.StatusOK, nil)
}


// func TestMain_uploadKvStore(t *testing.T) {
// 	// delete kv store for all
// 	requestToAllNodes(t, true, http.MethodPut, "/kvstore", map[string][]string{headers.ContentType: {"application/json"}, headers.Authorization: {config.BasicAuth("mockgo", configPassword)}}, "{}", http.StatusNoContent, nil)

// 	kvstore := `{"store1":{"mykey":"myvalue"}}`
// 	requestToNode(t, 0, true, http.MethodPut, "/kvstore", map[string][]string{headers.ContentType: {"application/json"}, headers.Authorization: {config.BasicAuth("mockgo", configPassword)}}, kvstore, http.StatusNoContent, nil)
// 	requestToNode(t, 0, true, http.MethodGet, "/kvstore", map[string][]string{headers.Accept: {"application/json"}, headers.Authorization: {config.BasicAuth("mockgo", configPassword)}}, "", http.StatusOK, func(responseBody string) {
// 		assert.Equal(t, kvstore, responseBody)
// 	})
// 	// upload kvstore not advertised
// 	requestToNode(t, 1, true, http.MethodGet, "/kvstore", map[string][]string{headers.Accept: {"application/json"}, headers.Authorization: {config.BasicAuth("mockgo", configPassword)}}, "", http.StatusOK, func(responseBody string) {
// 		assert.Equal(t, "{}", responseBody)
// 	})
// 	stopNode(1)
// 	startNode(1)
// 	time.Sleep(100 * time.Millisecond)
// 	// kvstore synced from node 0 at boot time
// 	requestToNode(t, 1, true, http.MethodGet, "/kvstore", map[string][]string{headers.Accept: {"application/json"}, headers.Authorization: {config.BasicAuth("mockgo", configPassword)}}, "", http.StatusOK, func(responseBody string) {
// 		assert.Equal(t, kvstore, responseBody)
// 	})
// }
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

// func TestMain_templateFunctionsKVStore(t *testing.T) {
// 	// set kvstore with a template func
// 	requestToNode(t, 0, false, http.MethodPut, "/setkvstore/maintest", nil, `{ "mainTest1": "mainTest1Value" }`, 200, func(responseBody string) {
// 		expectedResponseBody := `{
//     "message": "set kvstore successfully",
//     "key": "maintest",
//     "value": "{ \"mainTest1\": \"mainTest1Value\" }"
// }`
// 		assert.Equal(t, expectedResponseBody, responseBody)
// 	})

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

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func randString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Int63()%int64(len(letterBytes))]
	}
	return string(b)
}

func runAndCheckCoverage(testPackage string, m *testing.M, treshold float64) int {

	code := m.Run()

	if code == 0 && testing.CoverMode() != "" {
		coverage := testing.Coverage()
		if coverage < treshold {
			log.Printf("%s tests passed, but coverage must be above %2.2f%%, but it is %2.2f%%\n", testPackage, treshold*100, coverage*100)
			code = -1
		}
	}
	return code
}
