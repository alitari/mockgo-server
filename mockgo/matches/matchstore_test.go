package matches

import (
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var timeStamp = time.Date(2009, 11, 17, 20, 34, 58, 651387237, time.UTC)

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

func TestMatchstore_GetMatches(t *testing.T) {
	matchstore := NewInMemoryMatchstore()
	endpointId := "endpointId"
	matchstore.matches = map[string][]*Match{endpointId: {createMatch(endpointId)}}
	matches, err := matchstore.GetMatches(endpointId)
	assert.NoError(t, err)
	assert.Equal(t, matchstore.matches[endpointId], matches)
}

// func TestMatchstore_AddMatches(t *testing.T) {
// 	matchstore := NewInMemoryMatchstore(true,true)

// }
// func TestMatchstore_GetMatchesCount(t *testing.T) {
// 	matchstore := NewInMemoryMatchstore(true,true)

// }
// func TestMatchstore_AddMatchesCount(t *testing.T) {
// 	matchstore := NewInMemoryMatchstore(true,true)

// }
// func TestMatchstore_GetMismatches(t *testing.T) {
// 	matchstore := NewInMemoryMatchstore(true,true)

// }
// func TestMatchstore_AddMismatches(t *testing.T) {
// 	matchstore := NewInMemoryMatchstore(true,true)

// }
// func TestMatchstore_GetMismatchesCount(t *testing.T) {
// 	matchstore := NewInMemoryMatchstore(true,true)

// }
// func TestMatchstore_AddMismatchesCount(t *testing.T) {
// 	matchstore := NewInMemoryMatchstore(true,true)

// }
// func TestMatchstore_DeleteMatches(t *testing.T) {
// 	matchstore := NewInMemoryMatchstore(true,true)

// }
// func TestMatchstore_DeleteMismatches(t *testing.T) {
// 	matchstore := NewInMemoryMatchstore(true,true)

// }
