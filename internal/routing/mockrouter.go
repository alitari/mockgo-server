package routing

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig"
	"github.com/alitari/mockgo-server/internal/model"
	"github.com/alitari/mockgo-server/internal/utils"
	"gopkg.in/yaml.v2"

	"github.com/gorilla/mux"
)

type epSearchNode struct {
	searchNodes map[string]*epSearchNode
	endpoints   map[string][]*model.MockEndpoint
}

type MockRouter struct {
	mockDir             string
	mockFilepattern     string
	responseDir         string
	responseFilepattern string
	logger              *utils.Logger
	mocks               map[string]*model.Mock // mockfile -> mock
	endpoints           *epSearchNode
	responseFiles       map[string]string // responseFilename -> file content
	router              *mux.Router
}

func NewMockRouter(mockDir, mockFilepattern, responseDir, responseFilepattern string, logger *utils.Logger) (*MockRouter, error) {
	mockRouter := &MockRouter{
		mockDir:             mockDir,
		mockFilepattern:     mockFilepattern,
		responseDir:         responseDir,
		responseFilepattern: responseFilepattern,
		responseFiles:       make(map[string]string),
		logger:              logger,
		mocks:               make(map[string]*model.Mock),
		endpoints:           &epSearchNode{},
	}
	err := mockRouter.loadFiles()
	if err != nil {
		return nil, err
	}
	return mockRouter, nil
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
		r.mocks[mockFile] = &model.Mock{}
		if strings.HasSuffix(mockFile, ".yaml") || strings.HasSuffix(mockFile, ".yml") {
			err = yaml.Unmarshal(mockFileContent, r.mocks[mockFile])
		}
		if err != nil {
			return err
		}
		for i, endpoint := range r.mocks[mockFile].Endpoints {
			requestTpltSourceBytes, err := yaml.Marshal(endpoint.Request)
			if err != nil {
				return err
			}

			responseTpltSourceBytes, err := yaml.Marshal(endpoint.Response)
			if err != nil {
				return err
			}

			r.logger.LogWhenVerbose("requestTemplate source: '" + string(requestTpltSourceBytes) + "'")
			requestTplt, err := template.New(mockFile + "-request-" + strconv.Itoa(i)).Funcs(sprig.TxtFuncMap()).Parse(string(requestTpltSourceBytes))
			if err != nil {
				return err
			}
			endpoint.Request.Template = requestTplt
			responseTplt, err := template.New(mockFile + "-response-" + strconv.Itoa(i)).Funcs(sprig.TxtFuncMap()).Parse(string(responseTpltSourceBytes))
			if err != nil {
				return err
			}
			endpoint.Response.Template = responseTplt
		}
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
	var endPoint *model.MockEndpoint
	route := r.router.MatcherFunc(func(request *http.Request, match *mux.RouteMatch) bool {
		endPoint = r.matchRequestToEndpoint(request)
		return endPoint != nil
	})
	route.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		r.renderResponse(writer, request, endPoint)
	})
}

func (r *MockRouter) LoadMocks() error {
	r.endpoints = &epSearchNode{}
	for mockFile, mock := range r.mocks {
		for _, endpoint := range mock.Endpoints {
			requestExcecuted := &bytes.Buffer{}
			err := endpoint.Request.Template.Execute(requestExcecuted, "")
			if err != nil {
				return err
			}
			var request model.MatchRequest
			if strings.HasSuffix(mockFile, ".yaml") || strings.HasSuffix(mockFile, ".yml") {
				err = yaml.Unmarshal(requestExcecuted.Bytes(), &request)
			}
			if err != nil {
				return err
			}
			endpoint.Request = &request
			r.registerEndpoint(endpoint)
		}
	}
	return nil
}

