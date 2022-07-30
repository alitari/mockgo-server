package routing

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/alitari/mockgo-server/internal/kvstore"
	"github.com/alitari/mockgo-server/internal/utils"
	"github.com/go-http-utils/headers"
	"github.com/stretchr/testify/assert"
)

type configRouterTestCase struct {
	name                       string
	request                    *http.Request
	expectedResponseStatusCode int
	expectedResponseFile       string
}

func TestConfigRouter_Endpoints(t *testing.T) {
	mockRouter := createMockRouter("simplemocks", t)
	configRouter := NewConfigRouter(mockRouter, []string{}, &utils.Logger{Verbose: true, DebugResponseRendering: true})
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

func TestConfigRouter_UploadKVStore(t *testing.T) {
	mockRouter := createMockRouter("simplemocks", t)
	configRouter := NewConfigRouter(mockRouter, []string{}, &utils.Logger{Verbose: true, DebugResponseRendering: true})
	configRouter.newRouter()

	testCases := []*configRouterTestCase{
		{name: "UploadKVStore",
			request: createRequest(
				http.MethodPut,
				"http://somehost/kvstore/",
				"{ \"store1\" : { \"mykey1\" : \"myvalue1\"}, \"store2\" : { \"mykey2\" : \"myvalue2\"} }",
				map[string][]string{headers.ContentType: {"application/json"}},
				nil,
				t),
			expectedResponseStatusCode: http.StatusNoContent,
		},
	}
	assertConfigRouterResponse(configRouter.router.Get("uploadKVStore").GetHandler(), testCases, t)
	assert.Equal(t, map[string]*map[string]interface{}{"store1": {"mykey1": "myvalue1"}, "store2": {"mykey2": "myvalue2"}}, mockRouter.kvstore.GetAll())
}

func TestConfigRouter_DownloadKVStore(t *testing.T) {
	mockRouter := createMockRouter("simplemocks", t)
	configRouter := NewConfigRouter(mockRouter, []string{}, &utils.Logger{Verbose: true, DebugResponseRendering: true})
	configRouter.newRouter()
	store, err := kvstore.NewStoreWithContent("{ \"store1\" : { \"mykey1\" : \"myvalue1\"}, \"store2\" : { \"mykey2\" : \"myvalue2\"} }")
	assert.NoError(t, err)
	configRouter.mockRouter.kvstore = store

	testCases := []*configRouterTestCase{
		{name: "DownloadKVStore",
			request: createRequest(
				http.MethodGet,
				"http://somehost/kvstore/",
				"",
				map[string][]string{headers.Accept: {"application/json"}},
				nil,
				t),
			expectedResponseStatusCode: http.StatusOK,
			expectedResponseFile:       "../../test/expectedResponses/kvstoreAll.json",
		},
	}
	assertConfigRouterResponse(configRouter.router.Get("downloadKVStore").GetHandler(), testCases, t)
}

func TestConfigRouter_SetKVStore(t *testing.T) {
	mockRouter := createMockRouter("simplemocks", t)
	configRouter := NewConfigRouter(mockRouter, []string{}, &utils.Logger{Verbose: true, DebugResponseRendering: true})
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
	assert.NoError(t, err)
	assert.Equal(t, &map[string]interface{}{"mykey": "myvalue"}, value)
}

func TestConfigRouter_GetKVStore(t *testing.T) {
	mockRouter := createMockRouter("simplemocks", t)
	configRouter := NewConfigRouter(mockRouter, []string{}, &utils.Logger{Verbose: true, DebugResponseRendering: true})
	configRouter.newRouter()

	val := "{ \"myconfig\": \"is here!\" }"
	err := mockRouter.kvstore.Put("testapp", val)
	assert.NoError(t, err)

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

func TestConfigRouter_SyncWithCluster(t *testing.T) {
	var clusterNode1Request *http.Request
	var clusterNode2Request *http.Request
	clusterNode1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clusterNode1Request = r
		io.WriteString(w, `{ "store1": { "key1" : "value1" }}`)
	}))
	defer clusterNode1.Close()

	clusterNode2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clusterNode2Request = r
	}))
	defer clusterNode2.Close()

	mockRouter := createMockRouter("simplemocks", t)
	configRouter := NewConfigRouter(mockRouter, []string{clusterNode1.URL, clusterNode2.URL}, &utils.Logger{Verbose: true, DebugResponseRendering: true})
	configRouter.newRouter()
	configRouter.SyncWithCluster()

	assert.NotNil(t, clusterNode1Request, "clusterNode1Request must exist")
	assert.Nil(t, clusterNode2Request, "clusterNode2Request must not exist")
	assert.Equal(t, http.MethodGet, clusterNode1Request.Method, "clusterNode1Request.Method unexpected")
	assert.Equal(t, "/kvstore", clusterNode1Request.URL.Path, "clusterNode1Request path unexpected")
	assert.Equal(t, map[string]*map[string]interface{}{"store1": {"key1": "value1"}}, mockRouter.kvstore.GetAll())
}

func assertConfigRouterResponse(handler http.Handler, testCases []*configRouterTestCase, t *testing.T) {
	for _, testCase := range testCases {
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, testCase.request)
		assert.Equal(t, testCase.expectedResponseStatusCode, recorder.Result().StatusCode, "response status code unexpected")
		responseBody, err := io.ReadAll(recorder.Result().Body)
		assert.NoError(t, err)

		if len(testCase.expectedResponseFile) > 0 {
			expectedResponse, err := os.ReadFile(testCase.expectedResponseFile)
			assert.NoError(t, err)
			assert.Equal(t, expectedResponse, responseBody)
		}
	}
}
