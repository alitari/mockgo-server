package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"text/template"

	"github.com/alitari/mockgo-server/mockgo/kvstore"
	"github.com/alitari/mockgo-server/mockgo/logging"
	"github.com/alitari/mockgo-server/mockgo/matches"
	"github.com/alitari/mockgo-server/mockgo/mock"
	"github.com/gorilla/mux"
	"github.com/kelseyhightower/envconfig"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const banner = `
 __  __         _            
|  \/  |___  __| |______ ___ 
| |\/| / _ \/ _| / / _  / _ \
|_|  |_\___/\__|_\_\__, \___/
Standalone         |___/  %s
`

const versionTag = "latest"

/*
RequestHandler abstraction of a set of http handler funcs
*/
type RequestHandler interface {
	AddRoutes(router *mux.Router)
}

/*
Configuration iss the configuration model of the server which is defined via environment variables
*/
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
  Path prefix: '%s' ("API_PATH_PREFIX")
  BasicAuth User: '%s' ("API_USERNAME")
  BasicAuth Password: %s ("API_PASSWORD")
  LogLevel: '%s' ("LOGLEVEL_API")

Mock Server:
  Port: %v ("MOCK_PORT")
  Dir: '%s' ("MOCK_DIR")
  Filepattern: '%s' ("MOCK_FILEPATTERN")
  LogLevel: '%s' ("LOGLEVEL_MOCK")
  
Matches:
  Capacity: %d ("MATCHES_CAPACITY")
  `,
		c.APIPathPrefix, c.APIUsername, passwordMessage, logging.ParseLogLevel(c.LoglevelAPI).String(),
		c.MockPort, c.MockDir, c.MockFilepattern, logging.ParseLogLevel(c.LoglevelMock).String(),
		c.MatchesCapacity)
}

func main() {
	router, port, err := setupRouter()
	if err != nil {
		log.Fatalf("can't setup router : %v", err)
	}
	server := &http.Server{Addr: ":" + strconv.Itoa(port), Handler: router}
	log.Printf("serving on '%s' ...", server.Addr)

	err = server.ListenAndServe()
	if err != nil {
		log.Fatalf("can't serve : %v", err)
	}
}

func setupRouter() (*mux.Router, int, error) {
	log.Printf(banner, versionTag)
	configuration := createConfiguration()
	log.Print(configuration.info())
	matchStore := matches.NewInMemoryMatchstore(uint16(configuration.MatchesCapacity))
	matchHandler := createMatchHandler(configuration, matchStore)
	kvStoreHandler := createKVStoreHandler(configuration)
	mockHandler := createMockHandler(configuration, matchStore, kvStoreHandler.GetFuncMap())
	if err := mockHandler.LoadFiles(); err != nil {
		return nil, -1, err
	}
	if err := mock.RegisterMetrics(); err != nil {
		return nil, -1, err
	}
	return createRouter(matchHandler, kvStoreHandler, mockHandler), configuration.MockPort, nil
}

func createConfiguration() *Configuration {
	configuration := Configuration{}
	if err := envconfig.Process("", &configuration); err != nil {
		log.Fatal(err)
	}
	return &configuration
}

func createMatchHandler(configuration *Configuration, matchstore matches.Matchstore) *matches.RequestHandler {
	matchLogger := logging.NewLoggerUtil(logging.ParseLogLevel(configuration.LoglevelAPI))
	return matches.NewRequestHandler(configuration.APIPathPrefix, configuration.APIUsername, configuration.APIPassword,
		matchstore, matchLogger)
}

func createKVStoreHandler(configuration *Configuration) *kvstore.RequestHandler {
	kvstoreLogger := logging.NewLoggerUtil(logging.ParseLogLevel(configuration.LoglevelAPI))
	return kvstore.NewRequestHandler(configuration.APIPathPrefix, configuration.APIUsername, configuration.APIPassword, kvstore.NewInmemoryStorage(), kvstoreLogger)
}

func createMockHandler(configuration *Configuration, matchstore matches.Matchstore, funcMap template.FuncMap) *mock.RequestHandler {
	mockLogger := logging.NewLoggerUtil(logging.ParseLogLevel(configuration.LoglevelMock))
	mockHandler := mock.NewRequestHandler(configuration.MockDir, configuration.MockFilepattern, matchstore, funcMap, mockLogger)
	return mockHandler
}

func createRouter(requestHandlers ...RequestHandler) *mux.Router {
	router := mux.NewRouter()
	for _, handler := range requestHandlers {
		handler.AddRoutes(router)
	}
	router.NewRoute().Name("metrics").Path("/__/metrics").Handler(promhttp.Handler())
	return router
}
