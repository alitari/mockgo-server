package main

import (
	"github.com/alitari/mockgo-server/mockgo/kvstore"
	"github.com/alitari/mockgo-server/mockgo/matches"
	"github.com/alitari/mockgo-server/mockgo/starter"
	"github.com/gorilla/mux"
)

var versionTag = "v0.0.1"
var variant = "standalone"

var serve func(router *mux.Router)

func main() {
	matchStore := matches.NewInMemoryMatchstore(uint16(starter.BasicConfig.MatchesCapacity))
	kvstore := kvstore.NewInmemoryStorage()

	starter.SetupRouter(variant, versionTag, matchStore, kvstore, serve)
}
