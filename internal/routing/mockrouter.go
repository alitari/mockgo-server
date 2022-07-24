package routing

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig"
	"github.com/alitari/mockgo-server/internal/kvstore"
	"github.com/alitari/mockgo-server/internal/model"
	"github.com/alitari/mockgo-server/internal/utils"
	"gopkg.in/yaml.v2"

	"github.com/gorilla/mux"
)

type epSearchNode struct {
	searchNodes   map[string]*epSearchNode
	endpoints     map[string][]*model.MockEndpoint
	pathParamName string
}

type ResponseTemplateData struct {
	RequestPathParams map[string]string
	KVStore           map[string]interface{}
	RequestUrl        string
	RequestUser       string
	RequestPath       string
	RequestHost       string
	RequestBody       string
}

type MockRouter struct {
	mockDir             string
	mockFilepattern     string
	responseDir         string
	responseFilepattern string
	logger              *utils.Logger
	endpoints           *epSearchNode
	responseFiles       map[string]*template.Template // responseFilename -> template
	router              *mux.Router
	kvstore             *kvstore.KVStore
}

func NewMockRouter(mockDir, mockFilepattern, responseDir, responseFilepattern string, logger *utils.Logger) (*MockRouter, error) {
	mockRouter := &MockRouter{
		mockDir:             mockDir,
		mockFilepattern:     mockFilepattern,
		responseDir:         responseDir,
		responseFilepattern: responseFilepattern,
		responseFiles:       make(map[string]*template.Template),
		logger:              logger,
		endpoints:           &epSearchNode{},
		kvstore:             kvstore.NewStore(),
	}
	err := mockRouter.loadFiles()
	if err != nil {
		return nil, err
	}
	return mockRouter, nil
}

func (r *MockRouter) loadFiles() error {
	r.endpoints = &epSearchNode{}
	endPointCounter := 0
	mockFiles, err := utils.WalkMatch(r.mockDir, r.mockFilepattern)
	if err != nil {
		return err
	}
	r.logger.LogWhenVerbose(fmt.Sprintf("Found %v mock file(s):", len(mockFiles)))
	for _, mockFile := range mockFiles {
		mock, err := r.readMockFile(mockFile)
		if err != nil {
			return err
		}
		sort.SliceStable(mock.Endpoints, func(i, j int) bool {
			return mock.Endpoints[i].Prio < mock.Endpoints[j].Prio
		})
		for _, endpoint := range mock.Endpoints {
			endPointCounter++
			if len(endpoint.Id) == 0 {
				endpoint.Id = strconv.Itoa(endPointCounter)
			}
			endpoint.Mock = mock
			r.initResponseTemplate(endpoint)
			r.registerEndpoint(endpoint)
		}
	}
	r.newRouter()
	err = r.loadResponseFiles()
	if err != nil {
		return err
	}
	return nil
}

func (r *MockRouter) readMockFile(mockFile string) (*model.Mock, error) {
	r.logger.LogWhenVerbose(fmt.Sprintf("Reading mock file '%s' ...", mockFile))
	mockFileContent, err := ioutil.ReadFile(mockFile)
	if err != nil {
		return nil, err
	}
	mock := &model.Mock{}
	if strings.HasSuffix(mockFile, ".yaml") || strings.HasSuffix(mockFile, ".yml") {
		err = yaml.Unmarshal(mockFileContent, mock)
	}
	if err != nil {
		return nil, err
	}
	if len(mock.Name) == 0 {
		mock.Name = mockFile
	}
	return mock, nil
}

func (r *MockRouter) initResponseTemplate(endpoint *model.MockEndpoint) error {
	responseTpltSourceBytes, err := yaml.Marshal(endpoint.Response)
	if err != nil {
		return err
	}
	responseTplt, err := r.createTemplate(string(responseTpltSourceBytes))
	if err != nil {
		return err
	}
	endpoint.Response.Template = responseTplt
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
		responseFileTplt, err := r.createTemplate(string(responseFileContent))
		if err != nil {
			return err
		}
		r.responseFiles[filepath.Base(responseFile)] = responseFileTplt
	}
	return nil
}

func (r *MockRouter) createTemplate(content string) (*template.Template, error) {
	tplt, err := template.New("response").Funcs(sprig.TxtFuncMap()).Funcs(r.templateFuncMap()).Parse(string(content))
	if err != nil {
		return nil, err
	}
	return tplt, nil
}

func (r *MockRouter) newRouter() {
	r.router = mux.NewRouter()
	var endPoint *model.MockEndpoint
	var requestPathParam map[string]string
	route := r.router.MatcherFunc(func(request *http.Request, match *mux.RouteMatch) bool {
		endPoint, requestPathParam = r.matchRequestToEndpoint(request)
		return endPoint != nil
	})
	route.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		r.renderResponse(writer, request, endPoint, requestPathParam)
	})
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
		pathParamName := ""
		if strings.HasPrefix(pathSegment, "{") && strings.HasSuffix(pathSegment, "}") {
			pathParamName = pathSegment[1 : len(pathSegment)-1]
			pathSegment = "*"
		}
		sn.searchNodes[pathSegment] = &epSearchNode{}
		sn = sn.searchNodes[pathSegment]
		sn.pathParamName = pathParamName
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

