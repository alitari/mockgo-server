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

var kvstoreHandler *RequestHandler

func TestMain(m *testing.M) {
	kvstoreLogger := logging.NewLoggerUtil(logging.Debug)
	kvstoreHandler = NewRequestHandler("", username, password, NewInmemoryStorage(), kvstoreLogger)
	router := mux.NewRouter()
	kvstoreHandler.AddRoutes(router)
	testutil.StartServing(router)
	code := testutil.RunAndCheckCoverage("requestHandlerTest", m, 0.49)
	testutil.StopServing()
	os.Exit(code)
}

func TestKVStoreRequestHandler_health(t *testing.T) {
	request := testutil.CreateIncomingRequest(http.MethodGet, "/health", testutil.CreateHeader(), "")
	assert.NoError(t, testutil.AssertHandlerFunc(t, request, kvstoreHandler.handleHealth, func(response *http.Response, responseBody string) {
		assert.Equal(t, http.StatusOK, response.StatusCode)
	},
	))
}

func TestKVStoreRequestHandler_serving_health(t *testing.T) {
	assert.NoError(t, testutil.AssertResponseStatusOfRequestCall(t,
		testutil.CreateOutgoingRequest(t, http.MethodGet, "/health", testutil.CreateHeader(), ""), http.StatusOK))
}

func TestKVStoreRequestHandler_serving_putKVStore(t *testing.T) {
	store := randString(5)
	key := randString(5)
	err := kvstoreHandler.storage.Put(store, key, nil)
	assert.NoError(t, err)
	request := testutil.CreateOutgoingRequest(t, http.MethodPut, "/kvstore/"+store+"/"+key,
		testutil.CreateHeader().WithAuth(username, password).WithJSONContentType(),
		`{ "testkey":"testvalue"}`)
	assert.NoError(t, testutil.AssertResponseStatusOfRequestCall(t, request, http.StatusNoContent))
	val, err := kvstoreHandler.storage.Get(store, key)
	assert.NoError(t, err)
	assert.Equal(t, map[string]interface{}{"testkey": "testvalue"}, val)
}

func TestKVStoreRequestHandler_putKVStore_readingBytesError(t *testing.T) {
	request := testutil.CreateIncomingErrorReadingBodyRequest(http.MethodPut, "", testutil.CreateHeader())
	assert.NoError(t, testutil.AssertHandlerFunc(t, request, kvstoreHandler.handlePutKVStore, func(response *http.Response, responseBody string) {
		assert.Equal(t, http.StatusInternalServerError, response.StatusCode)
		assert.Equal(t, "Problem reading request body: error reading bytes\n", responseBody)
	},
	))
}

func TestKVStoreRequestHandler_putKVStore_NoJsonError(t *testing.T) {
	request := testutil.CreateIncomingRequest(http.MethodPut, "", testutil.CreateHeader(), `{ invalid json`)
	assert.NoError(t, testutil.AssertHandlerFunc(t, request, kvstoreHandler.handlePutKVStore, func(response *http.Response, responseBody string) {
		assert.Equal(t, http.StatusBadRequest, response.StatusCode)
		assert.Equal(t, "Can't parse request body: invalid character 'i' looking for beginning of object key string\n", responseBody)
	},
	))
}

func TestKVStoreRequestHandler_serving_getKVStore(t *testing.T) {
	store := randString(5)
	key := randString(5)
	err := kvstoreHandler.storage.Put(store, key, "expectedVal")
	assert.NoError(t, err)
	err = kvstoreHandler.storage.Put(store, "key2", "val2 not expected")
	assert.NoError(t, err)
	request := testutil.CreateOutgoingRequest(t, http.MethodGet, "/kvstore/"+store+"/"+key,
		testutil.CreateHeader().WithAuth(username, password).WithJSONAccept(), "")
	assert.NoError(t, testutil.AssertResponseOfRequestCall(t, request, func(response *http.Response, responseBody string) {
		assert.Equal(t, "expectedVal", responseBody)
	}))
}

func TestKVStoreRequestHandler_removeKVStore(t *testing.T) {
	store := randString(5)
	key := randString(5)
	err := kvstoreHandler.storage.Put(store, key, map[string]string{"deletepath": "deletzevalue"})
	assert.NoError(t, err)
	request := testutil.CreateOutgoingRequest(t, http.MethodDelete, "/kvstore/"+store+"/"+key,
		testutil.CreateHeader().WithAuth(username, password).WithJSONContentType(), "")
	assert.NoError(t, testutil.AssertResponseStatusOfRequestCall(t, request, http.StatusNoContent))
	val, err := kvstoreHandler.storage.Get(store, key)
	assert.NoError(t, err)
	assert.Nil(t, val)
}
