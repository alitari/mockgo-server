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
var endpointID1 = "endpointId1"
var endpointID2 = "endpointId2"

func createMatchesForEndpoint(endpointID string, count int) []*Match {
	var matches []*Match
	for i := 0; i < count; i++ {
		matches = append(matches, createMatch(endpointID))
	}
	return matches
}

func createMatch(endpointID string) *Match {
	request := &http.Request{Method: http.MethodGet, URL: &url.URL{Path: "http://myhost"}}
	return createMatchForRequest(endpointID, request)
}

func createMatchForRequest(endpointID string, request *http.Request) *Match {
	actualRequest := &ActualRequest{Method: request.Method, URL: request.URL.String(), Header: request.Header, Host: request.Host}
	match := &Match{EndpointID: endpointID, Timestamp: timeStamp, ActualRequest: actualRequest}
	return match
}

func createMismnatches(count int) []*Mismatch {
	var mismatches []*Mismatch
	for i := 0; i < count; i++ {
		mismatches = append(mismatches, createMismatch())
	}
	return mismatches
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
	matches, err := matchstore.GetMatches(endpointID1)
	assert.NoError(t, err)
	assert.Equal(t, []*Match{}, matches)
}

func TestInMemoryMatchstore_GetMisMatchesInit(t *testing.T) {
	matchstore := NewInMemoryMatchstore(5)
	mismatches, err := matchstore.GetMismatches()
	assert.NoError(t, err)
	assert.Equal(t, []*Mismatch{}, mismatches)
}

func TestInMemoryMatchstore_GetMatchesCountInit(t *testing.T) {
	matchstore := NewInMemoryMatchstore(5)
	matchesCount, err := matchstore.GetMatchesCount(endpointID1)
	assert.NoError(t, err)
	assert.Equal(t, uint64(0), matchesCount)
}

func TestInMemoryMatchstore_GetMismatchesCountInit(t *testing.T) {
	matchstore := NewInMemoryMatchstore(5)
	mismatchesCount, err := matchstore.GetMismatchesCount()
	assert.NoError(t, err)
	assert.Equal(t, uint64(0), mismatchesCount)
}

func TestInMemoryMatchstore_GetMatches(t *testing.T) {
	matchstore := NewInMemoryMatchstore(5)
	matchstore.matches = map[string]*list.List{endpointID1: list.New()}
	matches1 := createMatchesForEndpoint(endpointID1, 2)
	for _, match := range matches1 {
		matchstore.matches[endpointID1].PushBack(match)
	}
	matches, err := matchstore.GetMatches(endpointID1)
	assert.NoError(t, err)
	assert.Equal(t, matches1, matches)
}

func TestInMemoryMatchstore_GetMismatches(t *testing.T) {
	matchstore := NewInMemoryMatchstore(5)
	mismatches2 := createMismnatches(2)
	for _, mismatch := range mismatches2 {
		matchstore.mismatches.PushBack(mismatch)
	}
	mismatches, err := matchstore.GetMismatches()
	assert.NoError(t, err)
	assert.Equal(t, mismatches2, mismatches)
}

func TestInMemoryMatchstore_GetMatchesCount(t *testing.T) {
	matchstore := NewInMemoryMatchstore(5)
	count := rand.Uint64()
	matchstore.matchesCount = map[string]uint64{endpointID1: count}
	matchesCount, err := matchstore.GetMatchesCount(endpointID1)
	assert.NoError(t, err)
	assert.Equal(t, count, matchesCount)
}

func TestInMemoryMatchstore_GetMismatchesCount(t *testing.T) {
	matchstore := NewInMemoryMatchstore(5)
	count := rand.Uint64()
	matchstore.mismatchesCount = count
	mismatchesCount, err := matchstore.GetMismatchesCount()
	assert.NoError(t, err)
	assert.Equal(t, count, mismatchesCount)
}

func TestInMemoryMatchstore_AddMatch(t *testing.T) {
	matchstore := NewInMemoryMatchstore(3)
	matches1 := createMatchesForEndpoint(endpointID1, 4)
	for _, match := range matches1 {
		err := matchstore.AddMatch(endpointID1, match)
		assert.NoError(t, err)
	}
	matchList := matchstore.matches[endpointID1]
	assert.Equal(t, 3, matchList.Len())
	assert.Equal(t, uint64(4), matchstore.matchesCount[endpointID1])
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
	matchstore.matches = map[string]*list.List{endpointID1: list.New(), endpointID2: list.New()}
	matches1 := createMatchesForEndpoint(endpointID1, 2)
	for _, match := range matches1 {
		matchstore.matches[endpointID1].PushBack(match)
		matchstore.matchesCount[endpointID1]++
	}

	matches2 := createMatchesForEndpoint(endpointID2, 2)
	for _, match := range matches2 {
		matchstore.matches[endpointID2].PushBack(match)
		matchstore.matchesCount[endpointID2]++
	}
	err := matchstore.DeleteMatches(endpointID1)
	assert.NoError(t, err)
	assert.Equal(t, 0, matchstore.matches[endpointID1].Len())
	assert.Equal(t, uint64(0), matchstore.matchesCount[endpointID1])
	assert.Equal(t, 2, matchstore.matches[endpointID2].Len())
	assert.Equal(t, uint64(2), matchstore.matchesCount[endpointID2])
}

func TestInMemoryMatchstore_DeleteMismatches(t *testing.T) {
	matchstore := NewInMemoryMatchstore(5)
	matchstore.mismatches = list.New()
	mismatches1 := createMismnatches(2)
	for _, mismatch := range mismatches1 {
		matchstore.mismatches.PushBack(mismatch)
		matchstore.mismatchesCount++
	}
	err := matchstore.DeleteMismatches()
	assert.NoError(t, err)
	assert.Equal(t, 0, matchstore.mismatches.Len())
	assert.Equal(t, uint64(0), matchstore.mismatchesCount)
}
