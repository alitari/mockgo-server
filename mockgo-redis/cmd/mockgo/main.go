package main

import (
	"fmt"
	"github.com/alitari/mockgo-server/mockgo/kvstore"
	"github.com/alitari/mockgo-server/mockgo/logging"
	"github.com/alitari/mockgo-server/mockgo/matches"
	"github.com/alitari/mockgo-server/mockgo/mock"
	rediskvstore "github.com/alitari/mockgo-server/redis-kvstore/kvstore"
	redismatchstore "github.com/alitari/mockgo-server/redis-matchstore/matchstore"
	"github.com/gorilla/mux"
	"github.com/kelseyhightower/envconfig"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
	"strconv"
	"text/template"
)

const banner = `
 __  __         _            
|  \/  |___  __| |______ ___ 
| |\/| / _ \/ _| / / _  / _ \
|_|  |_\___/\__|_\_\__, \___/
Redis              |___/  %s
`

const versionTag = "latest"

/*
RequestHandler abstraction of a set of http handler funcs
*/
type RequestHandler interface {
	AddRoutes(router *mux.Router)
}

/*
Configuration is the configuration model of the server which is defined via environment variables
*/
type Configuration struct {
	LoglevelAPI        int    `default:"1" split_words:"true"`
	LoglevelMock       int    `default:"1" split_words:"true"`
	LoglevelMatchstore int    `default:"1" split_words:"true"`
	LoglevelKvstore    int    `default:"1" split_words:"true"`
	APIPathPrefix      string `default:"/__" split_words:"true"`
	APIUsername        string `default:"mockgo" split_words:"true"`
	APIPassword        string `default:"password" split_words:"true"`
	MockPort           int    `default:"8081" split_words:"true"`
	MockDir            string `default:"." split_words:"true"`
	MockFilepattern    string `default:"*-mock.*" split_words:"true"`
	RedisAddress       string `default:"localhost:6379" split_words:"true"`
	RedisPassword      string `default:"" split_words:"true"`
	MatchstoreRedisDB  int    `default:"0" split_words:"true"`
	KvstoreRedisDB     int    `default:"1" split_words:"true"`
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

API: 
  Path prefix: '%s' ("API_PATH_PREFIX")
  BasicAuth User: '%s' ("API_USERNAME")
  BasicAuth Password: %s ("API_PASSWORD")
  LogLevel: '%s' ("LOGLEVEL_API")

Mock Server:
  Port: %v ("MOCK_PORT")
  Dir: '%s' ("MOCK_DIR")
  Filepattern: '%s' ("MOCK_FILEPATTERN")
  LogLevel: '%s' ("LOGLEVEL_MOCK")

Matchstore:
  Redis Database: %d ("MATCHSTORE_REDIS_DB")

KVStore:
  Redis Database: %d ("KVSTORE_REDIS_DB")

Redis:
  Address: '%s' ("REDIS_ADDRESS")
  Password: '%s' ("REDIS_PASSWORD")
`,
		c.APIPathPrefix, c.APIUsername, passwordInfo(c.APIPassword), logging.ParseLogLevel(c.LoglevelAPI).String(),
		c.MockPort, c.MockDir, c.MockFilepattern, logging.ParseLogLevel(c.LoglevelMock).String(),
		c.MatchstoreRedisDB, c.KvstoreRedisDB, c.RedisAddress, passwordInfo(c.RedisPassword))

}

func main() {
	log.Printf(banner, versionTag)
	configuration := createConfiguration()
	log.Print(configuration.info())
	matchstore := createMatchstore(configuration)
	matchHandler := createMatchHandler(configuration, matchstore)
	kvStoreHandler := createKVStoreHandler(configuration)
	mockHandler := createMockHandler(configuration, matchstore, kvStoreHandler.GetFuncMap())
	if err := mockHandler.LoadFiles(); err != nil {
		log.Fatalf("can't load mock files: %v", err)
	}
	startServing(configuration, matchHandler, kvStoreHandler, mockHandler)
}

// a change
func createConfiguration() *Configuration {
	configuration := Configuration{}
	if err := envconfig.Process("", &configuration); err != nil {
		log.Fatal(err)
	}
	if err := configuration.validate(); err != nil {
		log.Fatal(err)
	}
	return &configuration
}

func createMatchstore(configuration *Configuration) matches.Matchstore {
	matchStore, err := redismatchstore.NewRedisMatchstore(configuration.RedisAddress, configuration.RedisPassword,
		configuration.MatchstoreRedisDB, 1000)
	if err != nil {
		log.Fatalf("can't initialize redis matchstore: %v", err)
	}
	return matchStore
}

func createMatchHandler(configuration *Configuration, matchstore matches.Matchstore) *matches.RequestHandler {
	matchLogger := logging.NewLoggerUtil(logging.ParseLogLevel(configuration.LoglevelAPI))
	return matches.NewRequestHandler(configuration.APIPathPrefix, configuration.APIUsername, configuration.APIPassword,
		matchstore, matchLogger)
}

func createKVStoreHandler(configuration *Configuration) *kvstore.RequestHandler {
	kvstoreLogger := logging.NewLoggerUtil(logging.ParseLogLevel(configuration.LoglevelKvstore))
	kvs, err := rediskvstore.NewRedisStorage(configuration.RedisAddress, configuration.RedisPassword,
		configuration.KvstoreRedisDB)
	if err != nil {
		log.Fatalf("can't initialize redis kvstore: %v", err)
	}
	return kvstore.NewRequestHandler(configuration.APIPathPrefix, configuration.APIUsername, configuration.APIPassword, kvs, kvstoreLogger)
}

func createMockHandler(configuration *Configuration, matchstore matches.Matchstore, funcMap template.FuncMap) *mock.RequestHandler {
	mockLogger := logging.NewLoggerUtil(logging.ParseLogLevel(configuration.LoglevelMock))
	mockHandler := mock.NewRequestHandler(configuration.APIPathPrefix, configuration.APIUsername, configuration.APIPassword,
		configuration.MockDir, configuration.MockFilepattern, matchstore, funcMap, mockLogger)
	return mockHandler
}

func startServing(configuration *Configuration, requestHandlers ...RequestHandler) error {
	router := mux.NewRouter()
	for _, handler := range requestHandlers {
		handler.AddRoutes(router)
	}
	router.NewRoute().Name("metrics").Path("/__/metrics").Handler(promhttp.Handler())
	server := &http.Server{Addr: ":" + strconv.Itoa(configuration.MockPort), Handler: router}
	log.Printf("Serving on '%s'", server.Addr)

	err := server.ListenAndServe()
	if err != nil {
		return err
	}
	return nil
}
