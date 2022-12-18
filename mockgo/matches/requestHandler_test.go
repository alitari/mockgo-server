package matches

import (
	"fmt"
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

type ErrorMatchstore struct {
}

func (s *ErrorMatchstore) GetMatches(endpointID string) ([]*Match, error) {
	return nil, fmt.Errorf("error in get matches")
}
func (s *ErrorMatchstore) GetMismatches() ([]*Mismatch, error) {
	return nil, fmt.Errorf("error in get mismatches")
}
func (s *ErrorMatchstore) GetMatchesCount(endpointID string) (uint64, error) {
	return 0, fmt.Errorf("error in get matches count")
}
func (s *ErrorMatchstore) AddMatch(endpointID string, match *Match) error {
	return fmt.Errorf("error in add match")
}
func (s *ErrorMatchstore) AddMismatch(*Mismatch) error {
	return fmt.Errorf("error in add mismatch")
}
func (s *ErrorMatchstore) GetMismatchesCount() (uint64, error) {
	return 0, fmt.Errorf("error in get mismatches count")
}
func (s *ErrorMatchstore) DeleteMatches(endpointID string) error {
	return fmt.Errorf("error in delete matches")
}
func (s *ErrorMatchstore) DeleteMismatches() error {
	return fmt.Errorf("error in delete mismatches")
}

var matchesRequestHandler = NewRequestHandler("", username, password, NewInMemoryMatchstore(uint16(100)), logging.NewLoggerUtil(logging.Debug))
var matchesRequestHandlerErroneous = NewRequestHandler("", username, password, &ErrorMatchstore{}, logging.NewLoggerUtil(logging.Debug))

func TestMain(m *testing.M) {
	router := mux.NewRouter()
	matchesRequestHandler.AddRoutes(router)
	testutil.StartServing(router)
	code := testutil.RunAndCheckCoverage("matchesRequestHandlerTest", m, 0.40)
	testutil.StopServing()
	os.Exit(code)
}

func TestMatchesRequestHandler_serving_health(t *testing.T) {
	assert.NoError(t, testutil.AssertResponseStatusOfRequestCall(t,
		testutil.CreateOutgoingRequest(t, http.MethodGet, "/health", testutil.CreateHeader(), ""), http.StatusOK))
}

func TestMatchesRequestHandler_serving_getMatches(t *testing.T) {
	endpointID := "myEndpointId"
	err := matchesRequestHandler.matchStore.DeleteMatches(endpointID)
	assert.NoError(t, err)
	err = matchesRequestHandler.matchStore.AddMatch(endpointID, createMatch(endpointID))
	assert.NoError(t, err)
	request := testutil.CreateOutgoingRequest(t, http.MethodGet, "/matches/"+endpointID,
		testutil.CreateHeader().WithAuth(username, password).WithJSONAccept(), "")
	testutil.AssertResponseOfRequestCall(t, request, func(response *http.Response, responseBody string) {
		assert.Equal(t, http.StatusOK, response.StatusCode)
		assert.Equal(t, `[{"endpointId":"myEndpointId","timestamp":"2009-11-17T20:34:58.651387237Z","actualRequest":{"method":"GET","url":"./http://myhost","header":null,"host":""},"actualResponse":null}]`, responseBody)
	})
}

func TestMatchesRequestHandler_getMatches_Error(t *testing.T) {
	request := testutil.CreateIncomingRequest(http.MethodGet, "/matches", testutil.CreateHeader(), "")
	assert.NoError(t, testutil.AssertHandlerFunc(t, request, matchesRequestHandlerErroneous.handleGetMatches, func(response *http.Response, responseBody string) {
		assert.Equal(t, http.StatusInternalServerError, response.StatusCode)
		assert.Equal(t, "error in get matches\n", responseBody)
	},
	))
}

func TestMatchesRequestHandler_serving_getMatchesCount(t *testing.T) {
	endpointID := "myEndpointId"
	err := matchesRequestHandler.matchStore.DeleteMatches(endpointID)
	assert.NoError(t, err)
	err = matchesRequestHandler.matchStore.AddMatch(endpointID, createMatch(endpointID))
	assert.NoError(t, err)
	request := testutil.CreateOutgoingRequest(t, http.MethodGet, "/matchesCount/"+endpointID,
		testutil.CreateHeader().WithAuth(username, password).WithJSONAccept(), "")
	testutil.AssertResponseOfRequestCall(t, request, func(response *http.Response, responseBody string) {
		assert.Equal(t, http.StatusOK, response.StatusCode)
		assert.Equal(t, `1`, responseBody)
	})
}

func TestMatchesRequestHandler_getMatchesCount_Error(t *testing.T) {
	request := testutil.CreateIncomingRequest(http.MethodGet, "/matchesCount", testutil.CreateHeader(), "")
	assert.NoError(t, testutil.AssertHandlerFunc(t, request, matchesRequestHandlerErroneous.handleGetMatchesCount, func(response *http.Response, responseBody string) {
		assert.Equal(t, http.StatusInternalServerError, response.StatusCode)
		assert.Equal(t, "error in get matches count\n", responseBody)
	},
	))
}

func TestMatchesRequestHandler_serving_getMismatches(t *testing.T) {
	err := matchesRequestHandler.matchStore.DeleteMismatches()
	assert.NoError(t, err)
	err = matchesRequestHandler.matchStore.AddMismatch(createMismatch())
	assert.NoError(t, err)
	request := testutil.CreateOutgoingRequest(t, http.MethodGet, "/mismatches",
		testutil.CreateHeader().WithAuth(username, password).WithJSONAccept(), "")
	testutil.AssertResponseOfRequestCall(t, request, func(response *http.Response, responseBody string) {
		assert.Equal(t, http.StatusOK, response.StatusCode)
		assert.Equal(t, `[{"MismatchDetails":"","timestamp":"2009-11-17T20:34:58.651387237Z","actualRequest":{"method":"GET","url":"./http://myhost","header":null,"host":""}}]`, responseBody)
	})
}

func TestMatchesRequestHandler_getMisMatches_Error(t *testing.T) {
	request := testutil.CreateIncomingRequest(http.MethodGet, "/mismatches", testutil.CreateHeader(), "")
	assert.NoError(t, testutil.AssertHandlerFunc(t, request, matchesRequestHandlerErroneous.handleGetMismatches, func(response *http.Response, responseBody string) {
		assert.Equal(t, http.StatusInternalServerError, response.StatusCode)
		assert.Equal(t, "error in get mismatches\n", responseBody)
	},
	))
}

func TestMatchesRequestHandler_serving_getMismatchesCount(t *testing.T) {
	err := matchesRequestHandler.matchStore.DeleteMismatches()
	assert.NoError(t, err)
	err = matchesRequestHandler.matchStore.AddMismatch(createMismatch())
	assert.NoError(t, err)
	request := testutil.CreateOutgoingRequest(t, http.MethodGet, "/mismatchesCount",
		testutil.CreateHeader().WithAuth(username, password).WithJSONAccept(), "")
	testutil.AssertResponseOfRequestCall(t, request, func(response *http.Response, responseBody string) {
		assert.Equal(t, http.StatusOK, response.StatusCode)
		assert.Equal(t, `1`, responseBody)
	})
}

func TestMatchesRequestHandler_getMismatchesCount_Error(t *testing.T) {
	request := testutil.CreateIncomingRequest(http.MethodGet, "/mismatchesCount", testutil.CreateHeader(), "")
	assert.NoError(t, testutil.AssertHandlerFunc(t, request, matchesRequestHandlerErroneous.handleGetMismatchesCount, func(response *http.Response, responseBody string) {
		assert.Equal(t, http.StatusInternalServerError, response.StatusCode)
		assert.Equal(t, "error in get mismatches count\n", responseBody)
	},
	))
}

func TestMatchesRequestHandler_serving_deleteMatches(t *testing.T) {
	endpointID := "myEndpointId"
	err := matchesRequestHandler.matchStore.AddMatch(endpointID, createMatch(endpointID))
	assert.NoError(t, err)
	request := testutil.CreateOutgoingRequest(t, http.MethodDelete, "/matches/"+endpointID,
		testutil.CreateHeader().WithAuth(username, password).WithJSONAccept(), "")
	testutil.AssertResponseStatusOfRequestCall(t, request, http.StatusOK)
	matches, err := matchesRequestHandler.matchStore.GetMatches(endpointID)
	assert.NoError(t, err)
	assert.Empty(t, matches)
}

func TestMatchesRequestHandler_deleteMatches_Error(t *testing.T) {
	request := testutil.CreateIncomingRequest(http.MethodDelete, "/matches", testutil.CreateHeader(), "")
	assert.NoError(t, testutil.AssertHandlerFunc(t, request, matchesRequestHandlerErroneous.handleDeleteMatches, func(response *http.Response, responseBody string) {
		assert.Equal(t, http.StatusInternalServerError, response.StatusCode)
		assert.Equal(t, "error in delete matches\n", responseBody)
	},
	))
}

func TestMatchesRequestHandler_serving_deleteMismatches(t *testing.T) {
	err := matchesRequestHandler.matchStore.AddMismatch(createMismatch())
	assert.NoError(t, err)
	request := testutil.CreateOutgoingRequest(t, http.MethodDelete, "/mismatches",
		testutil.CreateHeader().WithAuth(username, password).WithJSONAccept(), "")
	testutil.AssertResponseStatusOfRequestCall(t, request, http.StatusOK)
	mismatches, err := matchesRequestHandler.matchStore.GetMismatches()
	assert.NoError(t, err)
	assert.Empty(t, mismatches)
}

func TestMatchesRequestHandler_deleteMismatches_Error(t *testing.T) {
	request := testutil.CreateIncomingRequest(http.MethodGet, "/mismatches", testutil.CreateHeader(), "")
	assert.NoError(t, testutil.AssertHandlerFunc(t, request, matchesRequestHandlerErroneous.handleDeleteMismatches, func(response *http.Response, responseBody string) {
		assert.Equal(t, http.StatusInternalServerError, response.StatusCode)
		assert.Equal(t, "error in delete mismatches\n", responseBody)
	},
	))
}
