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
	kvstoreJSON := NewKVStoreJSON(NewInmemoryKVStore(), true)
	kvstoreHandler = NewRequestHandler("", username, password, kvstoreJSON, kvstoreLogger)
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

func TestKVStoreRequestHandler_serving_setKVStore(t *testing.T) {
	key := randString(5)
	err := kvstoreHandler.kvstore.Put(key, nil)
	assert.NoError(t, err)
	request := testutil.CreateOutgoingRequest(t, http.MethodPut, "/kvstore/"+key,
		testutil.CreateHeader().WithAuth(username, password).WithJSONContentType(),
		`{ "testkey":"testvalue"}`)
	assert.NoError(t, testutil.AssertResponseStatusOfRequestCall(t, request, http.StatusNoContent))
	val, err := kvstoreHandler.kvstore.Get(key)
	assert.NoError(t, err)
	assert.Equal(t, map[string]interface{}{"testkey": "testvalue"}, val)
}
func TestKVStoreRequestHandler_setKVStore_readingBytesError(t *testing.T) {
	request := testutil.CreateIncomingErrorReadingBodyRequest(http.MethodPut, "", testutil.CreateHeader())
	assert.NoError(t, testutil.AssertHandlerFunc(t, request, kvstoreHandler.handleSetKVStore, func(response *http.Response, responseBody string) {
		assert.Equal(t, http.StatusInternalServerError, response.StatusCode)
		assert.Equal(t, "Problem reading request body: error reading bytes\n", responseBody)
	},
	))
}

func TestKVStoreRequestHandler_setKVStore_NoJsonError(t *testing.T) {
	request := testutil.CreateIncomingRequest(http.MethodPut, "", testutil.CreateHeader(), `{ invalid json`)
	assert.NoError(t, testutil.AssertHandlerFunc(t, request, kvstoreHandler.handleSetKVStore, func(response *http.Response, responseBody string) {
		assert.Equal(t, http.StatusBadRequest, response.StatusCode)
		assert.Equal(t, "Problem with kvstore value, ( is it valid JSON?): invalid character 'i' looking for beginning of object key string\n", responseBody)
	},
	))
}

func TestKVStoreRequestHandler_serving_getKVStore(t *testing.T) {
	key := randString(5)
	err := kvstoreHandler.kvstore.Put(key, "expectedVal")
	assert.NoError(t, err)
	err = kvstoreHandler.kvstore.Put("key2", "val2 not expected")
	assert.NoError(t, err)
	request := testutil.CreateOutgoingRequest(t, http.MethodGet, "/kvstore/"+key,
		testutil.CreateHeader().WithAuth(username, password).WithJSONAccept(), "")
	assert.NoError(t, testutil.AssertResponseOfRequestCall(t, request, func(response *http.Response, responseBody string) {
		assert.Equal(t, "expectedVal", responseBody)
	}))
}

func TestKVStoreRequestHandler_serving_addKVStore(t *testing.T) {
	key := randString(5)
	err := kvstoreHandler.kvstore.Put(key, nil)
	assert.NoError(t, err)
	request := testutil.CreateOutgoingRequest(t, http.MethodPost, "/kvstore/"+key+"/add",
		testutil.CreateHeader().WithAuth(username, password).WithJSONContentType(),
		`{ "path": "/testpath", "value": "testvalue" }`)
	assert.NoError(t, testutil.AssertResponseStatusOfRequestCall(t, request, http.StatusNoContent))
	val, err := kvstoreHandler.kvstore.Get(key)
	assert.NoError(t, err)
	assert.Equal(t, map[string]interface{}{"testpath": "testvalue"}, val)
}

func TestKVStoreRequestHandler_addKVStore_readingBytesError(t *testing.T) {
	request := testutil.CreateIncomingErrorReadingBodyRequest(http.MethodPost, "", testutil.CreateHeader())
	assert.NoError(t, testutil.AssertHandlerFunc(t, request, kvstoreHandler.handleAddKVStore, func(response *http.Response, responseBody string) {
		assert.Equal(t, http.StatusInternalServerError, response.StatusCode)
		assert.Equal(t, "Problem reading request body: error reading bytes\n", responseBody)
	},
	))
}

