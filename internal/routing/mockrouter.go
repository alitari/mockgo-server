package routing

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/alitari/mockgo-server/internal/model"
	"github.com/alitari/mockgo-server/internal/utils"

	"github.com/gorilla/mux"
)

type MockRouter struct {
	mappingDir         string
	mappingFilepattern string
	logger             *utils.Logger
	mocks              map[string]*model.Mock
	router             *mux.Router
}

func NewMockRouter(mappingDir, mappingFilepattern string, logger *utils.Logger) (*MockRouter, error) {
	mockRouter := &MockRouter{mappingDir: mappingDir,
		mappingFilepattern: mappingFilepattern,
		logger:             logger,
		mocks:              make(map[string]*model.Mock)}
	router, err := mockRouter.load()
	if err != nil {
		return nil, err
	}
	mockRouter.router = router
	return mockRouter, nil
}

func (r *MockRouter) load() (*mux.Router, error) {
	mappingFiles, err := utils.WalkMatch(r.mappingDir, r.mappingFilepattern)
	if err != nil {
		return nil, err
	}

	r.logger.LogWhenVerbose(fmt.Sprintf("Found %v mapping file(s):", len(mappingFiles)))

	for _, mappingFile := range mappingFiles {
		r.logger.LogWhenVerbose(fmt.Sprintf("Reading mapping file '%s' ...", mappingFile))
		mockData, err := ioutil.ReadFile(mappingFile)
		if err != nil {
			return nil, err
		}
		mock := &model.Mock{}
		err = json.Unmarshal(mockData, mock)
		if err != nil {
			return nil, err
		}
		r.logger.LogWhenVerbose(fmt.Sprintf("Mock created with '%v' endpoints.", len(mock.Endpoints)))
		r.mocks[mappingFile] = mock
	}

	return r.newRouter()
}

func (r *MockRouter) newRouter() (*mux.Router, error) {
	router := mux.NewRouter()
	for _, mock := range r.mocks {
		for _, e := range mock.Endpoints {
			r := router.HandleFunc(e.Request.Path, handlerFactory(router, e))
			if e.Request.Method != "" {
				r.Methods(e.Request.Method)
			}
			if e.Request.Scheme != "" {
				r.Schemes(e.Request.Scheme)
			}
			if e.Request.Host != "" {
				r.Host(e.Request.Host)
			}
			if len(e.Request.Query) > 0 {
				for key, val := range e.Request.Query {
					r.Queries(key, val)
				}
			}
			if len(e.Request.Headers) > 0 {
				for key, val := range e.Request.Headers {
					r.Headers(key, val)
				}
			}
		}
	}
	return router, nil
}

func handlerFactory(r *mux.Router, e model.Endpoint) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if len(e.Response.Headers) > 0 {
			for key, val := range e.Response.Headers {
				w.Header().Set(key, val)
			}
		}
		if e.Response.Body != "" {
			if e.Response.StatusCode > 0 {
				w.WriteHeader(e.Response.StatusCode)
			} else {
				w.WriteHeader(http.StatusOK)
			}
			fmt.Fprint(w, e.Response.Body)
			return
		}
		if e.Response.BodyFileName != "" {
			http.ServeFile(w, r, e.Response.BodyFileName)
			return
		}
	}
}

func (r *MockRouter) ListenAndServe(port int) {
	r.logger.LogAlways(fmt.Sprintf("Serving mocks on port %v", port))
	http.ListenAndServe(":"+strconv.Itoa(port), r.router)
}
