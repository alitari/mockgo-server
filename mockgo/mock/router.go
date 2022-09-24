package mock

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/Masterminds/sprig"
	"github.com/alitari/mockgo/logging"
	"github.com/alitari/mockgo/matches"
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
	mockDir         string
	mockFilepattern string
	responseDir     string
	logger          *logging.LoggerUtil
	EpSearchNode    *EpSearchNode
	matchstore      matches.Matchstore
	router          *mux.Router
	ProxyPrefixPath string
	ProxyForHost    string
	httpClient      http.Client
}

func NewMockRouter(mockDir, mockFilepattern, responseDir string, matchstore matches.Matchstore, proxyPrefixPath string, proxyForHost string, httpClientTimeout time.Duration, logger *logging.LoggerUtil) *MockRouter {
	mockRouter := &MockRouter{
		mockDir:         mockDir,
		mockFilepattern: mockFilepattern,
		responseDir:     responseDir,
		logger:          logger,
		EpSearchNode:    &EpSearchNode{},
		matchstore:      matchstore,
		ProxyPrefixPath: proxyPrefixPath,
		ProxyForHost:    proxyForHost,
		httpClient:      createHttpClient(httpClientTimeout),
	}
	return mockRouter
}

func (r *MockRouter) LoadFiles(funcMap template.FuncMap) error {
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

func (r *MockRouter) readMockFile(mockFile string) (*Mock, error) {
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

func (r *MockRouter) initResponseTemplates(endpoint *MockEndpoint, funcMap template.FuncMap) error {
	endpoint.Response.Template = template.New(endpoint.Id).Funcs(sprig.TxtFuncMap()).Funcs(funcMap)
	body := ""
	if len(endpoint.Response.Body) > 0 {
		if len(endpoint.Response.BodyFilename) > 0 {
			return errors.New("error parsing endpoint id '%s' , response.body and response.bodyFilename can't be defined both")
		}
		body = endpoint.Response.Body
	} else {
		if len(endpoint.Response.BodyFilename) > 0 {
			bodyBytes, err := os.ReadFile(filepath.Join(r.responseDir, endpoint.Response.BodyFilename))
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
	var endPoint *MockEndpoint
	var match *matches.Match
	var requestPathParam map[string]string
	isProxyRequest := func(request *http.Request) bool {
		return len(r.ProxyPrefixPath) > 0 && strings.HasPrefix(request.URL.Path, r.ProxyPrefixPath)
	}
	route := r.router.MatcherFunc(func(request *http.Request, routematch *mux.RouteMatch) bool {

		if isProxyRequest(request) {
			return true
		}
		endPoint, match, requestPathParam = r.matchRequestToEndpoint(request)
		return endPoint != nil
	})
	route.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		r.logger.LogIncomingRequest(request)
		if r.logger.Level >= logging.Debug {
			writer = logging.NewLoggingResponseWriter(writer, r.logger, 2)
		}
		if isProxyRequest(request) {
			r.directToProxyForHost(writer, request)
		} else {
			r.renderResponse(writer, request, endPoint, match, requestPathParam)
		}
		if r.logger.Level >= logging.Debug {
			writer.(*logging.LoggingResponseWriter).Log()
		}

	})
}

func (r *MockRouter) directToProxyForHost(writer http.ResponseWriter, request *http.Request) {
	r.logger.LogWhenDebug(fmt.Sprintf("directToConfigRouter: incoming request: %s|%s", request.Method, request.URL.String()))
	request.RequestURI = ""
	request.URL.Scheme = "http"
	request.URL.Host = r.ProxyForHost
	pathSegments := strings.Split(request.URL.Path, "/")
	request.URL.Path = "/" + path.Join(pathSegments[2:]...)
	r.logger.LogWhenDebug(fmt.Sprintf("directToConfigRouter: calling request: %s|%s", request.Method, request.URL.String()))
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

func (r *MockRouter) registerEndpoint(endpoint *MockEndpoint) {
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
			sn.SearchNodes = make(map[string]*EpSearchNode)
		}
		pathParamName := ""
		if strings.HasPrefix(pathSegment, "{") && strings.HasSuffix(pathSegment, "}") {
			pathParamName = pathSegment[1 : len(pathSegment)-1]
			pathSegment = "*"
		}
		sn.SearchNodes[pathSegment] = &EpSearchNode{}
		sn = sn.SearchNodes[pathSegment]
		sn.PathParamName = pathParamName
	}
	if sn.Endpoints == nil {
		sn.Endpoints = make(map[string][]*MockEndpoint)
	}

	if sn.Endpoints[endpoint.Request.Method] == nil {
		sn.Endpoints[endpoint.Request.Method] = []*MockEndpoint{}
	}
	sn.Endpoints[endpoint.Request.Method] = append(sn.Endpoints[endpoint.Request.Method], endpoint)
	r.logger.LogWhenVerbose(fmt.Sprintf("register endpoint with id '%s' for path|method: %s|%s", endpoint.Id, endpoint.Request.Path, endpoint.Request.Method))
}

func (r *MockRouter) matchRequestToEndpoint(request *http.Request) (*MockEndpoint, *matches.Match, map[string]string) {
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

func (r *MockRouter) matchEndPointsAttributes(endPoints []*MockEndpoint, request *http.Request) (*MockEndpoint, *matches.Match) {
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

func (r *MockRouter) matchQueryParams(matchRequest *MatchRequest, request *http.Request) bool {
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

func (r *MockRouter) matchHeaderValues(matchRequest *MatchRequest, request *http.Request) bool {
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

func (r *MockRouter) matchBody(matchRequest *MatchRequest, request *http.Request) bool {
	if matchRequest.BodyRegexp != nil {
		if request.Body == nil {
			return false
		}
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

func (r *MockRouter) addMatch(endPoint *MockEndpoint, request *http.Request) *matches.Match {
	actualRequest := &matches.ActualRequest{Method: request.Method, URL: request.URL.String(), Header: request.Header, Host: request.Host}
	match := &matches.Match{EndpointId: endPoint.Id, Timestamp: time.Now(), ActualRequest: actualRequest}
	if r.matchstore.HasMatchesCountOnly() {
		r.matchstore.AddMatchesCount(map[string]int64{endPoint.Id: 1})
	} else {
		r.matchstore.AddMatches(map[string][]*matches.Match{endPoint.Id: {match}})
	}
	return match
}

func (r *MockRouter) addMismatch(sn *EpSearchNode, pathPos int, endpointMismatchDetails string, request *http.Request) {
	if r.matchstore.HasMismatchesCountOnly() {
		r.matchstore.AddMismatchesCount(1)
	} else {
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
		r.matchstore.AddMismatches([]*matches.Mismatch{mismatch})
	}
}

func (r *MockRouter) renderResponse(writer http.ResponseWriter, request *http.Request, endpoint *MockEndpoint, match *matches.Match, requestPathParams map[string]string) {
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

	match.ActualResponse = &matches.ActualResponse{StatusCode: responseStatus, Header: headers}
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

func createHttpClient(timeout time.Duration) http.Client {
	httpClient := http.Client{Timeout: timeout}
	return httpClient
}
