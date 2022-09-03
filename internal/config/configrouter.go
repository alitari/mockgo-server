package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/alitari/mockgo-server/internal/kvstore"
	"github.com/alitari/mockgo-server/internal/mock"
	"github.com/alitari/mockgo-server/internal/model"
	"github.com/alitari/mockgo-server/internal/utils"
	"github.com/google/uuid"

	"github.com/go-http-utils/headers"
	"github.com/gorilla/mux"
)

const NoAdvertiseHeader = "No-advertise"

type ConfigRouter struct {
	mockRouter          *mock.MockRouter
	router              *mux.Router
	server              *http.Server
	port                int
	id                  string
	clusterUrls         []string
	logger              *utils.Logger
	kvstore             *kvstore.KVStore
	httpClientTimeout   time.Duration
	basicAuthUsername   string
	basicAuthPassword   string
	transferringMatches bool
}

type EndpointsResponse struct {
	Endpoints []*model.MockEndpoint
}

func NewConfigRouter(username, password string, mockRouter *mock.MockRouter, port int, clusterUrls []string, kvstore *kvstore.KVStore, logger *utils.Logger) *ConfigRouter {
	configRouter := &ConfigRouter{
		mockRouter:          mockRouter,
		port:                port,
		clusterUrls:         clusterUrls,
		id:                  uuid.New().String(),
		logger:              logger,
		kvstore:             kvstore,
		httpClientTimeout:   1 * time.Second,
		basicAuthUsername:   username,
		basicAuthPassword:   password,
		transferringMatches: false,
	}
	configRouter.newRouter()
	return configRouter
}

func (r *ConfigRouter) Name() string {
	return "Configrouter"
}

func (r *ConfigRouter) Router() *mux.Router {
	return r.router
}

func (r *ConfigRouter) Server() *http.Server {
	return r.server
}

func (r *ConfigRouter) Port() int {
	return r.port
}

func (r *ConfigRouter) Logger() *utils.Logger {
	return r.logger
}

func (r *ConfigRouter) newRouter() {
	router := mux.NewRouter()
	router.NewRoute().Name("health").Path("/health").Methods(http.MethodGet).HandlerFunc(utils.RequestMustHave("", "", http.MethodGet, "", "", nil, r.health))
	router.NewRoute().Name("serverId").Path("/id").Methods(http.MethodGet).HandlerFunc(utils.RequestMustHave(r.basicAuthUsername, r.basicAuthPassword, http.MethodGet, "", "application/text", nil, r.serverId))
	router.NewRoute().Name("endpoints").Path("/endpoints").HandlerFunc(utils.RequestMustHave(r.basicAuthUsername, r.basicAuthPassword, http.MethodGet, "", "application/json", nil, r.endpoints))
	router.NewRoute().Name("setKVStore").Path("/kvstore/{key}").Methods(http.MethodPut).HandlerFunc(utils.RequestMustHave(r.basicAuthUsername, r.basicAuthPassword, http.MethodPut, "application/json", "", []string{"key"}, r.setKVStore))
	router.NewRoute().Name("getKVStore").Path("/kvstore/{key}").Methods(http.MethodGet).HandlerFunc(utils.RequestMustHave(r.basicAuthUsername, r.basicAuthPassword, http.MethodGet, "", "application/json", []string{"key"}, r.getKVStore))
	router.NewRoute().Name("uploadKVStore").Path("/kvstore").Methods(http.MethodPut).HandlerFunc(utils.RequestMustHave(r.basicAuthUsername, r.basicAuthPassword, http.MethodPut, "application/json", "", nil, r.uploadKVStore))
	router.NewRoute().Name("downloadKVStore").Path("/kvstore").Methods(http.MethodGet).HandlerFunc(utils.RequestMustHave(r.basicAuthUsername, r.basicAuthPassword, http.MethodGet, "", "application/json", nil, r.downloadKVStore))
	router.NewRoute().Name("getMatches").Path("/matches").Methods(http.MethodGet).HandlerFunc(utils.RequestMustHave(r.basicAuthUsername, r.basicAuthPassword, http.MethodGet, "", "application/json", nil, r.getMatchesFromAll))
	router.NewRoute().Name("addMatches").Path("/addmatches").Methods(http.MethodPost).HandlerFunc(utils.RequestMustHave(r.basicAuthUsername, r.basicAuthPassword, http.MethodPost, "application/json", "", nil, r.addMatches))
	router.NewRoute().Name("deleteMatches").Path("/matches").Methods(http.MethodDelete).HandlerFunc(utils.RequestMustHave(r.basicAuthUsername, r.basicAuthPassword, http.MethodDelete, "", "", nil, r.deleteMatchesFromAll))
	router.NewRoute().Name("transferMatches").Path("/transfermatches").Methods(http.MethodGet).HandlerFunc(utils.RequestMustHave("", "", http.MethodGet, "", "", nil, r.transferMatches))
	r.router = router
	r.server = &http.Server{Addr: ":" + strconv.Itoa(r.port), Handler: router}
}

