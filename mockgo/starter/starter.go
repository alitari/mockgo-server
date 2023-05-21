package starter

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/alitari/mockgo-server/mockgo/kvstore"
	"github.com/alitari/mockgo-server/mockgo/matches"
	"github.com/alitari/mockgo-server/mockgo/mock"
	"github.com/alitari/mockgo-server/mockgo/util"
	"github.com/gorilla/mux"
	"github.com/kelseyhightower/envconfig"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
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
	LoglevelAPI     string `default:"INFO" split_words:"true"`
	LoglevelMock    string `default:"INFO" split_words:"true"`
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

var server *http.Server

// Serving is the function which starts the server
var Serving func(router *mux.Router) error

// Shutdown is the function which stops the server
var Shutdown func() error

// logger is the basic logger
var logger *zap.Logger

// StopChannel is the channel which is used to stop the server
var StopChannel = make(chan os.Signal, 1)

func init() {
	BasicConfig = &BasicConfiguration{}
	if err := envconfig.Process("", BasicConfig); err != nil {
		log.Fatal("can't create configuration", zap.Error(err))
	}
}

// SetupRouter sets up the router with the given configuration, allows to control the server start
func SetupRouter(variant, versionTag, configInfo string, matchStore matches.Matchstore, kvStore kvstore.Storage) {

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		signal.Notify(StopChannel, os.Interrupt, syscall.SIGTERM)
		<-StopChannel
		cancel()
	}()

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
	router.NewRoute().Name("metrics").Path(BasicConfig.APIPathPrefix + "/metrics").Handler(promhttp.Handler())

	if Serving == nil {
		Serving = DefaultServing
	}
	if Shutdown == nil {
		Shutdown = DefaultShutdown
	}

	g, gCtx := errgroup.WithContext(ctx)
	g.Go(func() error {
		return Serving(router)
	})
	g.Go(func() error {
		<-gCtx.Done()

		logger.Info("shutting down matchstore ...")
		if err := matchStore.Shutdown(); err != nil {
			logger.Error("can't shutdown matchstore", zap.Error(err))
		}
		logger.Info("shutting down kvstore ...")
		if err := kvStore.Shutdown(); err != nil {
			logger.Error("can't shutdown kvstore", zap.Error(err))
		}
		logger.Info("shutting down http server ...")
		return Shutdown()
	})

	if err := g.Wait(); err != nil {
		logger.Info("exit ", zap.Error(err))
	}
}

// DefaultServing is the default server
func DefaultServing(router *mux.Router) error {
	server = &http.Server{Addr: ":" + strconv.Itoa(BasicConfig.MockPort), Handler: router}
	logger.Info("serving http  ...", zap.String("address", server.Addr))
	return server.ListenAndServe()
}

// DefaultShutdown is the default shutdown
func DefaultShutdown() error {
	return server.Shutdown(context.Background())
}
