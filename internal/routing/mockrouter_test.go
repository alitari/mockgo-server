package routing

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"text/template"

	"github.com/Masterminds/sprig"
	"github.com/alitari/mockgo-server/internal/model"
	"github.com/alitari/mockgo-server/internal/utils"
)

type matchingTestCase struct {
	name                      string
	request                   *http.Request
	expectedMatch             bool
	expectedRequestParams     map[string]string
	expectedMatchedEndpointId string
}

type renderingTestCase struct {
	name                       string
	responseTemplate           string
	requestParams              map[string]string
	request                    *http.Request
	kvstoreJson                string
	expectedResponseStatusCode int
	expectedResponseBody       string
	expectedResponseHeader     map[string]string
}

func TestMatchRequestToEndpoint_Simplemocks(t *testing.T) {
	mockRouter := createMockRouter("simplemocks", t)

	testCases := []*matchingTestCase{
		{name: "Minimal Mock: Match, full request",
			request: &http.Request{
				URL:    &url.URL{Scheme: "https", Host: "myhost", Path: "/minimal"},
				Method: "GET",
				Header: map[string][]string{"Accept": {"Something"}, "Authorization": {"Basic"}}},
			expectedMatch: true},
		{name: "Minimal Mock: Match, minimal ", request: &http.Request{URL: &url.URL{Path: "/minimal"}, Method: "GET"}, expectedMatch: true},
		{name: "Minimal Mock: No Match, wrong path name", request: &http.Request{URL: &url.URL{Path: "/minimals"}, Method: "GET"}, expectedMatch: false},
		{name: "Minimal Mock: No Match, wrong path length too long", request: &http.Request{URL: &url.URL{Path: "/minimal/foo"}, Method: "GET"}, expectedMatch: false},
		{name: "Minimal Mock: No Match, wrong path length too short", request: &http.Request{URL: &url.URL{Path: "/"}, Method: "GET"}, expectedMatch: false},
		{name: "Minimal Mock: No Match, wrong method", request: &http.Request{URL: &url.URL{Path: "/minimal"}, Method: "POST"}, expectedMatch: false},
		{name: "Maximal Mock: Match, exact",
			request: &http.Request{
				URL:    &url.URL{Scheme: "https", Host: "alexkrieg.com", Path: "/maximal", RawQuery: "firstQueryParam=value1&secondQueryParam=value2"},
				Method: "POST",
				Header: map[string][]string{"Content-Type": {"application/json"}, "Myheader": {"myheaderValue"}}},
			expectedMatch: true},
		{name: "Maximal Mock: Match, header and query superset",
			request: &http.Request{
				URL:    &url.URL{Scheme: "https", Host: "alexkrieg.com", Path: "/maximal", RawQuery: "firstQueryParam=value1&secondQueryParam=value2&thirdQueryParam=value3"},
				Method: "POST",
				Header: map[string][]string{"Content-Type": {"application/json"}, "Myheader": {"myheaderValue"}, "Anotherheader": {"anotherheaderValue"}}},
			expectedMatch: true},
		{name: "Maximal Mock: No Match, query subset",
			request: &http.Request{
				URL:    &url.URL{Scheme: "https", Host: "alexkrieg.com", Path: "/maximal", RawQuery: "firstQueryParam=value1&thirdQueryParam=value3"},
				Method: "POST",
				Header: map[string][]string{"Content-Type": {"application/json"}, "myheader": {"MyheaderValue"}, "Anotherheader": {"anotherheaderValue"}}},
			expectedMatch: false},
	}

	assertMatchRequestToEndpoint(mockRouter, testCases, t)

}

