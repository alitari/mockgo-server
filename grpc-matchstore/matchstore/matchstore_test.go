package matchstore

import (
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/alitari/mockgo-server/mockgo/logging"
	"github.com/alitari/mockgo-server/mockgo/matches"
	"github.com/alitari/mockgo-server/mockgo/testutil"
	"github.com/stretchr/testify/assert"
)

var clusterSize = 2
var startPort = 50051

var matchstores []*GrpcMatchstore

func TestMain(m *testing.M) {
	startMatchsroreCluster()
	code := testutil.RunAndCheckCoverage("main", m, 0.25)
	stopCluster()
	os.Exit(code)
}

func getClusterAdresses() []string {
	var clusterAddresses []string
	for i := 0; i < clusterSize; i++ {
		clusterAddresses = append(clusterAddresses, "localhost:"+strconv.Itoa(startPort+i))
	}
	return clusterAddresses
}

func startMatchsroreCluster() {
	addresses := getClusterAdresses()
	for i := 0; i < clusterSize; i++ {
		matchStore, err := NewGrpcMatchstore(addresses, startPort+i, uint16(100), logging.NewLoggerUtil(logging.Debug))
		if err != nil {
			log.Fatal(err)
		}
		matchstores = append(matchstores, matchStore)
	}
	time.Sleep(100 * time.Millisecond)
}

func stopCluster() {
	for _, matchstore := range matchstores {
		matchstore.StopServe()
	}
}

var timeStamp = time.Date(2009, 11, 17, 20, 34, 58, 651387237, time.UTC)

func addMatchesForEndpoint(storeNr int, endpointId string, count int) {
	for i := 0; i < count; i++ {
		matchstores[storeNr].AddMatch(endpointId, createMatch(endpointId))
	}
}

func addMismatches(storeNr int, count int) {
	for i := 0; i < count; i++ {
		matchstores[storeNr].AddMismatch(createMismatch())
	}
}

func createMatch(endpointId string) *matches.Match {
	request := &http.Request{Method: http.MethodGet, URL: &url.URL{Path: "http://myhost"}}
	return createMatchForRequest(endpointId, request)
}

func createMatchForRequest(endpointId string, request *http.Request) *matches.Match {
	actualRequest := &matches.ActualRequest{Method: request.Method, URL: request.URL.String(), Header: request.Header, Host: request.Host}
	actualResponse := &matches.ActualResponse{StatusCode: http.StatusOK, Header: map[string][]string{"key1": {"val1"}}}
	match := &matches.Match{EndpointId: endpointId, Timestamp: timeStamp, ActualRequest: actualRequest, ActualResponse: actualResponse}
	return match
}

func createMismatch() *matches.Mismatch {
	request := &http.Request{Method: http.MethodGet, URL: &url.URL{Path: "http://myhost"}}
	return createMismatchForRequest(request)
}

func createMismatchForRequest(request *http.Request) *matches.Mismatch {
	actualRequest := &matches.ActualRequest{Method: request.Method, URL: request.URL.String(), Header: request.Header, Host: request.Host}
	mismatch := &matches.Mismatch{Timestamp: timeStamp, ActualRequest: actualRequest}
	return mismatch
}

func TestMatchstore_GetMatchesSort(t *testing.T) {
	endpointId1 := "endpoint1"
	matchstores[0].DeleteMatches(endpointId1)
	err := matchstores[0].AddMatch(endpointId1, createMatchForRequest(endpointId1, &http.Request{Method: http.MethodGet, URL: &url.URL{Scheme: "http", Host: "host1"}}))
	assert.NoError(t, err)
	err = matchstores[0].AddMatch(endpointId1, createMatchForRequest(endpointId1, &http.Request{Method: http.MethodGet, URL: &url.URL{Scheme: "http", Host: "host2"}}))
	assert.NoError(t, err)
	err = matchstores[1].AddMatch(endpointId1, createMatchForRequest(endpointId1, &http.Request{Method: http.MethodGet, URL: &url.URL{Scheme: "http", Host: "host3"}}))
	assert.NoError(t, err)
	matches, err := matchstores[0].GetMatches(endpointId1)
	assert.NoError(t, err)
	assert.Equal(t, "http://host1", matches[0].ActualRequest.URL)
	assert.Equal(t, "http://host3", matches[1].ActualRequest.URL)
	assert.Equal(t, "http://host2", matches[2].ActualRequest.URL)
}

func TestMatchstore_GetMatches(t *testing.T) {
	endpointId1 := "endpoint1"
	endpointId2 := "endpoint2"
	matchstores[0].DeleteMatches(endpointId1)
	matchstores[0].DeleteMatches(endpointId2)
	addMatchesForEndpoint(0, endpointId1, 1)
	addMatchesForEndpoint(1, endpointId1, 1)
	addMatchesForEndpoint(0, endpointId2, 2)
	addMatchesForEndpoint(1, endpointId2, 1)
	matchesEndpoint1, err := matchstores[0].GetMatches(endpointId1)
	assert.NoError(t, err)
	assert.Len(t, matchesEndpoint1, 2)
	matchesEndpoint2, err := matchstores[0].GetMatches(endpointId2)
	assert.NoError(t, err)
	assert.Len(t, matchesEndpoint2, 3)

	matchesEndpoint1, err = matchstores[1].GetMatches(endpointId1)
	assert.NoError(t, err)
	assert.Len(t, matchesEndpoint1, 2)
	matchesEndpoint2, err = matchstores[1].GetMatches(endpointId2)
	assert.NoError(t, err)
	assert.Len(t, matchesEndpoint2, 3)
}

func TestMatchstore_GetMismatches(t *testing.T) {
	matchstores[0].DeleteMismatches()
	// matchstores[0].AddMismatches(createMismatches(1))
	addMismatches(0, 1)
	// matchstores[1].AddMismatches(createMismatches(2))
	addMismatches(1, 2)

	mismatches, err := matchstores[0].GetMismatches()
	assert.NoError(t, err)
	assert.Len(t, mismatches, 3)

	mismatches, err = matchstores[1].GetMismatches()
	assert.NoError(t, err)
	assert.Len(t, mismatches, 3)
}

func TestMatchstore_DeleteMatches(t *testing.T) {
	endpointId1 := "endpoint1"
	endpointId2 := "endpoint2"
	addMatchesForEndpoint(0, endpointId1, 4)
	addMatchesForEndpoint(1, endpointId2, 5)
	matchstores[0].DeleteMatches(endpointId2)
	matches, err := matchstores[1].GetMatches(endpointId2)
	assert.NoError(t, err)
	assert.Empty(t, matches)
	matchstores[1].DeleteMatches(endpointId1)
	matches, err = matchstores[0].GetMatches(endpointId1)
	assert.NoError(t, err)
	assert.Empty(t, matches)
}

func TestMatchstore_DeleteMismatches(t *testing.T) {
	addMismatches(0, 5)
	matchstores[1].DeleteMismatches()
	mismatches, err := matchstores[0].GetMismatches()
	assert.NoError(t, err)
	assert.Empty(t, mismatches)
}