func (r *MockRouter) matchRequestToEndpoint(request *http.Request) (*model.MockEndpoint, map[string]string) {
	requestPathParams := map[string]string{}
	sn := r.endpoints
	getPathSegment := func(segments []string, pos int) string {
		if pos < len(segments) {
			return segments[pos]
		} else {
			return ""
		}
	}
	pathSegments := strings.Split(request.URL.Path, "/")[1:]
	allMatch := false
	pathSegment := getPathSegment(pathSegments, 0)
	for pos := 1; pathSegment != ""; pos++ {
		if sn.searchNodes == nil {
			if allMatch {
				break
			} else {
				return nil, requestPathParams
			}
		} else {
			if allMatch {
				for i := pos; pathSegment != ""; i++ {
					if sn.searchNodes[pathSegment] != nil {
						pos = i
						break
					}
					pathSegment = getPathSegment(pathSegments, i)
				}
				allMatch = false
			}
			if sn.searchNodes[pathSegment] == nil {
				if sn.searchNodes["*"] == nil {
					if sn.searchNodes["**"] == nil {
						return nil, requestPathParams
					} else {
						allMatch = true
						sn = sn.searchNodes["**"]
					}
				} else {
					sn = sn.searchNodes["*"]
					if len(sn.pathParamName) > 0 {
						requestPathParams[sn.pathParamName] = pathSegment
					}
				}
			} else {
				sn = sn.searchNodes[pathSegment]
			}
			pathSegment = getPathSegment(pathSegments, pos)
		}
	}
	if sn != nil && sn.endpoints != nil && sn.endpoints[request.Method] != nil {
		return r.matchEndPointsAttributes(sn.endpoints[request.Method], request), requestPathParams
	}
	return nil, requestPathParams
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

func (r *MockRouter) renderResponse(writer http.ResponseWriter, request *http.Request, endpoint *model.MockEndpoint, requestPathParams map[string]string) {
	responseTemplateData, err := r.createResponseTemplateData(request, requestPathParams)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Header().Set("Mocked", "false")
		fmt.Fprintf(writer, "Error rendering response: %v", err)
		return
	}
	renderedResponse, err := r.executeResponseTemplate(endpoint, responseTemplateData)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Header().Set("Mocked", "false")
		fmt.Fprintf(writer, "Error rendering response: %v", err)
		return
	}
	if len(renderedResponse.Headers) > 0 {
		for key, val := range renderedResponse.Headers {
			writer.Header().Set(key, val)
		}
	}
	if renderedResponse.StatusCode > 0 {
		writer.WriteHeader(renderedResponse.StatusCode)
	} else {
		writer.WriteHeader(http.StatusOK)
	}
	bodyStr := ""
	if renderedResponse.Body != "" {
		bodyStr = renderedResponse.Body
	} else if renderedResponse.BodyFilename != "" {
		bodyStr, err = r.executeResponseFileTemplate(renderedResponse.BodyFilename, responseTemplateData)
		if err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			writer.Header().Set("Mocked", "false")
			fmt.Fprintf(writer, "Error rendering response: %v", err)
			return
		}
	}
	writer.Header().Set("Mocked", "true")
	fmt.Fprint(writer, bodyStr)
}

func (r *MockRouter) createResponseTemplateData(request *http.Request, requestPathParams map[string]string) (*ResponseTemplateData, error) {
	data := &ResponseTemplateData{
		RequestUrl:        request.URL.String(),
		RequestPathParams: requestPathParams,
		RequestPath:       request.URL.Path,
		RequestHost:       request.URL.Host,
	}
	if request.URL.User != nil {
		data.RequestUser = request.URL.User.Username()
	}
	if request.Body != nil {
		body := new(bytes.Buffer)
		_, err := body.ReadFrom(request.Body)
		if err != nil {
			return nil, err
		}
		data.RequestBody = body.String()
	}
	return data, nil
}

func (r *MockRouter) executeResponseTemplate(endpoint *model.MockEndpoint, responseTemplateData *ResponseTemplateData) (*model.MockResponse, error) {
	responseExcecuted := &bytes.Buffer{}
	err := endpoint.Response.Template.Execute(responseExcecuted, responseTemplateData)
	if err != nil {
		return nil, err
	}
	r.logger.LogWhenDebugRR(fmt.Sprintf("Rendered response:\n%s", responseExcecuted))
	renderedResponse := &model.MockResponse{}
	err = yaml.Unmarshal(responseExcecuted.Bytes(), renderedResponse)
	if err != nil {
		return nil, fmt.Errorf("could't unmarshall response yaml:\n'%s'\nerror: %v", responseExcecuted, err)
	}
	return renderedResponse, nil
}

func (r *MockRouter) executeResponseFileTemplate(responseFilename string, responseTemplateData *ResponseTemplateData) (string, error) {
	responseExcecuted := &bytes.Buffer{}
	err := r.responseFiles[responseFilename].Execute(responseExcecuted, responseTemplateData)
	if err != nil {
		return "", err
	}
	r.logger.LogWhenDebugRR(fmt.Sprintf("Rendered response file:\n%s", responseExcecuted))
	return responseExcecuted.String(), nil
}

func (r *MockRouter) ListenAndServe(port int) {
	r.logger.LogAlways(fmt.Sprintf("Serving mocks on port %v", port))
	err := http.ListenAndServe(":"+strconv.Itoa(port), r.router)
	if err != nil {
		log.Fatalf("Can't serve on port %v", port)
	}
}
