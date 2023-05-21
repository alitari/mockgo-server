package main

import (
	"github.com/alitari/mockgo-server/mockgo/kvstore"
	"github.com/alitari/mockgo-server/mockgo/matches"
	"github.com/alitari/mockgo-server/mockgo/starter"
)

var versionTag = "latest"
var variant = "standalone"

func main() {
	matchStore := matches.NewInMemoryMatchstore(uint16(starter.BasicConfig.MatchesCapacity))
	kvstore := kvstore.NewInmemoryStorage()
	starter.SetupRouter(variant, versionTag, "", matchStore, kvstore)
}