func (r *ConfigRouter) health(writer http.ResponseWriter, request *http.Request) {
	writer.WriteHeader(http.StatusOK)
}

func (r *ConfigRouter) createHttpClient() http.Client {
	httpClient := http.Client{Timeout: time.Duration(1) * time.Second}
	return httpClient
}

func (r *ConfigRouter) SyncKvstoreWithCluster() error {
	if len(r.clusterUrls) == 0 {
		return nil
	}
	httpClient := r.createHttpClient()
	for _, clusterUrl := range r.clusterUrls {
		r.logger.LogWhenVerbose(fmt.Sprintf("syncing with : '%s'...", clusterUrl))
		downloadKvstoreRequest, err := http.NewRequest(http.MethodGet, clusterUrl+"/kvstore", nil)
		if err != nil {
			r.logger.LogWhenVerbose(fmt.Sprintf("can't create request, error: %v", err))
			continue
		}
		downloadKvstoreRequest.Header.Add(headers.Accept, `application/json`)
		downloadKvstoreRequest.Header.Add(headers.Authorization, utils.BasicAuth(r.basicAuthUsername, r.basicAuthPassword))
		downloadKvstoreResponse, err := httpClient.Do(downloadKvstoreRequest)
		if err != nil {
			r.logger.LogWhenVerbose(fmt.Sprintf("cluster node can't process request, answered with error: %v", err))
			continue
		}
		defer downloadKvstoreResponse.Body.Close()
		if downloadKvstoreResponse.StatusCode != http.StatusOK {
			r.logger.LogWhenVerbose(fmt.Sprintf("cluster node can't process request, answered with status: %v\n", downloadKvstoreResponse.StatusCode))
			continue
		}

		storeBytes, err := ioutil.ReadAll(downloadKvstoreResponse.Body)
		if err != nil {
			r.logger.LogAlways(fmt.Sprintf("(ERROR) reading response from cluster url failed: %v ", err))
			return err
		}
		err = r.kvstore.PutAllJson(string(storeBytes))
		if err != nil {
			r.logger.LogAlways(fmt.Sprintf("(ERROR) creating new kvstore downloaded from clusterurl '%s' failed: %v ", clusterUrl, err))
			return err
		}

		r.logger.LogWhenVerbose(fmt.Sprintf("syncing completed, kvstore successfully downloaded from clusterurl '%s' ", clusterUrl))
		return nil
	}
	return nil
}

func (r *ConfigRouter) advertiseKVStore(key, value string) error {
	httpClient := r.createHttpClient()
	for _, clusterUrl := range r.clusterUrls {
		r.logger.LogWhenVerbose(fmt.Sprintf("syncing kvstore key '%s' with : '%s'...", key, clusterUrl))
		setKVstoreRequest, err := http.NewRequest(http.MethodPut, clusterUrl+"/kvstore/"+key, bytes.NewBufferString(value))
		if err != nil {
			r.logger.LogWhenVerbose(fmt.Sprintf("advertise KVStore: can't create request, error: %v", err))
			return err
		}
		setKVstoreRequest.Header.Add(NoAdvertiseHeader, "true")
		setKVstoreRequest.Header.Add(headers.ContentType, "application/json")
		setKVstoreRequest.Header.Add(headers.Authorization, utils.BasicAuth(r.basicAuthUsername, r.basicAuthPassword))
		setKVstoreResponse, err := httpClient.Do(setKVstoreRequest)
		if err != nil {
			r.logger.LogWhenVerbose(fmt.Sprintf("advertise KVStore: cluster node can't process request, answered with error: %v", err))
			return err
		}
		defer setKVstoreResponse.Body.Close()
		if setKVstoreResponse.StatusCode != http.StatusNoContent {
			r.logger.LogWhenVerbose(fmt.Sprintf("advertise KVStore: cluster node can't process request, answered with status: %v\n", setKVstoreResponse.StatusCode))
			return err
		}
		r.logger.LogWhenVerbose("synced successfully!")
	}
	return nil
}

func (r *ConfigRouter) serverId(writer http.ResponseWriter, request *http.Request) {
	_, err := io.WriteString(writer, r.id)
	if err != nil {
		http.Error(writer, fmt.Sprintf("Cannot write response: %v", err), http.StatusInternalServerError)
		return
	}
	writer.WriteHeader(http.StatusOK)
}

func (r *ConfigRouter) getMatches(writer http.ResponseWriter, request *http.Request) {
	utils.WriteEntity(writer, r.mockRouter.Matches)
}

