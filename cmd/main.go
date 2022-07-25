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
	Verbose             bool   `default:"true"`
	AdminPort           int    `default:"8081"`
	MockPort            int    `default:"8080"`
	MockDir             string `default:"."`
	MockFilepattern     string `default:"*-mock.*"`
	ResponseDir         string `default:"./responses"`
	ResponseFilepattern string `default:"*.*"`
}

func (c Configuration) info() string {
	return fmt.Sprintf(`Verbose: %v
Admin Port: %v
Mock Port: %v
Mock Dir: '%s'
Mock Filepattern: '%s'
Response Dir: '%s'
Response Filepattern: '%s'`, c.Verbose, c.AdminPort, c.MockPort, c.MockDir, c.MockFilepattern, c.ResponseDir, c.ResponseFilepattern)
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
		log.Fatalf("Can't load files: %v", err)
	}

	go routing.NewConfigRouter(mockRouter, logger).ListenAndServe(config.AdminPort)
	mockRouter.ListenAndServe(config.MockPort)
}
