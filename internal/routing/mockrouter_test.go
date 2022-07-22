package routing

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/alitari/mockgo-server/internal/utils"
)

type testCase struct {
	name          string
	request       *http.Request
	match         bool
	requestParams map[string]string
}

func TestMatchRequestToEndpoint_Simplemocks(t *testing.T) {
	mockRouter := createMockRouter("simplemocks", t)

	testCases := []*testCase{
		{name: "Minimal Mock: Match, full request",
			request: &http.Request{
				URL:    &url.URL{Scheme: "https", Host: "myhost", Path: "/minimal"},
				Method: "GET",
				Header: map[string][]string{"Accept": {"Something"}, "Authorization": {"Basic"}}},
			match: true},
		{name: "Minimal Mock: Match, minimal ", request: &http.Request{URL: &url.URL{Path: "/minimal"}, Method: "GET"}, match: true},
		{name: "Minimal Mock: No Match, wrong path name", request: &http.Request{URL: &url.URL{Path: "/minimals"}, Method: "GET"}, match: false},
		{name: "Minimal Mock: No Match, wrong path length too long", request: &http.Request{URL: &url.URL{Path: "/minimal/foo"}, Method: "GET"}, match: false},
		{name: "Minimal Mock: No Match, wrong path length too short", request: &http.Request{URL: &url.URL{Path: "/"}, Method: "GET"}, match: false},
		{name: "Minimal Mock: No Match, wrong method", request: &http.Request{URL: &url.URL{Path: "/minimal"}, Method: "POST"}, match: false},
		{name: "Maximal Mock: Match, exact",
			request: &http.Request{
				URL:    &url.URL{Scheme: "https", Host: "alexkrieg.com", Path: "/maximal", RawQuery: "firstQueryParam=value1&secondQueryParam=value2"},
				Method: "POST",
				Header: map[string][]string{"Content-Type": {"application/json"}, "Myheader": {"myheaderValue"}}},
			match: true},
		{name: "Maximal Mock: Match, header and query superset",
			request: &http.Request{
				URL:    &url.URL{Scheme: "https", Host: "alexkrieg.com", Path: "/maximal", RawQuery: "firstQueryParam=value1&secondQueryParam=value2&thirdQueryParam=value3"},
				Method: "POST",
				Header: map[string][]string{"Content-Type": {"application/json"}, "Myheader": {"myheaderValue"}, "Anotherheader": {"anotherheaderValue"}}},
			match: true},
		{name: "Maximal Mock: No Match, query subset",
			request: &http.Request{
				URL:    &url.URL{Scheme: "https", Host: "alexkrieg.com", Path: "/maximal", RawQuery: "firstQueryParam=value1&thirdQueryParam=value3"},
				Method: "POST",
				Header: map[string][]string{"Content-Type": {"application/json"}, "myheader": {"MyheaderValue"}, "Anotherheader": {"anotherheaderValue"}}},
			match: false},
	}

	assertMatchRequestToEndpoint(mockRouter, testCases, t)

}

func TestMatchRequestToEndpoint_Wildcardmocks(t *testing.T) {
	mockRouter := createMockRouter("wildcardmocks", t)

	testCases := []*testCase{
		{name: "Single wildcard Match 1 ", request: &http.Request{URL: &url.URL{Path: "/wildcard/bar/foo"}, Method: "GET"}, match: true},
		{name: "Single wildcard Match 2", request: &http.Request{URL: &url.URL{Path: "/wildcard/foo/foo"}, Method: "GET"}, match: true},
		{name: "Single wildcard No match, first path segment", request: &http.Request{URL: &url.URL{Path: "/wildcards/bar/foo"}, Method: "GET"}, match: false},
		{name: "Single wildcard No match, path too long ", request: &http.Request{URL: &url.URL{Path: "/wildcard/bar/foo/toolong"}, Method: "GET"}, match: false},
		{name: "Single wildcard No match, path too short ", request: &http.Request{URL: &url.URL{Path: "/bar/foo"}, Method: "GET"}, match: false},
		{name: "Multi wildcard Match", request: &http.Request{URL: &url.URL{Path: "/multiwildcard/bar/foo/bar"}, Method: "GET"}, match: true},
	}

	assertMatchRequestToEndpoint(mockRouter, testCases, t)
}

