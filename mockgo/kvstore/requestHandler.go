package kvstore

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/alitari/mockgo/logging"
	"github.com/alitari/mockgo/util"

	"github.com/gorilla/mux"
)

type KVStoreRequestHandler struct {
	pathPrefix        string
	logger            *logging.LoggerUtil
	kvstore           *KVStoreJSON
	basicAuthUsername string
	basicAuthPassword string
}

type AddKVStoreRequest struct {
	Path  string `json:"path"`
	Value string `json:"value"`
}

type RemoveKVStoreRequest struct {
	Path string `json:"path"`
}

func NewKVStoreRequestHandler(pathPrefix, username, password string, kvstore *KVStoreJSON, logger *logging.LoggerUtil) *KVStoreRequestHandler {
	configRouter := &KVStoreRequestHandler{
		pathPrefix:        pathPrefix,
		logger:            logger,
		kvstore:           kvstore,
		basicAuthUsername: username,
		basicAuthPassword: password,
	}
	return configRouter
}

func (r *KVStoreRequestHandler) AddRoutes(router *mux.Router) {
	router.NewRoute().Name("health").Path(r.pathPrefix + "/health").Methods(http.MethodGet).HandlerFunc(util.RequestMustHave(r.logger, "", "", http.MethodGet, "", "", nil, r.health))
	router.NewRoute().Name("setKVStore").Path(r.pathPrefix + "/kvstore/{key}").Methods(http.MethodPut).HandlerFunc(util.RequestMustHave(r.logger, r.basicAuthUsername, r.basicAuthPassword, http.MethodPut, "application/json", "", []string{"key"}, r.handleSetKVStore))
	router.NewRoute().Name("getKVStore").Path(r.pathPrefix + "/kvstore/{key}").Methods(http.MethodGet).HandlerFunc(util.RequestMustHave(r.logger, r.basicAuthUsername, r.basicAuthPassword, http.MethodGet, "", "application/json", []string{"key"}, r.handleGetKVStore))
	router.NewRoute().Name("addKVStore").Path(r.pathPrefix + "/kvstore/{key}/add").Methods(http.MethodPost).HandlerFunc(util.RequestMustHave(r.logger, r.basicAuthUsername, r.basicAuthPassword, http.MethodPost, "application/json", "", []string{"key"}, r.handleAddKVStore))
	router.NewRoute().Name("removeKVStore").Path(r.pathPrefix + "/kvstore/{key}/remove").Methods(http.MethodPost).HandlerFunc(util.RequestMustHave(r.logger, r.basicAuthUsername, r.basicAuthPassword, http.MethodPost, "application/json", "", []string{"key"}, r.handleRemoveKVStore))
	router.NewRoute().Name("uploadKVStore").Path(r.pathPrefix + "/kvstore").Methods(http.MethodPut).HandlerFunc(util.RequestMustHave(r.logger, r.basicAuthUsername, r.basicAuthPassword, http.MethodPut, "application/json", "", nil, r.handleUploadKVStore))
	router.NewRoute().Name("downloadKVStore").Path(r.pathPrefix + "/kvstore").Methods(http.MethodGet).HandlerFunc(util.RequestMustHave(r.logger, r.basicAuthUsername, r.basicAuthPassword, http.MethodGet, "", "application/json", nil, r.handleDownloadKVStore))
}

func (r *KVStoreRequestHandler) health(writer http.ResponseWriter, request *http.Request) {
	writer.WriteHeader(http.StatusOK)
}

func (r *KVStoreRequestHandler) handleDownloadKVStore(writer http.ResponseWriter, request *http.Request) {
	if kvs, err := r.kvstore.GetAll(); err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
	} else {
		util.WriteEntity(writer, kvs)
	}
}

func (r *KVStoreRequestHandler) handleUploadKVStore(writer http.ResponseWriter, request *http.Request) {
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

func (r *KVStoreRequestHandler) handleGetKVStore(writer http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	key := vars["key"]
	if val, err := r.kvstore.Get(key); err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
	} else {
		util.WriteEntity(writer, val)
	}
}

func (r *KVStoreRequestHandler) handleSetKVStore(writer http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	key := vars["key"]
	body, err := io.ReadAll(request.Body)
	if err != nil {
		http.Error(writer, "Problem reading request body: "+err.Error(), http.StatusInternalServerError)
		return
	}
	err = r.kvstore.PutAsJson(key, string(body))
	if err != nil {
		http.Error(writer, "Problem with kvstore value, ( is it valid JSON?): "+err.Error(), http.StatusBadRequest)
		return
	}
	writer.WriteHeader(http.StatusNoContent)
}

func (r *KVStoreRequestHandler) handleAddKVStore(writer http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	key := vars["key"]
	body, err := io.ReadAll(request.Body)
	if err != nil {
		http.Error(writer, "Problem reading request body: "+err.Error(), http.StatusInternalServerError)
		return
	}
	var addKvStoreRequest AddKVStoreRequest
	err = json.Unmarshal(body, &addKvStoreRequest)
	if err != nil {
		http.Error(writer, fmt.Sprintf("Can't parse request body '%s' : %v", body, err), http.StatusBadRequest)
		return
	}
	err = r.kvstore.PatchAdd(key, addKvStoreRequest.Path, addKvStoreRequest.Value)
	if err != nil {
		http.Error(writer, fmt.Sprintf("Problem adding kvstore path: '%s' value: '%s', : %v ", addKvStoreRequest.Path, addKvStoreRequest.Value, err), http.StatusBadRequest)
		return
	}
	writer.WriteHeader(http.StatusNoContent)
}

func (r *KVStoreRequestHandler) handleRemoveKVStore(writer http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	key := vars["key"]
	body, err := io.ReadAll(request.Body)
	if err != nil {
		http.Error(writer, "Problem reading request body: "+err.Error(), http.StatusInternalServerError)
		return
	}
	var removeKvStoreRequest RemoveKVStoreRequest
	err = json.Unmarshal(body, &removeKvStoreRequest)
	if err != nil {
		http.Error(writer, "Can't parse request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	err = r.kvstore.PatchRemove(key, removeKvStoreRequest.Path)
	if err != nil {
		http.Error(writer, fmt.Sprintf("Problem removing kvstore '%s', path: '%s' : %v ", key, removeKvStoreRequest.Path, err), http.StatusBadRequest)
		return
	}
	writer.WriteHeader(http.StatusNoContent)
}
