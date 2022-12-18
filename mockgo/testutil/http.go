package testutil

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-http-utils/headers"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

var testServer *httptest.Server

// StartServing creates httptest.Server
func StartServing(router *mux.Router) {
	testServer = httptest.NewServer(router)
}

// StopServing close httptest.Server
func StopServing() {
	testServer.Close()
}

// Header type for http header
type Header struct {
	entries map[string][]string
}

// CreateHeader creates a http header
func CreateHeader() Header {
	return Header{entries: map[string][]string{}}
}

// WithAuth adds BasicAuth to the header
func (h Header) WithAuth(username, password string) Header {
	h.entries[headers.Authorization] = []string{basicAuth(username, password)}
	return h
}

func basicAuth(username, password string) string {
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(username+":"+password))
}

// WithKeyValue adds key value to the header
func (h Header) WithKeyValue(key, value string) Header {
	h.entries[key] = []string{value}
	return h
}

// WithJSONContentType adds JSON Content-Type to the header
func (h Header) WithJSONContentType() Header {
	h.entries[headers.ContentType] = []string{"application/json"}
	return h
}

// WithJSONAccept adds JSON Accept to the header
func (h Header) WithJSONAccept() Header {
	h.entries[headers.Accept] = []string{"application/json"}
	return h
}

// CreateOutgoingRequest creates a http request
func CreateOutgoingRequest(t *testing.T, method, path string, header Header, body string) *http.Request {
	var bodyReader io.Reader
	if len(body) > 0 {
		bodyReader = bytes.NewBufferString(body)
	}

	request, err := http.NewRequest(method, fmt.Sprintf("%s%s", testServer.URL, path), bodyReader)
	assert.NoError(t, err)
	return setHeader(request, header)
}

func setHeader(request *http.Request, header Header) *http.Request {
	for k, v := range header.entries {
		request.Header.Add(k, v[0])
	}
	return request
}

// ReaderFunc proxy implementation
type ReaderFunc func(p []byte) (n int, err error)

func (f ReaderFunc) Read(p []byte) (n int, err error) {
	return f(p)
}

func errorReader(p []byte) (n int, err error) {
	return 0, fmt.Errorf("error reading bytes")
}

// CreateIncomingErrorReadingBodyRequest create test http request which fails when reading the request body
func CreateIncomingErrorReadingBodyRequest(method, path string, header Header) *http.Request {
	request := httptest.NewRequest(method, fmt.Sprintf("%s%s", testServer.URL, path), ReaderFunc(errorReader))
	return setHeader(request, header)
}

// CreateIncomingRequest create test http request for a an input in a request handler func
func CreateIncomingRequest(method, path string, header Header, body string) *http.Request {
	var bodyReader io.Reader
	if len(body) > 0 {
		bodyReader = bytes.NewBufferString(body)
	}

	request := httptest.NewRequest(method, fmt.Sprintf("%s%s", "http://localhost/", path), bodyReader)
	return setHeader(request, header)
}

// AssertHandlerFunc test a request handler func with a test http request, see CreateIncomingRequest
func AssertHandlerFunc(t *testing.T, request *http.Request, handlerFunc func(http.ResponseWriter, *http.Request), assertResponse func(response *http.Response, responseBody string)) error {
	recorder := httptest.NewRecorder()
	handlerFunc(recorder, request)
	response := recorder.Result()
	return callAssertResponse(response, assertResponse)
}

// AssertResponseOfRequestCall test a http request
func AssertResponseOfRequestCall(t *testing.T, request *http.Request, assertResponse func(response *http.Response, responseBody string)) error {
	t.Logf("call with request: %v ...", request)
	response, err := testServer.Client().Do(request)
	if err != nil {
		return err
	}
	return callAssertResponse(response, assertResponse)
}

func callAssertResponse(response *http.Response, assertResponse func(response *http.Response, responseBody string)) error {
	defer response.Body.Close()
	respBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}
	respBody := string(respBytes)
	assertResponse(response, respBody)
	return nil
}

// AssertResponseStatusOfRequestCall test a http request
func AssertResponseStatusOfRequestCall(t *testing.T, request *http.Request, expectedResponseStatus int) error {
	t.Logf("call with request: %v ...", request)
	response, err := testServer.Client().Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	t.Logf("response status: '%s'", response.Status)
	assert.Equal(t, expectedResponseStatus, response.StatusCode)
	return nil
}
