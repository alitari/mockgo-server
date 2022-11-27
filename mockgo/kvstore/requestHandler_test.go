package kvstore

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

var kvstoreHandler *KVStoreRequestHandler

func TestMain(m *testing.M) {
	kvstoreLogger := logging.NewLoggerUtil(logging.Debug)
	kvstoreJson := NewKVStoreJSON(NewInmemoryKVStore(), true)
	kvstoreHandler = NewKVStoreRequestHandler("", username, password, kvstoreJson, kvstoreLogger)
	router := mux.NewRouter()
	kvstoreHandler.AddRoutes(router)
	testutil.StartServing(router)
	code := testutil.RunAndCheckCoverage("requestHandlerTest", m, 0.49)
	testutil.StopServing()
	os.Exit(code)
}

func TestKVStoreRequestHandler_health(t *testing.T) {
	assert.NoError(t, testutil.AssertResponseStatusOfRequestCall(t,
		testutil.CreateRequest(t, http.MethodGet, "/health", testutil.CreateHeader(), ""), http.StatusOK))
}

func TestKVStoreRequestHandler_setKVStore(t *testing.T) {
	key := randString(5)
	err := kvstoreHandler.kvstore.Put(key, nil)
	assert.NoError(t, err)
	request := testutil.CreateRequest(t, http.MethodPut, "/kvstore/"+key,
		testutil.CreateHeader().WithAuth(username, password).WithJsonContentType(),
		`{ "testkey":"testvalue"}`)
	assert.NoError(t, testutil.AssertResponseStatusOfRequestCall(t, request, http.StatusNoContent))
	val, err := kvstoreHandler.kvstore.Get(key)
	assert.NoError(t, err)
	assert.Equal(t, map[string]interface{}{"testkey": "testvalue"}, val)
}

func TestKVStoreRequestHandler_getKVStore(t *testing.T) {
	key := randString(5)
	err := kvstoreHandler.kvstore.Put(key, "expectedVal")
	assert.NoError(t, err)
	err = kvstoreHandler.kvstore.Put("key2", "val2 not expected")
	assert.NoError(t, err)
	request := testutil.CreateRequest(t, http.MethodGet, "/kvstore/"+key,
		testutil.CreateHeader().WithAuth(username, password).WithJsonAccept(), "")
	assert.NoError(t, testutil.AssertResponseOfRequestCall(t, request, func(response *http.Response, responseBody string) {
		assert.Equal(t, "expectedVal", responseBody)
	}))
}

func TestKVStoreRequestHandler_addKVStore(t *testing.T) {
	key := randString(5)
	err := kvstoreHandler.kvstore.Put(key, nil)
	assert.NoError(t, err)
	request := testutil.CreateRequest(t, http.MethodPost, "/kvstore/"+key+"/add",
		testutil.CreateHeader().WithAuth(username, password).WithJsonContentType(),
		`{ "path": "/testpath", "value": "testvalue" }`)
	assert.NoError(t, testutil.AssertResponseStatusOfRequestCall(t, request, http.StatusNoContent))
	val, err := kvstoreHandler.kvstore.Get(key)
	assert.NoError(t, err)
	assert.Equal(t, map[string]interface{}{"testpath": "testvalue"}, val)
}

func TestKVStoreRequestHandler_removeKVStore(t *testing.T) {
	key := randString(5)
	err := kvstoreHandler.kvstore.Put(key, map[string]string{"deletepath": "deletzevalue"})
	assert.NoError(t, err)
	request := testutil.CreateRequest(t, http.MethodPost, "/kvstore/"+key+"/remove",
		testutil.CreateHeader().WithAuth(username, password).WithJsonContentType(), `{ "path": "/deletepath"}`)
	assert.NoError(t, testutil.AssertResponseStatusOfRequestCall(t, request, http.StatusNoContent))
	all, err := kvstoreHandler.kvstore.Get(key)
	assert.NoError(t, err)
	assert.Equal(t, map[string]interface{}{}, all)
}
