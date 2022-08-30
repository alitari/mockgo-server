package mock

import (
	"encoding/base64"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"text/template"
	"time"

	"github.com/Masterminds/sprig"
	"github.com/alitari/mockgo-server/internal/kvstore"
	"github.com/alitari/mockgo-server/internal/model"
	"github.com/alitari/mockgo-server/internal/utils"
	"github.com/go-http-utils/headers"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

type matchingTestCase struct {
	name                    string
	request                 *http.Request
	expectedMatchEndpointId string
	expectedRequestParams   map[string]string
}

type renderingTestCase struct {
	name                       string
	responseTemplate           string
	requestParams              map[string]string
	request                    *http.Request
	match                      *model.Match
	kvstoreJson                string
	expectedResponseStatusCode int
	expectedResponseBody       string
	expectedResponseHeader     map[string]string
}

func TestMain(m *testing.M) {
	os.Exit(utils.RunAndCheckCoverage("mockrouter", m, 0.65))
}

func TestMatchRequestToEndpoint_Simplemocks(t *testing.T) {
	mockRouter := createMockRouter("simplemocks", t)
	testCases := []*matchingTestCase{
		{name: "Minimal Mock: Match, full request",
			request: &http.Request{
				URL:    &url.URL{Scheme: "https", Host: "myhost", Path: "/minimal"},
				Method: http.MethodGet,
				Header: map[string][]string{"Accept": {"Something"}, "Authorization": {"Basic"}}},
			expectedMatchEndpointId: "minimal"},
		{name: "Minimal Mock: Match, minimal ", request: &http.Request{URL: &url.URL{Path: "/minimal"}, Method: http.MethodGet}, expectedMatchEndpointId: "minimal"},
		{name: "Minimal Mock: No Match, wrong path name", request: &http.Request{URL: &url.URL{Path: "/minimals"}, Method: http.MethodGet}, expectedMatchEndpointId: ""},
		{name: "Minimal Mock: No Match, wrong path length too long", request: &http.Request{URL: &url.URL{Path: "/minimal/foo"}, Method: http.MethodGet}, expectedMatchEndpointId: ""},
		{name: "Minimal Mock: No Match, wrong path length too short", request: &http.Request{URL: &url.URL{Path: "/"}, Method: http.MethodGet}, expectedMatchEndpointId: ""},
		{name: "Minimal Mock: No Match, wrong method", request: &http.Request{URL: &url.URL{Path: "/minimal"}, Method: "POST"}, expectedMatchEndpointId: ""},
		{name: "Maximal Mock: Match, exact",
			request: createRequest(http.MethodPost, "https://alexkrieg.com/maximal?firstQueryParam=value1&secondQueryParam=value2",
				"{\n  \"mybody\": \"is max\"\n}\n",
				map[string][]string{"Content-Type": {"application/json"}, "Myheader": {"myheaderValue"}}, nil, t),
			expectedMatchEndpointId: "maximal"},
		{name: "Maximal Mock: Match, header and query superset",
			request: createRequest(http.MethodPost, "https://alexkrieg.com/maximal?firstQueryParam=value1&secondQueryParam=value2&thirdQueryParam=value3",
				"{\n  \"mybody\": \"is max\"\n}\n",
				map[string][]string{"Content-Type": {"application/json"}, "Myheader": {"myheaderValue"}, "Anotherheader": {"anotherheaderValue"}}, nil, t),
			expectedMatchEndpointId: "maximal"},
		{name: "Maximal Mock: No Match, query subset",
			request: &http.Request{
				URL:    &url.URL{Scheme: "https", Host: "alexkrieg.com", Path: "/maximal", RawQuery: "firstQueryParam=value1&thirdQueryParam=value3"},
				Method: "POST",
				Header: map[string][]string{"Content-Type": {"application/json"}, "myheader": {"MyheaderValue"}, "Anotherheader": {"anotherheaderValue"}}},
			expectedMatchEndpointId: ""},
	}

	assertMatchRequestToEndpoint(mockRouter, testCases, t)

}

func TestMatchRequestToEndpoint_Matches(t *testing.T) {
	mockRouter := createMockRouter("simplemocks", t)

	request := createRequest(http.MethodGet, "http://host/minimal", "", nil, nil, t)
	ep, _, _ := mockRouter.matchRequestToEndpoint(request)
	assert.Equal(t, "minimal", ep.Id)
	endPointMatches := mockRouter.Matches[ep.Id]
	assert.NotNil(t, endPointMatches)
	assert.Equal(t, 1, len(endPointMatches))
	assert.Equal(t, ep.Id, endPointMatches[0].EndpointId)
	assert.NotNil(t, endPointMatches[0].ActualRequest)
	assert.Equal(t, http.MethodGet, endPointMatches[0].ActualRequest.Method)
	assert.Equal(t, "host", endPointMatches[0].ActualRequest.Host)
	assert.Equal(t, "http://host/minimal", endPointMatches[0].ActualRequest.URL)

}

func TestMatchRequestToEndpoint_Wildcardmocks(t *testing.T) {
	mockRouter := createMockRouter("wildcardmocks", t)

	testCases := []*matchingTestCase{
		{name: "Single wildcard Match 1 ", request: &http.Request{URL: &url.URL{Path: "/wildcard/bar/foo"}, Method: http.MethodGet}, expectedMatchEndpointId: "2"},
		{name: "Single wildcard Match 2", request: &http.Request{URL: &url.URL{Path: "/wildcard/foo/foo"}, Method: http.MethodGet}, expectedMatchEndpointId: "2"},
		{name: "Single wildcard No match, first path segment", request: &http.Request{URL: &url.URL{Path: "/wildcards/bar/foo"}, Method: http.MethodGet}, expectedMatchEndpointId: ""},
		{name: "Single wildcard No match, path too long ", request: &http.Request{URL: &url.URL{Path: "/wildcard/bar/foo/toolong"}, Method: http.MethodGet}, expectedMatchEndpointId: ""},
		{name: "Single wildcard No match, path too short ", request: &http.Request{URL: &url.URL{Path: "/bar/foo"}, Method: http.MethodGet}, expectedMatchEndpointId: ""},
		{name: "Multi wildcard Match", request: &http.Request{URL: &url.URL{Path: "/multiwildcard/bar/foo/bar"}, Method: http.MethodGet}, expectedMatchEndpointId: "1"},
	}

	assertMatchRequestToEndpoint(mockRouter, testCases, t)
}

func TestMatchRequestToEndpoint_AllMatchWildcardmocks(t *testing.T) {
	mockRouter := createMockRouter("allMatchWildcardMocks", t)
	testCases := []*matchingTestCase{
		{name: "Match 1 ", request: &http.Request{URL: &url.URL{Path: "/allmatchwildcardAtTheEnd/bar"}, Method: http.MethodGet}, expectedMatchEndpointId: "1"},
		{name: "Match 2 ", request: &http.Request{URL: &url.URL{Path: "/allmatchwildcardAtTheEnd/foo"}, Method: http.MethodGet}, expectedMatchEndpointId: "1"},
		{name: "Match path longer ", request: &http.Request{URL: &url.URL{Path: "/allmatchwildcardAtTheEnd/foo/bar"}, Method: http.MethodGet}, expectedMatchEndpointId: "1"},
		{name: "No Match, first path segment ", request: &http.Request{URL: &url.URL{Path: "/allmatchwildcardAtTheEnds/foo"}, Method: http.MethodGet}, expectedMatchEndpointId: ""},
		{name: "No Match, path shorter ", request: &http.Request{URL: &url.URL{Path: "/"}, Method: http.MethodGet}, expectedMatchEndpointId: ""},
		{name: "Match in the middle 1", request: &http.Request{URL: &url.URL{Path: "/allmatchwildcardInTheMiddle/bar/foo"}, Method: http.MethodGet}, expectedMatchEndpointId: "2"},
		{name: "Match in the middle 2", request: &http.Request{URL: &url.URL{Path: "/allmatchwildcardInTheMiddle/bar/ext/foo"}, Method: http.MethodGet}, expectedMatchEndpointId: "2"},
		{name: "Match in the middle 3", request: &http.Request{URL: &url.URL{Path: "/allmatchwildcardInTheMiddle/bar/ext/rem/foo"}, Method: http.MethodGet}, expectedMatchEndpointId: "2"},
		{name: "No Match endsegements", request: &http.Request{URL: &url.URL{Path: "/allmatchwildcardInTheMiddle/bar/foo/foo"}, Method: http.MethodGet}, expectedMatchEndpointId: ""},
		{name: "Match combined wildcards single segment ", request: &http.Request{URL: &url.URL{Path: "/combinedwildcards1/bar/foo/ext"}, Method: http.MethodGet}, expectedMatchEndpointId: "3"},
		{name: "Match combined wildcards multiple segment", request: &http.Request{URL: &url.URL{Path: "/combinedwildcards1/bar/a/b/c/foo/d"}, Method: http.MethodGet}, expectedMatchEndpointId: "3"},
		{name: "No Match combined wildcards last segment missing", request: &http.Request{URL: &url.URL{Path: "/combinedwildcards1/bar/a/b/c/foo"}, Method: http.MethodGet}, expectedMatchEndpointId: ""},
	}
	assertMatchRequestToEndpoint(mockRouter, testCases, t)
}

func TestMatchRequestToEndpoint_PathParamsmocks(t *testing.T) {
	mockRouter := createMockRouter("pathParamsMocks", t)
	testCases := []*matchingTestCase{
		{name: "Single pathparams, match ", request: &http.Request{URL: &url.URL{Path: "/pathParams/bar/foo"}, Method: http.MethodGet}, expectedMatchEndpointId: "2", expectedRequestParams: map[string]string{"pathParam": "bar"}},
		{name: "Single pathparams, No Match last segment does not match,  ", request: &http.Request{URL: &url.URL{Path: "/pathParams/bar/foos"}, Method: http.MethodGet}, expectedMatchEndpointId: ""},
		{name: "Multi pathparams, match ", request: &http.Request{URL: &url.URL{Path: "/multipathParams/val1/foo/val2"}, Method: http.MethodGet}, expectedMatchEndpointId: "1", expectedRequestParams: map[string]string{"pathParam1": "val1", "pathParam2": "val2"}},
	}
	assertMatchRequestToEndpoint(mockRouter, testCases, t)
}

func TestMatchRequestToEndpoint_Prio(t *testing.T) {
	mockRouter := createMockRouter("prioMocks", t)
	testCases := []*matchingTestCase{
		{name: "Simple prio, match ", request: &http.Request{URL: &url.URL{Path: "/prio"}, Method: http.MethodGet}, expectedMatchEndpointId: "mustwin"},
	}
	assertMatchRequestToEndpoint(mockRouter, testCases, t)
}

func TestMatchRequestToEndpoint_Regexp(t *testing.T) {
	mockRouter := createMockRouter("regexpmocks", t)
	testCases := []*matchingTestCase{
		{name: "Regexp Mock1: Match",
			request:                 createRequest(http.MethodPost, "https://mymock.com/regexp1", "{ alex }", nil, nil, t),
			expectedMatchEndpointId: "1"},
		{name: "Regexp Mock1: No Match",
			request:                 createRequest(http.MethodPost, "https://mymock.com/regexp1", "{ alex ", nil, nil, t),
			expectedMatchEndpointId: ""},
		{name: "Regexp Mock2: Match",
			request:                 createRequest(http.MethodPost, "https://mymock.com/regexp2", "{\n alex \n}", nil, nil, t),
			expectedMatchEndpointId: "2"},
		{name: "Regexp Mock3: Match",
			request:                 createRequest(http.MethodPost, "https://mymock.com/regexp3", `{ "email": "foo@bar.com" }`, nil, nil, t),
			expectedMatchEndpointId: "3"},
	}
	assertMatchRequestToEndpoint(mockRouter, testCases, t)
}

// map[string][]string{ headers.Authorization : {"Basic blabla"}}
func TestMatchRequest_Rendering(t *testing.T) {
	user := "Alex"
	password := "mysecretpassword"
	os.Setenv("USER", user)
	os.Setenv("PASSWORD", password)
	mockRouter := createMockRouter("requestRendering", t)
	basicAuth := base64.StdEncoding.EncodeToString([]byte(user + ":" + password))
	testCases := []*matchingTestCase{
		{name: "BasicAuth, match ",
			request: createRequest(http.MethodGet, "http://host/auth", "",
				map[string][]string{headers.ContentType: {"application/json"}, headers.Authorization: {"Basic " + basicAuth}}, nil, t),
			expectedMatchEndpointId: "basicauth"},
	}
	assertMatchRequestToEndpoint(mockRouter, testCases, t)
}

func TestRenderResponse_Delay(t *testing.T) {
	mockRouter := createMockRouter("responseRendering", t)

	testCases := []*renderingTestCase{
		{name: "delay", match: &model.Match{}, responseTemplate: "statusCode: 204 {{ delay 10 }}",
			expectedResponseStatusCode: 204},
	}
	ts := time.Now()
	assertRenderingResponse(mockRouter, testCases, t)
	assert.LessOrEqual(t, 10*time.Millisecond, time.Since(ts))
}

func TestRenderResponse_Matches(t *testing.T) {
	mockRouter := createMockRouter("responseRendering", t)
	request := createRequest(http.MethodGet, "http://host/one", "", map[string][]string{"headerKey": {"headerValue"}}, nil, t)
	ep, _, _ := mockRouter.matchRequestToEndpoint(request)
	assert.NotNil(t, ep)
	assert.Equal(t, "one", ep.Id)
	testCases := []*renderingTestCase{
		{name: "template matches",
			match: &model.Match{},
			responseTemplate: `body: |
  {{ range $i, $match := matches "one" -}}
  Match-EndpointId:  {{ $match.EndpointId }}
  Match-Request method:  {{ $match.ActualRequest.Method }}
  Match-Request URL:  {{ $match.ActualRequest.URL }}
  Match-Request Host:  {{ $match.ActualRequest.Host }}
  Match-Request header:
  {{ range $key, $value := $match.ActualRequest.Header -}}
  {{ $key }}: {{ $value }}
  {{ end }}
  {{ end }}`,
			expectedResponseStatusCode: 200,
			expectedResponseBody: `Match-EndpointId:  one
Match-Request method:  GET
Match-Request URL:  http://host/one
Match-Request Host:  host
Match-Request header:
headerKey: [headerValue]
`},
	}
	assertRenderingResponse(mockRouter, testCases, t)
}

func TestRenderResponse_Simple(t *testing.T) {
	kvstore.CreateTheStore()
	mockRouter := createMockRouter("responseRendering", t)
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
		{name: "status", match: &model.Match{}, responseTemplate: "statusCode: 204",
			expectedResponseStatusCode: 204},
		{name: "body", match: &model.Match{}, responseTemplate: "body: Hello",
			expectedResponseStatusCode: 200,
			expectedResponseBody:       "Hello"},
		{name: "body from response file", match: &model.Match{}, responseTemplate: "bodyFilename: simple-response.json",
			expectedResponseStatusCode: 200,
			expectedResponseBody:       "{\n    \"greets\": \"Hello\"\n}"},
		{name: "template RequestPathParams", match: &model.Match{}, responseTemplate: "body: {{ .RequestPathParams.param1 }}", requestParams: map[string]string{"param1": "Hello"},
			expectedResponseStatusCode: 200,
			expectedResponseBody:       "Hello"},
		{name: "template RequestParam statuscode", match: &model.Match{}, responseTemplate: "statusCode: {{ .RequestPathParams.param1 }}", requestParams: map[string]string{"param1": "202"},
			expectedResponseStatusCode: 202},
		{name: "template Request url", match: &model.Match{},
			responseTemplate: "body: \"incoming request url: '{{ .RequestUrl }}'\"",
			request: &http.Request{URL: &url.URL{User: url.User("alex"), Scheme: "https", Host: "myhost", Path: "/mypath"},
				Method: http.MethodGet,
				Header: map[string][]string{"headerKey": {"headerValue"}}},
			expectedResponseStatusCode: 200,
			expectedResponseBody:       "incoming request url: 'https://alex@myhost/mypath'"},
		{name: "template response file all request params", match: &model.Match{},
			responseTemplate:           "bodyFilename: request-template-response.json",
			request:                    createRequest("PUT", "https://coolhost.cooldomain.com/coolpath", "{ \"requestBodyKey\": \"requestBodyValue\" }", map[string][]string{"myheaderKey": {"myheaderValue"}}, nil, t),
			requestParams:              map[string]string{"myparam1": "myvalue"},
			expectedResponseStatusCode: 200,
			expectedResponseBody:       expectedResponseResult},
		{name: "template kvstore", match: &model.Match{},
			responseTemplate:           "body: |\n{{ kvStoreGet \"testkey\" | toPrettyJson | indent 2 }}",
			kvstoreJson:                `{ "myResponse" : "is Great!" }`,
			expectedResponseStatusCode: 200,
			expectedResponseBody:       "{\n  \"myResponse\": \"is Great!\"\n}"},
		{name: "template endpoints", match: &model.Match{},
			responseTemplate:           "body: |\n  {{ range $i, $ep := endpointIds }}\n  Endpoint: {{ $ep -}}\n  {{ end }}",
			expectedResponseStatusCode: 200,
			expectedResponseBody:       "\nEndpoint: one\nEndpoint: two"},
		{name: "response no yaml", match: &model.Match{}, responseTemplate: "statusCode: 204 this is no valid json",
			expectedResponseStatusCode: 500,
			expectedResponseBody:       "Error rendering response: could't unmarshall response yaml:\n'statusCode: 204 this is no valid json'\nerror: yaml: unmarshal errors:\n  line 1: cannot unmarshal !!str `204 thi...` into int",
		},
	}
	assertRenderingResponse(mockRouter, testCases, t)
}

