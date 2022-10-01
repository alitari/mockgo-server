package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	grpckvstore "github.com/alitari/mockgo-grpc-kvstore/kvstore"
	"github.com/alitari/mockgo-grpc-matchstore/matchstore"
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
Cluster-grpc       |___/     
`

type RequestHandler interface {
	AddRoutes(router *mux.Router)
}

type Configuration struct {
	LoglevelAPI         int      `default:"1" split_words:"true"`
	LoglevelMock        int      `default:"1" split_words:"true"`
	APIUsername         string   `default:"mockgo" split_words:"true"`
	APIPassword         string   `default:"password" split_words:"true"`
	MockPort            int      `default:"8081" split_words:"true"`
	MockDir             string   `default:"." split_words:"true"`
	MockFilepattern     string   `default:"*-mock.*" split_words:"true"`
	ResponseDir         string   `default:"." split_words:"true"`
	ClusterHostnames    []string ` split_words:"true"`
	MatchstorePort      int      `default:"50051" split_words:"true"`
	MatchesCountOnly    bool     `default:"true" split_words:"true"`
	MismatchesCountOnly bool     `default:"true" split_words:"true"`
	KvstorePort         int      `default:"50151" split_words:"true"`
}

func (c *Configuration) validate() error {
	if len( c.ClusterHostnames) < 2 {
		return fmt.Errorf("you must define multiple hostnames, but you have just '%v'",c.ClusterHostnames)
	}
	return nil
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
  BasicAuth User: '%s'
  BasicAuth Password: %s
  LogLevel: '%s'

Mock Server:
  Port: %v
  Dir: '%s'
  Filepattern: '%s'
  Response Dir: '%s'
  LogLevel: '%s'
  
Count only:
  Matches: %v
  Mismatches: %v

Cluster:
  Hostnames: %v
  Matchstore port: %d
  KVstore port: %d
  `,
		c.APIUsername, passwordMessage, logging.ParseLogLevel(c.LoglevelAPI).String(),
		c.MockPort, c.MockDir, c.MockFilepattern, c.ResponseDir, logging.ParseLogLevel(c.LoglevelMock).String(),
		c.MatchesCountOnly, c.MismatchesCountOnly, c.ClusterHostnames, c.MatchstorePort, c.KvstorePort)
}

func main() {
	log.Print(banner)
	configuration := createConfiguration()
	log.Print(configuration.info())
	matchstore := createMatchstore(configuration)
	matchHandler := createMatchHandler(configuration, matchstore)
	kvStoreHandler := createKVStoreHandler(configuration)
	mockHandler := createMockHandler(configuration, matchstore)
	if err := mockHandler.LoadFiles(nil); err != nil {
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
	if err := configuration.validate(); err != nil {
		log.Fatal(err)
	}
	return &configuration
}

func createMatchstore(configuration *Configuration) *matchstore.GrpcMatchstore {
	matchstoreLogger := logging.NewLoggerUtil(logging.ParseLogLevel(configuration.LoglevelAPI))
	addresses := []string{}

	for _, host := range configuration.ClusterHostnames {
		if strings.Contains(host, ":") {
			addresses = append(addresses, host)
		} else {
			addresses = append(addresses, fmt.Sprintf("%s:%d", host, configuration.MatchstorePort))
		}
	}
	matchStore, err := matchstore.NewGrpcMatchstore(addresses, configuration.MatchstorePort, matchstoreLogger)
	if err != nil {
		log.Fatalf("can't initialize grpc matchstore: %v", err)
	}
	return matchStore
}

func createMatchHandler(configuration *Configuration, matchstore matches.Matchstore) *matches.MatchesRequestHandler {
	matchLogger := logging.NewLoggerUtil(logging.ParseLogLevel(configuration.LoglevelAPI))
	return matches.NewMatchesRequestHandler(configuration.APIUsername, configuration.APIPassword, matchstore, matchLogger)
}

func createKVStoreHandler(configuration *Configuration) *kvstore.KVStoreRequestHandler {
	kvstoreLogger := logging.NewLoggerUtil(logging.ParseLogLevel(configuration.LoglevelAPI))
	addresses := []string{}
	for _, host := range configuration.ClusterHostnames {
		if strings.Contains(host, ":") {
			addresses = append(addresses, host)
		} else {
			addresses = append(addresses, fmt.Sprintf("%s:%d", host, configuration.KvstorePort))
		}
	}
	kvs, err := grpckvstore.NewGrpcKVstore(addresses, configuration.KvstorePort, kvstoreLogger)
	if err != nil {
		log.Fatalf("can't initialize grpc kvstore: %v", err)
	}
	kvstoreJson := kvstore.NewKVStoreJSON(kvs, logging.ParseLogLevel(configuration.LoglevelAPI) == logging.Debug)
	return kvstore.NewKVStoreRequestHandler(configuration.APIUsername, configuration.APIPassword, kvstoreJson, kvstoreLogger)
}

func createMockHandler(configuration *Configuration, matchstore matches.Matchstore) *mock.MockRequestHandler {
	mockLogger := logging.NewLoggerUtil(logging.ParseLogLevel(configuration.LoglevelMock))
	mockHandler := mock.NewMockRequestHandler(configuration.MockDir, configuration.MockFilepattern, configuration.ResponseDir, matchstore, mockLogger)
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
