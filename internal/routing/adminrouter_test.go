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

type adminRouterTestCase struct {
	name                       string
	request                    *http.Request
	expectedResponseStatusCode int
	expectedResponseFile       string
}

func TestAdminRouter_Endpoints(t *testing.T) {
	mockRouter := createMockRouter("simplemocks", t)
	adminRouter := NewAdminRouter(mockRouter, &utils.Logger{Verbose: true, DebugResponseRendering: true})

	testCases := []*adminRouterTestCase{
		{name: "Endpoints",
			request:                    &http.Request{},
			expectedResponseStatusCode: 200,
			expectedResponseFile:       "../../test/expectedResponses/endpoints.json",
		},
	}
	assertAdminRouterResponse(func(request *http.Request, recorder *httptest.ResponseRecorder) {
		adminRouter.endpoints(recorder, request)
	}, testCases, t)

}

func TestAdminRouter_KVStore(t *testing.T) {
	mockRouter := createMockRouter("simplemocks", t)
	adminRouter := NewAdminRouter(mockRouter, &utils.Logger{Verbose: true, DebugResponseRendering: true})
	testCases := []*adminRouterTestCase{
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
	assertAdminRouterResponse(func(request *http.Request, recorder *httptest.ResponseRecorder) {
		adminRouter.setKVStore(recorder, request)
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

func assertAdminRouterResponse(routerCall func(*http.Request, *httptest.ResponseRecorder), testCases []*adminRouterTestCase, t *testing.T) {
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