func TestMatchRequestToEndpoint_Wildcardmocks(t *testing.T) {
	mockRouter := createMockRouter("wildcardmocks", t)

	testCases := []*matchingTestCase{
		{name: "Single wildcard Match 1 ", request: &http.Request{URL: &url.URL{Path: "/wildcard/bar/foo"}, Method: "GET"}, expectedMatch: true},
		{name: "Single wildcard Match 2", request: &http.Request{URL: &url.URL{Path: "/wildcard/foo/foo"}, Method: "GET"}, expectedMatch: true},
		{name: "Single wildcard No match, first path segment", request: &http.Request{URL: &url.URL{Path: "/wildcards/bar/foo"}, Method: "GET"}, expectedMatch: false},
		{name: "Single wildcard No match, path too long ", request: &http.Request{URL: &url.URL{Path: "/wildcard/bar/foo/toolong"}, Method: "GET"}, expectedMatch: false},
		{name: "Single wildcard No match, path too short ", request: &http.Request{URL: &url.URL{Path: "/bar/foo"}, Method: "GET"}, expectedMatch: false},
		{name: "Multi wildcard Match", request: &http.Request{URL: &url.URL{Path: "/multiwildcard/bar/foo/bar"}, Method: "GET"}, expectedMatch: true},
	}

	assertMatchRequestToEndpoint(mockRouter, testCases, t)
}

func TestMatchRequestToEndpoint_AllMatchWildcardmocks(t *testing.T) {
	mockRouter := createMockRouter("allMatchWildcardMocks", t)
	testCases := []*matchingTestCase{
		{name: "Match 1 ", request: &http.Request{URL: &url.URL{Path: "/allmatchwildcardAtTheEnd/bar"}, Method: "GET"}, expectedMatch: true},
		{name: "Match 2 ", request: &http.Request{URL: &url.URL{Path: "/allmatchwildcardAtTheEnd/foo"}, Method: "GET"}, expectedMatch: true},
		{name: "Match path longer ", request: &http.Request{URL: &url.URL{Path: "/allmatchwildcardAtTheEnd/foo/bar"}, Method: "GET"}, expectedMatch: true},
		{name: "No Match, first path segment ", request: &http.Request{URL: &url.URL{Path: "/allmatchwildcardAtTheEnds/foo"}, Method: "GET"}, expectedMatch: false},
		{name: "No Match, path shorter ", request: &http.Request{URL: &url.URL{Path: "/"}, Method: "GET"}, expectedMatch: false},
		{name: "Match in the middle 1", request: &http.Request{URL: &url.URL{Path: "/allmatchwildcardInTheMiddle/bar/foo"}, Method: "GET"}, expectedMatch: true},
		{name: "Match in the middle 2", request: &http.Request{URL: &url.URL{Path: "/allmatchwildcardInTheMiddle/bar/ext/foo"}, Method: "GET"}, expectedMatch: true},
		{name: "Match in the middle 3", request: &http.Request{URL: &url.URL{Path: "/allmatchwildcardInTheMiddle/bar/ext/rem/foo"}, Method: "GET"}, expectedMatch: true},
		{name: "No Match endsegements", request: &http.Request{URL: &url.URL{Path: "/allmatchwildcardInTheMiddle/bar/foo/foo"}, Method: "GET"}, expectedMatch: false},
		{name: "Match combined wildcards single segment ", request: &http.Request{URL: &url.URL{Path: "/combinedwildcards1/bar/foo/ext"}, Method: "GET"}, expectedMatch: true},
		{name: "Match combined wildcards multiple segment", request: &http.Request{URL: &url.URL{Path: "/combinedwildcards1/bar/a/b/c/foo/d"}, Method: "GET"}, expectedMatch: true},
		{name: "No Match combined wildcards last segment missing", request: &http.Request{URL: &url.URL{Path: "/combinedwildcards1/bar/a/b/c/foo"}, Method: "GET"}, expectedMatch: false},
	}
	assertMatchRequestToEndpoint(mockRouter, testCases, t)
}

func TestMatchRequestToEndpoint_PathParamsmocks(t *testing.T) {
	mockRouter := createMockRouter("pathParamsMocks", t)
	testCases := []*matchingTestCase{
		{name: "Single pathparams, match ", request: &http.Request{URL: &url.URL{Path: "/pathParams/bar/foo"}, Method: "GET"}, expectedMatch: true, expectedRequestParams: map[string]string{"pathParam": "bar"}},
		{name: "Single pathparams, No Match last segment does not match,  ", request: &http.Request{URL: &url.URL{Path: "/pathParams/bar/foos"}, Method: "GET"}, expectedMatch: false},
		{name: "Multi pathparams, match ", request: &http.Request{URL: &url.URL{Path: "/multipathParams/val1/foo/val2"}, Method: "GET"}, expectedMatch: true, expectedRequestParams: map[string]string{"pathParam1": "val1", "pathParam2": "val2"}},
	}
	assertMatchRequestToEndpoint(mockRouter, testCases, t)
}

