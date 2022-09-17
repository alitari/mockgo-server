package mock

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"net/http"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/Masterminds/sprig"
	"github.com/alitari/mockgo-server/internal/kvstore"
	"github.com/alitari/mockgo-server/internal/model"
	"github.com/alitari/mockgo-server/internal/utils"
	"gopkg.in/yaml.v2"

	"github.com/gorilla/mux"
)

const TEMPLATE_NAME_RESPONSEBODY = "responseBody"
const TEMPLATE_NAME_RESPONSESTATUS = "responseStatus"
const TEMPLATE_NAME_RESPONSEHEADERS = "responseHeader"

const HEADER_KEY_ENDPOINT_ID = "endpoint-Id"

type ResponseTemplateData struct {
	RequestPathParams   map[string]string
	KVStore             map[string]interface{}
	RequestUrl          string
	RequestUser         string
	RequestPath         string
	RequestHost         string
	RequestBody         string
	RequestBodyJsonData map[string]interface{}
}

type MockRouter struct {
	mockDir               string
	mockFilepattern       string
	responseDir           string
	MatchesCountOnly      bool
	MismatchesCountOnly   bool
	port                  int
	logger                *utils.Logger
	EpSearchNode          *model.EpSearchNode
	router                *mux.Router
	server                *http.Server
	Matches               map[string][]*model.Match
	MatchesCount          map[string]int64
	Mismatches            []*model.Mismatch
	MismatchesCount       int64
	ProxyConfigRouterPath string
	ConfigRouterHost      string
	httpClient            http.Client
}

func NewMockRouter(mockDir, mockFilepattern, responseDir string, port int, kvstore *kvstore.KVStore, matchesCountOnly, mismatchesCountOnly bool, proxyConfigRouterPath string, configRouterPort int, httpClientTimeout time.Duration, logger *utils.Logger) *MockRouter {
	mockRouter := &MockRouter{
		mockDir:               mockDir,
		mockFilepattern:       mockFilepattern,
		responseDir:           responseDir,
		MatchesCountOnly:      matchesCountOnly,
		MismatchesCountOnly:   mismatchesCountOnly,
		port:                  port,
		logger:                logger,
		EpSearchNode:          &model.EpSearchNode{},
		Matches:               make(map[string][]*model.Match),
		MatchesCount:          make(map[string]int64),
		ProxyConfigRouterPath: proxyConfigRouterPath,
		ConfigRouterHost:      fmt.Sprintf("localhost:%d", configRouterPort),
		httpClient:            utils.CreateHttpClient(httpClientTimeout),
	}
	return mockRouter
}

func (r *MockRouter) Name() string {
	return "Mockrouter"
}

func (r *MockRouter) Router() *mux.Router {
	return r.router
}

func (r *MockRouter) Server() *http.Server {
	return r.server
}

func (r *MockRouter) Port() int {
	return r.port
}

func (r *MockRouter) Logger() *utils.Logger {
	return r.logger
}

func (r *MockRouter) LoadFiles(funcMap template.FuncMap) error {
	r.EpSearchNode = &model.EpSearchNode{}
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
			r.initResponseTemplates(endpoint, funcMap)
			r.registerEndpoint(endpoint)
		}
	}
	r.newRouter()

	return nil
}

func (r *MockRouter) readMockFile(mockFile string) (*model.Mock, error) {
	r.logger.LogWhenVerbose(fmt.Sprintf("Reading mock file '%s' ...", mockFile))
	mockFileContent, err := ioutil.ReadFile(mockFile)
	if err != nil {
		return nil, err
	}

	var mock model.Mock
	if strings.HasSuffix(mockFile, ".yaml") || strings.HasSuffix(mockFile, ".yml") {
		err = yaml.Unmarshal(mockFileContent, &mock)
	}
	if err != nil {
		return nil, err
	}
	if len(mock.Name) == 0 {
		mock.Name = filepath.Base(mockFile)
	}
	for _, endpoint := range mock.Endpoints {
		if len(endpoint.Request.Body) > 0 {
			bodyregexp, err := regexp.Compile(endpoint.Request.Body)
			if err != nil {
				return nil, err
			}
			endpoint.Request.BodyRegexp = bodyregexp
		}
	}
	return &mock, nil
}

