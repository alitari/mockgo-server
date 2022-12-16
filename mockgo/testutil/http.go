package testutil

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alitari/mockgo-server/mockgo/util"
	"github.com/go-http-utils/headers"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

var testServer *httptest.Server

func StartServing(router *mux.Router) {
	testServer = httptest.NewServer(router)
}

func StopServing() {
	testServer.Close()
}

type Header struct {
	entries map[string][]string
}

func CreateHeader() Header {
	return Header{entries: map[string][]string{}}
}

func (h Header) WithAuth(username, password string) Header {
	h.entries[headers.Authorization] = []string{util.BasicAuth(username, password)}
	return h
}

func (h Header) WithKeyValue(key, value string) Header {
	h.entries[key] = []string{value}
	return h
}

func (h Header) WithJsonContentType() Header {
	h.entries[headers.ContentType] = []string{"application/json"}
	return h
}

func (h Header) WithJsonAccept() Header {
	h.entries[headers.Accept] = []string{"application/json"}
	return h
}

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

type ReaderFunc func(p []byte) (n int, err error)

func (f ReaderFunc) Read(p []byte) (n int, err error) {
	return f(p)
}

func ErrorReader(p []byte) (n int, err error) {
	return 0, fmt.Errorf("error reading bytes")
}

func CreateIncomingErrorReadingBodyRequest(method, path string, header Header) *http.Request {
	request := httptest.NewRequest(method, fmt.Sprintf("%s%s", testServer.URL, path), ReaderFunc(ErrorReader))
	return setHeader(request, header)
}

func CreateIncomingRequest(method, path string, header Header, body string) *http.Request {
	var bodyReader io.Reader
	if len(body) > 0 {
		bodyReader = bytes.NewBufferString(body)
	}

	request := httptest.NewRequest(method, fmt.Sprintf("%s%s", testServer.URL, path), bodyReader)
	return setHeader(request, header)
}

func AssertHandlerFunc(t *testing.T, request *http.Request, handlerFunc func(http.ResponseWriter, *http.Request), assertResponse func(response *http.Response, responseBody string)) error {
	recorder := httptest.NewRecorder()
	handlerFunc(recorder, request)
	response := recorder.Result()
	return callAssertResponse(response, assertResponse)
}

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