func TestMatchRequestToEndpoint_Prio(t *testing.T) {
	mockRouter := createMockRouter("prioMocks", t)
	testCases := []*matchingTestCase{
		{name: "Simple prio, match ", request: &http.Request{URL: &url.URL{Path: "/prio"}, Method: "GET"}, expectedMatch: true, expectedMatchedEndpointId: "mustwin"},
	}
	assertMatchRequestToEndpoint(mockRouter, testCases, t)
}

func TestRenderResponse_Simple(t *testing.T) {
	mockRouter := createMockRouter("responseRendering", t)
	expectedResponseResult := `{
    "requestUrl": "https://coolhost.cooldomain.com/coolpath",
    "requestPathParams": "map[myparam1:myvalue]",
    "requestPathParam1": "myvalue",
    "requestPath": "/coolpath",
    "requestHost": "coolhost.cooldomain.com",
    "requestBody": "{ \"requestBody\": \"is there!\" }"
}`

	testCases := []*renderingTestCase{
		{name: "status", responseTemplate: "statusCode: 204",
			expectedResponseStatusCode: 204},
		{name: "body", responseTemplate: "body: Hello",
			expectedResponseStatusCode: 200,
			expectedResponseBody:       "Hello"},
		{name: "body from response file", responseTemplate: "bodyFilename: simple-response.json",
			expectedResponseStatusCode: 200,
			expectedResponseBody:       "{\n    \"greets\": \"Hello\"\n}"},
		{name: "template RequestPathParams", responseTemplate: "body: {{ .RequestPathParams.param1 }}", requestParams: map[string]string{"param1": "Hello"},
			expectedResponseStatusCode: 200,
			expectedResponseBody:       "Hello"},
		{name: "template Request url",
			responseTemplate: "body: \"incoming request url: '{{ .RequestUrl }}'\"",
			request: &http.Request{URL: &url.URL{User: url.User("alex"), Scheme: "https", Host: "myhost", Path: "/mypath"},
				Method: "GET",
				Header: map[string][]string{"headerKey": {"headerValue"}}},
			expectedResponseStatusCode: 200,
			expectedResponseBody:       "incoming request url: 'https://alex@myhost/mypath'"},
		{name: "template response file all request params",
			responseTemplate:           "bodyFilename: request-template-response.json",
			request:                    createRequest("PUT", "https://coolhost.cooldomain.com/coolpath", "{ \"requestBody\": \"is there!\" }", map[string][]string{"myheaderKey": {"myheaderValue"}}, t),
			requestParams:              map[string]string{"myparam1": "myvalue"},
			expectedResponseStatusCode: 200,
			expectedResponseBody:       expectedResponseResult},
		{name: "template kvstore",
			responseTemplate:           "body: |\n{{ kvStoreGet \"testkey\" | toPrettyJson | indent 2 }}",
			kvstoreJson:                `{ "myResponse" : "is Great!" }`,
			expectedResponseStatusCode: 200,
			expectedResponseBody:       "{\n  \"myResponse\": \"is Great!\"\n}"},
		{name: "response no yaml", responseTemplate: "statusCode: 204 this is no valid json",
			expectedResponseStatusCode: 500,
			expectedResponseBody:       "Error rendering response: could't unmarshall response yaml:\n'statusCode: 204 this is no valid json'\nerror: yaml: unmarshal errors:\n  line 1: cannot unmarshal !!str `204 thi...` into int",
		},
	}
	assertRenderingResponse(mockRouter, testCases, t)
}

func createRequest(method, url, bodyStr string, header map[string][]string, t *testing.T) *http.Request {
	body := io.NopCloser(strings.NewReader(bodyStr))
	request, err := http.NewRequest(method, url, body)
	if err != nil {
		t.Fatal(err)
	}
	request.Header = header
	return request
}

