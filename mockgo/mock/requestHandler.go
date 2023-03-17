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

const headerKeyEndpointID = "endpoint-Id"

type responseTemplateData struct {
	RequestPathParams   map[string]string
	RequestQueryParams  map[string]string
	RequestHeader       map[string]string
	KVStore             map[string]interface{}
	RequestURL          string
	RequestPath         string
	RequestHost         string
	RequestBody         string
	RequestBodyJSONData map[string]interface{}
}

/*
RequestHandler implements an http server for mock endpoints
*/
type RequestHandler struct {
	mockDir         string
	mockFilepattern string
	logger          *logging.LoggerUtil
	EpSearchNode    *epSearchNode
	matchstore      matches.Matchstore
}

/*
NewRequestHandler creates an instance of RequestHandler
*/
func NewRequestHandler(mockDir, mockFilepattern string, matchstore matches.Matchstore, logger *logging.LoggerUtil) *RequestHandler {
	mockRouter := &RequestHandler{
		mockDir:         mockDir,
		mockFilepattern: mockFilepattern,
		logger:          logger,
		EpSearchNode:    &epSearchNode{},
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

func (r *RequestHandler) registerMetrics() error {
	if err := prometheus.Register(matchesMetric); err != nil {
		return err
	}
	if err := prometheus.Register(mismatchesMetric); err != nil {
		return err
	}
	return nil
}

/*
LoadFiles reads the mockfiles from the mockDir and creates the datamodel for serving mock endpoints for http requests
*/
func (r *RequestHandler) LoadFiles(funcMap template.FuncMap) error {
	r.EpSearchNode = &epSearchNode{}
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
			if len(endpoint.ID) == 0 {
				endpoint.ID = strconv.Itoa(endPointCounter)
			}
			endpoint.Mock = mock
			err := r.initResponseTemplates(endpoint, funcMap)
			if err != nil {
				return err
			}
			r.registerEndpoint(endpoint)
		}
	}
	if err := r.registerMetrics(); err != nil {
		return err
	}
	return nil
}

func (r *RequestHandler) readMockFile(mockFile string) (*Mock, error) {
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

func (r *RequestHandler) initResponseTemplates(endpoint *Endpoint, funcMap template.FuncMap) error {
	endpoint.Response.Template = template.New(endpoint.ID).Funcs(sprig.TxtFuncMap()).Funcs(funcMap)
	body := ""
	if len(endpoint.Response.Body) > 0 {
		if len(endpoint.Response.BodyFilename) > 0 {
			return fmt.Errorf("error parsing endpoint id '%s' , response.body and response.bodyFilename can't be defined both", endpoint.ID)
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

/*
AddRoutes adds mux.Routes for the http API to a given mux.Router
*/
func (r *RequestHandler) AddRoutes(router *mux.Router) {
	var endPoint *Endpoint
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
			writer = logging.NewResponseWriter(writer, r.logger, 2)
		}
		r.renderResponse(writer, request, endPoint, match, requestPathParam, queryParams)
		if r.logger.Level >= logging.Debug {
			writer.(*logging.ResponseWriter).Log()
		}
	})
}

func (r *RequestHandler) registerEndpoint(endpoint *Endpoint) {
	if endpoint.Request.Method == "" {
		endpoint.Request.Method = "GET"
	}

	sn := r.EpSearchNode
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
		if sn.searchNodes[pathSegment] == nil {
			sn.searchNodes[pathSegment] = &epSearchNode{}
		}
		sn = sn.searchNodes[pathSegment]
		sn.pathParamName = pathParamName
	}
	endpointKey := endpoint.Request.Method
	if len(endpoint.Request.Host) > 0 {
		endpointKey = "+" + endpointKey + "-" + endpoint.Request.Host
	}
	if sn.endpoints == nil {
		sn.endpoints = make(map[string][]*Endpoint)
	}

	if sn.endpoints[endpointKey] == nil {
		sn.endpoints[endpointKey] = []*Endpoint{}
	}
	insertIndex := 0
	for i, ep := range sn.endpoints[endpointKey] {
		if endpoint.Prio > ep.Prio {
			insertIndex = i
			break
		}
	}
	if len(sn.endpoints[endpointKey]) == insertIndex {
		sn.endpoints[endpointKey] = append(sn.endpoints[endpointKey], endpoint)
	} else {
		sn.endpoints[endpointKey] = append(sn.endpoints[endpointKey][:insertIndex+1], sn.endpoints[endpointKey][insertIndex:]...)
		sn.endpoints[endpointKey][insertIndex] = endpoint
	}
	r.logger.LogWhenVerbose(fmt.Sprintf("register endpoint with id '%s' for path|method: %s|%s", endpoint.ID, endpoint.Request.Path, endpoint.Request.Method))
}

func getPathSegment(segments []string, pos int) string {
	if pos < len(segments) {
		return segments[pos]
	}
	return ""
}

