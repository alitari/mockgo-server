package routing

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/alitari/mockgo-server/internal/utils"
	"github.com/go-http-utils/headers"
)

type configRouterTestCase struct {
	name                       string
	request                    *http.Request
	expectedResponseStatusCode int
	expectedResponseFile       string
}

func TestConfigRouter_Endpoints(t *testing.T) {
	mockRouter := createMockRouter("simplemocks", t)
	configRouter := NewConfigRouter(mockRouter, &utils.Logger{Verbose: true, DebugResponseRendering: true})
	configRouter.newRouter()
	testCases := []*configRouterTestCase{
		{name: "Endpoints",
			request: createRequest(
				http.MethodGet,
				"http://somehost/endpoints",
				"",
				map[string][]string{headers.ContentType: {"application/json"}, headers.Accept: {"application/json"}},
				nil,
				t),
			expectedResponseStatusCode: 200,
			expectedResponseFile:       "../../test/expectedResponses/endpoints.json",
		},
	}
	assertConfigRouterResponse(configRouter.router.Get("endpoints").GetHandler(), testCases, t)
}

func TestConfigRouter_SetKVStore(t *testing.T) {
	mockRouter := createMockRouter("simplemocks", t)
	configRouter := NewConfigRouter(mockRouter, &utils.Logger{Verbose: true, DebugResponseRendering: true})
	configRouter.newRouter()

	testCases := []*configRouterTestCase{
		{name: "KVStore",
			request: createRequest(
				http.MethodPut,
				"http://somehost/kvstore/testapp",
				"{ \"mykey\": \"myvalue\" }",
				map[string][]string{headers.ContentType: {"application/json"}, headers.Accept: {"application/json"}},
				map[string]string{"key": "testapp"},
				t),
			expectedResponseStatusCode: http.StatusNoContent,
		},
	}
	assertConfigRouterResponse(configRouter.router.Get("setKVStore").GetHandler(), testCases, t)
	value, err := mockRouter.kvstore.Get("testapp")
	if err != nil {
		t.Fatal(err)
	}
	if fmt.Sprintf("%v", value) != "&map[mykey:myvalue]" {
		t.Errorf("Expected kv store value of key %s, is %s , but is %s", "myapp", "&map[mykey:myvalue]", fmt.Sprintf("%v", value))
	}
}

func TestConfigRouter_GetKVStore(t *testing.T) {
	mockRouter := createMockRouter("simplemocks", t)
	configRouter := NewConfigRouter(mockRouter, &utils.Logger{Verbose: true, DebugResponseRendering: true})
	configRouter.newRouter()

	val := "{ \"myconfig\": \"is here!\" }"
	err := mockRouter.kvstore.Put("testapp", val)
	if err != nil {
		t.Fatal(err)
	}
	testCases := []*configRouterTestCase{
		{name: "GetKVStore",
			request: createRequest(
				http.MethodGet,
				"http://somehost/kvstore/testapp",
				"",
				map[string][]string{headers.Accept: {"application/json"}},
				map[string]string{"key": "testapp"},
				t),
			expectedResponseStatusCode: http.StatusOK,
			expectedResponseFile:       "../../test/expectedResponses/kvstoreValue.json",
		},
	}
	assertConfigRouterResponse(configRouter.router.Get("getKVStore").GetHandler(), testCases, t)
}

func assertConfigRouterResponse(handler http.Handler, testCases []*configRouterTestCase, t *testing.T) {
	for _, testCase := range testCases {
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, testCase.request)
		if testCase.expectedResponseStatusCode != recorder.Result().StatusCode {
			t.Errorf("Expected status code %v, but is %v", testCase.expectedResponseStatusCode, recorder.Result().StatusCode)
		}
		responseBody, err := io.ReadAll(recorder.Result().Body)
		if err != nil {
			t.Fatal(err)
		}

		if len(testCase.expectedResponseFile) > 0 {
			expectedResponse, err := os.ReadFile(testCase.expectedResponseFile)
			if err != nil {
				t.Fatal(err)
			}
			if string(expectedResponse) != string(responseBody) {
				t.Errorf("Expected response  is:\n%s,\nbut is:\n%s", expectedResponse, responseBody)
			}
		}

		if !t.Failed() {
			t.Logf("testcase '%s':'%s' passed", t.Name(), testCase.name)
		}
	}
}