func assertRenderingResponse(mockRouter *MockRouter, testCases []*renderingTestCase, t *testing.T) {
	for _, testCase := range testCases {
		if len(testCase.kvstoreJson) > 0 {
			err := mockRouter.kvstore.Put("testkey", testCase.kvstoreJson)
			if err != nil {
				t.Fatal(err)
			}
		}
		recorder := httptest.NewRecorder()
		tplt, err := template.New("response").Funcs(sprig.TxtFuncMap()).Funcs(mockRouter.templateFuncMap()).Parse(testCase.responseTemplate)
		if err != nil {
			t.Fatal(err)
		}
		endpoint := &model.MockEndpoint{Response: &model.MockResponse{Template: tplt}}

		if testCase.request == nil {
			testCase.request = &http.Request{URL: &url.URL{}}
		}
		mockRouter.renderResponse(recorder, testCase.request, endpoint, testCase.requestParams)

		if testCase.expectedResponseStatusCode != recorder.Result().StatusCode {
			t.Errorf("Expected status code %v, but is %v", testCase.expectedResponseStatusCode, recorder.Result().StatusCode)
		}

		responseBody, err := io.ReadAll(recorder.Result().Body)
		if err != nil {
			t.Fatal(err)
		}

		if testCase.expectedResponseBody != string(responseBody) {
			t.Errorf("Expected response body is:\n%s,\nbut is:\n%s", testCase.expectedResponseBody, responseBody)
		}

		if testCase.expectedResponseHeader != nil {
			for expectedParamName, expectedParamValue := range testCase.expectedResponseHeader {
				if recorder.Header().Get(expectedParamName) != expectedParamValue {
					t.Errorf("testcase '%s' failed:  expect a response header param with key: '%s' and value: '%s'. but get '%s' ", testCase.name, expectedParamName, expectedParamValue, recorder.Header().Get(expectedParamName))
				}
			}
		}
		if !t.Failed() {
			t.Logf("testcase '%s':'%s' passed", t.Name(), testCase.name)
		}

	}
}

func assertMatchRequestToEndpoint(mockRouter *MockRouter, testCases []*matchingTestCase, t *testing.T) {
	for _, testCase := range testCases {
		ep, requestParams := mockRouter.matchRequestToEndpoint(testCase.request)
		if testCase.expectedMatch && ep == nil {
			t.Errorf("testcase '%s' failed:  expect a match for request: %v", testCase.name, testCase.request)
		} else if !testCase.expectedMatch && ep != nil {
			t.Errorf("testcase '%s' failed:  expect a no match for request: %v", testCase.name, testCase.request)
		}

		if testCase.expectedRequestParams != nil {
			for expectedParamName, expectedParamValue := range testCase.expectedRequestParams {
				if requestParams[expectedParamName] != expectedParamValue {
					t.Errorf("testcase '%s' failed:  expect a request param with key: '%s' and value: '%s'. but get '%s' ", testCase.name, expectedParamName, expectedParamValue, requestParams[expectedParamName])
				}
			}
		}
		if testCase.expectedMatchedEndpointId != "" {
			if ep.Id != testCase.expectedMatchedEndpointId {
				t.Errorf("testcase '%s' failed:  expect matched endpoint id is '%s' but is '%s' ", testCase.name, testCase.expectedMatchedEndpointId, ep.Id)
			}
		}
		if !t.Failed() {
			t.Logf("testcase '%s':'%s' passed", t.Name(), testCase.name)
		}
	}
}

func createMockRouter(testMockDir string, t *testing.T) *MockRouter {
	mockRouter, err := NewMockRouter("../../test/"+testMockDir, "*-mock.yaml", "../../test/"+testMockDir, "*-response.json", &utils.Logger{Verbose: true, DebugResponseRendering: true})
	if err != nil {
		t.Fatalf("Can't create mock router: %v", err)
	}
	if mockRouter == nil {
		t.Fatal("Mockrouter must not be nil")
	}
	if err != nil {
		t.Fatalf("Can't load mocks . %v", err)
	}
	return mockRouter
}
