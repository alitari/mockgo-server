package mock

import (
	"fmt"
	"text/template"

	"github.com/alitari/mockgo-server/internal/kvstore"
)

func (r *MockRouter) templateFuncMap() template.FuncMap {
	return template.FuncMap{
		"kvStoreGet": func(key string) interface{} {
			val, err := kvstore.TheKVStore.Get(key)
			if err != nil {
				return ""
			} else {
				return val
			}
		},
		"kvStorePut": func(key string, jsonValue string) string {
			err := kvstore.TheKVStore.Put(key, jsonValue)
			if err != nil {
				r.logger.LogAlways(fmt.Sprintf("Error storing key: '%s' with value:\n'%s' in kvStore: %v", key, jsonValue, err))
			}
			return ""
		},
	}
}
