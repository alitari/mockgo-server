package mock

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/Masterminds/sprig"
	"github.com/alitari/mockgo-server/mockgo/logging"
	"github.com/alitari/mockgo-server/mockgo/matches"
	"github.com/prometheus/client_golang/prometheus"

	"gopkg.in/yaml.v2"

	"github.com/gorilla/mux"
)

const templateResponseBody = "responseBody"
const templateResponseStatus = "responseStatus"
const templateResponseHeader = "responseHeader"

const headerKeyEndpointId = "endpoint-Id"

type ResponseTemplateData struct {
	RequestPathParams   map[string]string
	RequestQueryParams  map[string]string
	KVStore             map[string]interface{}
	RequestUrl          string
	RequestPath         string
	RequestHost         string
	RequestBody         string
	RequestBodyJsonData map[string]interface{}
}

type MockRequestHandler struct {
	mockDir         string
	mockFilepattern string
	logger          *logging.LoggerUtil
	EpSearchNode    *EpSearchNode
	matchstore      matches.Matchstore
}

func NewMockRequestHandler(mockDir, mockFilepattern string, matchstore matches.Matchstore, logger *logging.LoggerUtil) *MockRequestHandler {
	mockRouter := &MockRequestHandler{
		mockDir:         mockDir,
		mockFilepattern: mockFilepattern,
		logger:          logger,
		EpSearchNode:    &EpSearchNode{},
		matchstore:      matchstore,
	}
	return mockRouter
}

var (
	matchesMetric = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "matches",
			Help: "Number of matches of an endpoint",
		},
		[]string{"endpoint"},
	)
	mismatchesMetric = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "mismatches",
			Help: "Number of mismatches",
		},
	)
)

func (r *MockRequestHandler) RegisterMetrics() error {
	if err := prometheus.Register(matchesMetric); err != nil {
		return err
	}
	if err := prometheus.Register(mismatchesMetric); err != nil {
		return err
	}
	return nil
}

func (r *MockRequestHandler) LoadFiles(funcMap template.FuncMap) error {
	r.EpSearchNode = &EpSearchNode{}
	endPointCounter := 0
	mockFiles, err := walkMatch(r.mockDir, r.mockFilepattern)
	if err != nil {
		return err
	}
	r.logger.LogWhenVerbose(fmt.Sprintf("Found %v mock file(s):", len(mockFiles)))
	for _, mockFile := range mockFiles {
		mock, err := r.readMockFile(mockFile)
		if err != nil {
			return err
		}
		for _, endpoint := range mock.Endpoints {
			endPointCounter++
			if len(endpoint.Id) == 0 {
				endpoint.Id = strconv.Itoa(endPointCounter)
			}
			endpoint.Mock = mock
			err := r.initResponseTemplates(endpoint, funcMap)
			if err != nil {
				return err
			}
			r.registerEndpoint(endpoint)
		}
	}
	return nil
}

func (r *MockRequestHandler) readMockFile(mockFile string) (*Mock, error) {
	r.logger.LogWhenVerbose(fmt.Sprintf("Reading mock file '%s' ...", mockFile))
	mockFileContent, err := os.ReadFile(mockFile)
	if err != nil {
		return nil, err
	}

	var mock Mock
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

func (r *MockRequestHandler) initResponseTemplates(endpoint *MockEndpoint, funcMap template.FuncMap) error {
	endpoint.Response.Template = template.New(endpoint.Id).Funcs(sprig.TxtFuncMap()).Funcs(funcMap)
	body := ""
	if len(endpoint.Response.Body) > 0 {
		if len(endpoint.Response.BodyFilename) > 0 {
			return fmt.Errorf("error parsing endpoint id '%s' , response.body and response.bodyFilename can't be defined both", endpoint.Id)
		}
		body = endpoint.Response.Body
	} else {
		if len(endpoint.Response.BodyFilename) > 0 {
			bodyBytes, err := os.ReadFile(filepath.Join(r.mockDir, endpoint.Response.BodyFilename))
			if err != nil {
				return err
			}
			body = string(bodyBytes)
		}
	}
	_, err := endpoint.Response.Template.New(templateResponseBody).Parse(body)
	if err != nil {
		return err
	}
	if len(endpoint.Response.StatusCode) == 0 {
		endpoint.Response.StatusCode = strconv.Itoa(http.StatusOK)
	}
	_, err = endpoint.Response.Template.New(templateResponseStatus).Parse(endpoint.Response.StatusCode)
	if err != nil {
		return err
	}

	_, err = endpoint.Response.Template.New(templateResponseHeader).Parse(endpoint.Response.Headers)
	if err != nil {
		return err
	}

	return nil
}

func (r *MockRequestHandler) AddRoutes(router *mux.Router) {
	var endPoint *MockEndpoint
	var match *matches.Match
	var requestPathParam map[string]string
	var queryParams map[string]string
	route := router.MatcherFunc(func(request *http.Request, routematch *mux.RouteMatch) bool {
		endPoint, match, requestPathParam, queryParams = r.matchRequestToEndpoint(request)
		return endPoint != nil
	})
	route.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		r.logger.LogIncomingRequest(request)
		if r.logger.Level >= logging.Debug {
			writer = logging.NewLoggingResponseWriter(writer, r.logger, 2)
		}
		r.renderResponse(writer, request, endPoint, match, requestPathParam, queryParams)
		if r.logger.Level >= logging.Debug {
			writer.(*logging.LoggingResponseWriter).Log()
		}
	})
}

