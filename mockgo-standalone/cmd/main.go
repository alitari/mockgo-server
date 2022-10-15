package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/alitari/mockgo/kvstore"
	"github.com/alitari/mockgo/logging"
	"github.com/alitari/mockgo/matches"
	"github.com/alitari/mockgo/mock"
	"github.com/gorilla/mux"
	"github.com/kelseyhightower/envconfig"
)

const banner = `
 __  __         _            
|  \/  |___  __| |______ ___ 
| |\/| / _ \/ _| / / _  / _ \
|_|  |_\___/\__|_\_\__, \___/
Standalone         |___/     
`

type RequestHandler interface {
	AddRoutes(router *mux.Router)
}

type Configuration struct {
	LoglevelAPI     int    `default:"1" split_words:"true"`
	LoglevelMock    int    `default:"1" split_words:"true"`
	APIPathPrefix   string `default:"/__" split_words:"true"`
	APIUsername     string `default:"mockgo" split_words:"true"`
	APIPassword     string `default:"password" split_words:"true"`
	MockPort        int    `default:"8081" split_words:"true"`
	MockDir         string `default:"." split_words:"true"`
	MockFilepattern string `default:"*-mock.*" split_words:"true"`
	MatchesCapacity int    `default:"1000" split_words:"true"`
}

func (c *Configuration) info() string {
	var passwordMessage string
	if c.APIPassword == "password" {
		passwordMessage = "!! using UNSECURE password 'password'"
	} else {
		passwordMessage = c.APIPassword[:3] + "***"
	}
	return fmt.Sprintf(`

API: 
  Path prefix: '%s'
  BasicAuth User: '%s'
  BasicAuth Password: %s
  LogLevel: '%s'

Mock Server:
  Port: %v
  Dir: '%s'
  Filepattern: '%s'
  LogLevel: '%s'
  
Matches:
  Capacity: %d
  `,
		c.APIPathPrefix, c.APIUsername, passwordMessage, logging.ParseLogLevel(c.LoglevelAPI).String(),
		c.MockPort, c.MockDir, c.MockFilepattern, logging.ParseLogLevel(c.LoglevelMock).String(),
		c.MatchesCapacity)
}

func main() {
	log.Print(banner)
	configuration := createConfiguration()
	log.Print(configuration.info())
	matchStore := matches.NewInMemoryMatchstore(uint16(configuration.MatchesCapacity))
	matchHandler := createMatchHandler(configuration, matchStore)
	kvStoreHandler, kvs, logger := createKVStoreHandler(configuration)
	mockHandler := createMockHandler(configuration, matchStore)
	if err := mockHandler.LoadFiles(kvstore.KVStoreFuncMap(kvs, logger)); err != nil {
		log.Fatalf("can't load mock files: %v", err)
	}
	startServing(configuration, matchHandler, kvStoreHandler, mockHandler)
	// go startServe(configRouter)
	// startServe(mockRouter)
}

func createConfiguration() *Configuration {
	configuration := Configuration{}
	if err := envconfig.Process("", &configuration); err != nil {
		log.Fatal(err)
	}
	return &configuration
}

func createMatchHandler(configuration *Configuration, matchstore matches.Matchstore) *matches.MatchesRequestHandler {
	matchLogger := logging.NewLoggerUtil(logging.ParseLogLevel(configuration.LoglevelAPI))
	return matches.NewMatchesRequestHandler(configuration.APIPathPrefix, configuration.APIUsername, configuration.APIPassword,
		matchstore, matchLogger)
}

func createKVStoreHandler(configuration *Configuration) (*kvstore.KVStoreRequestHandler, *kvstore.KVStoreJSON, *logging.LoggerUtil) {
	kvstoreLogger := logging.NewLoggerUtil(logging.ParseLogLevel(configuration.LoglevelAPI))
	kvstoreJson := kvstore.NewKVStoreJSON(kvstore.NewInmemoryKVStore(), logging.ParseLogLevel(configuration.LoglevelAPI) == logging.Debug)
	return kvstore.NewKVStoreRequestHandler(configuration.APIPathPrefix, configuration.APIUsername, configuration.APIPassword, kvstoreJson, kvstoreLogger), kvstoreJson, kvstoreLogger
}

func createMockHandler(configuration *Configuration, matchstore matches.Matchstore) *mock.MockRequestHandler {
	mockLogger := logging.NewLoggerUtil(logging.ParseLogLevel(configuration.LoglevelMock))
	mockHandler := mock.NewMockRequestHandler(configuration.MockDir, configuration.MockFilepattern, matchstore, mockLogger)
	return mockHandler
}

func startServing(configuration *Configuration, requestHandlers ...RequestHandler) error {
	router := mux.NewRouter()
	for _, handler := range requestHandlers {
		handler.AddRoutes(router)
	}
	server := &http.Server{Addr: ":" + strconv.Itoa(configuration.MockPort), Handler: router}
	log.Printf("Serving on '%s'", server.Addr)

	err := server.ListenAndServe()
	if err != nil {
		return err
	}
	return nil
}
