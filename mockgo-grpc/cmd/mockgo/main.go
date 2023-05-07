package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"text/template"

	"github.com/alitari/mockgo-server/mockgo/util"

	grpckvstore "github.com/alitari/mockgo-server/grpc-kvstore/kvstore"
	"github.com/alitari/mockgo-server/grpc-matchstore/matchstore"
	"github.com/alitari/mockgo-server/mockgo/kvstore"
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
Cluster-grpc       |___/  %s
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
	LoglevelAPI        int      `default:"1" split_words:"true"`
	LoglevelMock       int      `default:"1" split_words:"true"`
	LoglevelMatchstore int      `default:"1" split_words:"true"`
	LoglevelKvstore    int      `default:"1" split_words:"true"`
	APIPathPrefix      string   `default:"/__" split_words:"true"`
	APIUsername        string   `default:"mockgo" split_words:"true"`
	APIPassword        string   `default:"password" split_words:"true"`
	MockPort           int      `default:"8081" split_words:"true"`
	MockDir            string   `default:"." split_words:"true"`
	MockFilepattern    string   `default:"*-mock.*" split_words:"true"`
	ClusterHostnames   []string ` split_words:"true"`
	MatchstorePort     int      `default:"50051" split_words:"true"`
	MatchesCapacity    int      `default:"1000" split_words:"true"`
	KvstorePort        int      `default:"50151" split_words:"true"`
}

func (c *Configuration) validate() error {

	if len(c.ClusterHostnames) < 2 {
		return fmt.Errorf("you must define multiple hostnames, but you have just '%v'", c.ClusterHostnames)
	}
	// fix empty string at the end
	if len(c.ClusterHostnames[len(c.ClusterHostnames)-1]) == 0 {
		c.ClusterHostnames = c.ClusterHostnames[:len(c.ClusterHostnames)-1]
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
  Path prefix: '%s' ("API_PATH_PREFIX")
  BasicAuth User: '%s' ("API_USERNAME")
  BasicAuth Password: %s ("API_PASSWORD")
  LogLevel: '%v' ("LOGLEVEL_API")

Mock Server:
  Port: %v ("MOCK_PORT")
  Dir: '%s' ("MOCK_DIR")
  Filepattern: '%s' ("MOCK_FILEPATTERN")
  LogLevel: '%v' ("LOGLEVEL_MOCK")
  
Matchstore:
  Port: %d ("MATCHSTORE_PORT")
  LogLevel: %v ("LOGLEVEL_MATCHSTORE")
  Capacity: %d ("MATCHES_CAPACITY")

KVStore:
  Port: %d ("KVSTORE_PORT")
  LogLevel: %v ("LOGLEVEL_KVSTORE")

Cluster:
  Hostnames: %v ("CLUSTER_HOSTNAMES")
`,
		c.APIPathPrefix, c.APIUsername, passwordMessage, c.LoglevelAPI,
		c.MockPort, c.MockDir, c.MockFilepattern, c.LoglevelMock,
		c.MatchstorePort, c.LoglevelMatchstore, c.MatchesCapacity,
		c.KvstorePort, c.LoglevelKvstore,
		c.ClusterHostnames)
}

func main() {
	log.Printf(banner, versionTag)
	configuration := createConfiguration()
	log.Print(configuration.info())
	matchstore := createMatchstore(configuration)
	matchHandler := createMatchHandler(configuration, matchstore)
	kvStoreHandler := createKVStoreHandler(configuration)
	mockHandler := createMockHandler(configuration, matchstore, nil)
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
	addresses := []string{}

	for _, host := range configuration.ClusterHostnames {
		if strings.Contains(host, ":") {
			addresses = append(addresses, host)
		} else {
			addresses = append(addresses, fmt.Sprintf("%s:%d", host, configuration.MatchstorePort))
		}
	}
	matchStore, err := matchstore.NewGrpcMatchstore(addresses, configuration.MatchstorePort, uint16(configuration.MatchesCapacity), configuration.LoglevelMatchstore)
	if err != nil {
		log.Fatalf("can't initialize grpc matchstore: %v", err)
	}
	return matchStore
}

func createMatchHandler(configuration *Configuration, matchstore matches.Matchstore) *matches.RequestHandler {
	return matches.NewRequestHandler(configuration.APIPathPrefix,
		matchstore, configuration.LoglevelAPI)
}

func createKVStoreHandler(configuration *Configuration) *kvstore.RequestHandler {
	addresses := []string{}
	for _, host := range configuration.ClusterHostnames {
		if strings.Contains(host, ":") {
			addresses = append(addresses, host)
		} else {
			addresses = append(addresses, fmt.Sprintf("%s:%d", host, configuration.KvstorePort))
		}
	}
	kvs, err := grpckvstore.NewGrpcStorage(addresses, configuration.KvstorePort, configuration.LoglevelKvstore)
	if err != nil {
		log.Fatalf("can't initialize grpc kvstore: %v", err)
	}
	return kvstore.NewRequestHandler(configuration.APIPathPrefix, kvs, configuration.LoglevelKvstore)
}

func createMockHandler(configuration *Configuration, matchstore matches.Matchstore, funcMap template.FuncMap) *mock.RequestHandler {
	mockHandler := mock.NewRequestHandler(configuration.APIPathPrefix,
		configuration.MockDir, configuration.MockFilepattern, matchstore, funcMap, configuration.LoglevelMock)
	return mockHandler
}

func startServing(configuration *Configuration, requestHandlers ...RequestHandler) error {
	router := mux.NewRouter()
	//router.Use(loggingMiddleware)
	router.Use(util.BasicAuthMiddleware(configuration.APIPathPrefix, configuration.APIUsername, configuration.APIPassword))
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
