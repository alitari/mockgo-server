package routing

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"text/template"

	"github.com/alitari/mockgo-server/internal/model"
	"github.com/alitari/mockgo-server/internal/utils"
	"gopkg.in/yaml.v2"

	"github.com/gorilla/mux"
)

type MockRouter struct {
	mockDir         string
	mockFilepattern string
	logger          *utils.Logger
	mockFiles       map[string]string                       // mockfile -> mockfile content
	mockTemplates   map[string]*template.Template           // mockfile -> template
	mocks           map[string]*model.Mock                  // mockfile -> mock
	endpoints       map[string]map[string][]*model.Endpoint // path -> method -> endpoint
	router          *mux.Router
}

func NewMockRouter(mockDir, mockFilepattern string, logger *utils.Logger) *MockRouter {
	mockRouter := &MockRouter{
		mockDir:         mockDir,
		mockFilepattern: mockFilepattern,
		logger:          logger,
		mocks:           make(map[string]*model.Mock),
		mockFiles:       make(map[string]string),
		mockTemplates:   make(map[string]*template.Template),

		endpoints: make(map[string]map[string][]*model.Endpoint),
	}
	mockRouter.loadFiles()
	return mockRouter
}

func (r *MockRouter) loadFiles() error {
	mockFiles, err := utils.WalkMatch(r.mockDir, r.mockFilepattern)
	if err != nil {
		return err
	}
	r.logger.LogWhenVerbose(fmt.Sprintf("Found %v mock file(s):", len(mockFiles)))
	for _, mockFile := range mockFiles {
		r.logger.LogWhenVerbose(fmt.Sprintf("Reading mock file '%s' ...", mockFile))
		mockFileContent, err := ioutil.ReadFile(mockFile)
		if err != nil {
			return err
		}
		r.mockFiles[mockFile] = string(mockFileContent)
		tplt, err := template.New(mockFile).Parse(string(mockFileContent))
		if err != nil {
			return err
		}
		r.mockTemplates[mockFile] = tplt
	}
	r.newRouter()
	return nil
}

func (r *MockRouter) LoadMocks() error {
	for mockFile, tplt := range r.mockTemplates {
		mockExcecuted := &bytes.Buffer{}
		err := tplt.Execute(mockExcecuted, nil)
		if err != nil {
			return err
		}
		mock := &model.Mock{}
		if strings.HasSuffix(mockFile, ".json") {
			err = json.Unmarshal(mockExcecuted.Bytes(), mock)
		} else if strings.HasSuffix(mockFile, ".yaml") || strings.HasSuffix(mockFile, ".yml") {
			err = yaml.Unmarshal(mockExcecuted.Bytes(), mock)
		}
		if err != nil {
			return err
		}
		r.mocks[mockFile] = mock
		for _, endpoint := range mock.Endpoints {
			if endpoint.Request.Path == "" {
				endpoint.Request.Path = "/"
			}
			if r.endpoints[endpoint.Request.Path] == nil {
				r.endpoints[endpoint.Request.Path] = make(map[string][]*model.Endpoint)
			}
			if endpoint.Request.Method == "" {
				endpoint.Request.Method = "GET"
			}
			endPoints := r.endpoints[endpoint.Request.Path][endpoint.Request.Method]
			if endPoints == nil {
				endPoints = []*model.Endpoint{}
			}
			endPoints = append(endPoints, &endpoint)
			r.endpoints[endpoint.Request.Path][endpoint.Request.Method] = endPoints
		}
	}
	return nil
}

func (r *MockRouter) newRouter() {
	r.router = mux.NewRouter()
	var endPoint *model.Endpoint
	route := r.router.MatcherFunc(func(request *http.Request, match *mux.RouteMatch) bool {
		endPoint = r.matchRequestToEndpoint(request)
		return endPoint != nil
	})
	route.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		r.renderResponse(writer, request, endPoint)
	})
}

