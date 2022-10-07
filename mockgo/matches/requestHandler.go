package matches

import (
	"net/http"

	"github.com/alitari/mockgo/logging"
	"github.com/alitari/mockgo/util"

	"github.com/gorilla/mux"
)

type MatchesRequestHandler struct {
	pathPrefix        string
	matchStore        Matchstore
	logger            *logging.LoggerUtil
	basicAuthUsername string
	basicAuthPassword string
}

func NewMatchesRequestHandler(pathPrefix, username, password string, matchStore Matchstore, logger *logging.LoggerUtil) *MatchesRequestHandler {
	configRouter := &MatchesRequestHandler{
		pathPrefix:        pathPrefix,
		matchStore:        matchStore,
		logger:            logger,
		basicAuthUsername: username,
		basicAuthPassword: password,
	}
	return configRouter
}

func (r *MatchesRequestHandler) AddRoutes(router *mux.Router) {

	router.NewRoute().Name("getMatches").Path(r.pathPrefix + "/matches/{endpointId}").Methods(http.MethodGet).
		HandlerFunc(util.RequestMustHave(r.logger, r.basicAuthUsername, r.basicAuthPassword, http.MethodGet, "", "application/json", []string{"endpointId"}, r.handleMatches))
	router.NewRoute().Name("getMatchesCount").Path(r.pathPrefix + "/matchesCount/{endpointId}").Methods(http.MethodGet).
		HandlerFunc(util.RequestMustHave(r.logger, r.basicAuthUsername, r.basicAuthPassword, http.MethodGet, "", "application/json", []string{"endpointId"}, r.handleMatchesCount))
	router.NewRoute().Name("getMismatches").Path(r.pathPrefix + "/mismatches").Methods(http.MethodGet).
		HandlerFunc(util.RequestMustHave(r.logger, r.basicAuthUsername, r.basicAuthPassword, http.MethodGet, "", "application/json", nil, r.handleMismatches))
	router.NewRoute().Name("getMismatchesCount").Path(r.pathPrefix + "/mismatchesCount").Methods(http.MethodGet).
		HandlerFunc(util.RequestMustHave(r.logger, r.basicAuthUsername, r.basicAuthPassword, http.MethodGet, "", "application/json", nil, r.handleMismatchesCount))
	router.NewRoute().Name("deleteMatches").Path(r.pathPrefix + "/matches/{endpointId}").Methods(http.MethodDelete).
		HandlerFunc(util.RequestMustHave(r.logger, r.basicAuthUsername, r.basicAuthPassword, http.MethodDelete, "", "", nil, r.handleDeleteMatches))
	router.NewRoute().Name("deleteMismatches").Path(r.pathPrefix + "/mismatches").Methods(http.MethodDelete).
		HandlerFunc(util.RequestMustHave(r.logger, r.basicAuthUsername, r.basicAuthPassword, http.MethodDelete, "", "", nil, r.handleDeleteMismatches))
	router.NewRoute().Name("health").Path(r.pathPrefix + "/health").Methods(http.MethodGet).
		HandlerFunc(util.RequestMustHave(r.logger, "", "", http.MethodGet, "", "", nil, r.health))
}

func (r *MatchesRequestHandler) health(writer http.ResponseWriter, request *http.Request) {
	writer.WriteHeader(http.StatusOK)
}

func (r *MatchesRequestHandler) handleMatches(writer http.ResponseWriter, request *http.Request) {
	endpointId := mux.Vars(request)["endpointId"]
	if matches, err := r.matchStore.GetMatches(endpointId); err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
	} else {
		util.WriteEntity(writer, matches)
	}
}

func (r *MatchesRequestHandler) handleMatchesCount(writer http.ResponseWriter, request *http.Request) {
	endpointId := mux.Vars(request)["endpointId"]
	if matchesCount, err := r.matchStore.GetMatchesCount(endpointId); err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
	} else {
		util.WriteEntity(writer, matchesCount)
	}
}

func (r *MatchesRequestHandler) handleMismatches(writer http.ResponseWriter, request *http.Request) {
	if mismatches, err := r.matchStore.GetMismatches(); err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
	} else {
		util.WriteEntity(writer, mismatches)
	}
}

func (r *MatchesRequestHandler) handleMismatchesCount(writer http.ResponseWriter, request *http.Request) {
	if mismatchesCount, err := r.matchStore.GetMismatchesCount(); err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
	} else {
		util.WriteEntity(writer, mismatchesCount)
	}
}

func (r *MatchesRequestHandler) handleDeleteMatches(writer http.ResponseWriter, request *http.Request) {
	endpointId := mux.Vars(request)["endpointId"]
	if err := r.matchStore.DeleteMatches(endpointId); err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
	} else {
		writer.WriteHeader(http.StatusOK)
	}
}

func (r *MatchesRequestHandler) handleDeleteMismatches(writer http.ResponseWriter, request *http.Request) {
	if err := r.matchStore.DeleteMismatches(); err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
	} else {
		writer.WriteHeader(http.StatusOK)
	}
}


