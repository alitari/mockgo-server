package matches

import (
	"net/http"
	"os"
	"testing"

	"github.com/alitari/mockgo-server/mockgo/logging"
	"github.com/alitari/mockgo-server/mockgo/testutil"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

const (
	username = "mockgo"
	password = "password"
)

var matchesRequestHandler *MatchesRequestHandler

func TestMain(m *testing.M) {
	logger := logging.NewLoggerUtil(logging.Debug)
	matchesRequestHandler = NewMatchesRequestHandler("", username, password, NewInMemoryMatchstore(uint16(100)), logger)
	router := mux.NewRouter()
	matchesRequestHandler.AddRoutes(router)
	testutil.StartServing(router)
	code := testutil.RunAndCheckCoverage("matchesRequestHandlerTest", m, 0.40)
	testutil.StopServing()
	os.Exit(code)
}

func TestMatchesRequestHandler_health(t *testing.T) {
	assert.NoError(t, testutil.AssertResponseStatusOfRequestCall(t,
		testutil.CreateOutgoingRequest(t, http.MethodGet, "/health", testutil.CreateHeader(), ""), http.StatusOK))
}

func TestMatchesRequestHandler_getMatches(t *testing.T) {
	endpointId := "myEndpointId"
	err := matchesRequestHandler.matchStore.AddMatch(endpointId, createMatch(endpointId))
	assert.NoError(t, err)
	request := testutil.CreateOutgoingRequest(t, http.MethodGet, "/matches/"+endpointId,
		testutil.CreateHeader().WithAuth(username, password).WithJsonAccept(), "")
	testutil.AssertResponseOfRequestCall(t, request, func(response *http.Response, responseBody string) {
		assert.Equal(t, http.StatusOK, response.StatusCode)
		assert.Equal(t, `[{"endpointId":"myEndpointId","timestamp":"2009-11-17T20:34:58.651387237Z","actualRequest":{"method":"GET","url":"./http://myhost","header":null,"host":""},"actualResponse":null}]`, responseBody)
	})

}

func TestMatchesRequestHandler_getMismatches(t *testing.T) {

}
func TestMatchesRequestHandler_addMatches(t *testing.T) {

}
func TestMatchesRequestHandler_addMismatches(t *testing.T) {

}
func TestMatchesRequestHandler_deleteMatches(t *testing.T) {

}
func TestMatchesRequestHandler_deleteMismatches(t *testing.T) {

}

// func TestKVStoreRequestHandler_setKVStore(t *testing.T) {
// 	err := kvstoreHandler.kvstore.PutAll(map[string]interface{}{})
// 	assert.NoError(t, err)
// 	key := randString(5)
// 	util.RequestCall(t, httpClient, http.MethodPut, urlPrefix+"/kvstore/"+key,
// 		map[string][]string{headers.Authorization: {util.BasicAuth(username, password)}, headers.ContentType: {"application/json"}},
// 		`{ "testkey":"testvalue"}`, http.StatusNoContent, nil)
// 	val, err := kvstoreHandler.kvstore.Get(key)
// 	assert.NoError(t, err)
// 	assert.Equal(t, map[string]interface{}{"testkey": "testvalue"}, val)
// }

// func TestKVStoreRequestHandler_getKVStore(t *testing.T) {
// 	key := randString(5)
// 	err := kvstoreHandler.kvstore.PutAll(map[string]interface{}{key: "expectedVal", "key2": "val2 not expected"})
// 	assert.NoError(t, err)
// 	util.RequestCall(t, httpClient, http.MethodGet, urlPrefix+"/kvstore/"+key,
// 		map[string][]string{headers.Authorization: {util.BasicAuth(username, password)}, headers.Accept: {"application/json"}},
// 		"", http.StatusOK,
// 		func(responseBody string) {
// 			assert.Equal(t, `"expectedVal"`, responseBody)
// 		})
// }

// func TestKVStoreRequestHandler_addKVStore(t *testing.T) {
// 	err := kvstoreHandler.kvstore.PutAll(map[string]interface{}{})
// 	assert.NoError(t, err)
// 	key := randString(5)
// 	util.RequestCall(t, httpClient, http.MethodPost, urlPrefix+"/kvstore/"+key+"/add",
// 		map[string][]string{headers.Authorization: {util.BasicAuth(username, password)}, headers.ContentType: {"application/json"}},
// 		`{ "path": "/testpath", "value": "testvalue" }`, http.StatusNoContent, nil)
// 	val, err := kvstoreHandler.kvstore.Get(key)
// 	assert.NoError(t, err)
// 	assert.Equal(t, map[string]interface{}{"testpath": "testvalue"}, val)
// }

// func TestKVStoreRequestHandler_removeKVStore(t *testing.T) {
// 	key := randString(5)
// 	err := kvstoreHandler.kvstore.PutAll(map[string]interface{}{key: map[string]string{"deletepath": "deletzevalue"}})
// 	assert.NoError(t, err)
// 	util.RequestCall(t, httpClient, http.MethodPost, urlPrefix+"/kvstore/"+key+"/remove",
// 		map[string][]string{headers.Authorization: {util.BasicAuth(username, password)}, headers.ContentType: {"application/json"}},
// 		`{ "path": "/deletepath"}`, http.StatusNoContent, nil)
// 	all, err := kvstoreHandler.kvstore.GetAll()
// 	assert.NoError(t, err)
// 	assert.Equal(t, map[string]interface{}{key: map[string]interface{}{}}, all)
// }

// func TestKVStoreRequestHandler_uploadKVStore(t *testing.T) {
// 	err := kvstoreHandler.kvstore.PutAll(map[string]interface{}{})
// 	assert.NoError(t, err)
// 	util.RequestCall(t, httpClient, http.MethodPut, urlPrefix+"/kvstore",
// 		map[string][]string{headers.Authorization: {util.BasicAuth(username, password)}, headers.ContentType: {"application/json"}},
// 		`{"store1":"store1value","store2":{"key2":"value2"}}`, http.StatusNoContent, nil)
// 	all, err := kvstoreHandler.kvstore.GetAll()
// 	assert.NoError(t, err)
// 	assert.Equal(t, map[string]interface{}{"store1": "store1value", "store2": map[string]interface{}{"key2": "value2"}}, all)
// }

// func TestKVStoreRequestHandler_downloadKVStore(t *testing.T) {
// 	err := kvstoreHandler.kvstore.PutAll(map[string]interface{}{"store1": "store1value", "store2": map[string]interface{}{"key2": "value2"}})
// 	assert.NoError(t, err)
// 	util.RequestCall(t, httpClient, http.MethodGet, urlPrefix+"/kvstore",
// 		map[string][]string{headers.Authorization: {util.BasicAuth(username, password)}, headers.Accept: {"application/json"}},
// 		"", http.StatusOK,
// 		func(responseBody string) {
// 			assert.Equal(t, `{"store1":"store1value","store2":{"key2":"value2"}}`, responseBody)
// 		})
// }
