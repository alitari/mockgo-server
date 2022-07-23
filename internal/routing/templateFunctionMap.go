package routing

import (
	"fmt"
	"text/template"
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
	}
}
