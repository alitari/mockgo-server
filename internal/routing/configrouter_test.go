package routing

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/alitari/mockgo-server/internal/utils"
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

	testCases := []*configRouterTestCase{
		{name: "Endpoints",
			request:                    &http.Request{},
			expectedResponseStatusCode: 200,
			expectedResponseFile:       "../../test/expectedResponses/endpoints.json",
		},
	}
	assertConfigRouterResponse(func(request *http.Request, recorder *httptest.ResponseRecorder) {
		configRouter.endpoints(recorder, request)
	}, testCases, t)

}

func TestConfigRouter_KVStore(t *testing.T) {
	mockRouter := createMockRouter("simplemocks", t)
	configRouter := NewConfigRouter(mockRouter, &utils.Logger{Verbose: true, DebugResponseRendering: true})
	testCases := []*configRouterTestCase{
		{name: "KVStore",
			request: createRequest(
				http.MethodPut,
				"http://somehost/kvstore/mypp",
				"{ \"mykey\": \"myvalue\" }",
				map[string][]string{"Content-Type": {"application/json"}},
				map[string]string{"key": "testapp"},
				t),
			expectedResponseStatusCode: http.StatusNoContent,
		},
	}
	assertConfigRouterResponse(func(request *http.Request, recorder *httptest.ResponseRecorder) {
		configRouter.setKVStore(recorder, request)
	}, testCases, t)
	value, err := mockRouter.kvstore.Get("testapp")
	if err != nil {
		t.Fatal(err)
	}
	switch value := value.(type) {
	case *map[string]interface{}:
		if fmt.Sprintf("%v", value) != "&map[mykey:myvalue]" {
			t.Errorf("Expected kv store value of key %s, is %s , but is %s", "mykey", "&map[mykey:myvalue]", fmt.Sprintf("%v", value))
		}
	default:
		t.Errorf("Wrong expected kvstore value type, expected is map[string]string, but is %T", value)
	}
}

func assertConfigRouterResponse(routerCall func(*http.Request, *httptest.ResponseRecorder), testCases []*configRouterTestCase, t *testing.T) {
	for _, testCase := range testCases {
		recorder := httptest.NewRecorder()
		routerCall(testCase.request, recorder)
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
