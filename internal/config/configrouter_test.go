package config

import (
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/alitari/mockgo-server/internal/kvstore"
	"github.com/alitari/mockgo-server/internal/mock"
	"github.com/alitari/mockgo-server/internal/model"
	"github.com/alitari/mockgo-server/internal/utils"
	"github.com/go-http-utils/headers"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

const password = "password"

type configRouterTestCase struct {
	name                       string
	request                    *http.Request
	expectedResponseStatusCode int
	expectedResponseFile       string
}

func TestMain(m *testing.M) {
	os.Exit(utils.RunAndCheckCoverage("configrouter", m, 0.30))
}

func TestConfigRouter_UploadKVStore(t *testing.T) {
	mockRouter := createMockRouter(t, "minmaxmocks", false, true)
	configRouter := NewConfigRouter("mockgo", password, mockRouter, 0, []string{}, kvstore.CreateTheStore(), &utils.Logger{Verbose: true, DebugResponseRendering: true})
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
	assert.Equal(t, map[string]interface{}{"store1": map[string]interface{}{"mykey1": "myvalue1"}, "store2": map[string]interface{}{"mykey2": "myvalue2"}}, kvstore.TheKVStore.GetAll())
}

func TestConfigRouter_DownloadKVStore(t *testing.T) {
	mockRouter := createMockRouter(t, "minmaxmocks", false, true)
	kvstore.CreateTheStore()
	err := kvstore.TheKVStore.PutAllJson("{ \"store1\" : { \"mykey1\" : \"myvalue1\"}, \"store2\" : { \"mykey2\" : \"myvalue2\"} }")
	assert.NoError(t, err)
	configRouter := NewConfigRouter("mockgo", password, mockRouter, 0, []string{}, kvstore.TheKVStore, &utils.Logger{Verbose: true, DebugResponseRendering: true})
	configRouter.newRouter()

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
	kvstore.CreateTheStore()
	mockRouter := createMockRouter(t, "minmaxmocks", false, true)
	configRouter := NewConfigRouter("mockgo", password, mockRouter, 0, []string{}, kvstore.TheKVStore, &utils.Logger{Verbose: true, DebugResponseRendering: true})
	configRouter.newRouter()

	testCases := []*configRouterTestCase{
		{name: "KVStore",
			request: createRequest(
				http.MethodPut,
				"http://somehost/kvstore/testapp",
				"{ \"mykey\": \"myvalue\" }",
				map[string][]string{headers.ContentType: {"application/json"}},
				map[string]string{"key": "testapp"},
				t),
			expectedResponseStatusCode: http.StatusNoContent,
		},
	}
	assertConfigRouterResponse(configRouter.router.Get("setKVStore").GetHandler(), testCases, t)
	value := kvstore.TheKVStore.Get("testapp")
	assert.Equal(t, map[string]interface{}{"mykey": "myvalue"}, value)
}

func TestConfigRouter_GetKVStore(t *testing.T) {
	mockRouter := createMockRouter(t, "minmaxmocks", false, true)
	kvstore.CreateTheStore()
	configRouter := NewConfigRouter("mockgo", password, mockRouter, 0, []string{}, kvstore.TheKVStore, &utils.Logger{Verbose: true, DebugResponseRendering: true})
	configRouter.newRouter()

	val := "{ \"myconfig\": \"is here!\" }"
	err := configRouter.kvstore.PutAsJson("testapp", val)
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

func TestConfigRouter_AddKVStore(t *testing.T) {
	mockRouter := createMockRouter(t, "minmaxmocks", false, true)
	kvstore.CreateTheStore()
	configRouter := NewConfigRouter("mockgo", password, mockRouter, 0, []string{}, kvstore.TheKVStore, &utils.Logger{Verbose: true, DebugResponseRendering: true})
	configRouter.newRouter()

	val := `{ "myconfig": "is here!" }`
	err := configRouter.kvstore.PutAsJson("testapp", val)
	assert.NoError(t, err)

	testCases := []*configRouterTestCase{
		{name: "AddKVStore",
			request: createRequest(
				http.MethodPost,
				"http://somehost/kvstore/testapp/add",
				`{ "path": "/myconfig2", "value": "is also here" }`,
				map[string][]string{headers.ContentType: {"application/json"}},
				map[string]string{"key": "testapp"},
				t),
			expectedResponseStatusCode: http.StatusNoContent,
		},
	}
	assertConfigRouterResponse(configRouter.router.Get("addKVStore").GetHandler(), testCases, t)
	value := kvstore.TheKVStore.Get("testapp")
	assert.Equal(t, map[string]interface{}{"myconfig": "is here!", "myconfig2": "is also here"}, value)
}

func TestConfigRouter_RemoveKVStore(t *testing.T) {
	mockRouter := createMockRouter(t, "minmaxmocks", false, true)
	kvstore.CreateTheStore()
	configRouter := NewConfigRouter("mockgo", password, mockRouter, 0, []string{}, kvstore.TheKVStore, &utils.Logger{Verbose: true, DebugResponseRendering: true})
	configRouter.newRouter()

	val := `{ "myconfig": "is here!" }`
	err := configRouter.kvstore.PutAsJson("testapp", val)
	assert.NoError(t, err)

	testCases := []*configRouterTestCase{
		{name: "RemoveKVStore",
			request: createRequest(
				http.MethodPost,
				"http://somehost/kvstore/testapp/remove",
				`{ "path": "/myconfig" }`,
				map[string][]string{headers.ContentType: {"application/json"}},
				map[string]string{"key": "testapp"},
				t),
			expectedResponseStatusCode: http.StatusNoContent,
		},
	}
	assertConfigRouterResponse(configRouter.router.Get("removeKVStore").GetHandler(), testCases, t)
	value := kvstore.TheKVStore.Get("testapp")
	assert.Empty(t, value)
}

func TestConfigRouter_GetMatches(t *testing.T) {
	mockRouter := createMockRouter(t, "minmaxmocks", false, true)

	actualRequest := &model.ActualRequest{Method: http.MethodGet, URL: "http://mytesturl", Header: map[string][]string{}, Host: "myhost"}
	match := &model.Match{EndpointId: "endpointId", Timestamp: time.Date(
		2009, 11, 17, 20, 34, 58, 651387237, time.UTC), ActualRequest: actualRequest}
	mockRouter.Matches["someEndpointId"] = append(mockRouter.Matches["someEndpointId"], match)

	kvstore.CreateTheStore()
	configRouter := NewConfigRouter("mockgo", password, mockRouter, 0, []string{}, kvstore.TheKVStore, &utils.Logger{Verbose: true, DebugResponseRendering: true})
	configRouter.newRouter()

	testCases := []*configRouterTestCase{
		{name: "GetMatches",
			request: createRequest(
				http.MethodGet,
				"http://somehost/matches",
				"",
				map[string][]string{headers.Accept: {"application/json"}},
				nil,
				t),
			expectedResponseStatusCode: http.StatusOK,
			expectedResponseFile:       "../../test/expectedResponses/matches.json",
		},
	}
	assertConfigRouterResponse(configRouter.router.Get("getMatches").GetHandler(), testCases, t)
}

func TestConfigRouter_GetMatchesCountOnly(t *testing.T) {
	mockRouter := createMockRouter(t, "minmaxmocks", true, true)
	mockRouter.MatchesCount["someEndpointId"] = 42
	kvstore.CreateTheStore()
	configRouter := NewConfigRouter("mockgo", password, mockRouter, 0, []string{}, kvstore.TheKVStore, &utils.Logger{Verbose: true, DebugResponseRendering: true})
	configRouter.newRouter()

	testCases := []*configRouterTestCase{
		{name: "GetMatches",
			request: createRequest(
				http.MethodGet,
				"http://somehost/matches",
				"",
				map[string][]string{headers.Accept: {"application/json"}},
				nil,
				t),
			expectedResponseStatusCode: http.StatusOK,
			expectedResponseFile:       "../../test/expectedResponses/matchesCountOnly.json",
		},
	}
	assertConfigRouterResponse(configRouter.router.Get("getMatches").GetHandler(), testCases, t)
}

func TestConfigRouter_DeleteMatches(t *testing.T) {
	mockRouter := createMockRouter(t, "minmaxmocks", false, true)

	actualRequest := &model.ActualRequest{Method: http.MethodGet, URL: "http://mytesturl", Header: map[string][]string{}, Host: "myhost"}
	match := &model.Match{EndpointId: "endpointId", Timestamp: time.Date(
		2009, 11, 17, 20, 34, 58, 651387237, time.UTC), ActualRequest: actualRequest}
	mockRouter.Matches["someEndpointId"] = append(mockRouter.Matches["someEndpointId"], match)

	kvstore.CreateTheStore()
	configRouter := NewConfigRouter("mockgo", password, mockRouter, 0, []string{}, kvstore.TheKVStore, &utils.Logger{Verbose: true, DebugResponseRendering: true})
	configRouter.newRouter()

	testCases := []*configRouterTestCase{
		{name: "DeleteMatches",
			request: createRequest(
				http.MethodDelete,
				"http://somehost/matches",
				"",
				map[string][]string{},
				nil,
				t),
			expectedResponseStatusCode: http.StatusOK,
			expectedResponseFile:       "../../test/expectedResponses/nocontent.json",
		},
	}
	assertConfigRouterResponse(configRouter.router.Get("deleteMatches").GetHandler(), testCases, t)
	assert.Empty(t, mockRouter.Matches)
}

func TestConfigRouter_AddMatches(t *testing.T) {
	mockRouter := createMockRouter(t, "minmaxmocks", false, true)
	kvstore.CreateTheStore()
	configRouter := NewConfigRouter("mockgo", password, mockRouter, 0, []string{}, kvstore.TheKVStore, &utils.Logger{Verbose: true, DebugResponseRendering: true})
	configRouter.newRouter()
	matchesToAdd1 := map[string][]*model.Match{"id1": {&model.Match{EndpointId: "id1",
		ActualRequest: &model.ActualRequest{Method: http.MethodGet, URL: "http://myurl"}, ActualResponse: &model.ActualResponse{StatusCode: http.StatusAccepted}}}}

	matchesToAddStr1, err := json.Marshal(matchesToAdd1)
	if err != nil {
		assert.NoError(t, err)
	}

	matchesToAdd2 := map[string][]*model.Match{"id2": {&model.Match{EndpointId: "id2",
		ActualRequest: &model.ActualRequest{Method: http.MethodGet, URL: "http://myurl2"}, ActualResponse: &model.ActualResponse{StatusCode: http.StatusOK}}}}

	matchesToAddStr2, err := json.Marshal(matchesToAdd2)
	if err != nil {
		assert.NoError(t, err)
	}

	matchesToAdd3 := map[string][]*model.Match{"id1": {&model.Match{EndpointId: "id1",
		ActualRequest: &model.ActualRequest{Method: http.MethodGet, URL: "http://myurl"}, ActualResponse: &model.ActualResponse{StatusCode: http.StatusAccepted}}}}

	matchesToAddStr3, err := json.Marshal(matchesToAdd3)
	if err != nil {
		assert.NoError(t, err)
	}

	testCases := []*configRouterTestCase{
		{name: "AddMatches1",
			request: createRequest(
				http.MethodPost,
				"http://somehost/addmatches",
				string(matchesToAddStr1),
				map[string][]string{headers.ContentType: {"application/json"}},
				nil,
				t),
			expectedResponseStatusCode: http.StatusOK,
			expectedResponseFile:       "",
		},
		{name: "AddMatches2",
			request: createRequest(
				http.MethodPost,
				"http://somehost/addmatches",
				string(matchesToAddStr2),
				map[string][]string{headers.ContentType: {"application/json"}},
				nil,
				t),
			expectedResponseStatusCode: http.StatusOK,
			expectedResponseFile:       "",
		},
		{name: "AddMatches3",
			request: createRequest(
				http.MethodPost,
				"http://somehost/addmatches",
				string(matchesToAddStr3),
				map[string][]string{headers.ContentType: {"application/json"}},
				nil,
				t),
			expectedResponseStatusCode: http.StatusOK,
			expectedResponseFile:       "",
		},
	}
	assertConfigRouterResponse(configRouter.router.Get("addMatches").GetHandler(), testCases, t)
	expectedMatches := map[string][]*model.Match{"id1": append(matchesToAdd1["id1"], matchesToAdd3["id1"]...), "id2": matchesToAdd2["id2"]}
	assert.EqualValues(t, expectedMatches, mockRouter.Matches)
}

func TestConfigRouter_AddMatchesCountOnly(t *testing.T) {
	mockRouter := createMockRouter(t, "minmaxmocks", true, true)
	kvstore.CreateTheStore()
	configRouter := NewConfigRouter("mockgo", password, mockRouter, 0, []string{}, kvstore.TheKVStore, &utils.Logger{Verbose: true, DebugResponseRendering: true})
	configRouter.newRouter()

	matchesCount := []int64{rand.Int63(), rand.Int63(), rand.Int63()}
	testCases := []*configRouterTestCase{
		{name: "AddMatches1",
			request: createRequest(
				http.MethodPost,
				"http://somehost/addmatches",
				fmt.Sprintf(`{ "id1":%d }`, matchesCount[0]),
				map[string][]string{headers.ContentType: {"application/json"}},
				nil,
				t),
			expectedResponseStatusCode: http.StatusOK,
			expectedResponseFile:       "",
		},
		{name: "AddMatches2",
			request: createRequest(
				http.MethodPost,
				"http://somehost/addmatches",
				fmt.Sprintf(`{ "id2":%d }`, matchesCount[1]),
				map[string][]string{headers.ContentType: {"application/json"}},
				nil,
				t),
			expectedResponseStatusCode: http.StatusOK,
			expectedResponseFile:       "",
		},
		{name: "AddMatches3",
			request: createRequest(
				http.MethodPost,
				"http://somehost/addmatches",
				fmt.Sprintf(`{ "id1":%d }`, matchesCount[2]),
				map[string][]string{headers.ContentType: {"application/json"}},
				nil,
				t),
			expectedResponseStatusCode: http.StatusOK,
			expectedResponseFile:       "",
		},
	}
	assertConfigRouterResponse(configRouter.router.Get("addMatches").GetHandler(), testCases, t)
	assert.Empty(t, mockRouter.Matches)
	assert.EqualValues(t, map[string]int64{"id1": matchesCount[0] + matchesCount[2], "id2": matchesCount[1]}, mockRouter.MatchesCount)
}

func TestConfigRouter_DownloadKVStoreFromCluster(t *testing.T) {
	kvstore.CreateTheStore()
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

	mockRouter := createMockRouter(t, "minmaxmocks", false, true)
	configRouter := NewConfigRouter("mockgo", password, mockRouter, 0, []string{clusterNode1.URL, clusterNode2.URL}, kvstore.TheKVStore, &utils.Logger{Verbose: true, DebugResponseRendering: true})
	configRouter.newRouter()
	configRouter.DownloadKVStoreFromCluster()

	assert.NotNil(t, clusterNode1Request, "clusterNode1Request must exist")
	assert.Nil(t, clusterNode2Request, "clusterNode2Request must not exist")
	assert.Equal(t, http.MethodGet, clusterNode1Request.Method, "clusterNode1Request.Method unexpected")
	assert.Equal(t, "/kvstore", clusterNode1Request.URL.Path, "clusterNode1Request path unexpected")
	assert.Equal(t, map[string]interface{}{"store1": map[string]interface{}{"key1": "value1"}}, kvstore.TheKVStore.GetAll())
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
			assert.Equal(t, string(expectedResponse), string(responseBody))
		}
	}
}

func createMockRouter(t *testing.T, testMockDir string, matchesCountOnly, matchesRecordMisMatch bool) *mock.MockRouter {
	mockRouter := mock.NewMockRouter("../../test/"+testMockDir, "*-mock.yaml", "../../test/"+testMockDir, 0, kvstore.TheKVStore, matchesCountOnly, matchesRecordMisMatch, &utils.Logger{Verbose: true, DebugResponseRendering: true})
	assert.NotNil(t, mockRouter, "Mockrouter must not be nil")
	err := mockRouter.LoadFiles(nil)
	assert.NoError(t, err)
	return mockRouter
}

func createRequest(method, url, bodyStr string, header map[string][]string, urlVars map[string]string, t *testing.T) *http.Request {
	body := io.NopCloser(strings.NewReader(bodyStr))
	request := httptest.NewRequest(method, url, body)
	header["Authorization"] = []string{utils.BasicAuth("mockgo", "password")}
	request.Header = header
	if urlVars != nil {
		request = mux.SetURLVars(request, urlVars)
	}
	return request
}