func (r *MockRouter) matchRequestToEndpoint(request *http.Request) *model.Endpoint {
	endpoints := r.endpoints[request.URL.Path][request.Method]
	if endpoints == nil {
		return nil
	}
	for _, ep := range endpoints {
		if len(ep.Request.Query) > 0 {
			for key, val := range ep.Request.Query {
				if request.URL.Query().Get(key) != val {
					continue
				}
			}
		}
		if len(ep.Request.Headers) > 0 {
			for key, val := range ep.Request.Headers {
				if request.Header.Get(key) != val {
					continue
				}
			}
		}
		return ep
	}
	return nil
}

func (r *MockRouter) renderResponse(writer http.ResponseWriter, request *http.Request, endpoint *model.Endpoint) {
	if len(endpoint.Response.Headers) > 0 {
		for key, val := range endpoint.Response.Headers {
			writer.Header().Set(key, val)
		}
	}
	if endpoint.Response.Body != "" {
		if endpoint.Response.StatusCode > 0 {
			writer.WriteHeader(endpoint.Response.StatusCode)
		} else {
			writer.WriteHeader(http.StatusOK)
		}
		fmt.Fprint(writer, endpoint.Response.Body)
		return
	}
	if endpoint.Response.BodyFileName != "" {
		http.ServeFile(writer, request, endpoint.Response.BodyFileName)
		return
	}
}

// func (r *MockRouter) handlerFactory() func(writer http.ResponseWriter, request *http.Request) {
// 	return func(writer http.ResponseWriter, request *http.Request) {
// 		val := request.Context().Value("mykey")
// 		r.logger.LogAlways(fmt.Sprintf("val=%v", val))
// 		var matchedEndpoint *model.Endpoint
// 		for _, mock := range r.mocks {
// 			endPointCandidate := r.matchedEndpoint(mock, request)
// 			if endPointCandidate != nil && (matchedEndpoint == nil || endPointCandidate.Prio > matchedEndpoint.Prio) {
// 				matchedEndpoint = endPointCandidate
// 			}
// 		}
// 		if matchedEndpoint == nil {
// 			writer.WriteHeader(http.StatusNotFound)
// 			return
// 		}

// 		if len(matchedEndpoint.Response.Headers) > 0 {
// 			for key, val := range matchedEndpoint.Response.Headers {
// 				writer.Header().Set(key, val)
// 			}
// 		}
// 		if matchedEndpoint.Response.Body != "" {
// 			if matchedEndpoint.Response.StatusCode > 0 {
// 				writer.WriteHeader(matchedEndpoint.Response.StatusCode)
// 			} else {
// 				writer.WriteHeader(http.StatusOK)
// 			}
// 			fmt.Fprint(writer, matchedEndpoint.Response.Body)
// 			return
// 		}
// 		if matchedEndpoint.Response.BodyFileName != "" {
// 			http.ServeFile(writer, request, matchedEndpoint.Response.BodyFileName)
// 			return
// 		}
// 	}
// }

// func (r *MockRouter) matchedEndpoint(mock *model.Mock, request *http.Request) *model.Endpoint {
// 	var matched *model.Endpoint
// 	for _, endPoint := range mock.Endpoints {
// 		if r.matchEndpoint(&endPoint, request) && (matched == nil || endPoint.Prio > matched.Prio) {
// 			matched = &endPoint
// 		}
// 	}
// 	return matched
// }

// func (r *MockRouter) matchEndpoint(endPoint *model.Endpoint, request *http.Request) bool {
// 	if endPoint.Request.Path != request.URL.Path {
// 		return false
// 	}

// 	if endPoint.Request.Method != "" && endPoint.Request.Method != request.Method {
// 		return false
// 	}
// 	if len(endPoint.Request.Query) > 0 {
// 		for key, val := range endPoint.Request.Query {
// 			if request.URL.Query().Get(key) != val {
// 				return false
// 			}
// 		}
// 	}
// 	if len(endPoint.Request.Headers) > 0 {
// 		for key, val := range endPoint.Request.Headers {
// 			if request.Header.Get(key) != val {
// 				return false
// 			}
// 		}
// 	}
// 	return true
// }

func (r *MockRouter) ListenAndServe(port int) {
	r.logger.LogAlways(fmt.Sprintf("Serving mocks on port %v", port))
	http.ListenAndServe(":"+strconv.Itoa(port), r.router)
}