func TestMatchRequestToEndpoint_AllMatchWildcardmocks(t *testing.T) {
	mockRouter := createMockRouter("allMatchWildcardMocks", t)
	testCases := []*testCase{
		{name: "Match 1 ", request: &http.Request{URL: &url.URL{Path: "/allmatchwildcardAtTheEnd/bar"}, Method: "GET"}, match: true},
		{name: "Match 2 ", request: &http.Request{URL: &url.URL{Path: "/allmatchwildcardAtTheEnd/foo"}, Method: "GET"}, match: true},
		{name: "Match path longer ", request: &http.Request{URL: &url.URL{Path: "/allmatchwildcardAtTheEnd/foo/bar"}, Method: "GET"}, match: true},
		{name: "No Match, first path segment ", request: &http.Request{URL: &url.URL{Path: "/allmatchwildcardAtTheEnds/foo"}, Method: "GET"}, match: false},
		{name: "No Match, path shorter ", request: &http.Request{URL: &url.URL{Path: "/"}, Method: "GET"}, match: false},
		{name: "Match in the middle 1", request: &http.Request{URL: &url.URL{Path: "/allmatchwildcardInTheMiddle/bar/foo"}, Method: "GET"}, match: true},
		{name: "Match in the middle 2", request: &http.Request{URL: &url.URL{Path: "/allmatchwildcardInTheMiddle/bar/ext/foo"}, Method: "GET"}, match: true},
		{name: "Match in the middle 3", request: &http.Request{URL: &url.URL{Path: "/allmatchwildcardInTheMiddle/bar/ext/rem/foo"}, Method: "GET"}, match: true},
		{name: "No Match endsegements", request: &http.Request{URL: &url.URL{Path: "/allmatchwildcardInTheMiddle/bar/foo/foo"}, Method: "GET"}, match: false},
		{name: "Match combined wildcards single segment ", request: &http.Request{URL: &url.URL{Path: "/combinedwildcards1/bar/foo/ext"}, Method: "GET"}, match: true},
		{name: "Match combined wildcards multiple segment", request: &http.Request{URL: &url.URL{Path: "/combinedwildcards1/bar/a/b/c/foo/d"}, Method: "GET"}, match: true},
		{name: "No Match combined wildcards last segment missing", request: &http.Request{URL: &url.URL{Path: "/combinedwildcards1/bar/a/b/c/foo"}, Method: "GET"}, match: false},
	}
	assertMatchRequestToEndpoint(mockRouter, testCases, t)
}

func TestMatchRequestToEndpoint_PathParamsmocks(t *testing.T) {
	mockRouter := createMockRouter("pathParamsMocks", t)
	testCases := []*testCase{
		{name: "Single pathparams, match ", request: &http.Request{URL: &url.URL{Path: "/pathParams/bar/foo"}, Method: "GET"}, match: true, requestParams: map[string]string{"pathParam": "bar"}},
		{name: "Single pathparams, No Match last segment does not match,  ", request: &http.Request{URL: &url.URL{Path: "/pathParams/bar/foos"}, Method: "GET"}, match: false},
		{name: "Multi pathparams, match ", request: &http.Request{URL: &url.URL{Path: "/multipathParams/val1/foo/val2"}, Method: "GET"}, match: true, requestParams: map[string]string{"pathParam1": "val1","pathParam2": "val2"}},
		
	}
	assertMatchRequestToEndpoint(mockRouter, testCases, t)
}

func assertMatchRequestToEndpoint(mockRouter *MockRouter, testCases []*testCase, t *testing.T) {
	for _, testCase := range testCases {
		ep, requestParams := mockRouter.matchRequestToEndpoint(testCase.request)
		if testCase.match && ep == nil {
			t.Errorf("testcase '%s' failed:  expect a match for request: %v", testCase.name, testCase.request)
		} else if !testCase.match && ep != nil {
			t.Errorf("testcase '%s' failed:  expect a no match for request: %v", testCase.name, testCase.request)
		} else {
			t.Logf("testcase '%s':'%s' passed", t.Name(), testCase.name)
		}
		if testCase.requestParams != nil {
			for expectedParamName, expectedParamValue := range testCase.requestParams {
				if requestParams[expectedParamName] != expectedParamValue {
					t.Errorf("testcase '%s' failed:  expect a request param with key: '%s' and value: '%s'. but get '%s' ", testCase.name, expectedParamName, expectedParamValue, requestParams[expectedParamName])
				}
			}
		}
	}

}

func createMockRouter(testMockDir string, t *testing.T) *MockRouter {
	mockRouter, err := NewMockRouter("../../test/"+testMockDir, "*-mock.yaml", "../../test/allMatchWildcardMocks", "*-response.json", &utils.Logger{Verbose: true})
	if err != nil {
		t.Fatalf("Can't create mock router: %v", err)
	}
	if mockRouter == nil {
		t.Fatal("Mockrouter must not be nil")
	}
	mockRouter.LoadMocks()
	if err != nil {
		t.Fatalf("Can't load mocks . %v", err)
	}
	return mockRouter
}