func TestKVStoreRequestHandler_addKVStore_NoJsonError(t *testing.T) {
	request := testutil.CreateIncomingRequest(http.MethodPost, "", testutil.CreateHeader(), `{ invalid json`)
	assert.NoError(t, testutil.AssertHandlerFunc(t, request, kvstoreHandler.handleAddKVStore, func(response *http.Response, responseBody string) {
		assert.Equal(t, http.StatusBadRequest, response.StatusCode)
		assert.Equal(t, "Can't parse request body '{ invalid json' : invalid character 'i' looking for beginning of object key string\n", responseBody)
	},
	))
}

func TestKVStoreRequestHandler_addKVStore_WrongPatchFormatError(t *testing.T) {
	request := testutil.CreateIncomingRequest(http.MethodPost, "", testutil.CreateHeader(), `{ "path": "missingslash", "value": "testvalue" }`)
	assert.NoError(t, testutil.AssertHandlerFunc(t, request, kvstoreHandler.handleAddKVStore, func(response *http.Response, responseBody string) {
		assert.Equal(t, http.StatusBadRequest, response.StatusCode)
		assert.Equal(t, "Problem adding kvstore path: 'missingslash' value: 'testvalue', : add operation does not apply: doc is missing path: \"missingslash\": missing value \n", responseBody)
	},
	))
}

func TestKVStoreRequestHandler_removeKVStore(t *testing.T) {
	key := randString(5)
	err := kvstoreHandler.kvstore.Put(key, map[string]string{"deletepath": "deletzevalue"})
	assert.NoError(t, err)
	request := testutil.CreateOutgoingRequest(t, http.MethodPost, "/kvstore/"+key+"/remove",
		testutil.CreateHeader().WithAuth(username, password).WithJSONContentType(), `{ "path": "/deletepath"}`)
	assert.NoError(t, testutil.AssertResponseStatusOfRequestCall(t, request, http.StatusNoContent))
	all, err := kvstoreHandler.kvstore.Get(key)
	assert.NoError(t, err)
	assert.Equal(t, map[string]interface{}{}, all)
}

func TestKVStoreRequestHandler_removeKVStore_readingBytesError(t *testing.T) {
	request := testutil.CreateIncomingErrorReadingBodyRequest(http.MethodPost, "", testutil.CreateHeader())
	assert.NoError(t, testutil.AssertHandlerFunc(t, request, kvstoreHandler.handleRemoveKVStore, func(response *http.Response, responseBody string) {
		assert.Equal(t, http.StatusInternalServerError, response.StatusCode)
		assert.Equal(t, "Problem reading request body: error reading bytes\n", responseBody)
	},
	))
}

func TestKVStoreRequestHandler_removeKVStore_NoJsonError(t *testing.T) {
	request := testutil.CreateIncomingRequest(http.MethodPost, "", testutil.CreateHeader(), `{ invalid json`)
	assert.NoError(t, testutil.AssertHandlerFunc(t, request, kvstoreHandler.handleRemoveKVStore, func(response *http.Response, responseBody string) {
		assert.Equal(t, http.StatusBadRequest, response.StatusCode)
		assert.Equal(t, "Can't parse request body: invalid character 'i' looking for beginning of object key string\n", responseBody)
	},
	))
}

func TestKVStoreRequestHandler_removeKVStore_WrongPatchFormatError(t *testing.T) {
	request := testutil.CreateIncomingRequest(http.MethodPost, "", testutil.CreateHeader(), `{ "path": "missingslash" }`)
	assert.NoError(t, testutil.AssertHandlerFunc(t, request, kvstoreHandler.handleRemoveKVStore, func(response *http.Response, responseBody string) {
		assert.Equal(t, http.StatusBadRequest, response.StatusCode)
		assert.Equal(t, "Problem removing kvstore '', path: 'missingslash' : remove operation does not apply: doc is missing path: \"missingslash\": missing value \n", responseBody)
	},
	))
}
