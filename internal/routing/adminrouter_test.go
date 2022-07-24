package routing

import (
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
			expectedResponseFile: "../../test/expectedResponses/endpoints.json",
		},
	}
	assertAdminRouterResponse(func(request *http.Request, recorder *httptest.ResponseRecorder) {
		adminRouter.endpoints(recorder, request)
	}, testCases, t)

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
