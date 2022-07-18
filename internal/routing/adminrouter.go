package routing

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/alitari/mockgo-server/internal/model"
	"github.com/alitari/mockgo-server/internal/utils"
	"gopkg.in/yaml.v2"

	"github.com/go-http-utils/headers"
	"github.com/gorilla/mux"
)

type AdminRouter struct {
	logger     *utils.Logger
	router     *mux.Router
	mockRouter *MockRouter
}

type InfoResponse struct {
	Mocks map[string]*model.Mock
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
	router.HandleFunc("/admin/info", r.info)
	r.router = router
}

func (r *AdminRouter) info(writer http.ResponseWriter, request *http.Request) {
	r.logger.LogAlways(fmt.Sprintf("Received request %v", request))
	infoData := &InfoResponse{Mocks: r.mockRouter.mocks}
	yamlStr, err := yaml.Marshal(infoData)
	if err != nil {
		mess := fmt.Sprintf("Cannot marshal data %v : %v", infoData, err)
		r.logger.LogAlways(mess)
		writer.Write([]byte(mess))
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	writer.Header().Set(headers.ContentType, "application/json")
	writer.Write(yamlStr)
	writer.WriteHeader(http.StatusOK)
}

func (r *AdminRouter) ListenAndServe(port int) {
	r.logger.LogAlways(fmt.Sprintf("Serving admin endpoints on port %v", port))
	http.ListenAndServe(":"+strconv.Itoa(port), r.router)
}
