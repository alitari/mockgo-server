package matches

import (
	"go.uber.org/zap"
	"net/http"

	"github.com/alitari/mockgo-server/mockgo/util"

	"github.com/gorilla/mux"
)

/*
RequestHandler implements an http API to access the Matchstore
*/
type RequestHandler struct {
	pathPrefix        string
	matchStore        Matchstore
	logger            *zap.Logger
	basicAuthUsername string
	basicAuthPassword string
}

/*
NewRequestHandler creates an instance of RequestHandler
*/
func NewRequestHandler(pathPrefix string, matchStore Matchstore, logLevel string) *RequestHandler {
	configRouter := &RequestHandler{
		pathPrefix: pathPrefix,
		matchStore: matchStore,
		logger:     util.CreateLogger(logLevel),
	}
	return configRouter
}

/*
AddRoutes adds mux.Routes for the http API to a given mux.Router
*/
func (r *RequestHandler) AddRoutes(router *mux.Router) {

	router.NewRoute().Name("getMatches").Path(r.pathPrefix + "/matches/{endpointId}").Methods(http.MethodGet).
		HandlerFunc(util.JSONAcceptRequest(util.PathParamRequest([]string{"endpointId"}, r.handleGetMatches)))
	router.NewRoute().Name("getMatchesCount").Path(r.pathPrefix + "/matchesCount/{endpointId}").Methods(http.MethodGet).
		HandlerFunc(util.JSONAcceptRequest(util.PathParamRequest([]string{"endpointId"}, r.handleGetMatchesCount)))
	router.NewRoute().Name("getMismatches").Path(r.pathPrefix + "/mismatches").Methods(http.MethodGet).
		HandlerFunc(util.JSONAcceptRequest(r.handleGetMismatches))
	router.NewRoute().Name("getMismatchesCount").Path(r.pathPrefix + "/mismatchesCount").Methods(http.MethodGet).
		HandlerFunc(util.JSONAcceptRequest(r.handleGetMismatchesCount))
	router.NewRoute().Name("deleteMatches").Path(r.pathPrefix + "/matches/{endpointId}").Methods(http.MethodDelete).
		HandlerFunc(util.PathParamRequest([]string{"endpointId"}, r.handleDeleteMatches))
	router.NewRoute().Name("deleteMismatches").Path(r.pathPrefix + "/mismatches").Methods(http.MethodDelete).
		HandlerFunc(r.handleDeleteMismatches)
}

func (r *RequestHandler) handleHealth(writer http.ResponseWriter, request *http.Request) {
	writer.WriteHeader(http.StatusOK)
}

func (r *RequestHandler) handleGetMatches(writer http.ResponseWriter, request *http.Request) {
	endpointID := mux.Vars(request)["endpointId"]
	if matches, err := r.matchStore.GetMatches(endpointID); err != nil {
		r.logger.Error("Error getting matches", zap.Error(err))
		http.Error(writer, err.Error(), http.StatusInternalServerError)
	} else {
		util.WriteEntity(writer, matches)
	}
}

func (r *RequestHandler) handleGetMatchesCount(writer http.ResponseWriter, request *http.Request) {
	endpointID := mux.Vars(request)["endpointId"]
	if matchesCount, err := r.matchStore.GetMatchesCount(endpointID); err != nil {
		r.logger.Error("Error getting matches count", zap.Error(err))
		http.Error(writer, err.Error(), http.StatusInternalServerError)
	} else {
		util.WriteEntity(writer, matchesCount)
	}
}

func (r *RequestHandler) handleGetMismatches(writer http.ResponseWriter, request *http.Request) {
	if mismatches, err := r.matchStore.GetMismatches(); err != nil {
		r.logger.Error("Error getting mismatches", zap.Error(err))
		http.Error(writer, err.Error(), http.StatusInternalServerError)
	} else {
		util.WriteEntity(writer, mismatches)
	}
}

func (r *RequestHandler) handleGetMismatchesCount(writer http.ResponseWriter, request *http.Request) {
	if mismatchesCount, err := r.matchStore.GetMismatchesCount(); err != nil {
		r.logger.Error("Error getting mismatches count", zap.Error(err))
		http.Error(writer, err.Error(), http.StatusInternalServerError)
	} else {
		util.WriteEntity(writer, mismatchesCount)
	}
}

func (r *RequestHandler) handleDeleteMatches(writer http.ResponseWriter, request *http.Request) {
	endpointID := mux.Vars(request)["endpointId"]
	if err := r.matchStore.DeleteMatches(endpointID); err != nil {
		r.logger.Error("Error deleting matches", zap.Error(err))
		http.Error(writer, err.Error(), http.StatusInternalServerError)
	} else {
		writer.WriteHeader(http.StatusOK)
	}
}

func (r *RequestHandler) handleDeleteMismatches(writer http.ResponseWriter, request *http.Request) {
	if err := r.matchStore.DeleteMismatches(); err != nil {
		r.logger.Error("Error deleting mismatches", zap.Error(err))
		http.Error(writer, err.Error(), http.StatusInternalServerError)
	} else {
		writer.WriteHeader(http.StatusOK)
	}
}
