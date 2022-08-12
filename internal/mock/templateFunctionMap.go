package mock

import (
	"fmt"
	"text/template"
	"time"

	"github.com/alitari/mockgo-server/internal/model"
)

func (r *MockRouter) templateFuncMap() template.FuncMap {
	return template.FuncMap{
		"kvStoreGet": func(key string) interface{} {
			val, err := r.kvstore.Get(key)
			if err != nil {
				return ""
			} else {
				return val
			}
		},
		"kvStorePut": func(key string, jsonValue string) string {
			err := r.kvstore.Put(key, jsonValue)
			if err != nil {
				r.logger.LogAlways(fmt.Sprintf("Error storing key: '%s' with value:\n'%s' in kvStore: %v", key, jsonValue, err))
			}
			return ""
		},
		"endpointIds": func() []string {
			var ids []string
			for _, ep := range r.AllEndpoints() {
				ids = append(ids, ep.Id)
			}
			return ids
		},
		"matches": func(id string) []*model.Match {
			return r.Matches[id]
		},
		"delay": func(millis int) string {
			time.Sleep(time.Duration(millis) * time.Millisecond)
			return ""
		},
	}
}
