package main

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/alitari/mockgo-server/internal/config"
	"github.com/alitari/mockgo-server/internal/kvstore"
	"github.com/alitari/mockgo-server/internal/mock"
	"github.com/alitari/mockgo-server/internal/model"
	"github.com/alitari/mockgo-server/internal/utils"
	"github.com/go-http-utils/headers"
	"github.com/stretchr/testify/assert"
)

var httpClient = http.Client{Timeout: time.Duration(1) * time.Second}
var clusterSize = 2
var startPort = 8080

var mockRouters []*mock.MockRouter
var configRouters []*config.ConfigRouter

var configRouterChan = make(chan *config.ConfigRouter)
var mockRouterChan = make(chan *mock.MockRouter)

func TestMain(m *testing.M) {
	startCluster()
	code := utils.RunAndCheckCoverage("main", m, 0.65)
	stopCluster()
	os.Exit(code)
}

func TestMain_health(t *testing.T) {
	requestToAllNodes(t, true, http.MethodGet, "/health", "", "", "", http.StatusOK, nil)
}

func TestMain_serverid(t *testing.T) {
	requestToAllNodes(t, true, http.MethodGet, "/id", "", "application/text", "", http.StatusOK, nil)
}

func TestMain_endpoints(t *testing.T) {
	requestToAllNodes(t, true, http.MethodGet, "/endpoints", "", "application/json", "", http.StatusOK, func(responseBody string) {
		assert.Equal(t, 1186, len(responseBody))
	})
}

func TestMain_setgetKvStore(t *testing.T) {
	requestToNode(t, 0, true, http.MethodPut, "/kvstore/store1", "application/json", "", `{ "mykey1": "myvalue1" }`, http.StatusNoContent, nil)
	requestToAllNodes(t, true, http.MethodGet, "/kvstore/store1", "", "application/json", "", http.StatusOK, func(responseBody string) {
		assert.Equal(t, "{\"mykey1\":\"myvalue1\"}", responseBody)
	})
}

func TestMain_getMatches(t *testing.T) {
	// get matches
	requestToNode(t, 0, true, http.MethodGet, "/matches", "", "application/json", "", http.StatusOK, func(responseBody string) {
		assert.Equal(t, "{}", responseBody)
	})
	// mock request
	requestToNode(t, 0, false, http.MethodGet, "/hello", "", "application/json", "", http.StatusOK, func(responseBody string) {
		assert.Equal(t, "{\n    \"hello\": \"from Mockgo!\"\n}", responseBody)
	})

	assertMatchesResponsesFunc := func(responseBody string) {
		var matchData map[string][]*model.Match
		err := json.Unmarshal([]byte(responseBody), &matchData)
		assert.NoError(t, err)
		matches := matchData["helloId"]
		assert.NotNil(t, matches)
		assert.Len(t, matches, 1)
		match := matches[0]
		assert.Equal(t, "helloId", match.EndpointId)
		assert.Greater(t, time.Now(), match.Timestamp)
		assert.Equal(t, http.MethodGet, match.ActualRequest.Method)
		assert.Equal(t, "localhost:8081", match.ActualRequest.Host)
		assert.Equal(t, "/hello", match.ActualRequest.URL)
		assert.Equal(t, map[string][]string{"Accept": {"application/json"}, "Accept-Encoding": {"gzip"}, "User-Agent": {"Go-http-client/1.1"}}, match.ActualRequest.Header)
		assert.Equal(t, http.StatusOK, match.ActualResponse.StatusCode)
		assert.Equal(t, map[string]string{}, match.ActualResponse.Header)
	}
	requestToAllNodes(t, true, http.MethodGet, "/matches", "", "application/json", "", http.StatusOK, assertMatchesResponsesFunc)

	requestToNode(t, 0, true, http.MethodDelete, "/matches", "", "", "", http.StatusOK, func(responseBody string) {
		assert.Empty(t, responseBody)
	})

	requestToAllNodes(t, true, http.MethodGet, "/matches", "", "application/json", "", http.StatusOK, func(responseBody string) {
		assert.Equal(t, "{}", responseBody)
	})

}

