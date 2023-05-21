package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/alitari/mockgo-server/mockgo/starter"

	"github.com/alitari/mockgo-server/mockgo-grpc/kvstore"
	"github.com/alitari/mockgo-server/mockgo-grpc/matchstore"
	"github.com/kelseyhightower/envconfig"
)

var versionTag = "latest"
var variant = "grpc      "
var config *Configuration

func init() {
	initialization()
}

func initialization() {
	config = &Configuration{}
	if err := envconfig.Process("", config); err != nil {
		log.Fatal(fmt.Sprintf("can't create configuration: %v", err))
	}
	if err := config.validate(); err != nil {
		log.Fatal(fmt.Sprintf("wrong configuration: %v", err))
	}
}

/*
Configuration is the configuration model of the server which is defined via environment variables
*/
type Configuration struct {
	ClusterHostnames []string `split_words:"true"`
	MatchstorePort   int      `default:"50051" split_words:"true"`
	KvstorePort      int      `default:"50151" split_words:"true"`
}

func (c *Configuration) validate() error {
	if len(c.ClusterHostnames) < 2 {
		return fmt.Errorf("you must define multiple hostnames, but you have just '%v'", c.ClusterHostnames)
	}
	// fix empty string at the end
	if len(c.ClusterHostnames[len(c.ClusterHostnames)-1]) == 0 {
		c.ClusterHostnames = c.ClusterHostnames[:len(c.ClusterHostnames)-1]
	}
	return nil
}

func (c *Configuration) info() string {
	return fmt.Sprintf(`
GRPC:
  Hostnames: %v ("CLUSTER_HOSTNAMES")
  KVStore Port: %d ("KVSTORE_PORT")
  Matchstore Port: %d ("MATCHSTORE_PORT")

`,
		c.ClusterHostnames,
		c.KvstorePort,
		c.MatchstorePort)
}

func main() {
	matchStore, err := matchstore.NewGrpcMatchstore(createAddresses(config.ClusterHostnames, config.MatchstorePort),
		config.MatchstorePort, uint16(starter.BasicConfig.MatchesCapacity), starter.BasicConfig.LoglevelAPI)
	if err != nil {
		log.Fatalf("can't initialize grpc matchstore: %v", err)
	}

	kvStore, err := kvstore.NewGrpcStorage(createAddresses(config.ClusterHostnames, config.MatchstorePort),
		config.KvstorePort, starter.BasicConfig.LoglevelAPI)
	if err != nil {
		log.Fatalf("can't initialize grpc kvstore: %v", err)
	}

	starter.SetupRouter(variant, versionTag, config.info(), matchStore, kvStore)
}

func createAddresses(hostNames []string, port int) []string {
	addresses := []string{}
	for _, host := range hostNames {
		if strings.Contains(host, ":") {
			addresses = append(addresses, host)
		} else {
			addresses = append(addresses, fmt.Sprintf("%s:%d", host, port))
		}
	}
	return addresses
}
