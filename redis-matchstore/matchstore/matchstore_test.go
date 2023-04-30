package matchstore

import (
	"encoding/json"
	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redismock/v9"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/alitari/mockgo-server/mockgo/matches"
)

var matchstore matches.Matchstore
var clientmock redismock.ClientMock

func createRedisMatchstore() {
	client, mock := redismock.NewClientMock()
	clientmock = mock
	matchstore = &redisMatchstore{
		client: client,
	}
}
func createMiniRedisMatchstore() {
	miniredis := miniredis.NewMiniRedis()
	err := miniredis.Start()
	if err != nil {
		panic(err)
	}
	matchstore, err = NewRedisMatchstore(miniredis.Addr(), "", 0, 1000)
	if err != nil {
		panic(err)
	}
}

var timeStamp = time.Date(2009, 11, 17, 20, 34, 58, 651387237, time.UTC)

func addMatchesForEndpoint(t *testing.T, endpointID string, count int) {
	for i := 0; i < count; i++ {
		err := matchstore.AddMatch(endpointID, createMatch(endpointID))
		assert.NoError(t, err)
	}
}

func addMismatches(t *testing.T, count int) {
	for i := 0; i < count; i++ {
		err := matchstore.AddMismatch(createMismatch())
		assert.NoError(t, err)
	}
}

func createMatch(endpointID string) *matches.Match {
	request := &http.Request{Method: http.MethodGet, URL: &url.URL{Path: "http://myhost"}}
	return createMatchForRequest(endpointID, request)
}

func createMatchString(match *matches.Match) string {
	matchStr, err := json.Marshal(match)
	if err != nil {
		panic(err)
	}
	return string(matchStr)
}

func createMismatchString(mismatch *matches.Mismatch) string {
	mismatchStr, err := json.Marshal(mismatch)
	if err != nil {
		panic(err)
	}
	return string(mismatchStr)
}

