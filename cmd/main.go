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
	Verbose            bool   `default:"true"`
	Port               int    `default:"8080"`
	MappingDir         string `default:"."`
	MappingFilepattern string `default:"*-mapping.json"`
}

func (c Configuration) info() string {
	return fmt.Sprintf(`Verbose: %v
Port: %v
Mapping Dir: %s
Mapping Filepattern: '%s'`, c.Verbose, c.Port, c.MappingDir, c.MappingFilepattern)
}

func main() {
	config := Configuration{}
	if err := envconfig.Process("", &config); err != nil {
		log.Fatal(err)
	}
	logger = &utils.Logger{Verbose: config.Verbose}

	logger.LogAlways(banner + config.info())

	mockRouter, err := routing.NewMockRouter(config.MappingDir, config.MappingFilepattern, logger)
	if err != nil {
		log.Fatalf("Error loading config file: %s", err)
	}
	mockRouter.ListenAndServe(config.Port)
}