func (r *ConfigRouter) deleteMatches() {
	r.mockRouter.Matches = make(map[string][]*model.Match)
}

func (r *ConfigRouter) downloadKVStore(writer http.ResponseWriter, request *http.Request) {
	utils.WriteEntity(writer, r.kvstore.GetAll())
}

func (r *ConfigRouter) uploadKVStore(writer http.ResponseWriter, request *http.Request) {
	body, err := io.ReadAll(request.Body)
	if err != nil {
		http.Error(writer, "Problem reading request body: "+err.Error(), http.StatusInternalServerError)
		return
	}
	err = r.kvstore.PutAllJson(string(body))
	if err != nil {
		http.Error(writer, "Problem creating kvstore: "+err.Error(), http.StatusInternalServerError)
		return
	}
	writer.WriteHeader(http.StatusNoContent)
}

func (r *ConfigRouter) getKVStore(writer http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	key := vars["key"]
	val, err := r.kvstore.Get(key)
	if err != nil {
		http.Error(writer, "Problem with getting kvstore value "+err.Error(), http.StatusInternalServerError)
		return
	}
	utils.WriteEntity(writer, val)
}

func (r *ConfigRouter) getMatchesFromAll(writer http.ResponseWriter, request *http.Request) {
	if len(r.clusterUrls) == 0 || request.Header.Get(NoAdvertiseHeader) == "true" {
		r.getMatches(writer, request)
	} else {
		allMatches := make(map[string][]*model.Match)
		httpClient := r.createHttpClient()
		for _, clusterUrl := range r.clusterUrls {
			matchUrl := clusterUrl + "/matches"
			r.logger.LogWhenVerbose(fmt.Sprintf("calling '%s'...", matchUrl))
			matchesRequest, err := http.NewRequest(http.MethodGet, matchUrl, nil)
			if err != nil {
				http.Error(writer, fmt.Sprintf("Cannot create match request : %v", err), http.StatusInternalServerError)
				return
			}
			matchesRequest.Header.Add(NoAdvertiseHeader, "true")
			matchesRequest.Header.Add(headers.Accept, "application/json")
			matchesRequest.Header.Add(headers.Authorization, utils.BasicAuth(r.basicAuthUsername, r.basicAuthPassword))
			matchesResponse, err := httpClient.Do(matchesRequest)
			if err != nil {
				http.Error(writer, fmt.Sprintf("Cannot send match request : %v", err), http.StatusInternalServerError)
				return
			}
			if matchesResponse.StatusCode != http.StatusOK {
				http.Error(writer, fmt.Sprintf("Unexpected response from %s : %d", clusterUrl, matchesResponse.StatusCode), http.StatusInternalServerError)
				return
			}
			body, err := io.ReadAll(matchesResponse.Body)
			if err != nil {
				http.Error(writer, "Problem reading response body: "+err.Error(), http.StatusInternalServerError)
				return
			}
			matchesResponse.Body.Close()
			bodyData := &map[string][]*model.Match{}
			err = json.Unmarshal(body, bodyData)
			if err != nil {
				http.Error(writer, "Problem marshalling response body: "+err.Error(), http.StatusInternalServerError)
				return
			}
			for k, v := range *bodyData {
				allMatches[k] = append(allMatches[k], v...)
			}
		}
		utils.WriteEntity(writer, allMatches)
	}
}

func (r *ConfigRouter) deleteMatchesFromAll(writer http.ResponseWriter, request *http.Request) {
	if len(r.clusterUrls) == 0 || request.Header.Get(NoAdvertiseHeader) == "true" {
		r.deleteMatches()
	} else {
		httpClient := r.createHttpClient()
		for _, clusterUrl := range r.clusterUrls {
			matchUrl := clusterUrl + "/matches"
			r.logger.LogWhenVerbose(fmt.Sprintf("calling '%s'...", matchUrl))
			matchesDeleteRequest, err := http.NewRequest(http.MethodDelete, matchUrl, nil)
			if err != nil {
				http.Error(writer, fmt.Sprintf("Cannot create match request : %v", err), http.StatusInternalServerError)
				return
			}
			matchesDeleteRequest.Header.Add(NoAdvertiseHeader, "true")
			matchesDeleteRequest.Header.Add(headers.Authorization, utils.BasicAuth(r.basicAuthUsername, r.basicAuthPassword))
			matchesResponse, err := httpClient.Do(matchesDeleteRequest)
			if err != nil {
				http.Error(writer, fmt.Sprintf("Cannot send match request : %v", err), http.StatusInternalServerError)
				return
			}
			if matchesResponse.StatusCode != http.StatusOK {
				http.Error(writer, fmt.Sprintf("Unexpected response from %s : %d", clusterUrl, matchesResponse.StatusCode), http.StatusInternalServerError)
				return
			}
		}
		writer.WriteHeader(http.StatusOK)
	}
}

