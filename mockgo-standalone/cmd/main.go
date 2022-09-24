package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/alitari/mockgo/kvstore"
	"github.com/alitari/mockgo/logging"
	"github.com/alitari/mockgo/matches"
	"github.com/alitari/mockgo/mock"
	"github.com/kelseyhightower/envconfig"
)

const banner = `
 __  __         _            
|  \/  |___  __| |______ ___ 
| |\/| / _ \/ _| / / _  / _ \
|_|  |_\___/\__|_\_\__, \___/
Standalone         |___/     
`

type Configuration struct {
	LoglevelAPI         int    `default:"1" split_words:"true"`
	LoglevelMock        int    `default:"1" split_words:"true"`
	APIPort             int    `default:"8080" split_words:"true"`
	APIUsername         string `default:"mockgo" split_words:"true"`
	APIPassword         string `default:"password" split_words:"true"`
	MockPort            int    `default:"8081" split_words:"true"`
	MockDir             string `default:"." split_words:"true"`
	MockFilepattern     string `default:"*-mock.*" split_words:"true"`
	MatchesCountOnly    bool   `default:"true" split_words:"true"`
	MismatchesCountOnly bool   `default:"true" split_words:"true"`
	ResponseDir         string `default:"./responses" split_words:"true"`
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
  Port: %v
  BasicAuth User: '%s'
  BasicAuth Password: %s
  LogLevel: '%s'

Mock Server:
  Port: %v
  Dir: '%s'
  Filepattern: '%s'
  Response Dir: '%s'
  LogLevel: '%s'

CountOnly:
  Matches: %v
  Mismatches: %v
`,
		c.APIPort, c.APIUsername, passwordMessage, logging.ParseLogLevel(c.LoglevelAPI).String(),
		c.MockPort, c.MockDir, c.MockFilepattern, c.ResponseDir, logging.ParseLogLevel(c.LoglevelMock).String(),
		c.MatchesCountOnly, c.MismatchesCountOnly)
}

func main() {
	configuration := createConfiguration()
	matchStore := matches.NewInMemoryMatchstore(configuration.MatchesCountOnly, configuration.MismatchesCountOnly)
	matchRouter := createMatchAPIRouter(configuration, matchStore)
	kvStoreAPIRouter := createKVStoreAPIRouter(configuration)
	mockRouter := createMockRouter(configuration, matchStore)
	// go startServe(configRouter)
	// startServe(mockRouter)
}

func createMatchAPIRouter(configuration *Configuration, matchstore matches.Matchstore) *matches.MatchAPIRouter {
	matchLogger := logging.NewLoggerUtil(logging.ParseLogLevel(configuration.LoglevelAPI))
	return matches.NewMatchAPIRouter(configuration.APIUsername, configuration.APIPassword, matchStore, matchLogger)
}

func createKVStoreAPIRouter(configuration *Configuration) *kvstore.KVStoreAPIRouter {
	kvstoreLogger := logging.NewLoggerUtil(logging.ParseLogLevel(configuration.LoglevelAPI))
	kvstoreJson := kvstore.NewKVStoreJSON(kvstore.NewKVStoreInMemory(), logging.ParseLogLevel(configuration.LoglevelAPI) == logging.Debug)
	return kvstore.NewKVStoreAPIRouter(configuration.APIUsername, configuration.APIPassword, kvstoreJson, kvstoreLogger)
}

func createConfiguration() *Configuration {
	configuration := Configuration{}
	if err := envconfig.Process("", &configuration); err != nil {
		log.Fatal(err)
	}
	return &configuration
}

func createMockRouter(configuration *Configuration, matchstore matches.Matchstore) *mock.MockRouter {
	mockLogger := logging.NewLoggerUtil(logging.ParseLogLevel(configuration.LoglevelMock))
	mockRouter := mock.NewMockRouter(configuration.MockDir, configuration.MockFilepattern, configuration.ResponseDir, matchstore, "", "", time.Second, mockLogger)
	return mockRouter
}

func startServeAPI(configuration *Configuration, matchRouter *matches.MatchAPIRouter, kvStoreRouter *kvstore.KVStoreAPIRouter) error {
	log.Printf("Serving %s on port %v", "startServeAPI", configuration.APIPort)
	
	server := &http.Server{Addr: ":" + strconv.Itoa(configuration.APIPort), Handler: matchRouter.Router}
	err := server.ListenAndServe()
	if err != nil {
		return err
	}
	return nil
}

func stopServe(serving model.Serving) {
	log.Printf("Stop Serving %s on port %d", serving.Name(), serving.Port())
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := serving.Server().Shutdown(ctx); err != nil {
		log.Fatalf("Can't stop server %v", err)
	}
}

func createInMemoryStore() *kvstore.KVStoreJSON {
	kvstoreImpl := kvstore.NewKVStoreInMemory()
	return kvstore.NewKVStoreJSON(&kvstoreImpl, true)
}
