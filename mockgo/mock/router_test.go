package mock

import (
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/alitari/mockgo-server/mockgo/logging"
	"github.com/alitari/mockgo-server/mockgo/matches"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

type matchingTestCase struct {
	name                       string
	request                    *http.Request
	expectedMatchEndpointId    string
	expectedRequestPathParams  map[string]string
	expectedRequestQueryParams map[string]string
}

type mismatchingTestCase struct {
	name                    string
	request                 *http.Request
	expectedMismatchDetails string
}

type renderingTestCase struct {
	name                       string
	requestPathParams          map[string]string
	requestQueryParams         map[string]string
	request                    *http.Request
	response                   *MockResponse
	expectedResponseStatusCode int
	expectedResponseBody       string
	expectedResponseHeader     map[string]string
}

// func TestMain(m *testing.M) {
// 	os.Exit(runAndCheckCoverage("mockrouter", m, 0.65))
// }

// func TestMatchRequestToEndpoint_MatchMinmaxMocks(t *testing.T) {
// 	mockRouter := createMockRouter(t, "minmaxmocks", false, false)
// 	testCases := []*matchingTestCase{
// 		{name: "minimal ", request: createRequest(http.MethodGet, "https://somehost:110/minimal", "", nil, nil), expectedMatchEndpointId: "minimal"},
// 		{name: "minimal more attributes",
// 			request: createRequest(http.MethodGet, "http://doesntcare:7777/minimal", "",
// 				map[string][]string{"Accept": {"Something"}, "Authorization": {"Basic"}}, nil),
// 			expectedMatchEndpointId: "minimal"},

// 		{name: "maximal",
// 			request: createRequest(http.MethodPost, "https://alexkrieg.com/maximal?firstQueryParam=value1&secondQueryParam=value2",
// 				"{\n  \"mybody\": \"is max\"\n}\n",
// 				map[string][]string{"Content-Type": {"application/json"}, "Myheader": {"myheaderValue"}}, nil),
// 			expectedMatchEndpointId: "maximal"},
// 		{name: "maximal header and query superset",
// 			request: createRequest(http.MethodPost, "https://alexkrieg.com/maximal?firstQueryParam=value1&secondQueryParam=value2&thirdQueryParam=value3",
// 				"{\n  \"mybody\": \"is max\"\n}\n",
// 				map[string][]string{"Content-Type": {"application/json"}, "Myheader": {"myheaderValue"}, "Anotherheader": {"anotherheaderValue"}}, nil),
// 			expectedMatchEndpointId: "maximal"},
// 	}
// 	assertMatchRequestToEndpoint(t, mockRouter, testCases)
// }

// func TestMatchRequestToEndpoint_MismatchMinmaxMocks(t *testing.T) {
// 	mockRouter := createMockRouter(t, "minmaxmocks", false, false)
// 	testCases := []*mismatchingTestCase{
// 		{name: "wrong path name", request: &http.Request{URL: &url.URL{Path: "/minimals"}, Method: http.MethodGet},
// 			expectedMismatchDetails: "path '/minimals' not matched, subpath which matched: ''"},
// 		{name: "wrong path length too long", request: &http.Request{URL: &url.URL{Path: "/minimal/foo"}, Method: http.MethodGet},
// 			expectedMismatchDetails: "path '/minimal/foo' not matched, subpath which matched: 'minimal'"},
// 		{name: "wrong path length too short", request: &http.Request{URL: &url.URL{Path: "/"}, Method: http.MethodGet},
// 			expectedMismatchDetails: "path '/' not matched, subpath which matched: '/'"},
// 		{name: "wrong method", request: &http.Request{URL: &url.URL{Path: "/minimal"}, Method: "POST"},
// 			expectedMismatchDetails: "path '/minimal' matched, but no endpoint found with method 'POST'"},
// 		{name: "wrong query params",
// 			request: createRequest(http.MethodPost, "https://alexkrieg.com/maximal?firstQueryParam=value1&thirdQueryParam=value3", "",
// 				map[string][]string{"Content-Type": {"application/json"}, "Myheader": {"myheaderValue"}, "Anotherheader": {"anotherheaderValue"}}, nil),
// 			expectedMismatchDetails: "path '/maximal' matched, but , endpointId 'maximal' not matched because of wanted query params: map[firstQueryParam:value1 secondQueryParam:value2]"},
// 		{name: "wrong header value",
// 			request: createRequest(http.MethodPost, "https://alexkrieg.com/maximal?firstQueryParam=value1&secondQueryParam=value2", "",
// 				map[string][]string{"Content-Type": {"application/json"}, "Myheader": {"myheaderValueWrong"}, "Anotherheader": {"anotherheaderValue"}}, nil),
// 			expectedMismatchDetails: "path '/maximal' matched, but , endpointId 'maximal' not matched because of wanted header: map[Content-Type:application/json Myheader:myheaderValue]"},
// 		{name: "wrong body",
// 			request: createRequest(http.MethodPost, "https://alexkrieg.com/maximal?firstQueryParam=value1&secondQueryParam=value2", "",
// 				map[string][]string{"Content-Type": {"application/json"}, "Myheader": {"myheaderValue"}, "Anotherheader": {"anotherheaderValue"}}, nil),
// 			expectedMismatchDetails: "path '/maximal' matched, but , endpointId 'maximal' not matched because of wanted body: '{\n  \"mybody\": \"is max\"\n}\n'"},
// 	}

// 	assertMismatchRequestToEndpoint(t, mockRouter, testCases)

// }

// func TestMatchRequestToEndpoint_MatchWildcards(t *testing.T) {
// 	mockRouter := createMockRouter(t, "wildcardmocks", false, false)

// 	testCases := []*matchingTestCase{
// 		{name: "Single wildcard 1 ", request: &http.Request{URL: &url.URL{Path: "/wildcard/bar/foo"}, Method: http.MethodGet}, expectedMatchEndpointId: "single"},
// 		{name: "Single wildcard 2", request: &http.Request{URL: &url.URL{Path: "/wildcard/foo/foo"}, Method: http.MethodGet}, expectedMatchEndpointId: "single"},
// 		// {name: "Single wildcard No match, first path segment", request: &http.Request{URL: &url.URL{Path: "/wildcards/bar/foo"}, Method: http.MethodGet}, expectedMatchEndpointId: ""},
// 		// {name: "Single wildcard No match, path too long ", request: &http.Request{URL: &url.URL{Path: "/wildcard/bar/foo/toolong"}, Method: http.MethodGet}, expectedMatchEndpointId: ""},
// 		// {name: "Single wildcard No match, path too short ", request: &http.Request{URL: &url.URL{Path: "/bar/foo"}, Method: http.MethodGet}, expectedMatchEndpointId: ""},
// 		{name: "Multi wildcard", request: &http.Request{URL: &url.URL{Path: "/multiwildcard/bar/foo/bar"}, Method: http.MethodGet}, expectedMatchEndpointId: "multi"},
// 	}

// 	assertMatchRequestToEndpoint(t, mockRouter, testCases)
// }

// func TestMatchRequestToEndpoint_MismatchWildcards(t *testing.T) {
// 	mockRouter := createMockRouter(t, "wildcardmocks", false, false)
// 	testCases := []*mismatchingTestCase{
// 		{name: "first path segment", request: createRequest(http.MethodGet, "http://somehost/wildcards/bar/foo", "", nil, nil),
// 			expectedMismatchDetails: "path '/wildcards/bar/foo' not matched, subpath which matched: ''"},
// 		{name: "path too long", request: createRequest(http.MethodGet, "http://somehost/wildcard/bar/foo/toolong", "", nil, nil),
// 			expectedMismatchDetails: "path '/wildcard/bar/foo/toolong' not matched, subpath which matched: 'wildcard/bar/foo'"},
// 		{name: "path too short", request: createRequest(http.MethodGet, "http://somehost/bar/foo", "", nil, nil),
// 			expectedMismatchDetails: "path '/bar/foo' not matched, subpath which matched: ''"},
// 	}
// 	assertMismatchRequestToEndpoint(t, mockRouter, testCases)
// }

// func TestMatchRequestToEndpoint_MatchAllMatchWildcardmocks(t *testing.T) {
// 	mockRouter := createMockRouter(t, "allMatchWildcardMocks", false, false)
// 	testCases := []*matchingTestCase{
// 		{name: "the end 1", request: &http.Request{URL: &url.URL{Path: "/allmatchwildcardAtTheEnd/bar"}, Method: http.MethodGet}, expectedMatchEndpointId: "1"},
// 		{name: "the end 2", request: &http.Request{URL: &url.URL{Path: "/allmatchwildcardAtTheEnd/foo"}, Method: http.MethodGet}, expectedMatchEndpointId: "1"},
// 		{name: "path longer ", request: &http.Request{URL: &url.URL{Path: "/allmatchwildcardAtTheEnd/foo/bar"}, Method: http.MethodGet}, expectedMatchEndpointId: "1"},
// 		{name: "in the middle 1", request: &http.Request{URL: &url.URL{Path: "/allmatchwildcardInTheMiddle/bar/foo"}, Method: http.MethodGet}, expectedMatchEndpointId: "2"},
// 		{name: "in the middle 2", request: &http.Request{URL: &url.URL{Path: "/allmatchwildcardInTheMiddle/bar/ext/foo"}, Method: http.MethodGet}, expectedMatchEndpointId: "2"},
// 		{name: "in the middle 3", request: &http.Request{URL: &url.URL{Path: "/allmatchwildcardInTheMiddle/bar/ext/rem/foo"}, Method: http.MethodGet}, expectedMatchEndpointId: "2"},
// 		{name: "combined wildcards single segment ", request: &http.Request{URL: &url.URL{Path: "/combinedwildcards1/bar/foo/ext"}, Method: http.MethodGet}, expectedMatchEndpointId: "3"},
// 		{name: "combined wildcards multiple segment", request: &http.Request{URL: &url.URL{Path: "/combinedwildcards1/bar/a/b/c/foo/d"}, Method: http.MethodGet}, expectedMatchEndpointId: "3"},
// 	}
// 	assertMatchRequestToEndpoint(t, mockRouter, testCases)
// }

// func TestMatchRequestToEndpoint_MismatchAllMatchWildcardmocks(t *testing.T) {
// 	mockRouter := createMockRouter(t, "allMatchWildcardMocks", false, false)
// 	testCases := []*mismatchingTestCase{
// 		{name: "first path segment", request: &http.Request{URL: &url.URL{Path: "/allmatchwildcardAtTheEnds/foo"}, Method: http.MethodGet},
// 			expectedMismatchDetails: "path '/allmatchwildcardAtTheEnds/foo' not matched, subpath which matched: ''"},
// 		{name: "path shorter", request: &http.Request{URL: &url.URL{Path: "/"}, Method: http.MethodGet},
// 			expectedMismatchDetails: "path '/' not matched, subpath which matched: '/'"},
// 		{name: "allmatchwildcardInTheMiddle endsegments wrong", request: &http.Request{URL: &url.URL{Path: "/allmatchwildcardInTheMiddle/bar/foo/foo"}, Method: http.MethodGet},
// 			expectedMismatchDetails: "path '/allmatchwildcardInTheMiddle/bar/foo/foo' not matched, subpath which matched: 'allmatchwildcardInTheMiddle/bar/foo'"},
// 		{name: "combined wildcards last segment missing", request: &http.Request{URL: &url.URL{Path: "/combinedwildcards1/bar/a/b/c/foo"}, Method: http.MethodGet},
// 			expectedMismatchDetails: "path '/combinedwildcards1/bar/a/b/c/foo' not matched, subpath which matched: '/combinedwildcards1/bar/a/b/c/foo'"},
// 	}
// 	assertMismatchRequestToEndpoint(t, mockRouter, testCases)
// }

// func TestMatchRequestToEndpoint_MatchPathParams(t *testing.T) {
// 	mockRouter := createMockRouter(t, "pathParamsMocks", false, false)
// 	testCases := []*matchingTestCase{
// 		{name: "single",
// 			request:                   &http.Request{URL: &url.URL{Path: "/pathParams/bar/foo"}, Method: http.MethodGet},
// 			expectedMatchEndpointId:   "2",
// 			expectedRequestPathParams: map[string]string{"pathParam": "bar"}},
// 		{name: "multi",
// 			request:                   &http.Request{URL: &url.URL{Path: "/multipathParams/val1/foo/val2"}, Method: http.MethodGet},
// 			expectedMatchEndpointId:   "1",
// 			expectedRequestPathParams: map[string]string{"pathParam1": "val1", "pathParam2": "val2"}},
// 	}
// 	assertMatchRequestToEndpoint(t, mockRouter, testCases)
// }

// func TestMatchRequestToEndpoint_MismatchPathParams(t *testing.T) {
// 	mockRouter := createMockRouter(t, "pathParamsMocks", false, false)
// 	testCases := []*mismatchingTestCase{
// 		{name: "last segment",
// 			request:                 &http.Request{URL: &url.URL{Path: "/pathParams/bar/foos"}, Method: http.MethodGet},
// 			expectedMismatchDetails: "path '/pathParams/bar/foos' not matched, subpath which matched: 'pathParams/bar'"},
// 	}
// 	assertMismatchRequestToEndpoint(t, mockRouter, testCases)
// }

// func TestMatchRequestToEndpoint_Prio(t *testing.T) {
// 	mockRouter := createMockRouter(t, "prioMocks", false, false)
// 	testCases := []*matchingTestCase{
// 		{name: "simple prio",
// 			request:                 &http.Request{URL: &url.URL{Path: "/prio"}, Method: http.MethodGet},
// 			expectedMatchEndpointId: "mustwin"},
// 	}
// 	assertMatchRequestToEndpoint(t, mockRouter, testCases)
// }

// func TestMatchRequestToEndpoint_MatchBodyRegexp(t *testing.T) {
// 	mockRouter := createMockRouter(t, "regexpmocks", false, false)
// 	testCases := []*matchingTestCase{
// 		{name: "regexp 1",
// 			request:                 createRequest(http.MethodPost, "https://mymock.com/regexp1", "{ alex }", nil, nil),
// 			expectedMatchEndpointId: "1"},
// 		{name: "regexp 2",
// 			request:                 createRequest(http.MethodPost, "https://mymock.com/regexp2", "{\n alex \n}", nil, nil),
// 			expectedMatchEndpointId: "2"},
// 		{name: "regexp 3",
// 			request:                 createRequest(http.MethodPost, "https://mymock.com/regexp3", `{ "email": "foo@bar.com" }`, nil, nil),
// 			expectedMatchEndpointId: "3"},
// 	}
// 	assertMatchRequestToEndpoint(t, mockRouter, testCases)
// }

// func TestMatchRequestToEndpoint_MismatchBodyRegexp(t *testing.T) {
// 	mockRouter := createMockRouter(t, "regexpmocks", false, false)
// 	testCases := []*mismatchingTestCase{
// 		{name: "regexp 1",
// 			request:                 createRequest(http.MethodPost, "https://mymock.com/regexp1", "{ alex ", nil, nil),
// 			expectedMismatchDetails: "path '/regexp1' matched, but , endpointId '1' not matched because of wanted body: '^{.*}$'"},
// 	}
// 	assertMismatchRequestToEndpoint(t, mockRouter, testCases)
// }

func TestRenderResponse_Simple(t *testing.T) {
	mockRouter := createMockRouter(t, "responseRendering", false, false)

	expectedResponseResult := `{
    "requestUrl": "https://coolhost.cooldomain.com/coolpath",
    "requestPathParams": "map[myparam1:myvalue]",
    "requestPathParam1": "myvalue",
    "requestPath": "/coolpath",
    "requestHost": "coolhost.cooldomain.com",
    "requestBody": "{ \"requestBodyKey\": \"requestBodyValue\" }",
    "requestBodyKey" : "requestBodyValue"
}`

	testCases := []*renderingTestCase{
		{name: "status", response: &MockResponse{StatusCode: "204"},
			expectedResponseStatusCode: 204},
		{name: "body", response: &MockResponse{Body: "Hello"},
			expectedResponseStatusCode: 200,
			expectedResponseBody:       "Hello"},
		{name: "body from response file", response: &MockResponse{BodyFilename: "simple-response.json"},
			expectedResponseStatusCode: 200,
			expectedResponseBody:       "{\n    \"greets\": \"Hello\"\n}"},
		{name: "template RequestPathParams", response: &MockResponse{Body: "{{ .RequestPathParams.param1 }}"}, requestPathParams: map[string]string{"param1": "Hello"},
			expectedResponseStatusCode: 200,
			expectedResponseBody:       "Hello"},
		{name: "template RequestParam statuscode", response: &MockResponse{StatusCode: "{{ .RequestPathParams.param1 }}"}, requestPathParams: map[string]string{"param1": "202"},
			expectedResponseStatusCode: 202},
		{name: "template RequestParam headers", response: &MockResponse{Headers: "requestParam: {{ .RequestPathParams.param1 }}"}, requestPathParams: map[string]string{"param1": "param1HeaderValue"},
			expectedResponseStatusCode: 200,
			expectedResponseHeader:     map[string]string{"requestParam": "param1HeaderValue"}},
		{name: "template Request url",
			response: &MockResponse{Body: "incoming request url: '{{ .RequestUrl }}'"},
			request: &http.Request{URL: &url.URL{User: url.User("alex"), Scheme: "https", Host: "myhost", Path: "/mypath"},
				Method: http.MethodGet,
				Header: map[string][]string{"headerKey": {"headerValue"}}},
			expectedResponseStatusCode: 200,
			expectedResponseBody:       "incoming request url: 'https://alex@myhost/mypath'"},
		{name: "template response file all request params",
			response:                   &MockResponse{BodyFilename: "request-template-response.json"},
			request:                    createRequest("PUT", "https://coolhost.cooldomain.com/coolpath", "{ \"requestBodyKey\": \"requestBodyValue\" }", map[string][]string{"myheaderKey": {"myheaderValue"}}, nil),
			requestPathParams:          map[string]string{"myparam1": "myvalue"},
			expectedResponseStatusCode: 200,
			expectedResponseBody:       expectedResponseResult},
	}
	assertRenderingResponse(mockRouter, testCases, t)
}

func createRequest(method, url, bodyStr string, header map[string][]string, urlVars map[string]string) *http.Request {
	body := io.NopCloser(strings.NewReader(bodyStr))
	request := httptest.NewRequest(method, url, body)
	request.Header = header
	if urlVars != nil {
		request = mux.SetURLVars(request, urlVars)
	}
	return request
}

func assertRenderingResponse(mockRouter *MockRequestHandler, testCases []*renderingTestCase, t *testing.T) {
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			if testCase.request == nil {
				testCase.request = createRequest(http.MethodGet, "http://host/minimal", "", nil, nil)
			}
			endpoint := &MockEndpoint{Id: testCase.name, Response: testCase.response}
			err := mockRouter.initResponseTemplates(endpoint, nil)
			assert.NoError(t, err, "testcase: '"+testCase.name+"': error in initResonseTemplates")
			mockRouter.renderResponse(recorder, testCase.request, endpoint, &matches.Match{}, testCase.requestPathParams, testCase.requestQueryParams)

			assert.Equal(t, testCase.expectedResponseStatusCode, recorder.Result().StatusCode, "testcase: '"+testCase.name+"': unexpected statuscode")

			responseBody, err := io.ReadAll(recorder.Result().Body)
			assert.NoError(t, err, "testcase: '"+testCase.name+"': error reading responseBody")

			assert.Equal(t, testCase.expectedResponseBody, string(responseBody), "testcase: '"+testCase.name+"': unexpected responseBody")

			if testCase.expectedResponseHeader != nil {
				for expectedParamName, expectedParamValue := range testCase.expectedResponseHeader {
					assert.Equal(t, recorder.Header().Get(expectedParamName), expectedParamValue, "testcase: '"+testCase.name+"': unexpected header param")
				}
			}
			if !t.Failed() {
				t.Logf("testcase '%s':'%s' passed", t.Name(), testCase.name)
			}
		})
	}
}

