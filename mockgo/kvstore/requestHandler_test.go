package kvstore

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/alitari/mockgo/logging"
	"github.com/alitari/mockgo/util"
	"github.com/go-http-utils/headers"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

const (
	username = "mockgo"
	password = "password"
	port     = 8080
)

var urlPrefix = fmt.Sprintf("http://localhost:%d", port)
var httpClient = http.Client{Timeout: time.Duration(1) * time.Second}

var kvstoreHandler *KVStoreRequestHandler

func TestMain(m *testing.M) {
	go startServing()
	time.Sleep(200 * time.Millisecond)
	code := util.RunAndCheckCoverage("requestHandlerTest", m, 0.50)
	os.Exit(code)
}

func startServing() {
	kvstoreLogger := logging.NewLoggerUtil(logging.Debug)
	kvstoreJson := NewKVStoreJSON(NewInmemoryKVStore(), true)
	kvstoreHandler = NewKVStoreRequestHandler("", username, password, kvstoreJson, kvstoreLogger)
	router := mux.NewRouter()
	kvstoreHandler.AddRoutes(router)
	server := &http.Server{Addr: ":" + strconv.Itoa(port), Handler: router}
	log.Printf("Serving on '%s'", server.Addr)

	err := server.ListenAndServe()
	if err != nil {
		log.Fatalf("Can't serve: %v", err)
	}
}

func TestKVStoreRequestHandler_health(t *testing.T) {
	util.RequestCall(t, httpClient, http.MethodGet, urlPrefix+"/health", map[string][]string{}, "", http.StatusOK, nil)
}

func TestKVStoreRequestHandler_setKVStore(t *testing.T) {
	key := randString(5)
	err := kvstoreHandler.kvstore.Put(key, nil)
	assert.NoError(t, err)
	util.RequestCall(t, httpClient, http.MethodPut, urlPrefix+"/kvstore/"+key,
		map[string][]string{headers.Authorization: {util.BasicAuth(username, password)}, headers.ContentType: {"application/json"}},
		`{ "testkey":"testvalue"}`, http.StatusNoContent, nil)
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
	util.RequestCall(t, httpClient, http.MethodGet, urlPrefix+"/kvstore/"+key,
		map[string][]string{headers.Authorization: {util.BasicAuth(username, password)}, headers.Accept: {"application/json"}},
		"", http.StatusOK,
		func(responseBody string) {
			assert.Equal(t, `"expectedVal"`, responseBody)
		})
}

func TestKVStoreRequestHandler_addKVStore(t *testing.T) {
	key := randString(5)
	err := kvstoreHandler.kvstore.Put(key, nil)
	assert.NoError(t, err)
	util.RequestCall(t, httpClient, http.MethodPost, urlPrefix+"/kvstore/"+key+"/add",
		map[string][]string{headers.Authorization: {util.BasicAuth(username, password)}, headers.ContentType: {"application/json"}},
		`{ "path": "/testpath", "value": "testvalue" }`, http.StatusNoContent, nil)
	val, err := kvstoreHandler.kvstore.Get(key)
	assert.NoError(t, err)
	assert.Equal(t, map[string]interface{}{"testpath": "testvalue"}, val)
}

func TestKVStoreRequestHandler_removeKVStore(t *testing.T) {
	key := randString(5)
	err := kvstoreHandler.kvstore.Put(key, map[string]string{"deletepath": "deletzevalue"})
	assert.NoError(t, err)
	util.RequestCall(t, httpClient, http.MethodPost, urlPrefix+"/kvstore/"+key+"/remove",
		map[string][]string{headers.Authorization: {util.BasicAuth(username, password)}, headers.ContentType: {"application/json"}},
		`{ "path": "/deletepath"}`, http.StatusNoContent, nil)
	all, err := kvstoreHandler.kvstore.Get(key)
	assert.NoError(t, err)
	assert.Equal(t, map[string]interface{}{}, all)
}