func (r *MockRouter) registerEndpoint(endpoint *model.MockEndpoint) {
	if endpoint.Request.Path == "" {
		endpoint.Request.Path = "/"
	}
	if endpoint.Request.Method == "" {
		endpoint.Request.Method = "GET"
	}

	sn := r.endpoints
	pathSegments := strings.Split(endpoint.Request.Path, "/")
	for _, pathSegment := range pathSegments[1:] {
		if sn.searchNodes == nil {
			sn.searchNodes = make(map[string]*epSearchNode)
		}
		sn.searchNodes[pathSegment] = &epSearchNode{}
		sn = sn.searchNodes[pathSegment]
	}
	if sn.endpoints == nil {
		sn.endpoints = make(map[string][]*model.MockEndpoint)
	}

	if sn.endpoints[endpoint.Request.Method] == nil {
		sn.endpoints[endpoint.Request.Method] = []*model.MockEndpoint{}
	}
	sn.endpoints[endpoint.Request.Method] = append(sn.endpoints[endpoint.Request.Method], endpoint)
	r.logger.LogWhenVerbose(fmt.Sprintf("register endpoint for path|method: %s|%s", endpoint.Request.Path, endpoint.Request.Method))
}

func (r *MockRouter) matchRequestToEndpoint(request *http.Request) *model.MockEndpoint {
	sn := r.endpoints
	getPathSegment := func(segments []string, pos int, skip bool) string {
		if pos < len(segments) {
			return segments[pos]
		} else {
			return ""
		}
	}
	pathSegments := strings.Split(request.URL.Path, "/")[1:]
	allMatch := false
	pathSegment := getPathSegment(pathSegments, 0, allMatch)
	for pos := 1; pathSegment != ""; pos++ {
		if sn.searchNodes == nil {
			if allMatch {
				break
			} else {
				return nil
			}
		} else {
			if allMatch {
				for i := pos ; pathSegment != ""; i++ {
					if sn.searchNodes[pathSegment] != nil {
						pos = i
						break
					}
					pathSegment = getPathSegment(pathSegments, i, allMatch)
				}
				allMatch = false
			}
			if sn.searchNodes[pathSegment] == nil {
				if sn.searchNodes["*"] == nil {
					if sn.searchNodes["**"] == nil {
						return nil
					} else {
						allMatch = true
						sn = sn.searchNodes["**"]
					}
				} else {
					sn = sn.searchNodes["*"]
				}
			} else {
				sn = sn.searchNodes[pathSegment]
			}
			pathSegment = getPathSegment(pathSegments, pos, allMatch)
		}
	}
	if sn != nil && sn.endpoints != nil && sn.endpoints[request.Method] != nil {
		return r.matchEndPointsAttributes(sn.endpoints[request.Method], request)
	}

	return nil
}

func (r *MockRouter) matchEndPointsAttributes(endPoints []*model.MockEndpoint, request *http.Request) *model.MockEndpoint {
	for _, ep := range endPoints {
		if !r.matchQueryParams(ep.Request, request) {
			continue
		}
		if !r.matchHeaderValues(ep.Request, request) {
			continue
		}
		return ep
	}
	return nil
}

func (r *MockRouter) matchQueryParams(matchRequest *model.MatchRequest, request *http.Request) bool {
	if len(matchRequest.Query) > 0 {
		for key, val := range matchRequest.Query {
			if request.URL.Query().Get(key) != val {
				return false
			}
		}
		return true
	} else {
		return true
	}
}

func (r *MockRouter) matchHeaderValues(matchRequest *model.MatchRequest, request *http.Request) bool {
	if len(matchRequest.Headers) > 0 {
		for key, val := range matchRequest.Headers {
			if request.Header.Get(key) != val {
				return false
			}
		}
		return true
	} else {
		return true
	}
}

func (r *MockRouter) renderResponse(writer http.ResponseWriter, request *http.Request, endpoint *model.MockEndpoint) {
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
	err := http.ListenAndServe(":"+strconv.Itoa(port), r.router)
	if err != nil {
		log.Fatalf("Can't serve on port %v", port)
	}
}
