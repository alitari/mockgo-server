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

func (h Header) WithJsonContentType() Header {
	h.entries[headers.ContentType] = []string{"application/json"}
	return h
}

func (h Header) WithJsonAccept() Header {
	h.entries[headers.Accept] = []string{"application/json"}
	return h
}

func CreateRequest(t *testing.T, method, path string, header Header, body string) *http.Request {
	var bodyReader io.Reader
	if len(body) > 0 {
		bodyReader = bytes.NewBufferString(body)
	}

	request, err := http.NewRequest(method, fmt.Sprintf("%s%s", testServer.URL, path), bodyReader)
	assert.NoError(t, err)
	for k, v := range header.entries {
		request.Header.Add(k, v[0])
	}
	return request
}

func AssertResponseOfRequestCall(t *testing.T, request *http.Request, assertResponse func(response *http.Response, responseBody string)) error {
	t.Logf("call with request: %v ...", request)
	response, err := testServer.Client().Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	t.Logf("response status: '%s'", response.Status)
	respBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}
	respBody := string(respBytes)
	t.Logf("response body:\n '%s'", respBody)
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
