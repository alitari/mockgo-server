package routing

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/alitari/mockgo-server/internal/model"

	"github.com/gorilla/mux"
)

func Load(configFile string) (*mux.Router, error) {

	data, err := ioutil.ReadFile(configFile)
	if err != nil {
		return nil, err
	}

	m := &model.Mock{}
	err = json.Unmarshal(data, m)
	if err != nil {
		return nil, err
	}

	router := mux.NewRouter()
	for _, e := range m.Endpoints {
		r := router.HandleFunc(e.Request.Path, handlerFactory(router, e))
		if e.Request.Method != "" {
			r.Methods(e.Request.Method)
		}
		if e.Request.Scheme != "" {
			r.Schemes(e.Request.Scheme)
		}
		if e.Request.Host != "" {
			r.Host(e.Request.Host)
		}
		if len(e.Request.Query) > 0 {
			for key, val := range e.Request.Query {
				r.Queries(key, val)
			}
		}
		if len(e.Request.Headers) > 0 {
			for key, val := range e.Request.Headers {
				r.Headers(key, val)
			}
		}
	}

	return router, nil
}

func handlerFactory(r *mux.Router, e model.Endpoint) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if len(e.Response.Headers) > 0 {
			for key, val := range e.Response.Headers {
				w.Header().Set(key, val)
			}
		}
		if e.Response.Body != "" {
			if e.Response.StatusCode > 0 {
				w.WriteHeader(e.Response.StatusCode)
			} else {
				w.WriteHeader(http.StatusOK)
			}
			fmt.Fprint(w, e.Response.Body)
			return
		}
		if e.Response.BodyFileName != "" {
			http.ServeFile(w, r, e.Response.BodyFileName)
			return
		}
	}
}