func createRequest(method, url, bodyStr string, header map[string][]string, urlVars map[string]string, t *testing.T) *http.Request {
	body := io.NopCloser(strings.NewReader(bodyStr))
	request := httptest.NewRequest(method, url, body)
	request.Header = header
	if urlVars != nil {
		request = mux.SetURLVars(request, urlVars)
	}
	return request
}

func assertRenderingResponse(mockRouter *MockRouter, testCases []*renderingTestCase, t *testing.T) {
	for _, testCase := range testCases {
		if len(testCase.kvstoreJson) > 0 {
			err := kvstore.TheKVStore.Put("testkey", testCase.kvstoreJson)
			assert.NoError(t, err)
		}
		recorder := httptest.NewRecorder()
		tplt, err := template.New("response").Funcs(sprig.TxtFuncMap()).Funcs(mockRouter.templateFuncMap()).Parse(testCase.responseTemplate)
		assert.NoError(t, err)
		endpoint := &model.MockEndpoint{Response: &model.MockResponse{Template: tplt}}

		if testCase.request == nil {
			testCase.request = &http.Request{URL: &url.URL{}}
		}
		mockRouter.renderResponse(recorder, testCase.request, endpoint, testCase.match, testCase.requestParams)

		assert.Equal(t, testCase.expectedResponseStatusCode, recorder.Result().StatusCode, "unexpected statuscode")

		responseBody, err := io.ReadAll(recorder.Result().Body)
		assert.NoError(t, err)

		assert.Equal(t, testCase.expectedResponseBody, string(responseBody), "unexpected responseBody")

		if testCase.expectedResponseHeader != nil {
			for expectedParamName, expectedParamValue := range testCase.expectedResponseHeader {
				assert.Equal(t, recorder.Header().Get(expectedParamName), expectedParamValue, "unexpected header param")
			}
		}
		if !t.Failed() {
			t.Logf("testcase '%s':'%s' passed", t.Name(), testCase.name)
		}
	}
}

