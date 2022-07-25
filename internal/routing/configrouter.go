package routing

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

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
	router.HandleFunc("/endpoints", r.endpoints)
	router.HandleFunc("/kvstore/{key}", r.setKVStore)
	r.router = router
}

func (r *ConfigRouter) setKVStore(writer http.ResponseWriter, request *http.Request) {
	if request.Method == http.MethodPut {
		if request.Header.Get("Content-Type") == "application/json" {
			vars := mux.Vars(request)
			key := vars["key"]
			if len(key) == 0 {
				http.Error(writer, "Need a key", http.StatusNotFound)
				return
			}
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
		} else {
			http.Error(writer, "Content-Type must be json", http.StatusUnsupportedMediaType)
			return
		}
	} else {
		http.Error(writer, "Only PUT is allowed", http.StatusMethodNotAllowed)
	}
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
	r.logger.LogAlways(fmt.Sprintf("Received request %v", request))
	endpoints := []*model.MockEndpoint{}
	endpoints = r.getEndpoints(endpoints, r.mockRouter.endpoints)
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
