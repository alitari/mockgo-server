package routing

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/alitari/mockgo-server/internal/utils"
)

type testCase struct {
	name    string
	request *http.Request
	match   bool
}

func TestMatchRequestToEndpoint_Simplemocks(t *testing.T) {
	mockRouter, err := NewMockRouter("../../test/data/simplemocks", "*-mock.yaml", "../../test/data/simplemocks", "*-response.json", &utils.Logger{Verbose: true})
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

	testCases := []testCase{
		{name: "Minimal: Match, full request",
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

	for _, testCase := range testCases {

		ep := mockRouter.matchRequestToEndpoint(testCase.request)
		if testCase.match && ep == nil {
			t.Errorf("testcase '%s' failed:  expect a match for request: %v", testCase.name, testCase.request)
		} else if !testCase.match && ep != nil {
			t.Errorf("testcase '%s' failed:  expect a no match for request: %v", testCase.name, testCase.request)
		} else {
			t.Logf("testcase '%s':'%s' passed", t.Name(), testCase.name)
		}
	}

}

func TestMatchRequestToEndpoint_Wildcardmocks(t *testing.T) {
	mockRouter, err := NewMockRouter("../../test/data/wildcardmocks", "*-mock.yaml", "../../test/data/wildcardmocks", "*-response.json", &utils.Logger{Verbose: true})
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

	testCases := []testCase{
		{name: "Single wildcard Mock: Match 1 ", request: &http.Request{URL: &url.URL{Path: "/wildcard/bar/foo"}, Method: "GET"}, match: true},
		{name: "Single wildcard Mock: Match 2", request: &http.Request{URL: &url.URL{Path: "/wildcard/foo/foo"}, Method: "GET"}, match: true},
		{name: "Single wildcard Mock: No match, first path segment", request: &http.Request{URL: &url.URL{Path: "/wildcards/bar/foo"}, Method: "GET"}, match: false},
		{name: "Single wildcard Mock: No match, path too long ", request: &http.Request{URL: &url.URL{Path: "/wildcard/bar/foo/toolong"}, Method: "GET"}, match: false},
		{name: "Single wildcard Mock: No match, path too short ", request: &http.Request{URL: &url.URL{Path: "/bar/foo"}, Method: "GET"}, match: false},
	}

	for _, testCase := range testCases {

		ep := mockRouter.matchRequestToEndpoint(testCase.request)
		if testCase.match && ep == nil {
			t.Errorf("testcase '%s' failed:  expect a match for request: %v", testCase.name, testCase.request)
		} else if !testCase.match && ep != nil {
			t.Errorf("testcase '%s' failed:  expect a no match for request: %v", testCase.name, testCase.request)
		} else {
			t.Logf("testcase '%s':'%s' passed", t.Name(), testCase.name)
		}
	}

}