func assertMatchRequestToEndpoint(mockRouter *MockRouter, testCases []*matchingTestCase, t *testing.T) {
	for _, testCase := range testCases {
		ep, match, requestParams := mockRouter.matchRequestToEndpoint(testCase.request)

		if len(testCase.expectedMatchEndpointId) > 0 {
			assert.NotNil(t, match, "%s : expected match, but is none", testCase.name)
			assert.Equal(t, testCase.expectedMatchEndpointId, match.EndpointId, "%s : unexpected matchEndpointId", testCase)
			assert.NotNil(t, ep, "%s : expect a match for request: %v", testCase.name, testCase.request)
		} else {
			assert.Nil(t, ep, "%s : expect no match for request: %v", testCase.name, testCase.request)
		}

		if testCase.expectedRequestParams != nil {
			for expectedParamName, expectedParamValue := range testCase.expectedRequestParams {
				assert.Equal(t, requestParams[expectedParamName], expectedParamValue, "unexpected request param")
			}
		}
	}
}

func createMockRouter(testMockDir string, t *testing.T) *MockRouter {
	mockRouter, err := NewMockRouter("../../test/"+testMockDir, "*-mock.yaml", "../../test/"+testMockDir, "*-response.json", 0, kvstore.TheKVStore, &utils.Logger{Verbose: true, DebugResponseRendering: true})
	assert.NoError(t, err)
	assert.NotNil(t, mockRouter, "Mockrouter must not be nil")
	return mockRouter
}
