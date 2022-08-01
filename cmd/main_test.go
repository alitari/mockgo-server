package main

import (
	"bytes"
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
	code := m.Run()
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
		assert.Equal(t, 582, len(responseBody))
	})
}

func TestMain_setgetKvStore(t *testing.T) {
	requestToNode(t, 0, true, http.MethodPut, "/kvstore/store1", "application/json", "", `{ "mykey1": "myvalue1" }`, http.StatusNoContent, nil)
	requestToAllNodes(t, true, http.MethodGet, "/kvstore/store1", "", "application/json", "", http.StatusOK, func(responseBody string) {
		assert.Equal(t, "{\n    \"mykey1\": \"myvalue1\"\n}", responseBody)
	})
}

func TestMain_uploadKvStore(t *testing.T) {
	// delete kv store for all
	requestToAllNodes(t, true, http.MethodPut, "/kvstore", "application/json", "", "{}", http.StatusNoContent, nil)

	kvstore := `{"store1":{"mykey":"myvalue"}}`
	requestToNode(t, 0, true, http.MethodPut, "/kvstore", "application/json", "", kvstore, http.StatusNoContent, nil)
	requestToNode(t, 0,true, http.MethodGet, "/kvstore", "", "application/json", "", http.StatusOK, func(responseBody string) {
		assert.Equal(t, kvstore, responseBody)
	})
	requestToNode(t, 1,true, http.MethodGet, "/kvstore", "", "application/json", "", http.StatusOK, func(responseBody string) {
		assert.Equal(t, "{}", responseBody)
	})

	
}

func requestToAllNodes(t *testing.T, config bool, method, path, contentType, acceptHeader, body string, expectedStatus int, assertBody func(responseBody string)) {
	for i := 0; i < len(mockRouters); i++ {
		requestToNode(t, i, config, method, path, contentType, acceptHeader, body, expectedStatus, assertBody)
	}
}

func requestToNode(t *testing.T, nodeNr int, config bool, method, path, contentType, acceptHeader, body string, expectedStatus int, assertBody func(responseBody string)) {
	port := startPort + nodeNr*2
	if config {
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
		clusterUrls = append(clusterUrls, "http://localhost:"+strconv.Itoa(startPort+(i*2)+1))
	}
	return strings.Join(clusterUrls, ",")
}

func startCluster() {
	for i := 0; i < clusterSize; i++ {
		go startNode(i)
		mockRouters = append(mockRouters, <-mockRouterChan)
		configRouters = append(configRouters, <-configRouterChan)
	}
}

func startNode(nodeNr int) {
	mockPort := startPort+(nodeNr*2)
	configPort := mockPort + 1
	os.Setenv("VERBOSE", "true")
	os.Setenv("MOCK_PORT", strconv.Itoa(mockPort))
	os.Setenv("CONFIG_PORT", strconv.Itoa(configPort))
	os.Setenv("MOCK_DIR", "../test/main")
	os.Setenv("RESPONSE_DIR", "../test/main/responses")
	os.Setenv("CLUSTER_URLS", getClusterUrls())

	configuration := createConfiguration()
	logger := &utils.Logger{Verbose: configuration.Verbose}
	logger.LogAlways("Mockgo SERVER Nr. " + strconv.Itoa(len(mockRouters)+1) + "\n" + configuration.info())

	kvstore := kvstore.NewStore()
	mockRouter := createMockRouter(configuration, kvstore, logger)
	configRouter := createConfigRouter(configuration, mockRouter, kvstore, logger)
	mockRouterChan <- mockRouter
	configRouterChan <- configRouter
	go startServe(configRouter)
	startServe(mockRouter)
}

// func stopNode() {

func stopCluster() {
	for i, mockRouter := range mockRouters {
		stopServe(mockRouter)
		stopServe(configRouters[i])
	}
}
