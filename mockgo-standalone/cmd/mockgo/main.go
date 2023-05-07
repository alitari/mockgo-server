package main

import (
	"github.com/alitari/mockgo-server/mockgo/kvstore"
	"github.com/alitari/mockgo-server/mockgo/matches"
	"github.com/alitari/mockgo-server/mockgo/starter"
	"github.com/alitari/mockgo-server/mockgo/util"
	"github.com/gorilla/mux"
	"github.com/kelseyhightower/envconfig"
	"go.uber.org/zap"
)

var versionTag = "v0.0.1"
var variant = "standalone"
var logger *zap.Logger
var config *starter.BasicConfiguration

var startServe func(*mux.Router)

func init() {
	initializations()
}

func initializations() {
	config = createConfiguration()
	logger = util.CreateLogger(config.LoglevelAPI)
	startServe = func(router *mux.Router) {
		starter.Serve(config, router, logger)
	}

}

func main() {
	matchStore := matches.NewInMemoryMatchstore(uint16(config.MatchesCapacity))
	kvstore := kvstore.NewInmemoryStorage()
	router, err := starter.SetupRouter(config, variant, versionTag, logger, matchStore, kvstore)
	if err != nil {
		logger.Fatal("can't create router", zap.Error(err))
	}
	startServe(router)
}

func createConfiguration() *starter.BasicConfiguration {
	configuration := starter.BasicConfiguration{}
	if err := envconfig.Process("", &configuration); err != nil {
		logger.Fatal("can't create configuration", zap.Error(err))
	}
	return &configuration
}