func (r *MockRouter) initResponseTemplates(endpoint *model.MockEndpoint, funcMap template.FuncMap) error {
	endpoint.Response.Template = template.New(endpoint.Id).Funcs(sprig.TxtFuncMap()).Funcs(funcMap)
	body := ""
	if len(endpoint.Response.Body) > 0 {
		if len(endpoint.Response.BodyFilename) > 0 {
			return errors.New("error parsing endpoint id '%s' , response.body and response.bodyFilename can't be defined both")
		}
		body = endpoint.Response.Body
	} else {
		if len(endpoint.Response.BodyFilename) > 0 {
			bodyBytes, err := ioutil.ReadFile(filepath.Join(r.responseDir, endpoint.Response.BodyFilename))
			if err != nil {
				return err
			}
			body = string(bodyBytes)
		}
	}
	_, err := endpoint.Response.Template.New(TEMPLATE_NAME_RESPONSEBODY).Parse(body)
	if err != nil {
		return err
	}
	if len(endpoint.Response.StatusCode) == 0 {
		endpoint.Response.StatusCode = strconv.Itoa(http.StatusOK)
	}
	_, err = endpoint.Response.Template.New(TEMPLATE_NAME_RESPONSESTATUS).Parse(endpoint.Response.StatusCode)
	if err != nil {
		return err
	}

	_, err = endpoint.Response.Template.New(TEMPLATE_NAME_RESPONSEHEADERS).Parse(endpoint.Response.Headers)
	if err != nil {
		return err
	}

	return nil
}

func (r *MockRouter) newRouter() {
	r.router = mux.NewRouter()
	var endPoint *model.MockEndpoint
	var match *model.Match
	var requestPathParam map[string]string
	isConfigRequest := func(request *http.Request) bool {
		return len(r.ProxyConfigRouterPath) > 0 && strings.HasPrefix(request.URL.Path, r.ProxyConfigRouterPath)
	}
	route := r.router.MatcherFunc(func(request *http.Request, routematch *mux.RouteMatch) bool {
		if isConfigRequest(request) {
			return true
		}
		endPoint, match, requestPathParam = r.matchRequestToEndpoint(request)
		return endPoint != nil
	})
	route.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if isConfigRequest(request) {
			r.directToConfigRouter(writer, request)
		} else {
			r.renderResponse(writer, request, endPoint, match, requestPathParam)
		}
	})
	r.server = &http.Server{Addr: ":" + strconv.Itoa(r.port), Handler: r.router}
}

func (r *MockRouter) directToConfigRouter(writer http.ResponseWriter, request *http.Request) {
	r.logger.LogWhenDebugRR(fmt.Sprintf("directToConfigRouter: incoming request: %s|%s", request.Method, request.URL.String()))
	request.RequestURI = ""
	request.URL.Scheme = "http"
	request.URL.Host = r.ConfigRouterHost
	pathSegments := strings.Split(request.URL.Path, "/")
	request.URL.Path = "/" + path.Join(pathSegments[2:]...)
	r.logger.LogWhenDebugRR(fmt.Sprintf("directToConfigRouter: calling request: %s|%s", request.Method, request.URL.String()))
	response, err := r.httpClient.Do(request)
	if err != nil {
		http.Error(writer, fmt.Sprintf("error calling configRouter with request: %v, eror: %v", request, err), http.StatusInternalServerError)
		return
	}
	defer response.Body.Close()
	// copy header
	for k, vv := range response.Header {
		for _, v := range vv {
			writer.Header().Add(k, v)
		}
	}
	writer.WriteHeader(response.StatusCode)
	io.Copy(writer, response.Body)
}