func (r *MockRequestHandler) registerEndpoint(endpoint *MockEndpoint) {
	if endpoint.Request.Method == "" {
		endpoint.Request.Method = "GET"
	}

	sn := r.EpSearchNode
	pathSegments := strings.Split(endpoint.Request.Path, "/")
	for _, pathSegment := range pathSegments[1:] {
		if sn.SearchNodes == nil {
			sn.SearchNodes = make(map[string]*EpSearchNode)
		}
		pathParamName := ""
		if strings.HasPrefix(pathSegment, "{") && strings.HasSuffix(pathSegment, "}") {
			pathParamName = pathSegment[1 : len(pathSegment)-1]
			pathSegment = "*"
		}
		if sn.SearchNodes[pathSegment] == nil {
			sn.SearchNodes[pathSegment] = &EpSearchNode{}
		}
		sn = sn.SearchNodes[pathSegment]
		sn.PathParamName = pathParamName
	}
	if sn.Endpoints == nil {
		sn.Endpoints = make(map[string][]*MockEndpoint)
	}

	if sn.Endpoints[endpoint.Request.Method] == nil {
		sn.Endpoints[endpoint.Request.Method] = []*MockEndpoint{}
	}
	insertIndex := 0
	for i, ep := range sn.Endpoints[endpoint.Request.Method] {
		if endpoint.Prio > ep.Prio {
			insertIndex = i
			break
		}
	}
	if len(sn.Endpoints[endpoint.Request.Method]) == insertIndex {
		sn.Endpoints[endpoint.Request.Method] = append(sn.Endpoints[endpoint.Request.Method], endpoint)
	} else {
		sn.Endpoints[endpoint.Request.Method] = append(sn.Endpoints[endpoint.Request.Method][:insertIndex+1], sn.Endpoints[endpoint.Request.Method][insertIndex:]...)
		sn.Endpoints[endpoint.Request.Method][insertIndex] = endpoint
	}
	r.logger.LogWhenVerbose(fmt.Sprintf("register endpoint with id '%s' for path|method: %s|%s", endpoint.Id, endpoint.Request.Path, endpoint.Request.Method))
}

func getPathSegment(segments []string, pos int) string {
	if pos < len(segments) {
		return segments[pos]
	} else {
		return ""
	}
}

