package kvstore

import (
	"fmt"
	"text/template"

	"github.com/alitari/mockgo/logging"
)

func KVStoreFuncMap(kvs *KVStoreJSON, logger * logging.LoggerUtil) template.FuncMap {
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
		// "kvStoreAdd": func(key, path, value string) string {
		// 	err := r.addKVStoreToCluster(key, path, value)
		// 	if err != nil {
		// 		r.logger.LogAlways(fmt.Sprintf("Error adding value: '%s' with path '%s' in kvStore: '%s' in kvStore: %v", value, path, key, err))
		// 	}
		// 	return ""
		// },
		// "kvStoreRemove": func(key, path string) string {
		// 	err := r.removeKVStoreToCluster(key, path)
		// 	if err != nil {
		// 		r.logger.LogAlways(fmt.Sprintf("Error removing value on path : '%s'  in kvStore: '%s':  %v", path, key, err))
		// 	}
		// 	return ""
		// },
		// "kvStoreJsonPath": func(key, jsonPath string) interface{} {
		// 	value, err := r.kvstore.LookUp(key, jsonPath)
		// 	if err != nil {
		// 		r.logger.LogAlways(fmt.Sprintf("Error get value with Jsonpath '%s' in kvStore: '%s' : %v", jsonPath, key, err))
		// 		return ""
		// 	}
		// 	return value
		// },
		// "matches": func(id string) []*model.Match {
		// 	return r.mockRouter.Matches[id]
		// },
	}
}