func TestMain_uploadKvStore(t *testing.T) {
	// delete kv store for all
	requestToAllNodes(t, true, http.MethodPut, "/kvstore", "application/json", "", "{}", http.StatusNoContent, nil)

	kvstore := `{"store1":{"mykey":"myvalue"}}`
	requestToNode(t, 0, true, http.MethodPut, "/kvstore", "application/json", "", kvstore, http.StatusNoContent, nil)
	requestToNode(t, 0, true, http.MethodGet, "/kvstore", "", "application/json", "", http.StatusOK, func(responseBody string) {
		assert.Equal(t, kvstore, responseBody)
	})
	// upload kvstore not advertised
	requestToNode(t, 1, true, http.MethodGet, "/kvstore", "", "application/json", "", http.StatusOK, func(responseBody string) {
		assert.Equal(t, "{}", responseBody)
	})
	stopNode(1)
	startNode(1)
	time.Sleep(100 * time.Millisecond)
	// kvstore synced from node 0 at boot time
	requestToNode(t, 1, true, http.MethodGet, "/kvstore", "", "application/json", "", http.StatusOK, func(responseBody string) {
		assert.Equal(t, kvstore, responseBody)
	})
}

// func TestMain_matchstatistics(t *testing.T) {
// 	// TODO clean match statistics
// 	requestToNode(t, 0, false, http.MethodGet, "/matchstatistics", "","application/json", "", http.StatusOK, func(responseBody string) {
// 		assert.Equal(t, "", responseBody)
// 	})
// }

func requestToAllNodes(t *testing.T, config bool, method, path, contentType, acceptHeader, body string, expectedStatus int, assertBody func(responseBody string)) {
	for i := 0; i < len(mockRouters); i++ {
		requestToNode(t, i, config, method, path, contentType, acceptHeader, body, expectedStatus, assertBody)
	}
}

func requestToNode(t *testing.T, nodeNr int, config bool, method, path, contentType, acceptHeader, body string, expectedStatus int, assertBody func(responseBody string)) {
	port := startPort + nodeNr*2
	if !config {
		port++
	}
	url := "http://localhost:" + strconv.Itoa(port) + path
	t.Logf("calling url: '%s' ...", url)

	var bodyReader io.Reader
	if len(body) > 0 {
		bodyReader = bytes.NewBufferString(body)
	}
	request, err := http.NewRequest(method, url, bodyReader)
	assert.NoError(t, err)
	if len(contentType) > 0 {
		request.Header.Add(headers.ContentType, contentType)
	}
	if len(acceptHeader) > 0 {
		request.Header.Add(headers.Accept, acceptHeader)
	}
	response, err := httpClient.Do(request)
	assert.NoError(t, err)
	defer response.Body.Close()
	t.Logf("response status: '%s'", response.Status)
	assert.Equal(t, expectedStatus, response.StatusCode)
	respBytes, err := ioutil.ReadAll(response.Body)
	respBody := string(respBytes)
	assert.NoError(t, err)
	t.Logf("response body:\n '%s'", respBody)
	if assertBody != nil {
		assertBody(respBody)
	}
}

func getClusterUrls() string {
	var clusterUrls []string
	for i := 0; i < clusterSize; i++ {
		clusterUrls = append(clusterUrls, "http://localhost:"+strconv.Itoa(startPort+(i*2)))
	}
	return strings.Join(clusterUrls, ",")
}

func startCluster() {
	for i := 0; i < clusterSize; i++ {
		startNode(i)
	}
}

func stopCluster() {
	for i := 0; i < len(mockRouters); i++ {
		stopServe(mockRouters[i])
		stopServe(configRouters[i])
	}
}

func stopNode(nodeNr int) {
	stopServe(mockRouters[nodeNr])
	stopServe(configRouters[nodeNr])
	mockRouters = append(mockRouters[:nodeNr], mockRouters[nodeNr+1:]...)
	configRouters = append(configRouters[:nodeNr], configRouters[nodeNr+1:]...)
}

func startNode(nodeNr int) {
	go serveNode(nodeNr)
	mockRouters = append(mockRouters, <-mockRouterChan)
	configRouters = append(configRouters, <-configRouterChan)
}

func serveNode(nodeNr int) {
	configPort := startPort + (nodeNr * 2)
	mockPort := configPort + 1
	os.Setenv("VERBOSE", "true")
	os.Setenv("MOCK_PORT", strconv.Itoa(mockPort))
	os.Setenv("CONFIG_PORT", strconv.Itoa(configPort))
	os.Setenv("MOCK_DIR", "../test/main")
	os.Setenv("RESPONSE_DIR", "../test/main/responses")
	os.Setenv("CLUSTER_URLS", getClusterUrls())

	configuration := createConfiguration()
	logger := &utils.Logger{Verbose: configuration.Verbose}
	mockRouter, configRouter := createRouters(kvstore.CreateTheStore(), logger)
	mockRouterChan <- mockRouter
	configRouterChan <- configRouter
	go startServe(configRouter)
	startServe(mockRouter)
}

// func stopNode() {
