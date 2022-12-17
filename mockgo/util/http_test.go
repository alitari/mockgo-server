package util

import (
	"net/http"
	"testing"

	"github.com/alitari/mockgo-server/mockgo/testutil"
	"github.com/stretchr/testify/assert"
)

var impl = func(writer http.ResponseWriter, request *http.Request) {
	writer.WriteHeader(http.StatusOK)
	writer.Write([]byte("OK"))
}

func TestBasicAuthRequest(t *testing.T) {
	request := testutil.CreateIncomingRequest(http.MethodGet, "/hello", testutil.CreateHeader(), "")
	assert.NoError(t, testutil.AssertHandlerFunc(t, request, BasicAuthRequest("alex", "alexpassword", impl),
		func(response *http.Response, responseBody string) {
			assert.Equal(t, http.StatusUnauthorized, response.StatusCode)
			assert.Equal(t, "Unauthorized\n", responseBody)
		}))

	request = testutil.CreateIncomingRequest(http.MethodGet, "/hello", testutil.CreateHeader().WithAuth("alex", "wrongpass"), "")
	assert.NoError(t, testutil.AssertHandlerFunc(t, request, BasicAuthRequest("alex", "alexpassword", impl),
		func(response *http.Response, responseBody string) {
			assert.Equal(t, http.StatusUnauthorized, response.StatusCode)
			assert.Equal(t, "Authorization with username 'alex' failed. \n", responseBody)
		}))
	request = testutil.CreateIncomingRequest(http.MethodGet, "/hello", testutil.CreateHeader().WithAuth("alex", "alexpassword"), "")
	assert.NoError(t, testutil.AssertHandlerFunc(t, request, BasicAuthRequest("alex", "alexpassword", impl),
		func(response *http.Response, responseBody string) {
			assert.Equal(t, http.StatusOK, response.StatusCode)
			assert.Equal(t, "OK", responseBody)
		}))
}

func TestJsonContentTypeRequest(t *testing.T) {
	request := testutil.CreateIncomingRequest(http.MethodGet, "/hello", testutil.CreateHeader(), "")
	assert.NoError(t, testutil.AssertHandlerFunc(t, request, JsonContentTypeRequest(impl),
		func(response *http.Response, responseBody string) {
			assert.Equal(t, http.StatusUnsupportedMediaType, response.StatusCode)
			assert.Equal(t, "wrong request headers: Content-Type must be application/json, but is ''\n", responseBody)
		}))
	request = testutil.CreateIncomingRequest(http.MethodGet, "/hello", testutil.CreateHeader().WithJsonContentType(), "")
	assert.NoError(t, testutil.AssertHandlerFunc(t, request, JsonContentTypeRequest(impl),
		func(response *http.Response, responseBody string) {
			assert.Equal(t, http.StatusOK, response.StatusCode)
			assert.Equal(t, "OK", responseBody)
		}))
}

func TestJsonAcceptRequest(t *testing.T) {
	request := testutil.CreateIncomingRequest(http.MethodGet, "/hello", testutil.CreateHeader(), "")
	assert.NoError(t, testutil.AssertHandlerFunc(t, request, JsonAcceptRequest(impl),
		func(response *http.Response, responseBody string) {
			assert.Equal(t, http.StatusUnsupportedMediaType, response.StatusCode)
			assert.Equal(t, "wrong request headers: Accept must be application/json, but is ''\n", responseBody)
		}))
	request = testutil.CreateIncomingRequest(http.MethodGet, "/hello", testutil.CreateHeader().WithJsonAccept(), "")
	assert.NoError(t, testutil.AssertHandlerFunc(t, request, JsonAcceptRequest(impl),
		func(response *http.Response, responseBody string) {
			assert.Equal(t, http.StatusOK, response.StatusCode)
			assert.Equal(t, "OK", responseBody)
		}))
}
