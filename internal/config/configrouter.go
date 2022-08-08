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
	mockRouter  *mock.MockRouter
	router      *mux.Router
	server      *http.Server
	port        int
	id          string
	clusterUrls []string
	logger      *utils.Logger
	kvstore     *kvstore.KVStore
}

type EndpointsResponse struct {
	Endpoints []*model.MockEndpoint
}

func NewConfigRouter(mockRouter *mock.MockRouter, port int, clusterUrls []string, kvstore *kvstore.KVStore, logger *utils.Logger) *ConfigRouter {
	configRouter := &ConfigRouter{
		mockRouter:  mockRouter,
		port:        port,
		clusterUrls: clusterUrls,
		id:          uuid.New().String(),
		logger:      logger,
		kvstore:     kvstore,
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
	router.NewRoute().Name("health").Path("/health").Methods(http.MethodGet).HandlerFunc(utils.RequestMustHave(http.MethodGet, "", "", nil, r.health))
	router.NewRoute().Name("serverId").Path("/id").Methods(http.MethodGet).HandlerFunc(utils.RequestMustHave(http.MethodGet, "", "application/text", nil, r.serverId))
	router.NewRoute().Name("endpoints").Path("/endpoints").HandlerFunc(utils.RequestMustHave(http.MethodGet, "", "application/json", nil, r.endpoints))
	router.NewRoute().Name("setKVStore").Path("/kvstore/{key}").Methods(http.MethodPut).HandlerFunc(utils.RequestMustHave(http.MethodPut, "application/json", "", []string{"key"}, r.setKVStore))
	router.NewRoute().Name("getKVStore").Path("/kvstore/{key}").Methods(http.MethodGet).HandlerFunc(utils.RequestMustHave(http.MethodGet, "", "application/json", []string{"key"}, r.getKVStore))
	router.NewRoute().Name("uploadKVStore").Path("/kvstore").Methods(http.MethodPut).HandlerFunc(utils.RequestMustHave(http.MethodPut, "application/json", "", nil, r.uploadKVStore))
	router.NewRoute().Name("downloadKVStore").Path("/kvstore").Methods(http.MethodGet).HandlerFunc(utils.RequestMustHave(http.MethodGet, "", "application/json", nil, r.downloadKVStore))
	r.router = router
	r.server = &http.Server{Addr: ":" + strconv.Itoa(r.port), Handler: router}
}

func (r *ConfigRouter) health(writer http.ResponseWriter, request *http.Request) {
	writer.WriteHeader(http.StatusOK)
}

func (r *ConfigRouter) SyncWithCluster() error {
	if len(r.clusterUrls) == 0 {
		return nil
	}
	httpClient := http.Client{Timeout: time.Duration(1) * time.Second}
	for _, clusterUrl := range r.clusterUrls {
		r.logger.LogWhenVerbose(fmt.Sprintf("syncing with : '%s'...", clusterUrl))
		downloadKvstoreRequest, err := http.NewRequest(http.MethodGet, clusterUrl+"/kvstore", nil)
		if err != nil {
			r.logger.LogWhenVerbose(fmt.Sprintf("can't create request, error: %v", err))
			continue
		}
		downloadKvstoreRequest.Header.Add(headers.Accept, `application/json`)
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
	httpClient := http.Client{Timeout: time.Duration(1) * time.Second}
	for _, clusterUrl := range r.clusterUrls {
		r.logger.LogWhenVerbose(fmt.Sprintf("syncing kvstore key '%s' with : '%s'...", key, clusterUrl))
		setKVstoreRequest, err := http.NewRequest(http.MethodPut, clusterUrl+"/kvstore/"+key, bytes.NewBufferString(value))
		if err != nil {
			r.logger.LogWhenVerbose(fmt.Sprintf("advertise KVStore: can't create request, error: %v", err))
			return err
		}
		setKVstoreRequest.Header.Add(NoAdvertiseHeader, "true")
		setKVstoreRequest.Header.Add(headers.ContentType, "application/json")
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

func (r *ConfigRouter) downloadKVStore(writer http.ResponseWriter, request *http.Request) {
	store := r.kvstore.GetAll()
	storeBytes, err := json.Marshal(store)
	if err != nil {
		http.Error(writer, fmt.Sprintf("Cannot marshall response: %v", err), http.StatusInternalServerError)
		return
	}
	_, err = io.WriteString(writer, string(storeBytes))
	if err != nil {
		http.Error(writer, fmt.Sprintf("Cannot write response: %v", err), http.StatusInternalServerError)
		return
	}
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
	resp, err := json.MarshalIndent(val, "", "    ")
	if err != nil {
		http.Error(writer, fmt.Sprintf("Cannot marshall response: %v", err), http.StatusInternalServerError)
		return
	}
	_, err = io.WriteString(writer, string(resp))
	if err != nil {
		http.Error(writer, fmt.Sprintf("Cannot write response: %v", err), http.StatusInternalServerError)
		return
	}
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
