package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/alitari/mockgo-server/internal/config"
	"github.com/alitari/mockgo-server/internal/kvstore"
	"github.com/alitari/mockgo-server/internal/mock"
	"github.com/alitari/mockgo-server/internal/model"
	"github.com/alitari/mockgo-server/internal/utils"
	"github.com/kelseyhightower/envconfig"
)

const banner = `
 __  __         _            
|  \/  |___  __| |______ ___ 
| |\/| / _ \/ _| / / _  / _ \
|_|  |_\___/\__|_\_\__, \___/
                   |___/     
`

type Configuration struct {
	LoglevelConfig        int           `default:"1" split_words:"true"`
	LoglevelMock          int           `default:"1" split_words:"true"`
	ConfigPort            int           `default:"8080" split_words:"true"`
	ConfigUsername        string        `default:"mockgo" split_words:"true"`
	ConfigPassword        string        `default:"password" split_words:"true"`
	MockPort              int           `default:"8081" split_words:"true"`
	MockDir               string        `default:"." split_words:"true"`
	MockFilepattern       string        `default:"*-mock.*" split_words:"true"`
	MatchesCountOnly      bool          `default:"true" split_words:"true"`
	MismatchesCountOnly   bool          `default:"true" split_words:"true"`
	ResponseDir           string        `default:"./responses" split_words:"true"`
	ClusterUrls           []string      `default:"" split_words:"true"`
	ClusterPodLabelValue  string        `default:"" split_words:"true"`
	HttpClientTimeout     time.Duration `default:"1s" split_words:"true"`
	ProxyConfigRouterPath string        `default:"" split_words:"true"`
}

func (c *Configuration) validateAndFix() *Configuration {
	if len(c.ClusterUrls) > 0 && c.ClusterUrls[len(c.ClusterUrls)-1] == "" {
		c.ClusterUrls = c.ClusterUrls[:len(c.ClusterUrls)-1]
	}
	return c
}

func (c *Configuration) info() string {
	var passwordMessage string
	if c.ConfigPassword == "password" {
		passwordMessage = "!! using UNSECURE password 'password'"
	} else {
		passwordMessage = c.ConfigPassword[:3] + "***"
	}
	clusterSetup := config.NewClusterSetup(c.ClusterUrls)

	return fmt.Sprintf(`

Config API: 
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

Cluster:
  Setup: %v
  ClusterUrls: %v
  HttpClient timeout: '%v'
  ClusterPodLabelValue: '%s'
  ProxyConfigRouterPath: '%s'`,
		c.ConfigPort, c.ConfigUsername, passwordMessage, utils.ParseLogLevel(c.LoglevelConfig).String(),
		c.MockPort, c.MockDir, c.MockFilepattern, c.ResponseDir, utils.ParseLogLevel(c.LoglevelMock).String(),
		c.MatchesCountOnly, c.MismatchesCountOnly,
		clusterSetup.String(), c.ClusterUrls, c.HttpClientTimeout, c.ClusterPodLabelValue, c.ProxyConfigRouterPath)
}

func main() {
	mockRouter, configRouter := createRouters(kvstore.CreateTheStore())
	go startServe(configRouter)
	startServe(mockRouter)
}

func createRouters(kvstore *kvstore.KVStore) (*mock.MockRouter, *config.ConfigRouter) {
	configuration := createConfiguration().validateAndFix()
	configLogger := utils.NewLoggerUtil(utils.ParseLogLevel(configuration.LoglevelConfig))
	configLogger.LogAlways(banner + configuration.info())

	mockRouter := createMockRouter(configuration, kvstore, utils.NewLoggerUtil(utils.ParseLogLevel(configuration.LoglevelMock)))
	configRouter := createConfigRouter(configuration, mockRouter, kvstore, configLogger)
	err := mockRouter.LoadFiles(configRouter.TemplateFuncMap())
	if err != nil {
		log.Fatalf("(FATAL) Can't load files: %v", err)
	}
	return mockRouter, configRouter

}

func createConfiguration() *Configuration {
	configuration := Configuration{}
	if err := envconfig.Process("", &configuration); err != nil {
		log.Fatal(err)
	}
	return &configuration
}

func createMockRouter(configuration *Configuration, kvstore *kvstore.KVStore, logger *utils.LoggerUtil) *mock.MockRouter {
	mockRouter := mock.NewMockRouter(configuration.MockDir, configuration.MockFilepattern, configuration.ResponseDir,
		configuration.MockPort, kvstore, configuration.MatchesCountOnly, configuration.MismatchesCountOnly,
		configuration.ProxyConfigRouterPath, configuration.ConfigPort, configuration.HttpClientTimeout, logger)
	return mockRouter
}

func createConfigRouter(configuration *Configuration, mockRouter *mock.MockRouter, kvStore *kvstore.KVStore, logger *utils.LoggerUtil) *config.ConfigRouter {
	configRouter := config.NewConfigRouter(configuration.ConfigUsername, configuration.ConfigPassword, mockRouter,
		configuration.ConfigPort, configuration.ClusterUrls, configuration.ClusterPodLabelValue, kvStore, configuration.HttpClientTimeout, logger)
	err := configRouter.DownloadKVStoreFromCluster()
	if err != nil {
		log.Fatalf("(FATAL) Can't sync with cluster: %v\n", err)
	}
	return configRouter
}

func startServe(serving model.Serving) error {
	serving.Logger().LogAlways(fmt.Sprintf("Serving %s on port %v", serving.Name(), serving.Port()))
	s := serving.Server()
	err := s.ListenAndServe()
	if err != nil {
		return err
	}
	return nil
}

func stopServe(serving model.Serving) {
	serving.Logger().LogAlways(fmt.Sprintf("Stop Serving %s on port %d", serving.Name(), serving.Port()))
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := serving.Server().Shutdown(ctx); err != nil {
		log.Fatalf("Can't stop server %v", err)
	}
}