func (r *RequestHandler) matchRequestToEndpoint(request *http.Request) (*Endpoint, *matches.Match, map[string]string, map[string]string) {
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
		if sn.searchNodes == nil {
			if allMatch {
				break
			} else {
				r.addMismatch(sn, pos, "", request)
				return nil, nil, requestPathParams, queryParams
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
						r.addMismatch(sn, pos, "", request)
						return nil, nil, requestPathParams, queryParams
					}
					allMatch = true
					sn = sn.searchNodes["**"]
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
	if sn != nil && sn.endpoints != nil {
		for endpointKey, endpoint := range sn.endpoints {
			if request.Method == endpointKey || "+"+request.Method+"-"+strings.Split(request.Host, ":")[0] == endpointKey {
				ep, match := r.matchEndPointsAttributes(endpoint, request)
				return ep, match, requestPathParams, queryParams
			}
		}
		r.addMismatch(nil, -1, fmt.Sprintf("no endpoint found with method '%s'", request.Method), request)
		return nil, nil, requestPathParams, queryParams
	}
	r.addMismatch(sn, math.MaxInt, "", request)
	return nil, nil, requestPathParams, queryParams
}

func (r *RequestHandler) matchEndPointsAttributes(endPoints []*Endpoint, request *http.Request) (*Endpoint, *matches.Match) {
	mismatchMessage := ""
	// sort.SliceStable(endPoints, func(i, j int) bool {
	// 	return endPoints[i].Prio > endPoints[j].Prio
	// })
	for _, ep := range endPoints {
		if !r.matchQueryParams(ep.Request, request) {
			mismatchMessage = mismatchMessage + fmt.Sprintf(", endpointId '%s' not matched because of wanted query params: %v", ep.ID, ep.Request.Query)
			continue
		}
		if !r.matchHeaderValues(ep.Request, request) {
			mismatchMessage = mismatchMessage + fmt.Sprintf(", endpointId '%s' not matched because of wanted header: %v", ep.ID, ep.Request.Headers)
			continue
		}
		if !r.matchBody(ep.Request, request) {
			mismatchMessage = mismatchMessage + fmt.Sprintf(", endpointId '%s' not matched because of wanted body: '%s'", ep.ID, ep.Request.Body)
			continue
		}
		match := r.addMatch(ep, request)
		return ep, match
	}
	r.addMismatch(nil, -1, mismatchMessage, request)
	return nil, nil
}

func (r *RequestHandler) matchQueryParams(matchRequest *MatchRequest, request *http.Request) bool {
	if len(matchRequest.Query) > 0 {
		for key, val := range matchRequest.Query {
			if request.URL.Query().Get(key) != val {
				return false
			}
		}
		return true
	}
	return true
}

func (r *RequestHandler) matchHeaderValues(matchRequest *MatchRequest, request *http.Request) bool {
	if len(matchRequest.Headers) > 0 {
		for key, val := range matchRequest.Headers {
			if request.Header.Get(key) != val {
				return false
			}
		}
		return true
	}
	return true
}

func (r *RequestHandler) matchBody(matchRequest *MatchRequest, request *http.Request) bool {
	if matchRequest.BodyRegexp != nil {
		reqBodyBytes, err := io.ReadAll(request.Body)
		if err != nil {
			r.logger.LogError("no match, error reading request body", err)
			return false
		}
		return matchRequest.BodyRegexp.Match(reqBodyBytes)
	}
	return true
}

func (r *RequestHandler) addMatch(endPoint *Endpoint, request *http.Request) *matches.Match {
	actualRequest := &matches.ActualRequest{Method: request.Method, URL: request.URL.String(), Header: request.Header, Host: request.Host}
	match := &matches.Match{EndpointID: endPoint.ID, Timestamp: time.Now(), ActualRequest: actualRequest}
	r.matchstore.AddMatch(endPoint.ID, match)
	matchesMetric.With(prometheus.Labels{"endpoint": endPoint.ID}).Inc()
	return match
}

func (r *RequestHandler) addMismatch(sn *epSearchNode, pathPos int, endpointMismatchDetails string, request *http.Request) {
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

func (r *RequestHandler) renderResponse(writer http.ResponseWriter, request *http.Request, endpoint *Endpoint, match *matches.Match, requestPathParams, queryParams map[string]string) {
	writer.Header().Add(headerKeyEndpointID, endpoint.ID)
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

func (r *RequestHandler) createResponseTemplateData(request *http.Request, requestPathParams, queryParams map[string]string) (*responseTemplateData, error) {
	data := &responseTemplateData{
		RequestURL:         request.URL.String(),
		RequestPathParams:  requestPathParams,
		RequestHeader:      make(map[string]string),
		RequestQueryParams: queryParams,
		RequestPath:        request.URL.Path,
		RequestHost:        request.URL.Host,
	}

	for k, v := range request.Header {
		data.RequestHeader[k] = v[0]
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
			data.RequestBodyJSONData = *bodyData
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