func (r *ConfigRouter) addMatches(writer http.ResponseWriter, request *http.Request) {
	r.logger.LogWhenVerbose("adding matches...")
	if r.transferringMatches {
		http.Error(writer, "Already transferring", http.StatusLocked)
		return
	}
	body, err := io.ReadAll(request.Body)
	if err != nil {
		http.Error(writer, "Problem reading request body: "+err.Error(), http.StatusInternalServerError)
		return
	}
	matchData := &map[string][]*model.Match{}
	err = json.Unmarshal(body, matchData)
	if err != nil {
		http.Error(writer, "Problem marshalling response body: "+err.Error(), http.StatusInternalServerError)
		return
	}
	currentMatches := r.mockRouter.Matches
	for k, v := range *matchData {
		currentMatches[k] = append(currentMatches[k], v...)
	}
	r.logger.LogWhenVerbose(fmt.Sprintf("added %d matches sucessfully", r.matchesCount(*matchData)))
	writer.WriteHeader(http.StatusOK)
}

func (r *ConfigRouter) transferMatches(writer http.ResponseWriter, request *http.Request) {
	r.logger.LogWhenVerbose("transfer matches...")
	r.transferringMatches = true
	defer func() {
		r.transferringMatches = false
	}()
	matches, err := json.Marshal(r.mockRouter.Matches)
	if err != nil {
		http.Error(writer, "Problem marshalling matches: "+err.Error(), http.StatusInternalServerError)
		return
	}
	httpClient := r.createHttpClient()
	for _, clusterUrl := range r.clusterUrls {
		r.logger.LogWhenVerbose(fmt.Sprintf("adding matches to : '%s'...", clusterUrl))
		addMatchesRequest, err := http.NewRequest(http.MethodPost, clusterUrl+"/addmatches", bytes.NewBuffer(matches))
		if err != nil {
			r.logger.LogWhenVerbose(fmt.Sprintf("can't create request, error: %v", err))
			continue
		}
		addMatchesRequest.Header.Add(headers.ContentType, `application/json`)
		addMatchesRequest.Header.Add(headers.Authorization, utils.BasicAuth(r.basicAuthUsername, r.basicAuthPassword))
		addMatchesResponse, err := httpClient.Do(addMatchesRequest)
		if err != nil {
			r.logger.LogWhenVerbose(fmt.Sprintf("can't add matches, to: %s ,error: %v", clusterUrl, err))
			continue
		}
		if addMatchesResponse.StatusCode == http.StatusOK {
			r.logger.LogWhenVerbose(fmt.Sprintf("%d matches successfully transferred to: %s", r.matchesCount(r.mockRouter.Matches), clusterUrl))
			r.deleteMatches()
			writer.WriteHeader(http.StatusOK)
			return
		} else {
			r.logger.LogWhenVerbose(fmt.Sprintf("response: %d ,matches not transferred", addMatchesResponse.StatusCode))
		}
	}
	http.Error(writer, "Couldn't transfer matches.", http.StatusGone)
}

func (r *ConfigRouter) matchesCount( matches map[string][]*model.Match) int {
	sum := 0
	for _, matches := range matches {
		sum = sum + len(matches)
	}
	return sum
}

func (r *ConfigRouter) setKVStore(writer http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	key := vars["key"]
	body, err := io.ReadAll(request.Body)
	if err != nil {
		http.Error(writer, "Problem reading request body: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if len(r.clusterUrls) == 0 || request.Header.Get(NoAdvertiseHeader) == "true" {
		err = r.kvstore.Put(key, string(body))
		if err != nil {
			http.Error(writer, "Problem with kvstore value, ( is it valid JSON?): "+err.Error(), http.StatusBadRequest)
			return
		}
		writer.WriteHeader(http.StatusNoContent)
	} else {
		err := r.advertiseKVStore(key, string(body))
		if err != nil {
			http.Error(writer, "Problem advertising kvstore value : "+err.Error(), http.StatusInternalServerError)
			return
		}
		writer.WriteHeader(http.StatusNoContent)
	}
}

func (r *ConfigRouter) endpoints(writer http.ResponseWriter, request *http.Request) {
	endPointResponse := &EndpointsResponse{Endpoints: r.mockRouter.AllEndpoints()}
	writer.Header().Set(headers.ContentType, "application/json")
	resp, err := json.MarshalIndent(endPointResponse, "", "    ")
	if err != nil {
		io.WriteString(writer, fmt.Sprintf("Cannot marshall response: %v", err))
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	r.logger.LogWhenDebugRR(fmt.Sprintf("%v", endPointResponse))
	_, err = io.WriteString(writer, string(resp))
	if err != nil {
		io.WriteString(writer, fmt.Sprintf("Cannot write response: %v", err))
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}
}
