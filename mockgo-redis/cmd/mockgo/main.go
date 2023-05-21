package main

import (
	"fmt"
	"log"

	"github.com/alitari/mockgo-server/mockgo-redis/kvstore"
	"github.com/alitari/mockgo-server/mockgo-redis/matchstore"
	"github.com/alitari/mockgo-server/mockgo/starter"
	"github.com/kelseyhightower/envconfig"
)

var versionTag = "latest"
var variant = "redis     "
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
	RedisAddress      string `default:"localhost:6379" split_words:"true"`
	RedisPassword     string `default:"" split_words:"true"`
	MatchstoreRedisDB int    `default:"0" split_words:"true"`
	KvstoreRedisDB    int    `default:"1" split_words:"true"`
}

func (c *Configuration) validate() error {
	if c.MatchstoreRedisDB == c.KvstoreRedisDB {
		return fmt.Errorf("redis db for matchstore and kvstore must be different")
	}
	return nil
}

func passwordInfo(realPassword string) string {
	if realPassword == "password" || len(realPassword) < 4 {
		return fmt.Sprintf("!! using UNSECURE password '%s' ", realPassword)
	}
	return realPassword[:3] + "***"
}

func (c *Configuration) info() string {
	return fmt.Sprintf(`

Redis:
  Address: '%s' ("REDIS_ADDRESS")
  Password: '%s' ("REDIS_PASSWORD")
  Matchstore Database: %d ("MATCHSTORE_REDIS_DB")
  KVStore Database: %d ("KVSTORE_REDIS_DB")

Redis:
`,
		c.RedisAddress, passwordInfo(c.RedisPassword), c.MatchstoreRedisDB, c.KvstoreRedisDB)

}

func main() {
	matchStore, err := matchstore.NewRedisMatchstore(config.RedisAddress, config.RedisPassword,
		config.MatchstoreRedisDB, uint16(starter.BasicConfig.MatchesCapacity))
	if err != nil {
		log.Fatalf("can't initialize redis matchstore: %v", err)
	}

	kvStore, err := kvstore.NewRedisStorage(config.RedisAddress, config.RedisPassword,
		config.KvstoreRedisDB)
	if err != nil {
		log.Fatalf("can't initialize redis kvstore: %v", err)
	}

	starter.SetupRouter(variant, versionTag, config.info(), matchStore, kvStore)

}
