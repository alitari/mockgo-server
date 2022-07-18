package routing

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"

	"github.com/alitari/mockgo-server/internal/model"
	"github.com/alitari/mockgo-server/internal/utils"
	"gopkg.in/yaml.v2"

	"github.com/gorilla/mux"
)

type MockRouter struct {
	mockDir             string
	mockFilepattern     string
	responseDir         string
	responseFilepattern string
	logger              *utils.Logger
	mocks               map[string]*model.Mock                  // mockfile -> mock
	endpoints           map[string]map[string][]*model.Endpoint // path -> method -> endpoint
	responseFiles       map[string]string                       // responseFilename -> file content
	router              *mux.Router
}

func NewMockRouter(mockDir, mockFilepattern, responseDir, responseFilepattern string, logger *utils.Logger) *MockRouter {
	mockRouter := &MockRouter{
		mockDir:             mockDir,
		mockFilepattern:     mockFilepattern,
		responseDir:         responseDir,
		responseFilepattern: responseFilepattern,
		responseFiles:       make(map[string]string),
		logger:              logger,
		mocks:               make(map[string]*model.Mock),
		endpoints:           make(map[string]map[string][]*model.Endpoint),
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
		mock := &model.Mock{}
		isYaml := false
		if strings.HasSuffix(mockFile, ".json") {
			err = json.Unmarshal(mockFileContent, mock)

		} else if strings.HasSuffix(mockFile, ".yaml") || strings.HasSuffix(mockFile, ".yml") {
			err = yaml.Unmarshal(mockFileContent, mock)
			isYaml = true
		}
		if err != nil {
			return err
		}
		for i, endpoint := range mock.Endpoints {
			var requestTpltSource []byte
			var responseTpltSource []byte

			if isYaml {
				requestTpltSource, err = yaml.Marshal(endpoint.Request)
				if err != nil {
					return err
				}
				responseTpltSource, err = yaml.Marshal(endpoint.Response)
				if err != nil {
					return err
				}
			} else {
				requestTpltSource, err = json.MarshalIndent(endpoint.Request, "", "  ")
				if err != nil {
					return err
				}
				responseTpltSource, err = json.MarshalIndent(endpoint.Response, "", "  ")
				if err != nil {
					return err
				}
			}
			requestTplt, err := template.New(mockFile + "-request-" + strconv.Itoa(i)).Parse(string(requestTpltSource))
			if err != nil {
				return err
			}
			endpoint.Request.Template = requestTplt
			responseTplt, err := template.New(mockFile + "-response-" + strconv.Itoa(i)).Parse(string(responseTpltSource))
			if err != nil {
				return err
			}
			endpoint.Response.Template = responseTplt
		}
		r.mocks[mockFile] = mock
	}
	r.newRouter()
	err = r.loadResponseFiles()
	if err != nil {
		return err
	}
	return nil
}

func (r *MockRouter) loadResponseFiles() error {
	responseFiles, err := utils.WalkMatch(r.responseDir, r.responseFilepattern)
	if err != nil {
		return err
	}
	r.logger.LogWhenVerbose(fmt.Sprintf("Found %v response file(s):", len(responseFiles)))
	for _, responseFile := range responseFiles {
		r.logger.LogWhenVerbose(fmt.Sprintf("Reading response file '%s' ...", responseFile))
		responseFileContent, err := ioutil.ReadFile(responseFile)
		if err != nil {
			return err
		}
		r.responseFiles[filepath.Base(responseFile)] = string(responseFileContent)
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

func (r *MockRouter) LoadMocks() error {
	r.endpoints = map[string]map[string][]*model.Endpoint{}
	for mockFile, mock := range r.mocks {
		for _, endpoint := range mock.Endpoints {
			requestExcecuted := &bytes.Buffer{}
			err := endpoint.Request.Template.Execute(requestExcecuted, nil)
			if err != nil {
				return err
			}
			var request model.Request
			if strings.HasSuffix(mockFile, ".json") {
				err = json.Unmarshal(requestExcecuted.Bytes(), &request)

			} else if strings.HasSuffix(mockFile, ".yaml") || strings.HasSuffix(mockFile, ".yml") {
				err = yaml.Unmarshal(requestExcecuted.Bytes(), &request)
			}
			if err != nil {
				return err
			}
			endpoint.Request = &request
			r.storeEndpoint(endpoint)
		}
	}
	return nil
}

func (r *MockRouter) storeEndpoint(endpoint *model.Endpoint) {
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
	endPoints = append(endPoints, endpoint)
	r.endpoints[endpoint.Request.Path][endpoint.Request.Method] = endPoints
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
	if endpoint.Response.StatusCode > 0 {
		writer.WriteHeader(endpoint.Response.StatusCode)
	} else {
		writer.WriteHeader(http.StatusOK)
	}
	bodyStr := ""
	if endpoint.Response.Body != "" {
		bodyStr = endpoint.Response.Body
	} else if endpoint.Response.BodyFileName != "" {
		bodyStr = r.responseFiles[endpoint.Response.BodyFileName]
	}
	fmt.Fprint(writer, bodyStr)
}

func (r *MockRouter) ListenAndServe(port int) {
	r.logger.LogAlways(fmt.Sprintf("Serving mocks on port %v", port))
	http.ListenAndServe(":"+strconv.Itoa(port), r.router)
}
