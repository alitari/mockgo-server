package util

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alitari/mockgo-server/mockgo/testutil"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

var impl = func(writer http.ResponseWriter, request *http.Request) {
	writer.WriteHeader(http.StatusOK)
	writer.Write([]byte("OK"))
}

//func TestBasicAuthRequest(t *testing.T) {
//	request := testutil.CreateIncomingRequest(http.MethodGet, "/hello", testutil.CreateHeader(), "")
//	assert.NoError(t, testutil.AssertHandlerFunc(t, request, BasicAuthRequest("alex", "alexpassword", impl),
//		func(response *http.Response, responseBody string) {
//			assert.Equal(t, http.StatusUnauthorized, response.StatusCode)
//			assert.Equal(t, "Unauthorized\n", responseBody)
//		}))
//
//	request = testutil.CreateIncomingRequest(http.MethodGet, "/hello", testutil.CreateHeader().WithAuth("alex", "wrongpass"), "")
//	assert.NoError(t, testutil.AssertHandlerFunc(t, request, BasicAuthRequest("alex", "alexpassword", impl),
//		func(response *http.Response, responseBody string) {
//			assert.Equal(t, http.StatusUnauthorized, response.StatusCode)
//			assert.Equal(t, "Authorization with username 'alex' failed. \n", responseBody)
//		}))
//	request = testutil.CreateIncomingRequest(http.MethodGet, "/hello", testutil.CreateHeader().WithAuth("alex", "alexpassword"), "")
//	assert.NoError(t, testutil.AssertHandlerFunc(t, request, BasicAuthRequest("alex", "alexpassword", impl),
//		func(response *http.Response, responseBody string) {
//			assert.Equal(t, http.StatusOK, response.StatusCode)
//			assert.Equal(t, "OK", responseBody)
//		}))
//}

func TestJsonContentTypeRequest(t *testing.T) {
	request := testutil.CreateIncomingRequest(http.MethodGet, "/hello", testutil.CreateHeader(), "")
	assert.NoError(t, testutil.AssertHandlerFunc(t, request, JSONContentTypeRequest(impl),
		func(response *http.Response, responseBody string) {
			assert.Equal(t, http.StatusUnsupportedMediaType, response.StatusCode)
			assert.Equal(t, "wrong request headers: Content-Type must be application/json, but is ''\n", responseBody)
		}))
	request = testutil.CreateIncomingRequest(http.MethodGet, "/hello", testutil.CreateHeader().WithJSONContentType(), "")
	assert.NoError(t, testutil.AssertHandlerFunc(t, request, JSONContentTypeRequest(impl),
		func(response *http.Response, responseBody string) {
			assert.Equal(t, http.StatusOK, response.StatusCode)
			assert.Equal(t, "OK", responseBody)
		}))
}

func TestJsonAcceptRequest(t *testing.T) {
	request := testutil.CreateIncomingRequest(http.MethodGet, "/hello", testutil.CreateHeader(), "")
	assert.NoError(t, testutil.AssertHandlerFunc(t, request, JSONAcceptRequest(impl),
		func(response *http.Response, responseBody string) {
			assert.Equal(t, http.StatusUnsupportedMediaType, response.StatusCode)
			assert.Equal(t, "wrong request headers: Accept must be application/json, but is ''\n", responseBody)
		}))
	request = testutil.CreateIncomingRequest(http.MethodGet, "/hello", testutil.CreateHeader().WithJSONAccept(), "")
	assert.NoError(t, testutil.AssertHandlerFunc(t, request, JSONAcceptRequest(impl),
		func(response *http.Response, responseBody string) {
			assert.Equal(t, http.StatusOK, response.StatusCode)
			assert.Equal(t, "OK", responseBody)
		}))
}

const (
	varsKey int = iota
)

func TestPathParamRequest(t *testing.T) {
	request := testutil.CreateIncomingRequest(http.MethodGet, "/hello", testutil.CreateHeader(), "")
	assert.NoError(t, testutil.AssertHandlerFunc(t, request, PathParamRequest([]string{"param1"}, impl),
		func(response *http.Response, responseBody string) {
			assert.Equal(t, http.StatusNotFound, response.StatusCode)
			assert.Equal(t, "url path param 'param1' is not set\n", responseBody)
		}))
	request = testutil.CreateIncomingRequest(http.MethodGet, "/hello", testutil.CreateHeader(), "")
	request = mux.SetURLVars(request, map[string]string{"param1": "value1"})
	assert.NoError(t, testutil.AssertHandlerFunc(t, request, PathParamRequest([]string{"param1"}, impl),
		func(response *http.Response, responseBody string) {
			assert.Equal(t, http.StatusOK, response.StatusCode)
			assert.Equal(t, "OK", responseBody)
		}))
}

//func TestLoggingRequest(t *testing.T) {
//	request := testutil.CreateIncomingRequest(http.MethodGet, "/hello", testutil.CreateHeader(), "")
//	assert.NoError(t, testutil.AssertHandlerFunc(t, request, LoggingRequest(logging.NewLoggerUtil(logging.Debug), impl),
//		func(response *http.Response, responseBody string) {
//			assert.Equal(t, http.StatusOK, response.StatusCode)
//			assert.Equal(t, "OK", responseBody)
//		}))
//}

func TestWriteEntityString(t *testing.T) {
	entity := "Hello"
	writer := httptest.NewRecorder()
	WriteEntity(writer, entity)
	assert.Equal(t, http.StatusOK, writer.Result().StatusCode)
	respBytes, err := io.ReadAll(writer.Result().Body)
	assert.NoError(t, err)
	assert.Equal(t, entity, string(respBytes))
}

func TestWriteEntityStruct(t *testing.T) {
	entity := struct {
		Foo string
		Bar int
	}{"test", 42}
	writer := httptest.NewRecorder()
	WriteEntity(writer, entity)
	assert.Equal(t, http.StatusOK, writer.Result().StatusCode)
	respBytes, err := io.ReadAll(writer.Result().Body)
	assert.NoError(t, err)
	assert.Equal(t, `{"Foo":"test","Bar":42}`, string(respBytes))
}

func TestWriteEntity_Marshall_error(t *testing.T) {
	entity := make(chan int)
	writer := httptest.NewRecorder()
	WriteEntity(writer, entity)
	assert.Equal(t, http.StatusInternalServerError, writer.Result().StatusCode)
	respBytes, err := io.ReadAll(writer.Result().Body)
	assert.NoError(t, err)
	assert.Equal(t, "Cannot marshall response: json: unsupported type: chan int\n", string(respBytes))
}

func TestBasicAuth(t *testing.T) {
	assert.Equal(t, "Basic YWxleDpwYXNzd29yZA==", BasicAuth("alex", "password"))
}
