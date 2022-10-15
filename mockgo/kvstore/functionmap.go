package kvstore

import (
	"fmt"
	"text/template"

	"github.com/alitari/mockgo/logging"
)

func KVStoreFuncMap(kvs *KVStoreJSON, logger *logging.LoggerUtil) template.FuncMap {
	return template.FuncMap{
		"kvStoreGet": func(key string) interface{} {
			if val, err := kvs.Get(key); err != nil {
				return nil
			} else {
				return val
			}
		},
		"kvStorePut": func(key string, jsonValue string) string {
			err := kvs.Put(key, jsonValue)
			if err != nil {
				logger.LogError(fmt.Sprintf("Error storing value:'%s' in kvStore '%s'", jsonValue, key), err)
			}
			return ""
		},
		"kvStoreAdd": func(key, path, value string) string {
			err := kvs.PatchAdd(key, path, value)
			if err != nil {
				logger.LogError(fmt.Sprintf("Error adding value: '%s' with path '%s' in kvStore: '%s' in kvStore", value, path, key), err)
			}
			return ""
		},
		"kvStoreRemove": func(key, path string) string {
			err := kvs.PatchRemove(key, path)
			if err != nil {
				logger.LogError(fmt.Sprintf("Error removing value on path : '%s'  in kvStore: '%s'", path, key), err)
			}
			return ""
		},
		"kvStoreLookup": func(key, jsonPath string) interface{} {
			value, err := kvs.LookUp(key, jsonPath)
			if err != nil {
				logger.LogError(fmt.Sprintf("Error get value with Jsonpath '%s' in kvStore: '%s'", jsonPath, key), err)
				return ""
			}
			return value
		},
	}
}
