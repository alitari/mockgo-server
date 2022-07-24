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

type AdminRouter struct {
	logger     *utils.Logger
	router     *mux.Router
	mockRouter *MockRouter
}

type EndpointsResponse struct {
	Endpoints []*model.MockEndpoint
}

func NewAdminRouter(mockRouter *MockRouter, logger *utils.Logger) *AdminRouter {
	adminRouter := &AdminRouter{
		mockRouter: mockRouter,
		logger:     logger,
	}
	adminRouter.newRouter()
	return adminRouter
}

func (r *AdminRouter) newRouter() {
	router := mux.NewRouter()
	router.HandleFunc("/endpoints", r.endpoints)
	r.router = router
}

func (r *AdminRouter) getEndpoints(endpoints []*model.MockEndpoint, sn *epSearchNode) []*model.MockEndpoint {
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

func (r *AdminRouter) endpoints(writer http.ResponseWriter, request *http.Request) {
	r.logger.LogAlways(fmt.Sprintf("Received request %v", request))
	endpoints := []*model.MockEndpoint{}
	endpoints = r.getEndpoints(endpoints, r.mockRouter.endpoints)
	endPointResponse := &EndpointsResponse{Endpoints: endpoints}
	writer.Header().Set(headers.ContentType, "application/json")
	resp, err := json.MarshalIndent(endPointResponse, "", "  ")
	if err != nil {
		io.WriteString(writer, fmt.Sprintf("Cannot marshall response: %v", err))
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	r.logger.LogWhenDebugRR(fmt.Sprintf("%v", endPointResponse))
	io.WriteString(writer, string(resp))
}

func (r *AdminRouter) ListenAndServe(port int) {
	r.logger.LogAlways(fmt.Sprintf("Serving admin endpoints on port %v", port))
	http.ListenAndServe(":"+strconv.Itoa(port), r.router)
}
