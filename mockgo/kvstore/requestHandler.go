package kvstore

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"text/template"

	"github.com/alitari/mockgo-server/mockgo/logging"
	"github.com/alitari/mockgo-server/mockgo/util"

	"github.com/gorilla/mux"
)

type RequestHandler struct {
	pathPrefix        string
	logger            *logging.LoggerUtil
	kvstore           *jsonStorage
	basicAuthUsername string
	basicAuthPassword string
}

type AddRequest struct {
	Path  string `json:"path"`
	Value string `json:"value"`
}

type RemoveRequest struct {
	Path string `json:"path"`
}

func NewRequestHandler(pathPrefix, username, password string, storage Storage, logger *logging.LoggerUtil) *RequestHandler {
	kvstore := newJSONStorage(storage, logger.Level == logging.Debug)
	configRouter := &RequestHandler{
		pathPrefix:        pathPrefix,
		logger:            logger,
		kvstore:           kvstore,
		basicAuthUsername: username,
		basicAuthPassword: password,
	}
	return configRouter
}

func (r *RequestHandler) GetFuncMap() template.FuncMap {
	return r.kvstore.templateFuncMap()
}

func (r *RequestHandler) AddRoutes(router *mux.Router) {
	router.NewRoute().Name("health").Path(r.pathPrefix + "/health").Methods(http.MethodGet).
		HandlerFunc(r.handleHealth)
	router.NewRoute().Name("setKVStore").Path(r.pathPrefix + "/kvstore/{key}").Methods(http.MethodPut).
		HandlerFunc(util.BasicAuthRequest(r.basicAuthUsername, r.basicAuthPassword, util.JSONContentTypeRequest(util.PathParamRequest([]string{"key"}, r.handleSetKVStore))))
	router.NewRoute().Name("getKVStore").Path(r.pathPrefix + "/kvstore/{key}").Methods(http.MethodGet).
		HandlerFunc(util.BasicAuthRequest(r.basicAuthUsername, r.basicAuthPassword, util.JSONAcceptRequest(util.PathParamRequest([]string{"key"}, r.handleGetKVStore))))
	router.NewRoute().Name("addKVStore").Path(r.pathPrefix + "/kvstore/{key}/add").Methods(http.MethodPost).
		HandlerFunc(util.BasicAuthRequest(r.basicAuthUsername, r.basicAuthPassword, util.JSONContentTypeRequest(util.PathParamRequest([]string{"key"}, r.handleAddKVStore))))
	router.NewRoute().Name("removeKVStore").Path(r.pathPrefix + "/kvstore/{key}/remove").Methods(http.MethodPost).
		HandlerFunc(util.BasicAuthRequest(r.basicAuthUsername, r.basicAuthPassword, util.PathParamRequest([]string{"key"}, r.handleRemoveKVStore)))
}

func (r *RequestHandler) handleHealth(writer http.ResponseWriter, request *http.Request) {
	writer.WriteHeader(http.StatusOK)
}

func (r *RequestHandler) handleGetKVStore(writer http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	key := vars["key"]
	if val, err := r.kvstore.get(key); err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
	} else {
		util.WriteEntity(writer, val)
	}
}

func (r *RequestHandler) handleSetKVStore(writer http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	key := vars["key"]
	body, err := io.ReadAll(request.Body)
	if err != nil {
		http.Error(writer, "Problem reading request body: "+err.Error(), http.StatusInternalServerError)
		return
	}
	err = r.kvstore.putAsJSON(key, string(body))
	if err != nil {
		http.Error(writer, "Problem with kvstore value, ( is it valid JSON?): "+err.Error(), http.StatusBadRequest)
		return
	}
	writer.WriteHeader(http.StatusNoContent)
}

func (r *RequestHandler) handleAddKVStore(writer http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	key := vars["key"]
	body, err := io.ReadAll(request.Body)
	if err != nil {
		http.Error(writer, "Problem reading request body: "+err.Error(), http.StatusInternalServerError)
		return
	}
	var addKvStoreRequest AddRequest
	err = json.Unmarshal(body, &addKvStoreRequest)
	if err != nil {
		http.Error(writer, fmt.Sprintf("Can't parse request body '%s' : %v", body, err), http.StatusBadRequest)
		return
	}
	err = r.kvstore.patchAdd(key, addKvStoreRequest.Path, addKvStoreRequest.Value)
	if err != nil {
		http.Error(writer, fmt.Sprintf("Problem adding kvstore path: '%s' value: '%s', : %v ", addKvStoreRequest.Path, addKvStoreRequest.Value, err), http.StatusBadRequest)
		return
	}
	writer.WriteHeader(http.StatusNoContent)
}

func (r *RequestHandler) handleRemoveKVStore(writer http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	key := vars["key"]
	body, err := io.ReadAll(request.Body)
	if err != nil {
		http.Error(writer, "Problem reading request body: "+err.Error(), http.StatusInternalServerError)
		return
	}
	var removeKvStoreRequest RemoveRequest
	err = json.Unmarshal(body, &removeKvStoreRequest)
	if err != nil {
		http.Error(writer, "Can't parse request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	err = r.kvstore.patchRemove(key, removeKvStoreRequest.Path)
	if err != nil {
		http.Error(writer, fmt.Sprintf("Problem removing kvstore '%s', path: '%s' : %v ", key, removeKvStoreRequest.Path, err), http.StatusBadRequest)
		return
	}
	writer.WriteHeader(http.StatusNoContent)
}