func (r *MockRouter) registerEndpoint(endpoint *model.MockEndpoint) {
	if endpoint.Request.Path == "" {
		endpoint.Request.Path = "/"
	}
	if endpoint.Request.Method == "" {
		endpoint.Request.Method = "GET"
	}

	sn := r.EpSearchNode
	pathSegments := strings.Split(endpoint.Request.Path, "/")
	for _, pathSegment := range pathSegments[1:] {
		if sn.SearchNodes == nil {
			sn.SearchNodes = make(map[string]*model.EpSearchNode)
		}
		pathParamName := ""
		if strings.HasPrefix(pathSegment, "{") && strings.HasSuffix(pathSegment, "}") {
			pathParamName = pathSegment[1 : len(pathSegment)-1]
			pathSegment = "*"
		}
		sn.SearchNodes[pathSegment] = &model.EpSearchNode{}
		sn = sn.SearchNodes[pathSegment]
		sn.PathParamName = pathParamName
	}
	if sn.Endpoints == nil {
		sn.Endpoints = make(map[string][]*model.MockEndpoint)
	}

	if sn.Endpoints[endpoint.Request.Method] == nil {
		sn.Endpoints[endpoint.Request.Method] = []*model.MockEndpoint{}
	}
	sn.Endpoints[endpoint.Request.Method] = append(sn.Endpoints[endpoint.Request.Method], endpoint)
	r.logger.LogWhenVerbose(fmt.Sprintf("register endpoint with id '%s' for path|method: %s|%s", endpoint.Id, endpoint.Request.Path, endpoint.Request.Method))
}

func (r *MockRouter) matchRequestToEndpoint(request *http.Request) (*model.MockEndpoint, *model.Match, map[string]string) {
	requestPathParams := map[string]string{}
	sn := r.EpSearchNode
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
		if sn.SearchNodes == nil {
			if allMatch {
				break
			} else {
				r.addMismatch(sn, pos, "", request)
				return nil, nil, requestPathParams
			}
		} else {
			if allMatch {
				for i := pos; pathSegment != ""; i++ {
					if sn.SearchNodes[pathSegment] != nil {
						pos = i
						break
					}
					pathSegment = getPathSegment(pathSegments, i)
				}
				allMatch = false
			}
			if sn.SearchNodes[pathSegment] == nil {
				if sn.SearchNodes["*"] == nil {
					if sn.SearchNodes["**"] == nil {
						r.addMismatch(sn, pos, "", request)
						return nil, nil, requestPathParams
					} else {
						allMatch = true
						sn = sn.SearchNodes["**"]
					}
				} else {
					sn = sn.SearchNodes["*"]
					if len(sn.PathParamName) > 0 {
						requestPathParams[sn.PathParamName] = pathSegment
					}
				}
			} else {
				sn = sn.SearchNodes[pathSegment]
			}
			pathSegment = getPathSegment(pathSegments, pos)
		}
	}
	if sn != nil && sn.Endpoints != nil {
		if sn.Endpoints[request.Method] != nil {
			ep, match := r.matchEndPointsAttributes(sn.Endpoints[request.Method], request)
			return ep, match, requestPathParams
		} else {
			r.addMismatch(nil, -1, fmt.Sprintf("no endpoint found with method '%s'", request.Method), request)
			return nil, nil, requestPathParams
		}
	}
	r.addMismatch(sn, math.MaxInt, "", request)
	return nil, nil, requestPathParams
}