func assertMatchRequestToEndpoint(t *testing.T, mockRouter *MockRequestHandler, testCases []*matchingTestCase) {
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			timeStamp := time.Now()
			matchCountBefore, err := mockRouter.matchstore.GetMatchesCount(testCase.expectedMatchEndpointId)
			assert.NoError(t, err)
			ep, match, requestPathParams, requestQueryParams := mockRouter.matchRequestToEndpoint(testCase.request)
			matchCountAfter := matchCountBefore + 1
			assert.NotNil(t, match, "expect a match")
			assert.NotNil(t, ep, "for a match, we expect an endpoint")
			assert.Equal(t, testCase.expectedMatchEndpointId, match.EndpointId)
			actualCount, err := mockRouter.matchstore.GetMatchesCount(testCase.expectedMatchEndpointId)
			assert.NoError(t, err)
			assert.Equal(t, matchCountAfter, actualCount, "expect matches are counted")
			currentMatches, err := mockRouter.matchstore.GetMatches(testCase.expectedMatchEndpointId)
			assert.NoError(t, err)
			assert.Len(t, currentMatches, int(matchCountAfter), "expect a mismatch object stored")

			currentMatch := currentMatches[matchCountAfter-1]
			actualRequest := currentMatch.ActualRequest
			assert.Equal(t, testCase.request.Method, actualRequest.Method)
			assert.Equal(t, testCase.request.Host, actualRequest.Host)
			assert.Equal(t, testCase.request.URL.String(), actualRequest.URL)

			assert.Equal(t, testCase.expectedMatchEndpointId, currentMatch.EndpointId)

			assert.LessOrEqual(t, timeStamp, currentMatch.Timestamp)

			if testCase.expectedRequestPathParams != nil {
				for expectedParamName, expectedParamValue := range testCase.expectedRequestPathParams {
					assert.Equal(t, requestPathParams[expectedParamName], expectedParamValue)
				}
			}

			if testCase.expectedRequestQueryParams != nil {
				for expectedParamName, expectedParamValue := range testCase.expectedRequestQueryParams {
					assert.Equal(t, requestQueryParams[expectedParamName], expectedParamValue)
				}
			}
		})
	}
}

