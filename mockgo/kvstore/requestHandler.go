package kvstore

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"go.uber.org/zap"

	"github.com/alitari/mockgo-server/mockgo/util"

	"github.com/gorilla/mux"
)

/*
RequestHandler implements an http API to access a Storage
*/
type RequestHandler struct {
	pathPrefix        string
	logger            *zap.Logger
	storage           Storage
	basicAuthUsername string
	basicAuthPassword string
}

type addRequest struct {
	Path  string `json:"path"`
	Value string `json:"value"`
}

type removeRequest struct {
	Path string `json:"path"`
}

const (
	pathParamStore = "store"
	pathParamKey   = "key"
)

/*
NewRequestHandler creates an instance of RequestHandler
*/
func NewRequestHandler(pathPrefix string, storage Storage, logLevel string) *RequestHandler {
	configRouter := &RequestHandler{
		pathPrefix: pathPrefix,
		logger:     util.CreateLogger(logLevel),
		storage:    storage,
	}
	return configRouter
}

/*
AddRoutes adds mux.Routes for the http API to a given mux.Router
*/
func (r *RequestHandler) AddRoutes(router *mux.Router) {
	baseStorePath := r.pathPrefix + "/kvstore"
	storePath := baseStorePath + "/{" + pathParamStore + "}"
	storePathWithKey := storePath + "/{" + pathParamKey + "}"
	router.NewRoute().Name("putKVStore").Path(storePathWithKey).Methods(http.MethodPut).
		HandlerFunc(util.JSONContentTypeRequest(util.PathParamRequest([]string{pathParamStore, pathParamKey}, r.handlePutKVStore)))
	router.NewRoute().Name("getKVStore").Path(storePathWithKey).Methods(http.MethodGet).
		HandlerFunc(util.JSONAcceptRequest(util.PathParamRequest([]string{pathParamStore, pathParamKey}, r.handleGetKVStore)))
	router.NewRoute().Name("getAllKVStore").Path(storePath).Methods(http.MethodGet).
		HandlerFunc(util.JSONAcceptRequest(util.PathParamRequest([]string{pathParamStore}, r.handleGetAllKVStore)))
	router.NewRoute().Name("removeKVStore").Path(storePathWithKey).Methods(http.MethodDelete).
		HandlerFunc(util.PathParamRequest([]string{pathParamStore, pathParamKey}, r.handleRemoveKVStore))
}

func (r *RequestHandler) handleGetKVStore(writer http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	if val, err := r.storage.Get(vars[pathParamStore], vars[pathParamKey]); err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
	} else {
		if val == nil {
			writer.WriteHeader(http.StatusNotFound)
			return
		}
		util.WriteEntity(writer, val)
	}
}

func (r *RequestHandler) handleGetAllKVStore(writer http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	if val, err := r.storage.GetAll(vars[pathParamStore]); err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
	} else {
		if val == nil {
			writer.WriteHeader(http.StatusNotFound)
			return
		}
		util.WriteEntity(writer, val)
	}
}

func (r *RequestHandler) handlePutKVStore(writer http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	body, err := io.ReadAll(request.Body)
	if err != nil {
		http.Error(writer, "Problem reading request body: "+err.Error(), http.StatusInternalServerError)
		return
	}
	var bodyVal interface{}
	if err := json.Unmarshal(body, &bodyVal); err != nil {
		http.Error(writer, "Can't parse request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	err = r.storage.Put(vars[pathParamStore], vars[pathParamKey], bodyVal)
	if err != nil {
		http.Error(writer, "Can't store kvstore value "+err.Error(), http.StatusInternalServerError)
		return
	}
	writer.WriteHeader(http.StatusNoContent)
}

func (r *RequestHandler) handleRemoveKVStore(writer http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	if err := r.storage.Put(vars[pathParamStore], vars[pathParamKey], nil); err != nil {
		http.Error(writer, fmt.Sprintf("Can't remove value from kvstore: %s", err.Error()), http.StatusInternalServerError)
		return
	}
	writer.WriteHeader(http.StatusNoContent)
}