func createMatchForRequest(endpointID string, request *http.Request) *matches.Match {
	actualRequest := &matches.ActualRequest{Method: request.Method, URL: request.URL.String(), Header: request.Header, Host: request.Host}
	actualResponse := &matches.ActualResponse{StatusCode: http.StatusOK, Header: map[string][]string{"key1": {"val1"}}}
	match := &matches.Match{EndpointID: endpointID, Timestamp: timeStamp, ActualRequest: actualRequest, ActualResponse: actualResponse}
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

func TestRedisMatchstore_GetMatches(t *testing.T) {
	createRedisMatchstore()
	endpoint := "myendpoint"
	match := createMatch(endpoint)
	clientmock.ExpectLRange(endpoint, 0, -1).SetVal([]string{createMatchString(match)})
	getMatches, err := matchstore.GetMatches(endpoint)
	assert.NoError(t, err)
	assert.EqualValues(t, []*matches.Match{match}, getMatches)
}

func TestRedisMatchstore_GetMatchesCount(t *testing.T) {
	createRedisMatchstore()
	endpoint := "myendpoint"
	length := int64(76)
	clientmock.ExpectLLen(endpoint).SetVal(length)
	getMatchesCount, err := matchstore.GetMatchesCount(endpoint)
	assert.NoError(t, err)
	assert.EqualValues(t, uint64(length), getMatchesCount)
}

func TestRedisMatchstore_GetMismatches(t *testing.T) {
	createRedisMatchstore()
	mismatch := createMismatch()
	clientmock.ExpectLRange(matchstore.(*redisMatchstore).mismatchesKey, 0, -1).SetVal([]string{createMismatchString(mismatch)})
	getMismatches, err := matchstore.GetMismatches()
	assert.NoError(t, err)
	assert.EqualValues(t, []*matches.Mismatch{mismatch}, getMismatches)
}

func TestRedisMatchstore_GetMisMatchesCount(t *testing.T) {
	createRedisMatchstore()
	length := int64(76)
	clientmock.ExpectLLen(matchstore.(*redisMatchstore).mismatchesKey).SetVal(length)
	getMismatchesCount, err := matchstore.GetMismatchesCount()
	assert.NoError(t, err)
	assert.EqualValues(t, uint64(length), getMismatchesCount)
}

func TestRedisMatchstore_AddMatch(t *testing.T) {
	createRedisMatchstore()
	endpoint := "myendpoint"
	match := createMatch(endpoint)
	clientmock.ExpectRPush(endpoint, []byte(createMatchString(match))).SetVal(0)
	err := matchstore.AddMatch(endpoint, match)
	assert.NoError(t, err)
}

func TestRedisMatchstore_AddMismatch(t *testing.T) {
	createRedisMatchstore()
	mismatch := createMismatch()
	clientmock.ExpectRPush(matchstore.(*redisMatchstore).mismatchesKey, []byte(createMismatchString(mismatch))).SetVal(0)
	err := matchstore.AddMismatch(mismatch)
	assert.NoError(t, err)
}

func TestRedisMatchstore_DeleteMatches(t *testing.T) {
	createRedisMatchstore()
	endpoint := "myendpoint"
	clientmock.ExpectDel(endpoint).SetVal(0)
	err := matchstore.DeleteMatches(endpoint)
	assert.NoError(t, err)
}

func TestRedisMatchstore_DeleteMismatches(t *testing.T) {
	createRedisMatchstore()
	clientmock.ExpectDel(matchstore.(*redisMatchstore).mismatchesKey).SetVal(0)
	err := matchstore.DeleteMismatches()
	assert.NoError(t, err)
}

func TestRedisMatchstore_GetMatchesSort(t *testing.T) {
	createMiniRedisMatchstore()
	endpointID1 := "endpoint1"
	err := matchstore.DeleteMatches(endpointID1)
	assert.NoError(t, err)
	err = matchstore.AddMatch(endpointID1, createMatchForRequest(endpointID1, &http.Request{Method: http.MethodGet, URL: &url.URL{Scheme: "http", Host: "host1"}}))
	assert.NoError(t, err)
	err = matchstore.AddMatch(endpointID1, createMatchForRequest(endpointID1, &http.Request{Method: http.MethodGet, URL: &url.URL{Scheme: "http", Host: "host2"}}))
	assert.NoError(t, err)
	err = matchstore.AddMatch(endpointID1, createMatchForRequest(endpointID1, &http.Request{Method: http.MethodGet, URL: &url.URL{Scheme: "http", Host: "host3"}}))
	assert.NoError(t, err)
	matches, err := matchstore.GetMatches(endpointID1)
	assert.NoError(t, err)
	assert.Len(t, matches, 3)
	assert.Equal(t, "http://host1", matches[0].ActualRequest.URL)
	assert.Equal(t, "http://host2", matches[1].ActualRequest.URL)
	assert.Equal(t, "http://host3", matches[2].ActualRequest.URL)
}

func TestRedisMatchstore_MatchesCount(t *testing.T) {
	createMiniRedisMatchstore()
	endpointID1 := "endpoint1"
	endpointID2 := "endpoint2"
	err := matchstore.DeleteMatches(endpointID1)
	assert.NoError(t, err)
	err = matchstore.DeleteMatches(endpointID2)
	assert.NoError(t, err)
	addMatchesForEndpoint(t, endpointID1, 1)
	addMatchesForEndpoint(t, endpointID1, 1)
	addMatchesForEndpoint(t, endpointID2, 2)
	addMatchesForEndpoint(t, endpointID2, 1)
	matchesCountEndpoint1, err := matchstore.GetMatchesCount(endpointID1)
	assert.NoError(t, err)
	assert.Equal(t, uint64(2), matchesCountEndpoint1)
	matchesCountEndpoint2, err := matchstore.GetMatchesCount(endpointID2)
	assert.NoError(t, err)
	assert.Equal(t, uint64(3), matchesCountEndpoint2)
}

func TestRedisMatchstore_GetMismatchesSort(t *testing.T) {
	createMiniRedisMatchstore()
	err := matchstore.DeleteMismatches()
	assert.NoError(t, err)
	err = matchstore.AddMismatch(createMismatchForRequest(&http.Request{Method: http.MethodGet, URL: &url.URL{Scheme: "http", Host: "host1"}}))
	assert.NoError(t, err)
	err = matchstore.AddMismatch(createMismatchForRequest(&http.Request{Method: http.MethodGet, URL: &url.URL{Scheme: "http", Host: "host2"}}))
	assert.NoError(t, err)
	err = matchstore.AddMismatch(createMismatchForRequest(&http.Request{Method: http.MethodGet, URL: &url.URL{Scheme: "http", Host: "host3"}}))
	assert.NoError(t, err)
	mismatches, err := matchstore.GetMismatches()
	assert.NoError(t, err)
	assert.Len(t, mismatches, 3)
	assert.Equal(t, "http://host1", mismatches[0].ActualRequest.URL)
	assert.Equal(t, "http://host2", mismatches[1].ActualRequest.URL)
	assert.Equal(t, "http://host3", mismatches[2].ActualRequest.URL)
}

func TestRedisMatchstore_Delete(t *testing.T) {
	createMiniRedisMatchstore()
	endpointID1 := "endpoint1"
	endpointID2 := "endpoint2"
	addMatchesForEndpoint(t, endpointID1, 4)
	addMatchesForEndpoint(t, endpointID2, 5)
	l1, err := matchstore.GetMatchesCount(endpointID1)
	assert.NoError(t, err)
	assert.Equal(t, uint64(4), l1)
	l2, err := matchstore.GetMatchesCount(endpointID2)
	assert.NoError(t, err)
	assert.Equal(t, uint64(5), l2)
	err = matchstore.DeleteMatches(endpointID1)
	assert.NoError(t, err)
	l1, err = matchstore.GetMatchesCount(endpointID1)
	assert.NoError(t, err)
	assert.Equal(t, uint64(0), l1)
	l2, err = matchstore.GetMatchesCount(endpointID2)
	assert.NoError(t, err)
	assert.Equal(t, uint64(5), l2)
}

func TestRedisMatchstore_Delete2(t *testing.T) {
	createMiniRedisMatchstore()
	addMismatches(t, 4)
	l, err := matchstore.GetMismatchesCount()
	assert.NoError(t, err)
	assert.Equal(t, uint64(4), l)
	err = matchstore.DeleteMismatches()
	assert.NoError(t, err)
	l, err = matchstore.GetMismatchesCount()
	assert.NoError(t, err)
	assert.Equal(t, uint64(0), l)
}
