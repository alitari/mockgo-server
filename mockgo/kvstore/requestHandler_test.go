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
	code := runAndCheckCoverage("main", m, 0.65)
	os.Exit(code)
}

func startServing() {
	kvstoreLogger := logging.NewLoggerUtil(logging.Debug)
	kvstoreJson := NewKVStoreJSON(NewInmemoryKVStore(), true)
	kvstoreHandler = NewKVStoreRequestHandler(username, password, kvstoreJson, kvstoreLogger)
	router := mux.NewRouter()
	kvstoreHandler.AddRoutes(router)
	server := &http.Server{Addr: ":" + strconv.Itoa(port), Handler: router}
	log.Printf("Serving on '%s'", server.Addr)

	err := server.ListenAndServe()
	if err != nil {
		log.Fatalf("Can't serve: %v", err)
	}
}

func runAndCheckCoverage(testPackage string, m *testing.M, treshold float64) int {
	code := m.Run()
	if code == 0 && testing.CoverMode() != "" {
		coverage := testing.Coverage()
		if coverage < treshold {
			log.Printf("%s tests passed, but coverage must be above %2.2f%%, but it is %2.2f%%\n", testPackage, treshold*100, coverage*100)
			code = -1
		}
	}
	return code
}

func TestKVStoreRequestHandler_health(t *testing.T) {
	util.RequestCall(t, httpClient, http.MethodGet, urlPrefix+"/health", map[string][]string{}, "", http.StatusOK, nil)
}

func TestKVStoreRequestHandler_setKVStore(t *testing.T) {
	err := kvstoreHandler.kvstore.PutAll(map[string]interface{}{})
	assert.NoError(t, err)
	key := randString(5)
	util.RequestCall(t, httpClient, http.MethodPut, urlPrefix+"/kvstore/"+key,
		map[string][]string{headers.Authorization: {util.BasicAuth(username, password)}, headers.ContentType: {"application/json"}},
		`{ "testkey":"testvalue"}`, http.StatusNoContent, nil)
	val, err := kvstoreHandler.kvstore.Get(key)
	assert.NoError(t, err)
	assert.Equal(t, map[string]interface{}{"testkey": "testvalue"}, val)
}

func TestKVStoreRequestHandler_getKVStore(t *testing.T) {
	key := randString(5)
	err := kvstoreHandler.kvstore.PutAll(map[string]interface{}{key: "expectedVal", "key2": "val2 not expected"})
	assert.NoError(t, err)
	util.RequestCall(t, httpClient, http.MethodGet, urlPrefix+"/kvstore/"+key,
		map[string][]string{headers.Authorization: {util.BasicAuth(username, password)}, headers.Accept: {"application/json"}},
		"", http.StatusOK,
		func(responseBody string) {
			assert.Equal(t, `"expectedVal"`, responseBody)
		})
}

func TestKVStoreRequestHandler_addKVStore(t *testing.T) {
	err := kvstoreHandler.kvstore.PutAll(map[string]interface{}{})
	assert.NoError(t, err)
	key := randString(5)
	util.RequestCall(t, httpClient, http.MethodPost, urlPrefix+"/kvstore/"+key+"/add",
		map[string][]string{headers.Authorization: {util.BasicAuth(username, password)}, headers.ContentType: {"application/json"}},
		`{ "path": "/testpath", "value": "testvalue" }`, http.StatusNoContent, nil)
	val, err := kvstoreHandler.kvstore.Get(key)
	assert.NoError(t, err)
	assert.Equal(t, map[string]interface{}{"testkey": "testvalue"}, val)
}

func TestKVStoreRequestHandler_removeKVStore(t *testing.T) {
}
func TestKVStoreRequestHandler_uploadKVStore(t *testing.T) {
}
func TestKVStoreRequestHandler_downloadKVStore(t *testing.T) {
}

// func TestKVStoreRequestHandler_uploadKvStore(t *testing.T) {
// 	// delete kv store for all
// 	requestToAllNodes(t, true, http.MethodPut, "/kvstore", map[string][]string{headers.ContentType: {"application/json"}, headers.Authorization: {config.BasicAuth("mockgo", configPassword)}}, "{}", http.StatusNoContent, nil)

// 	kvstore := `{"store1":{"mykey":"myvalue"}}`
// 	requestToNode(t, 0, true, http.MethodPut, "/kvstore", map[string][]string{headers.ContentType: {"application/json"}, headers.Authorization: {config.BasicAuth("mockgo", configPassword)}}, kvstore, http.StatusNoContent, nil)
// 	requestToNode(t, 0, true, http.MethodGet, "/kvstore", map[string][]string{headers.Accept: {"application/json"}, headers.Authorization: {config.BasicAuth("mockgo", configPassword)}}, "", http.StatusOK, func(responseBody string) {
// 		assert.Equal(t, kvstore, responseBody)
// 	})
// 	// upload kvstore not advertised
// 	requestToNode(t, 1, true, http.MethodGet, "/kvstore", map[string][]string{headers.Accept: {"application/json"}, headers.Authorization: {config.BasicAuth("mockgo", configPassword)}}, "", http.StatusOK, func(responseBody string) {
// 		assert.Equal(t, "{}", responseBody)
// 	})
// 	stopNode(1)
// 	startNode(1)
// 	time.Sleep(100 * time.Millisecond)
// 	// kvstore synced from node 0 at boot time
// 	requestToNode(t, 1, true, http.MethodGet, "/kvstore", map[string][]string{headers.Accept: {"application/json"}, headers.Authorization: {config.BasicAuth("mockgo", configPassword)}}, "", http.StatusOK, func(responseBody string) {
// 		assert.Equal(t, kvstore, responseBody)
// 	})
// }
