package main

import (
	"fmt"
	"github.com/alitari/mockgo-server/mockgo/kvstore"
	"github.com/alitari/mockgo-server/mockgo/matches"
	"github.com/alitari/mockgo-server/mockgo/mock"
	"github.com/alitari/mockgo-server/mockgo/util"
	"github.com/gorilla/mux"
	"github.com/kelseyhightower/envconfig"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"net/http"
	"strconv"
)

const banner = `
 __  __         _            
|  \/  |___  __| |______ ___ 
| |\/| / _ \/ _| / / _  / _ \
|_|  |_\___/\__|_\_\__, \___/
Standalone         |___/  %s
`

const versionTag = "latest"

var logger *zap.Logger

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
  LogLevel: '%v' ("LOGLEVEL_API")

Mock Server:
  Port: %v ("MOCK_PORT")
  Dir: '%s' ("MOCK_DIR")
  Filepattern: '%s' ("MOCK_FILEPATTERN")
  LogLevel: '%v' ("LOGLEVEL_MOCK")
  
Matches:
  Capacity: %d ("MATCHES_CAPACITY")
  `,
		c.APIPathPrefix, c.APIUsername, passwordMessage, c.LoglevelAPI,
		c.MockPort, c.MockDir, c.MockFilepattern, c.LoglevelMock,
		c.MatchesCapacity)
}

func main() {
	router, port, err := setupRouter()
	if err != nil {
		logger.Fatal("can't setup router", zap.Error(err))
	}
	server := &http.Server{Addr: ":" + strconv.Itoa(port), Handler: router}
	logger.Info("serving http  ...", zap.String("address", server.Addr))

	err = server.ListenAndServe()
	if err != nil {
		logger.Fatal("can't start server", zap.Error(err))
	}
}

func setupRouter() (*mux.Router, int, error) {
	fmt.Printf(banner, versionTag)
	configuration := createConfiguration()
	logger = util.CreateLogger(configuration.LoglevelAPI)
	logger.Info(configuration.info())
	matchStore := matches.NewInMemoryMatchstore(uint16(configuration.MatchesCapacity))
	matchHandler := matches.NewRequestHandler(configuration.APIPathPrefix, configuration.APIUsername, configuration.APIPassword,
		matchStore, configuration.LoglevelAPI)
	kvStoreHandler := kvstore.NewRequestHandler(configuration.APIPathPrefix, configuration.APIUsername, configuration.APIPassword, kvstore.NewInmemoryStorage(), configuration.LoglevelAPI)
	mockHandler := mock.NewRequestHandler(configuration.APIPathPrefix, configuration.APIUsername,
		configuration.APIPassword, configuration.MockDir, configuration.MockFilepattern, matchStore, kvStoreHandler.GetFuncMap(), configuration.LoglevelMock)
	if err := mockHandler.LoadFiles(); err != nil {
		return nil, -1, fmt.Errorf("can't load mockfiles: %v", err)
	}
	if err := mock.RegisterMetrics(); err != nil {
		return nil, -1, err
	}
	return createRouter(matchHandler, kvStoreHandler, mockHandler), configuration.MockPort, nil
}

func createConfiguration() *Configuration {
	configuration := Configuration{}
	if err := envconfig.Process("", &configuration); err != nil {
		logger.Fatal("can't create configuration", zap.Error(err))
	}
	return &configuration
}

func createRouter(requestHandlers ...RequestHandler) *mux.Router {
	router := mux.NewRouter()
	for _, handler := range requestHandlers {
		handler.AddRoutes(router)
	}
	router.NewRoute().Name("metrics").Path("/__/metrics").Handler(promhttp.Handler())
	return router
}
