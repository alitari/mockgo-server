package utils

import (
	"fmt"
	"net/http"

	"github.com/go-http-utils/headers"
	"github.com/gorilla/mux"
)

func RequestMustHave(method, contentType, acceptType string, urlPathParams []string, impl func(writer http.ResponseWriter, request *http.Request)) func(http.ResponseWriter, *http.Request) {
	f := func(w http.ResponseWriter, r *http.Request) {
		if r.Method == method {
			if len(contentType) == 0 || r.Header.Get(headers.ContentType) == contentType {
				if len(acceptType) == 0 || r.Header.Get(headers.Accept) == acceptType {
					if urlPathParams != nil {
						vars := mux.Vars(r)
						for _, urlPathParam := range urlPathParams {
							if vars[urlPathParam] == "" {
								http.Error(w, fmt.Sprintf("url path param '%s' is not set", urlPathParam), http.StatusNotFound)
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
	}
	return f
}
