package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/alitari/mockgo-server/internal/config"
	"github.com/alitari/mockgo-server/internal/kvstore"
	"github.com/alitari/mockgo-server/internal/mock"
	"github.com/alitari/mockgo-server/internal/model"
	"github.com/alitari/mockgo-server/internal/utils"
	"github.com/kelseyhightower/envconfig"
)

const banner = `
 __  __         _            
|  \/  |___  __| |_____ ___ 
| |\/| / _ \/ _| / / _  / _ \
|_|  |_\___/\__|_\_\__, \___/
Configuration:     |___/     
==============
`

var logger *utils.Logger

type Configuration struct {
	Verbose             bool     `default:"true"`
	ConfigPort          int      `default:"8081" split_words:"true"`
	MockPort            int      `default:"8080" split_words:"true"`
	MockDir             string   `default:"." split_words:"true"`
	MockFilepattern     string   `default:"*-mock.*" split_words:"true"`
	ResponseDir         string   `default:"./responses" split_words:"true"`
	ResponseFilepattern string   `default:"*.*" split_words:"true"`
	ClusterUrls         []string `default:"" split_words:"true"`
}

func (c Configuration) info() string {
	return fmt.Sprintf(`Verbose: %v
Config Port: %v
Mock Port: %v
Mock Dir: '%s'
Mock Filepattern: '%s'
Response Dir: '%s'
Response Filepattern: '%s'
Cluster URLs: '%v'`, c.Verbose, c.ConfigPort, c.MockPort, c.MockDir, c.MockFilepattern, c.ResponseDir, c.ResponseFilepattern, c.ClusterUrls)
}

func main() {
	configuration := createConfiguration()
	logger = &utils.Logger{Verbose: configuration.Verbose}
	logger.LogAlways(banner + configuration.info())
	kvstore.CreateTheStore()

	mockRouter := createMockRouter(configuration, kvstore.TheKVStore, logger)
	configRouter := createConfigRouter(configuration, mockRouter, kvstore.TheKVStore, logger)

	go startServe(configRouter)
	startServe(mockRouter)
}

func createConfiguration() *Configuration {
	configuration := Configuration{}
	if err := envconfig.Process("", &configuration); err != nil {
		log.Fatal(err)
	}
	return &configuration
}

func createMockRouter(configuration *Configuration, kvstore *kvstore.KVStore, logger *utils.Logger) *mock.MockRouter {
	mockRouter, err := mock.NewMockRouter(configuration.MockDir, configuration.MockFilepattern, configuration.ResponseDir, configuration.ResponseFilepattern, configuration.MockPort, kvstore, logger)
	if err != nil {
		log.Fatalf("(FATAL) Can't load files: %v", err)
	}
	return mockRouter
}

func createConfigRouter(configuration *Configuration, mockRouter *mock.MockRouter, kvStore *kvstore.KVStore, logger *utils.Logger) *config.ConfigRouter {
	configRouter := config.NewConfigRouter(mockRouter, configuration.ConfigPort, configuration.ClusterUrls, kvStore, logger)
	err := configRouter.SyncWithCluster()
	if err != nil {
		log.Fatalf("(FATAL) Can't sync with cluster: %v\n", err)
	}
	return configRouter
}

func startServe(serving model.Serving) error {
	logger.LogAlways(fmt.Sprintf("Serving on port %v", serving.Port()))
	s := serving.Server()
	err := s.ListenAndServe()
	if err != nil {
		return err
	}
	return nil
}

func stopServe(serving model.Serving) {
	logger.LogAlways(fmt.Sprintf("Stop Serving on port %d", serving.Port()))
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := serving.Server().Shutdown(ctx); err != nil {
		log.Fatalf("Can't stop server %v", err)
	}
}
