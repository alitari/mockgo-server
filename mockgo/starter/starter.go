package starter

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/alitari/mockgo-server/mockgo/kvstore"
	"github.com/alitari/mockgo-server/mockgo/matches"
	"github.com/alitari/mockgo-server/mockgo/mock"
	"github.com/alitari/mockgo-server/mockgo/util"
	"github.com/gorilla/mux"
	"github.com/kelseyhightower/envconfig"
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

// BasicConfig is the basic mock configuration
var BasicConfig *BasicConfiguration

// logger is the basic logger
var logger *zap.Logger

func init() {
	BasicConfig = &BasicConfiguration{}
	if err := envconfig.Process("", BasicConfig); err != nil {
		log.Fatal("can't create configuration", zap.Error(err))
	}
}

// SetupRouter sets up the router with the given configuration, allows to control the server start
func SetupRouter(variant, versionTag, configInfo string, matchStore matches.Matchstore, kvStore kvstore.Storage, serve func(router *mux.Router)) {
	logger = util.CreateLogger(BasicConfig.LoglevelAPI)
	fmt.Printf(banner, variant, versionTag)
	logger.Info(BasicConfig.Info() + configInfo)
	router := mux.NewRouter()
	router.NewRoute().Name("health").Path(BasicConfig.APIPathPrefix + "/health").Methods(http.MethodGet).
		HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			writer.WriteHeader(http.StatusOK)
			writer.Write([]byte("OK"))
		})

	router.Use(util.BasicAuthMiddleware(BasicConfig.APIPathPrefix, BasicConfig.APIUsername, BasicConfig.APIPassword))
	mockHandler := mock.NewRequestHandler(BasicConfig.APIPathPrefix, BasicConfig.MockDir, BasicConfig.MockFilepattern, matchStore,
		kvstore.NewKVStoreTemplateFuncMap(kvStore), BasicConfig.LoglevelMock)
	if err := mockHandler.LoadFiles(); err != nil {
		logger.Fatal("can't load mockfiles", zap.Error(err))
	}
	matchHandler := matches.NewRequestHandler(BasicConfig.APIPathPrefix, matchStore, BasicConfig.LoglevelAPI)
	kvHandler := kvstore.NewRequestHandler(BasicConfig.APIPathPrefix, kvStore, BasicConfig.LoglevelAPI)

	mockHandler.AddRoutes(router)
	matchHandler.AddRoutes(router)
	kvHandler.AddRoutes(router)

	mock.RegisterMetrics()
	router.NewRoute().Name("metrics").Path("/__/metrics").Handler(promhttp.Handler())

	if serve == nil {
		serve = DefaultServe
	}
	serve(router)
}

// DefaultServe starts the default http server
func DefaultServe(router *mux.Router) {
	server := &http.Server{Addr: ":" + strconv.Itoa(BasicConfig.MockPort), Handler: router}
	logger.Info("serving http  ...", zap.String("address", server.Addr))
	err := server.ListenAndServe()
	if err != nil {
		logger.Fatal("can't start server", zap.Error(err))
	}
}
