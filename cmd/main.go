package main

import (
	"fmt"
	"log"

	"github.com/alitari/mockgo-server/internal/routing"
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
	config := Configuration{}
	if err := envconfig.Process("", &config); err != nil {
		log.Fatal(err)
	}
	logger = &utils.Logger{Verbose: config.Verbose}

	logger.LogAlways(banner + config.info())

	mockRouter, err := routing.NewMockRouter(config.MockDir, config.MockFilepattern, config.ResponseDir, config.ResponseFilepattern, logger)
	if err != nil {
		log.Fatalf("(FATAL) Can't load files: %v", err)
	}

	configRouter := routing.NewConfigRouter(mockRouter, logger)
	err = configRouter.SyncWithCluster(config.ClusterUrls)
	if err != nil {
		log.Fatalf("(FATAL) Can't sync with cluster: %v\n", err)
	}

	go configRouter.ListenAndServe(config.ConfigPort)

	mockRouter.ListenAndServe(config.MockPort)
}
