package matches

import (
	"net/http"

	"github.com/alitari/mockgo-server/mockgo/logging"
	"github.com/alitari/mockgo-server/mockgo/util"

	"github.com/gorilla/mux"
)

type RequestHandler struct {
	pathPrefix        string
	matchStore        Matchstore
	logger            *logging.LoggerUtil
	basicAuthUsername string
	basicAuthPassword string
}

func NewRequestHandler(pathPrefix, username, password string, matchStore Matchstore, logger *logging.LoggerUtil) *RequestHandler {
	configRouter := &RequestHandler{
		pathPrefix:        pathPrefix,
		matchStore:        matchStore,
		logger:            logger,
		basicAuthUsername: username,
		basicAuthPassword: password,
	}
	return configRouter
}

func (r *RequestHandler) AddRoutes(router *mux.Router) {
	router.NewRoute().Name("health").Path(r.pathPrefix + "/health").Methods(http.MethodGet).
		HandlerFunc(r.health)
	router.NewRoute().Name("getMatches").Path(r.pathPrefix + "/matches/{endpointId}").Methods(http.MethodGet).
		HandlerFunc(util.BasicAuthRequest(r.basicAuthUsername, r.basicAuthPassword, util.JSONAcceptRequest(util.PathParamRequest([]string{"endpointId"}, r.handleGetMatches))))
	router.NewRoute().Name("getMatchesCount").Path(r.pathPrefix + "/matchesCount/{endpointId}").Methods(http.MethodGet).
		HandlerFunc(util.BasicAuthRequest(r.basicAuthUsername, r.basicAuthPassword, util.JSONAcceptRequest(util.PathParamRequest([]string{"endpointId"}, r.handleGetMatchesCount))))
	router.NewRoute().Name("getMismatches").Path(r.pathPrefix + "/mismatches").Methods(http.MethodGet).
		HandlerFunc(util.BasicAuthRequest(r.basicAuthUsername, r.basicAuthPassword, util.JSONAcceptRequest(r.handleGetMismatches)))
	router.NewRoute().Name("getMismatchesCount").Path(r.pathPrefix + "/mismatchesCount").Methods(http.MethodGet).
		HandlerFunc(util.BasicAuthRequest(r.basicAuthUsername, r.basicAuthPassword, util.JSONAcceptRequest(r.handleGetMismatchesCount)))
	router.NewRoute().Name("deleteMatches").Path(r.pathPrefix + "/matches/{endpointId}").Methods(http.MethodDelete).
		HandlerFunc(util.BasicAuthRequest(r.basicAuthUsername, r.basicAuthPassword, util.PathParamRequest([]string{"endpointId"}, r.handleDeleteMatches)))
	router.NewRoute().Name("deleteMismatches").Path(r.pathPrefix + "/mismatches").Methods(http.MethodDelete).
		HandlerFunc(util.BasicAuthRequest(r.basicAuthUsername, r.basicAuthPassword, r.handleDeleteMismatches))

	// 	router.NewRoute().Name("getMatches").Path(r.pathPrefix + "/matches/{endpointId}").Methods(http.MethodGet).
	// 	HandlerFunc(util.RequestMustHave(r.logger, r.basicAuthUsername, r.basicAuthPassword, http.MethodGet, "", "application/json", []string{"endpointId"}, r.handleGetMatches))
	// router.NewRoute().Name("getMatchesCount").Path(r.pathPrefix + "/matchesCount/{endpointId}").Methods(http.MethodGet).
	// 	HandlerFunc(util.RequestMustHave(r.logger, r.basicAuthUsername, r.basicAuthPassword, http.MethodGet, "", "application/json", []string{"endpointId"}, r.handleGetMatchesCount))
	// router.NewRoute().Name("getMismatches").Path(r.pathPrefix + "/mismatches").Methods(http.MethodGet).
	// 	HandlerFunc(util.RequestMustHave(r.logger, r.basicAuthUsername, r.basicAuthPassword, http.MethodGet, "", "application/json", nil, r.handleGetMismatches))
	// router.NewRoute().Name("getMismatchesCount").Path(r.pathPrefix + "/mismatchesCount").Methods(http.MethodGet).
	// 	HandlerFunc(util.RequestMustHave(r.logger, r.basicAuthUsername, r.basicAuthPassword, http.MethodGet, "", "application/json", nil, r.handleGetMismatchesCount))
	// router.NewRoute().Name("deleteMatches").Path(r.pathPrefix + "/matches/{endpointId}").Methods(http.MethodDelete).
	// 	HandlerFunc(util.RequestMustHave(r.logger, r.basicAuthUsername, r.basicAuthPassword, http.MethodDelete, "", "", nil, r.handleDeleteMatches))
	// router.NewRoute().Name("deleteMismatches").Path(r.pathPrefix + "/mismatches").Methods(http.MethodDelete).
	// 	HandlerFunc(util.RequestMustHave(r.logger, r.basicAuthUsername, r.basicAuthPassword, http.MethodDelete, "", "", nil, r.handleDeleteMismatches))
	// router.NewRoute().Name("health").Path(r.pathPrefix + "/health").Methods(http.MethodGet).
	// 	HandlerFunc(util.RequestMustHave(r.logger, "", "", http.MethodGet, "", "", nil, r.health))

}

func (r *RequestHandler) health(writer http.ResponseWriter, request *http.Request) {
	writer.WriteHeader(http.StatusOK)
}

func (r *RequestHandler) handleGetMatches(writer http.ResponseWriter, request *http.Request) {
	endpointID := mux.Vars(request)["endpointId"]
	if matches, err := r.matchStore.GetMatches(endpointID); err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
	} else {
		util.WriteEntity(writer, matches)
	}
}

func (r *RequestHandler) handleGetMatchesCount(writer http.ResponseWriter, request *http.Request) {
	endpointID := mux.Vars(request)["endpointId"]
	if matchesCount, err := r.matchStore.GetMatchesCount(endpointID); err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
	} else {
		util.WriteEntity(writer, matchesCount)
	}
}

func (r *RequestHandler) handleGetMismatches(writer http.ResponseWriter, request *http.Request) {
	if mismatches, err := r.matchStore.GetMismatches(); err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
	} else {
		util.WriteEntity(writer, mismatches)
	}
}

func (r *RequestHandler) handleGetMismatchesCount(writer http.ResponseWriter, request *http.Request) {
	if mismatchesCount, err := r.matchStore.GetMismatchesCount(); err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
	} else {
		util.WriteEntity(writer, mismatchesCount)
	}
}

func (r *RequestHandler) handleDeleteMatches(writer http.ResponseWriter, request *http.Request) {
	endpointID := mux.Vars(request)["endpointId"]
	if err := r.matchStore.DeleteMatches(endpointID); err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
	} else {
		writer.WriteHeader(http.StatusOK)
	}
}

func (r *RequestHandler) handleDeleteMismatches(writer http.ResponseWriter, request *http.Request) {
	if err := r.matchStore.DeleteMismatches(); err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
	} else {
		writer.WriteHeader(http.StatusOK)
	}
}
