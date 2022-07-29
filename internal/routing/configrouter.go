package routing

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"

	"github.com/alitari/mockgo-server/internal/kvstore"
	"github.com/alitari/mockgo-server/internal/model"
	"github.com/alitari/mockgo-server/internal/utils"

	"github.com/go-http-utils/headers"
	"github.com/gorilla/mux"
)

type ConfigRouter struct {
	logger     *utils.Logger
	router     *mux.Router
	mockRouter *MockRouter
}

type EndpointsResponse struct {
	Endpoints []*model.MockEndpoint
}

func NewConfigRouter(mockRouter *MockRouter, logger *utils.Logger) *ConfigRouter {
	configRouter := &ConfigRouter{
		mockRouter: mockRouter,
		logger:     logger,
	}
	configRouter.newRouter()
	return configRouter
}

func (r *ConfigRouter) newRouter() {
	router := mux.NewRouter()
	router.NewRoute().Name("endpoints").Path("/endpoints").HandlerFunc(utils.RequestMustHave(http.MethodGet, "", "application/json", nil, r.endpoints))
	router.NewRoute().Name("setKVStore").Path("/kvstore/{key}").Methods(http.MethodPut).HandlerFunc(utils.RequestMustHave(http.MethodPut, "application/json", "", []string{"key"}, r.setKVStore))
	router.NewRoute().Name("getKVStore").Path("/kvstore/{key}").Methods(http.MethodGet).HandlerFunc(utils.RequestMustHave(http.MethodGet, "", "application/json", []string{"key"}, r.getKVStore))
	router.NewRoute().Name("uploadKVStore").Path("/kvstore/").Methods(http.MethodPut).HandlerFunc(utils.RequestMustHave(http.MethodPut, "application/json", "", nil, r.uploadKVStore))
	router.NewRoute().Name("downloadKVStore").Path("/kvstore/").Methods(http.MethodPut).HandlerFunc(utils.RequestMustHave(http.MethodGet, "", "application/json", nil, r.downloadKVStore))
	r.router = router
}

func (r *ConfigRouter) downloadKVStore(writer http.ResponseWriter, request *http.Request) {
	store := r.mockRouter.kvstore.GetAll()
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
	store, err := kvstore.NewStoreWithContent(string(body))
	if err != nil {
		http.Error(writer, "Problem creating kvstore: "+err.Error(), http.StatusInternalServerError)
		return
	}
	r.mockRouter.kvstore = store
	writer.WriteHeader(http.StatusNoContent)
}

func (r *ConfigRouter) getKVStore(writer http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	key := vars["key"]
	val, err := r.mockRouter.kvstore.Get(key)
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
	err = r.mockRouter.kvstore.Put(key, string(body))
	if err != nil {
		http.Error(writer, "Problem with kvstore value, ( is it valid JSON?): "+err.Error(), http.StatusBadRequest)
		return
	}
	writer.WriteHeader(http.StatusNoContent)
}

func (r *ConfigRouter) getEndpoints(endpoints []*model.MockEndpoint, sn *epSearchNode) []*model.MockEndpoint {
	for _, sns := range sn.searchNodes {
		if sns.endpoints != nil {
			for _, epMethodMap := range sns.endpoints {
				endpoints = append(endpoints, epMethodMap...)
			}
		}
		if sns.searchNodes != nil {
			endpoints = append(endpoints, r.getEndpoints(endpoints, sns)...)
		}
	}
	return endpoints
}

func (r *ConfigRouter) endpoints(writer http.ResponseWriter, request *http.Request) {
	endpoints := []*model.MockEndpoint{}
	endpoints = r.getEndpoints(endpoints, r.mockRouter.endpoints)
	sort.SliceStable(endpoints, func(i, j int) bool {
		return endpoints[i].Id < endpoints[j].Id
	})
	endPointResponse := &EndpointsResponse{Endpoints: endpoints}
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

func (r *ConfigRouter) ListenAndServe(port int) {
	r.logger.LogAlways(fmt.Sprintf("Serving admin endpoints on port %v", port))
	http.ListenAndServe(":"+strconv.Itoa(port), r.router)
}
