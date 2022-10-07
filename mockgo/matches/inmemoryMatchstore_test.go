package matches

import (
	"container/list"
	"math/rand"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var timeStamp = time.Date(2009, 11, 17, 20, 34, 58, 651387237, time.UTC)
var endpointId1 = "endpointId1"
var endpointId2 = "endpointId2"

func createMatchesForEndpoint(endpointId string, count int) []*Match {
	var matches []*Match
	for i := 0; i < count; i++ {
		matches = append(matches, createMatch(endpointId))
	}
	return matches
}

func createMatch(endpointId string) *Match {
	request := &http.Request{Method: http.MethodGet, URL: &url.URL{Path: "http://myhost"}}
	return createMatchForRequest(endpointId, request)
}

func createMatchForRequest(endpointId string, request *http.Request) *Match {
	actualRequest := &ActualRequest{Method: request.Method, URL: request.URL.String(), Header: request.Header, Host: request.Host}
	match := &Match{EndpointId: endpointId, Timestamp: timeStamp, ActualRequest: actualRequest}
	return match
}

func createMismatch() *Mismatch {
	request := &http.Request{Method: http.MethodGet, URL: &url.URL{Path: "http://myhost"}}
	return createMismatchForRequest(request)
}

func createMismatchForRequest(request *http.Request) *Mismatch {
	actualRequest := &ActualRequest{Method: request.Method, URL: request.URL.String(), Header: request.Header, Host: request.Host}
	mismatch := &Mismatch{Timestamp: timeStamp, ActualRequest: actualRequest}
	return mismatch
}

func TestInMemoryMatchstore_GetMatchesInit(t *testing.T) {
	matchstore := NewInMemoryMatchstore(5)
	matches, err := matchstore.GetMatches(endpointId1)
	assert.NoError(t, err)
	assert.Equal(t, []*Match{}, matches)
}

func TestInMemoryMatchstore_GetMatchesCountInit(t *testing.T) {
	matchstore := NewInMemoryMatchstore(5)
	matchesCount, err := matchstore.GetMatchesCount(endpointId1)
	assert.NoError(t, err)
	assert.Equal(t, uint64(0), matchesCount)
}

func TestInMemoryMatchstore_GetMatches(t *testing.T) {
	matchstore := NewInMemoryMatchstore(5)
	matchstore.matches = map[string]*list.List{endpointId1: list.New()}
	matches1 := createMatchesForEndpoint(endpointId1, 2)
	for _, match := range matches1 {
		matchstore.matches[endpointId1].PushBack(match)
	}
	matches, err := matchstore.GetMatches(endpointId1)
	assert.NoError(t, err)
	assert.Equal(t, matches1, matches)
}

func TestInMemoryMatchstore_GetMatchesCount(t *testing.T) {
	matchstore := NewInMemoryMatchstore(5)
	count := rand.Uint64()
	matchstore.matchesCount = map[string]uint64{endpointId1: count}
	matchesCount, err := matchstore.GetMatchesCount(endpointId1)
	assert.NoError(t, err)
	assert.Equal(t, count, matchesCount)
}

func TestInMemoryMatchstore_AddMatch(t *testing.T) {
	matchstore := NewInMemoryMatchstore(3)
	matches1 := createMatchesForEndpoint(endpointId1, 4)
	for _, match := range matches1 {
		err := matchstore.AddMatch(endpointId1, match)
		assert.NoError(t, err)
	}
	matchList := matchstore.matches[endpointId1]
	assert.Equal(t, 3, matchList.Len())
	assert.Equal(t, uint64(4), matchstore.matchesCount[endpointId1])
}

func TestInMemoryMatchstore_AddMisMatch(t *testing.T) {
	matchstore := NewInMemoryMatchstore(3)
	for i := 0; i < 5; i++ {
		err := matchstore.AddMismatch(createMismatch())
		assert.NoError(t, err)
	}
	mismatchList := matchstore.mismatches
	assert.Equal(t, 3, mismatchList.Len())
	assert.Equal(t, uint64(5), matchstore.mismatchesCount)
}

func TestInMemoryMatchstore_DeleteMatches(t *testing.T) {
	matchstore := NewInMemoryMatchstore(5)
	matchstore.matches = map[string]*list.List{endpointId1: list.New(), endpointId2: list.New()}
	matches1 := createMatchesForEndpoint(endpointId1, 2)
	for _, match := range matches1 {
		matchstore.matches[endpointId1].PushBack(match)
		matchstore.matchesCount[endpointId1]++
	}

	matches2 := createMatchesForEndpoint(endpointId2, 2)
	for _, match := range matches2 {
		matchstore.matches[endpointId2].PushBack(match)
		matchstore.matchesCount[endpointId2]++
	}
	err := matchstore.DeleteMatches(endpointId1)
	assert.NoError(t, err)
	assert.Equal(t, 0, matchstore.matches[endpointId1].Len())
	assert.Equal(t, uint64(0), matchstore.matchesCount[endpointId1])
	assert.Equal(t, 2, matchstore.matches[endpointId2].Len())
	assert.Equal(t, uint64(2), matchstore.matchesCount[endpointId2])
}