func (r *MockRequestHandler) matchRequestToEndpoint(request *http.Request) (*MockEndpoint, *matches.Match, map[string]string, map[string]string) {
	requestPathParams := map[string]string{}
	queryParams := map[string]string{}

	for k, v := range request.URL.Query() {
		queryParams[k] = v[0]
	}

	sn := r.EpSearchNode
	pathSegments := strings.Split(request.URL.Path, "/")[1:]
	allMatch := false
	pathSegment := getPathSegment(pathSegments, 0)
	for pos := 1; pathSegment != ""; pos++ {
		if sn.SearchNodes == nil {
			if allMatch {
				break
			} else {
				r.addMismatch(sn, pos, "", request)
				return nil, nil, requestPathParams, queryParams
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
						return nil, nil, requestPathParams, queryParams
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
			return ep, match, requestPathParams, queryParams
		} else {
			r.addMismatch(nil, -1, fmt.Sprintf("no endpoint found with method '%s'", request.Method), request)
			return nil, nil, requestPathParams, queryParams
		}
	}
	r.addMismatch(sn, math.MaxInt, "", request)
	return nil, nil, requestPathParams, queryParams
}

func (r *MockRequestHandler) matchEndPointsAttributes(endPoints []*MockEndpoint, request *http.Request) (*MockEndpoint, *matches.Match) {
	mismatchMessage := ""
	// sort.SliceStable(endPoints, func(i, j int) bool {
	// 	return endPoints[i].Prio > endPoints[j].Prio
	// })
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

func (r *MockRequestHandler) matchQueryParams(matchRequest *MatchRequest, request *http.Request) bool {
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

func (r *MockRequestHandler) matchHeaderValues(matchRequest *MatchRequest, request *http.Request) bool {
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

func (r *MockRequestHandler) matchBody(matchRequest *MatchRequest, request *http.Request) bool {
	if matchRequest.BodyRegexp != nil {
		reqBodyBytes, err := io.ReadAll(request.Body)
		if err != nil {
			r.logger.LogError("no match, error reading request body", err)
			return false
		}
		return matchRequest.BodyRegexp.Match(reqBodyBytes)
	} else {
		return true
	}
}

func (r *MockRequestHandler) addMatch(endPoint *MockEndpoint, request *http.Request) *matches.Match {
	actualRequest := &matches.ActualRequest{Method: request.Method, URL: request.URL.String(), Header: request.Header, Host: request.Host}
	match := &matches.Match{EndpointId: endPoint.Id, Timestamp: time.Now(), ActualRequest: actualRequest}
	r.matchstore.AddMatch(endPoint.Id, match)
	matchesMetric.With(prometheus.Labels{"endpoint": endPoint.Id}).Inc()
	return match
}

func (r *MockRequestHandler) addMismatch(sn *EpSearchNode, pathPos int, endpointMismatchDetails string, request *http.Request) {
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
	actualRequest := &matches.ActualRequest{Method: request.Method, URL: request.URL.String(), Header: request.Header, Host: request.Host}
	mismatch := &matches.Mismatch{
		MismatchDetails: mismatchDetails,
		Timestamp:       time.Now(),
		ActualRequest:   actualRequest}
	r.matchstore.AddMismatch(mismatch)
	mismatchesMetric.Inc()
}

func (r *MockRequestHandler) renderResponse(writer http.ResponseWriter, request *http.Request, endpoint *MockEndpoint, match *matches.Match, requestPathParams, queryParams map[string]string) {
	writer.Header().Add(headerKeyEndpointId, endpoint.Id)
	responseTemplateData, err := r.createResponseTemplateData(request, requestPathParams, queryParams)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(writer, "Error rendering response: %v", err)
		return
	}

	var renderedHeaders bytes.Buffer
	err = endpoint.Response.Template.ExecuteTemplate(&renderedHeaders, templateResponseHeader, responseTemplateData)
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
	err = endpoint.Response.Template.ExecuteTemplate(&renderedStatus, templateResponseStatus, responseTemplateData)
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

	var renderedBody bytes.Buffer
	err = endpoint.Response.Template.ExecuteTemplate(&renderedBody, templateResponseBody, responseTemplateData)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(writer, "Error rendering response body: %v", err)
		return
	}
	writer.WriteHeader(responseStatus)
	writer.Write(renderedBody.Bytes())

	//TODO: handle headers
	match.ActualResponse = &matches.ActualResponse{StatusCode: responseStatus, Header: make(map[string][]string)}
}

func (r *MockRequestHandler) createResponseTemplateData(request *http.Request, requestPathParams, queryParams map[string]string) (*ResponseTemplateData, error) {
	data := &ResponseTemplateData{
		RequestUrl:         request.URL.String(),
		RequestPathParams:  requestPathParams,
		RequestQueryParams: queryParams,
		RequestPath:        request.URL.Path,
		RequestHost:        request.URL.Host,
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

func walkMatch(root, pattern string) ([]string, error) {
	var matches []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		info.Mode()
		if matched, err := filepath.Match(pattern, filepath.Base(path)); err != nil {
			return err
		} else if matched {
			if filepath.IsAbs(root) {
				if filepath.Dir(path) == root {
					matches = append(matches, path)
				} else {
					return nil
				}
			} else {
				matches = append(matches, path)
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return matches, nil
}
