package util

import (
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/alitari/mockgo-server/mockgo/logging"
	"github.com/go-http-utils/headers"
	"github.com/gorilla/mux"
)

/*
BasicAuthRequest checks whether the BasicAuth header of the request contains the expected username and password
*/
func BasicAuthRequest(expectedUsername, expectedPassword string, impl func(writer http.ResponseWriter, request *http.Request)) func(http.ResponseWriter, *http.Request) {
	f := func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		if ok {
			usernameMatch := username == expectedUsername
			passwordMatch := subtle.ConstantTimeCompare([]byte(password), []byte(expectedPassword)) == 1
			if usernameMatch && passwordMatch {
				impl(w, r)
			} else {
				http.Error(w, fmt.Sprintf("Authorization with username '%s' failed. ", username), http.StatusUnauthorized)
			}
		} else {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
		}
	}
	return f
}

/*
JSONContentTypeRequest checks whether the Content-type header of the request is application/json
*/
func JSONContentTypeRequest(impl func(writer http.ResponseWriter, request *http.Request)) func(http.ResponseWriter, *http.Request) {
	f := func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get(headers.ContentType) == "application/json" {
			impl(w, r)
		} else {
			http.Error(w, fmt.Sprintf("wrong request headers: Content-Type must be application/json, but is '%s'", r.Header.Get(headers.ContentType)), http.StatusUnsupportedMediaType)
		}
	}
	return f
}

/*
JSONAcceptRequest checks whether the Accept header of the request is application/json
*/
func JSONAcceptRequest(impl func(writer http.ResponseWriter, request *http.Request)) func(http.ResponseWriter, *http.Request) {
	f := func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get(headers.Accept) == "application/json" {
			impl(w, r)
		} else {
			http.Error(w, fmt.Sprintf("wrong request headers: Accept must be application/json, but is '%s'", r.Header.Get(headers.ContentType)), http.StatusUnsupportedMediaType)
		}
	}
	return f
}

/*
PathParamRequest checks whether the request path contains expected path parameters
*/
func PathParamRequest(expectedPathParams []string, impl func(writer http.ResponseWriter, request *http.Request)) func(http.ResponseWriter, *http.Request) {
	f := func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		for _, expectedPathParam := range expectedPathParams {
			if vars[expectedPathParam] == "" {
				http.Error(w, fmt.Sprintf("url path param '%s' is not set", expectedPathParam), http.StatusNotFound)
				return
			}
		}
		impl(w, r)
	}
	return f
}

/*
LoggingRequest writes details of request and response
*/
func LoggingRequest(loggerUtil *logging.LoggerUtil, impl func(writer http.ResponseWriter, request *http.Request)) func(http.ResponseWriter, *http.Request) {
	f := func(w http.ResponseWriter, r *http.Request) {
		loggerUtil.LogIncomingRequest(r)
		if loggerUtil.Level >= logging.Debug {
			w = logging.NewResponseWriter(w, loggerUtil, 2)
		}
		impl(w, r)
		if loggerUtil.Level >= logging.Debug {
			w.(*logging.ResponseWriter).Log()
		}
	}
	return f
}

/*
WriteEntity marshals an entity and writes the output to the http response
*/
func WriteEntity(writer http.ResponseWriter, entity interface{}) {
	entityString, isString := entity.(string)
	if !isString {
		entityBytes, err := json.Marshal(entity)
		if err != nil {
			http.Error(writer, fmt.Sprintf("Cannot marshall response: %v", err), http.StatusInternalServerError)
			return
		}
		entityString = string(entityBytes)
	}
	_, err := io.WriteString(writer, entityString)
	if err != nil {
		http.Error(writer, fmt.Sprintf("Cannot write response: %v", err), http.StatusInternalServerError)
		return
	}
}

/*
BasicAuth create a BasicAuth header value for given user credentials
*/
func BasicAuth(username, password string) string {
	auth := username + ":" + password
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(auth))
}