func assertMismatchRequestToEndpoint(t *testing.T, mockRouter *MockRequestHandler, testCases []*mismatchingTestCase) {
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			timeStamp := time.Now()
			mismatchCountBefore, err := mockRouter.matchstore.GetMismatchesCount()
			assert.NoError(t, err)
			_, match, _, _ := mockRouter.matchRequestToEndpoint(testCase.request)
			mismatchCountAfter := mismatchCountBefore + 1
			assert.Nil(t, match, "expected not a match")
			mismatchCount, err := mockRouter.matchstore.GetMismatchesCount()
			assert.NoError(t, err)
			assert.Equal(t, mismatchCountAfter, mismatchCount, "expect mismatches are counted")
			mismatches, err := mockRouter.matchstore.GetMismatches()
			assert.NoError(t, err)
			assert.Len(t, mismatches, int(mismatchCountAfter), "expect a mismatch object stored")

			actualRequest := mismatches[mismatchCountAfter-1].ActualRequest
			assert.Equal(t, testCase.request.Method, actualRequest.Method)
			assert.Equal(t, testCase.request.Host, actualRequest.Host)
			assert.Equal(t, testCase.request.URL.String(), actualRequest.URL)

			assert.Equal(t, testCase.expectedMismatchDetails, mismatches[mismatchCountAfter-1].MismatchDetails)

			assert.LessOrEqual(t, timeStamp, mismatches[mismatchCountAfter-1].Timestamp)
		})
	}
}

func createMockRouter(t *testing.T, testMockDir string, matchesCountOnly, mismatchesCountOnly bool) *MockRequestHandler {
	mockRouter := NewMockRequestHandler("../../test/"+testMockDir, "*-mock.yaml", createInMemoryMatchStore(), logging.NewLoggerUtil(logging.Debug))
	assert.NotNil(t, mockRouter, "Mockrouter must not be nil")
	err := mockRouter.LoadFiles(nil)
	assert.NoError(t, err)
	return mockRouter
}

func createInMemoryMatchStore() matches.Matchstore {
	return matches.NewInMemoryMatchstore(uint16(10))
}

func runAndCheckCoverage(testPackage string, m *testing.M, treshold float64) int {

	code := m.Run()

	if code == 0 && testing.CoverMode() != "" {
		coverage := testing.Coverage()
		if coverage < treshold {
			log.Printf("%s tests passed, but coverage must be above %2.2f%%, but it is %2.2f%%\n", testPackage, treshold*100, coverage*100)
			code = -1
		}
	}
	return code
}
