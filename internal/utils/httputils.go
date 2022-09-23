package utils

import (
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/go-http-utils/headers"
	"github.com/gorilla/mux"
)

func RequestMustHave(loggerUtil *LoggerUtil, expectedUsername, expectedPassword, method, contentType, acceptType string, urlPathParams []string, impl func(writer http.ResponseWriter, request *http.Request)) func(http.ResponseWriter, *http.Request) {
	f := func(w http.ResponseWriter, r *http.Request) {
		loggerUtil.LogIncomingRequest(r)
		if loggerUtil.Level >= Debug {
			w = NewLoggingResponseWriter(w, loggerUtil,2)
		}
		noAuth := len(expectedUsername) == 0 && len(expectedPassword) == 0
		username, password, ok := r.BasicAuth()
		if ok || noAuth {
			usernameMatch := noAuth || username == expectedUsername
			passwordMatch := noAuth || subtle.ConstantTimeCompare([]byte(password), []byte(expectedPassword)) == 1

			if usernameMatch && passwordMatch {
				if r.Method == method {
					if len(contentType) == 0 || r.Header.Get(headers.ContentType) == contentType {
						if len(acceptType) == 0 || r.Header.Get(headers.Accept) == acceptType {
							if urlPathParams != nil {
								vars := mux.Vars(r)
								for _, urlPathParam := range urlPathParams {
									if vars[urlPathParam] == "" {
										http.Error(w, fmt.Sprintf("url path param '%s' is not set", urlPathParam), http.StatusNotFound)
										if loggerUtil.Level >= Debug {
											w.(*LoggingResponseWriter).Log()
										}
										return
									}
								}
							}
							impl(w, r)
						} else {
							http.Error(w, fmt.Sprintf("wrong request headers: Accept must be %s, but is %s ", acceptType, r.Header.Get(headers.Accept)), http.StatusUnsupportedMediaType)
						}
					} else {
						http.Error(w, fmt.Sprintf("wrong request headers: Content-Type must be %s, but is %s ", contentType, r.Header.Get(headers.ContentType)), http.StatusUnsupportedMediaType)
					}
				} else {
					http.Error(w, fmt.Sprintf("wrong http request method: must be %s ", method), http.StatusBadRequest)
				}
			} else {
				http.Error(w, fmt.Sprintf("Authorization with username '%s' failed. ", username), http.StatusUnauthorized)
			}
		} else {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
		}
		if loggerUtil.Level >= Debug {
			w.(*LoggingResponseWriter).Log()
		}
	}
	return f
}

func WriteEntity(writer http.ResponseWriter, entity interface{}) {
	entityBytes, err := json.Marshal(entity)
	if err != nil {
		http.Error(writer, fmt.Sprintf("Cannot marshall response: %v", err), http.StatusInternalServerError)
		return
	}
	_, err = io.WriteString(writer, string(entityBytes))
	if err != nil {
		http.Error(writer, fmt.Sprintf("Cannot write response: %v", err), http.StatusInternalServerError)
		return
	}
}

func BasicAuth(username, password string) string {
	auth := username + ":" + password
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(auth))
}

func CreateHttpClient(timeout time.Duration) http.Client {
	httpClient := http.Client{Timeout: timeout}
	return httpClient
}

// func containKeys(amap map[string]string, keys []string) bool {
// 	for _, key := range keys {
// 		if len(amap[key]) == 0 {
// 			return false
// 		}
// 	}
// 	return true

// }
