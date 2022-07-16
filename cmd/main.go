package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/alitari/mockgo-server/internal/routing"
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

type Configuration struct {
	Verbose bool `default:"true"`
	Port    int  `default:"8080"`
	ConfigFile    string  `default:"./config.json"`
}

func (c Configuration) info() string {
	return fmt.Sprintf(`Verbose: %v
Port: %v`, c.Verbose, c.Port)
}

func main() {
	config := Configuration{}
	if err := envconfig.Process("", &config); err != nil {
		log.Fatal(err)
	}

	log.Print(banner + config.info())

	r, err := routing.Load(config.ConfigFile)
	if err != nil {
		log.Fatalf("Error loading config file: %s", err)
	}

	log.Printf("Running on port [%v]\n", config.Port)
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(config.Port), r))
}
