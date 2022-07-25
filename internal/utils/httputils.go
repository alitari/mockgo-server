package utils

import (
	"fmt"
	"net/http"

	"github.com/go-http-utils/headers"
)

func RequestMustBe(method, contentType, acceptType string, impl func(writer http.ResponseWriter, request *http.Request)) func(http.ResponseWriter, *http.Request) {

	f := func(w http.ResponseWriter, r *http.Request) {
		if r.Method == method {
			if (len(contentType) == 0 || r.Header.Get(headers.ContentType) == contentType) && (len(acceptType) == 0 || r.Header.Get(headers.Accept) == acceptType) {
				impl(w, r)
			} else {
				http.Error(w, fmt.Sprintf("wrong request headers: Content-Type must be %s and Accept must be %s ", contentType, acceptType), http.StatusUnsupportedMediaType)
			}
		} else {
			http.Error(w, fmt.Sprintf("wrong http request method: must be %s ", method), http.StatusBadRequest)
		}
	}
	return f

}
