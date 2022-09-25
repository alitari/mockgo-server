package matches

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/alitari/mockgo/logging"
	"github.com/alitari/mockgo/util"

	"github.com/gorilla/mux"
)

type MatchesRequestHandler struct {
	matchStore        Matchstore
	logger            *logging.LoggerUtil
	basicAuthUsername string
	basicAuthPassword string
}

func NewMatchesRequestHandler(username, password string, matchStore Matchstore, logger *logging.LoggerUtil) *MatchesRequestHandler {
	configRouter := &MatchesRequestHandler{
		matchStore:        matchStore,
		logger:            logger,
		basicAuthUsername: username,
		basicAuthPassword: password,
	}
	return configRouter
}

func (r *MatchesRequestHandler) AddRoutes(router *mux.Router) {
	router.NewRoute().Name("health").Path("/health").Methods(http.MethodGet).
		HandlerFunc(util.RequestMustHave(r.logger, "", "", http.MethodGet, "", "", nil, r.health))
	router.NewRoute().Name("getMatches").Path("/matches/{endpointId}").Methods(http.MethodGet).
		HandlerFunc(util.RequestMustHave(r.logger, r.basicAuthUsername, r.basicAuthPassword, http.MethodGet, "", "application/json", []string{"endpointId"}, r.handleMatches))
	router.NewRoute().Name("getMismatches").Path("/mismatches").Methods(http.MethodGet).
		HandlerFunc(util.RequestMustHave(r.logger, r.basicAuthUsername, r.basicAuthPassword, http.MethodGet, "", "application/json", nil, r.handleMismatches))
	router.NewRoute().Name("addMatches").Path("/addmatches/").Methods(http.MethodPost).
		HandlerFunc(util.RequestMustHave(r.logger, r.basicAuthUsername, r.basicAuthPassword, http.MethodPost, "application/json", "", nil, r.handleAddMatches))
	router.NewRoute().Name("addMismatches").Path("/addmismatches/").Methods(http.MethodPost).
		HandlerFunc(util.RequestMustHave(r.logger, r.basicAuthUsername, r.basicAuthPassword, http.MethodPost, "application/json", "", nil, r.handleAddMismatches))
	router.NewRoute().Name("deleteMatches").Path("/matches").Methods(http.MethodDelete).
		HandlerFunc(util.RequestMustHave(r.logger, r.basicAuthUsername, r.basicAuthPassword, http.MethodDelete, "", "", nil, r.handleDeleteMatches))
	router.NewRoute().Name("deleteMismatches").Path("/mismatches").Methods(http.MethodDelete).
		HandlerFunc(util.RequestMustHave(r.logger, r.basicAuthUsername, r.basicAuthPassword, http.MethodDelete, "", "", nil, r.handleDeleteMismatches))
	// router.NewRoute().Name("transferMatches").Path("/transfermatches").Methods(http.MethodGet).HandlerFunc(util.RequestMustHave(r.logger, "", "", http.MethodGet, "", "", nil, r.transferMatchesHandler))
}

func (r *MatchesRequestHandler) health(writer http.ResponseWriter, request *http.Request) {
	writer.WriteHeader(http.StatusOK)
}

func (r *MatchesRequestHandler) handleMatches(writer http.ResponseWriter, request *http.Request) {
	endpointId := mux.Vars(request)["endpointId"]
	if r.matchStore.HasMatchesCountOnly() {
		if matchesCount, err := r.matchStore.GetMatchesCount(endpointId); err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
		} else {
			util.WriteEntity(writer, matchesCount)
		}
	} else {
		if matches, err := r.matchStore.GetMatches(endpointId); err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
		} else {
			util.WriteEntity(writer, matches)
		}
	}
}

func (r *MatchesRequestHandler) handleMismatches(writer http.ResponseWriter, request *http.Request) {
	if r.matchStore.HasMismatchesCountOnly() {
		if mismatchesCount, err := r.matchStore.GetMismatchesCount(); err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
		} else {
			util.WriteEntity(writer, mismatchesCount)
		}
	} else {
		if mismatches, err := r.matchStore.GetMismatches(); err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
		} else {
			util.WriteEntity(writer, mismatches)
		}
	}
}

func (r *MatchesRequestHandler) handleDeleteMatches(writer http.ResponseWriter, request *http.Request) {
	if err := r.matchStore.DeleteMatches(); err != nil {
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

func (r *MatchesRequestHandler) handleAddMatches(writer http.ResponseWriter, request *http.Request) {
	body, err := io.ReadAll(request.Body)
	if err != nil {
		http.Error(writer, "Problem reading request body: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if r.matchStore.HasMatchesCountOnly() {
		var matchData map[string]int64
		err = json.Unmarshal(body, &matchData)
		if err != nil {
			http.Error(writer, "Problem marshalling matches response body: "+err.Error(), http.StatusInternalServerError)
			return
		}
		if err := r.matchStore.AddMatchesCount(matchData); err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		} else {
			r.logger.LogWhenVerbose(fmt.Sprintf("added matchesCount from %d endpoints sucessfully", len(matchData)))
		}

	} else {
		var matchData map[string][]*Match
		err = json.Unmarshal(body, &matchData)
		if err != nil {
			http.Error(writer, "Problem marshalling matches response body: "+err.Error(), http.StatusInternalServerError)
			return
		}
		if err := r.matchStore.AddMatches(matchData); err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		} else {
			r.logger.LogWhenVerbose(fmt.Sprintf("added matches from %d endpoints sucessfully", len(matchData)))
		}
	}
	writer.WriteHeader(http.StatusOK)
}

func (r *MatchesRequestHandler) handleAddMismatches(writer http.ResponseWriter, request *http.Request) {
	body, err := io.ReadAll(request.Body)
	if err != nil {
		http.Error(writer, "Problem reading request body: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if r.matchStore.HasMismatchesCountOnly() {
		var matchData int64
		err = json.Unmarshal(body, &matchData)
		if err != nil {
			http.Error(writer, "Problem marshalling mismatches response body: "+err.Error(), http.StatusInternalServerError)
			return
		}
		if err := r.matchStore.AddMismatchesCount(matchData); err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		r.logger.LogWhenVerbose(fmt.Sprintf("added %d mismatchesCount sucessfully", matchData))
	} else {
		var matchData []*Mismatch
		err = json.Unmarshal(body, &matchData)
		if err != nil {
			http.Error(writer, "Problem marshalling mismatches response body: "+err.Error(), http.StatusInternalServerError)
			return
		}
		if err := r.matchStore.AddMismatches(matchData); err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		r.logger.LogWhenVerbose(fmt.Sprintf("added %d mismatches sucessfully", len(matchData)))
	}
	writer.WriteHeader(http.StatusOK)
}
