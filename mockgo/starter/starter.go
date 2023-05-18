package starter

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/alitari/mockgo-server/mockgo/kvstore"
	"github.com/alitari/mockgo-server/mockgo/matches"
	"github.com/alitari/mockgo-server/mockgo/mock"
	"github.com/alitari/mockgo-server/mockgo/util"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

const banner = `
 __  __         _            
|  \/  |___  __| |______ ___ 
| |\/| / _ \/ _| / / _  / _ \
|_|  |_\___/\__|_\_\__, \___/
%s         |___/  %s

`

/*
RequestHandler abstraction of a set of http handler funcs
*/
type RequestHandler interface {
	AddRoutes(router *mux.Router)
}

/*
Configuration is the configuration model of the server which is defined via environment variables
*/
type Configuration interface {
	Info() string
	Port() int
	LogLevelAPI() int
	LogLevelMock() int
	PathPrefix() string
	User() string
	Password() string
	MockDirectory() string
	MockFilePattern() string
}

// BasicConfiguration is the basic configuration model of the server which is defined via environment variables
type BasicConfiguration struct {
	LoglevelAPI     int    `default:"-1" split_words:"true"`
	LoglevelMock    int    `default:"-1" split_words:"true"`
	MockPort        int    `default:"8081" split_words:"true"`
	MockDir         string `default:"." split_words:"true"`
	MockFilepattern string `default:"*-mock.*" split_words:"true"`
	MatchesCapacity int    `default:"1000" split_words:"true"`
	APIPathPrefix   string `default:"/__" split_words:"true"`
	APIUsername     string `default:"mockgo" split_words:"true"`
	APIPassword     string `default:"password" split_words:"true"`
}

// Info returns a string with the configuration info
func (c *BasicConfiguration) Info() string {
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

// Port returns the port of the server
func (c *BasicConfiguration) Port() int {
	return c.MockPort
}

// LogLevelAPI returns the loglevel of the api
func (c *BasicConfiguration) LogLevelAPI() int {
	return c.LoglevelAPI
}

// LogLevelMock returns the loglevel of the mock server
func (c *BasicConfiguration) LogLevelMock() int {
	return c.LoglevelMock
}

// PathPrefix returns the path prefix of the api
func (c *BasicConfiguration) PathPrefix() string {
	return c.APIPathPrefix
}

// MockDirectory returns the directory where the mock files are stored
func (c *BasicConfiguration) MockDirectory() string {
	return c.MockDir
}

// MockFilePattern returns the file pattern of the mock files
func (c *BasicConfiguration) MockFilePattern() string {
	return c.MockFilepattern
}

// User returns the user of the basic auth
func (c *BasicConfiguration) User() string {
	return c.APIUsername
}

// Password returns the password of the basic auth
func (c *BasicConfiguration) Password() string {
	return c.APIPassword
}

// SetupRouter sets up the router with the given configuration
func SetupRouter(config Configuration, variant, versionTag string, logger *zap.Logger, matchStore matches.Matchstore, kvStore kvstore.Storage) (*mux.Router, error) {
	fmt.Printf(banner, variant, versionTag)
	logger.Info(config.Info())
	router := mux.NewRouter()
	router.NewRoute().Name("health").Path(config.PathPrefix() + "/health").Methods(http.MethodGet).
		HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			writer.WriteHeader(http.StatusOK)
			writer.Write([]byte("OK"))
		})

	router.Use(util.BasicAuthMiddleware(config.PathPrefix(), config.User(), config.Password()))
	mockHandler := mock.NewRequestHandler(config.PathPrefix(), config.MockDirectory(), config.MockFilePattern(), matchStore,
		kvstore.NewKVStoreTemplateFuncMap(kvStore), config.LogLevelMock())
	if err := mockHandler.LoadFiles(); err != nil {
		return nil, fmt.Errorf("can't load mockfiles: %v", err)
	}
	matchHandler := matches.NewRequestHandler(config.PathPrefix(), matchStore, config.LogLevelAPI())
	kvHandler := kvstore.NewRequestHandler(config.PathPrefix(), kvStore, config.LogLevelAPI())

	mockHandler.AddRoutes(router)
	matchHandler.AddRoutes(router)
	kvHandler.AddRoutes(router)

	mock.RegisterMetrics()
	router.NewRoute().Name("metrics").Path("/__/metrics").Handler(promhttp.Handler())

	return router, nil
}

// Serve starts the server with the given configuration
func Serve(config Configuration, router *mux.Router, logger *zap.Logger) {
	server := &http.Server{Addr: ":" + strconv.Itoa(config.Port()), Handler: router}
	logger.Info("serving http  ...", zap.String("address", server.Addr))
	err := server.ListenAndServe()
	if err != nil {
		logger.Fatal("can't start server", zap.Error(err))
	}
}
