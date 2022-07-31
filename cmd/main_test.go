package main

import (
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/alitari/mockgo-server/internal/config"
	"github.com/alitari/mockgo-server/internal/mock"
	"github.com/alitari/mockgo-server/internal/utils"
)

var mockRouters []*mock.MockRouter
var configRouters []*config.ConfigRouter

var configRouterChan = make(chan *config.ConfigRouter)
var mockRouterChan = make(chan *mock.MockRouter)

func Test_main(t *testing.T) {
	startCluster(8080, 2)
	time.Sleep(5 * time.Second)
	stopAllRouters()
}

func startCluster(startPort, nodeCount int) {
	for i := 0; i < nodeCount; i++ {
		go startRouters(startPort + 1)
		mockRouters = append(mockRouters, <-mockRouterChan)
		configRouters = append(configRouters, <-configRouterChan)
	}
}

func startRouters(mockPort int) {
	configPort := mockPort + 1
	os.Setenv("VERBOSE", "true")
	os.Setenv("MOCK_PORT", strconv.Itoa(mockPort))
	os.Setenv("CONFIG_PORT", strconv.Itoa(configPort))
	os.Setenv("MOCK_DIR", "../test/main")
	os.Setenv("RESPONSE_DIR", "../test/main/responses")
	os.Setenv("CLUSTER_URLS", "")

	configuration := createConfiguration()
	logger := &utils.Logger{Verbose: configuration.Verbose}
	logger.LogAlways("Mockgo SERVER Nr. " + strconv.Itoa(len(mockRouters)+1) + "\n" + configuration.info())

	mockRouter := createMockRouter(configuration, logger)
	configRouter := createConfigRouter(configuration, mockRouter, logger)
	mockRouterChan <- mockRouter
	configRouterChan <- configRouter
	go startServe(configRouter)
	startServe(mockRouter)
}

func stopAllRouters() {
	for i, mockRouter := range mockRouters {
		stopServe(mockRouter)
		stopServe(configRouters[i])
	}
}