func (r *MockRouter) matchEndPointsAttributes(endPoints []*model.MockEndpoint, request *http.Request) (*model.MockEndpoint, *model.Match) {
	mismatchMessage := ""
	for _, ep := range endPoints {
		if !r.matchQueryParams(ep.Request, request) {
			mismatchMessage = mismatchMessage + fmt.Sprintf(", endpointId '%s' not matched because of wanted query params: %v", ep.Id, ep.Request.Query)
			continue
		}
		if !r.matchHeaderValues(ep.Request, request) {
			mismatchMessage = mismatchMessage + fmt.Sprintf(", endpointId '%s' not matched because of wanted header: %v", ep.Id, ep.Request.Headers)
			continue
		}
		if !r.matchBody(ep.Request, request) {
			mismatchMessage = mismatchMessage + fmt.Sprintf(", endpointId '%s' not matched because of wanted body: '%s'", ep.Id, ep.Request.Body)
			continue
		}
		match := r.addMatch(ep, request)
		return ep, match
	}
	r.addMismatch(nil, -1, mismatchMessage, request)
	return nil, nil
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

func (r *MockRouter) matchBody(matchRequest *model.MatchRequest, request *http.Request) bool {
	if matchRequest.BodyRegexp != nil {
		if request.Body == nil {
			return false
		}
		reqBodyBytes, err := ioutil.ReadAll(request.Body)
		if err != nil {
			r.logger.LogAlways(fmt.Sprintf("No match, error reading request body: %v", err))
			return false
		}
		return matchRequest.BodyRegexp.Match(reqBodyBytes)
	} else {
		return true
	}
}

func (r *MockRouter) addMatch(endPoint *model.MockEndpoint, request *http.Request) *model.Match {
	r.MatchesCount[endPoint.Id] = r.MatchesCount[endPoint.Id] + 1
	actualRequest := &model.ActualRequest{Method: request.Method, URL: request.URL.String(), Header: request.Header, Host: request.Host}
	match := &model.Match{EndpointId: endPoint.Id, Timestamp: time.Now(), ActualRequest: actualRequest}
	if !r.MatchesCountOnly {
		r.Matches[endPoint.Id] = append(r.Matches[endPoint.Id], match)
	}
	return match
}

func (r *MockRouter) addMismatch(sn *model.EpSearchNode, pathPos int, endpointMismatchDetails string, request *http.Request) {
	r.MismatchesCount++
	if !r.MismatchesCountOnly {
		var mismatchDetails string
		if sn == nil { // node found -> path matched
			mismatchDetails = fmt.Sprintf("path '%s' matched, but %s", request.URL.Path, endpointMismatchDetails)
		} else {
			var matchedSubPath string
			if pathPos == math.MaxInt {
				matchedSubPath = request.URL.Path
			} else {
				pathSegments := strings.Split(request.URL.Path, "/")[1:]
				matchedSubPath = strings.Join(pathSegments[:pathPos-1], "/")
			}
			mismatchDetails = fmt.Sprintf("path '%s' not matched, subpath which matched: '%s'", request.URL.Path, matchedSubPath)
		}
		actualRequest := &model.ActualRequest{Method: request.Method, URL: request.URL.String(), Header: request.Header, Host: request.Host}
		mismatch := &model.Mismatch{
			MismatchDetails: mismatchDetails,
			Timestamp:       time.Now(),
			ActualRequest:   actualRequest}
		r.Mismatches = append(r.Mismatches, mismatch)
	}
}

func (r *MockRouter) renderResponse(writer http.ResponseWriter, request *http.Request, endpoint *model.MockEndpoint, match *model.Match, requestPathParams map[string]string) {
	writer.Header().Add(HEADER_KEY_ENDPOINT_ID, endpoint.Id)
	responseTemplateData, err := r.createResponseTemplateData(request, requestPathParams)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(writer, "Error rendering response: %v", err)
		return
	}

	var renderedHeaders bytes.Buffer
	err = endpoint.Response.Template.ExecuteTemplate(&renderedHeaders, TEMPLATE_NAME_RESPONSEHEADERS, responseTemplateData)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(writer, "Error rendering response headers: %v", err)
		return
	}

	var headers map[string]string
	err = yaml.Unmarshal(renderedHeaders.Bytes(), &headers)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(writer, "Error unmarshalling response headers: %v", err)
		return
	}
	for key, val := range headers {
		writer.Header().Add(key, val)
	}

	var renderedStatus bytes.Buffer
	err = endpoint.Response.Template.ExecuteTemplate(&renderedStatus, TEMPLATE_NAME_RESPONSESTATUS, responseTemplateData)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(writer, "Error rendering response status: %v", err)
		return
	}
	responseStatus, err := strconv.Atoi(renderedStatus.String())
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(writer, "Error converting response status: %v", err)
		return
	}
	writer.WriteHeader(responseStatus)

	var renderedBody bytes.Buffer
	err = endpoint.Response.Template.ExecuteTemplate(&renderedBody, TEMPLATE_NAME_RESPONSEBODY, responseTemplateData)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(writer, "Error rendering response body: %v", err)
		return
	}

	_, err = writer.Write(renderedBody.Bytes())
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(writer, "Error writing response body: %v", err)
		return
	}

	match.ActualResponse = &model.ActualResponse{StatusCode: responseStatus, Header: headers}
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
		bodyData := &map[string]interface{}{}
		err = json.Unmarshal(body.Bytes(), bodyData)
		if err == nil { // ignore when no json
			data.RequestBodyJsonData = *bodyData
		}
	}
	return data, nil
}
