package kvstore

import "text/template"

/*
Storage is *the* interface for the key value store

There can be implementations using:
- a local in-memory storage ,
- multiple servers for a distributed
- a database like redis
*/
type Storage interface {
	GetAll(store string) (map[string]interface{}, error)
	Put(store, key string, val interface{}) error
	Get(store, key string) (interface{}, error)
}

/*
InmemoryStorage is an implementation of Storage using local in-memory storage
*/
type InmemoryStorage struct {
	store map[string]map[string]interface{}
}

/*
GetAll returns all values for a given store
*/
func (s *InmemoryStorage) GetAll(store string) (map[string]interface{}, error) {
	st := s.store[store]
	if st == nil {
		return nil, nil
	}
	return st, nil
}

/*
Get returns a value for a given key
*/
func (s *InmemoryStorage) Get(store, key string) (interface{}, error) {
	st := s.store[store]
	if st == nil {
		return nil, nil
	}
	return st[key], nil
}

/*
Put stores a value for a given key
*/
func (s *InmemoryStorage) Put(store, key string, val interface{}) error {
	st := s.store[store]
	if st == nil {
		s.store[store] = map[string]interface{}{}
		st = s.store[store]
	}
	st[key] = val
	return nil
}

/*
NewInmemoryStorage creates a new instance of InmemoryStorage
*/
func NewInmemoryStorage() *InmemoryStorage {
	st := map[string]map[string]interface{}{}
	return &InmemoryStorage{store: st}
}

//	func newJSONStorage(kvStore Storage, log bool) *jsonStorage {
//		return &jsonStorage{store: kvStore, log: log}
//	}
//
//	func (s *jsonStorage) putAsJSON(key, jsonStr string) error {
//		var storeVal interface{}
//		err := json.Unmarshal([]byte(jsonStr), &storeVal)
//		if err != nil {
//			return err
//		}
//		err = s.put(key, storeVal)
//		return err
//	}
//
//	func (s *jsonStorage) get(key string) (interface{}, error) {
//		return s.store.GetVal(key)
//	}
//
//	func (s *jsonStorage) put(key string, storeVal interface{}) error {
//		return s.store.PutVal(key, storeVal)
//	}
//
//	func (s *jsonStorage) getAsJSON(key string) (string, error) {
//		storeVal, err := s.get(key)
//		if err != nil {
//			return "", err
//		}
//		if storeVal == nil {
//			return "{}", nil
//		}
//		storeJSON, err := json.Marshal(storeVal)
//		if err != nil {
//			return "", err
//		}
//		return string(storeJSON), nil
//	}
//
//	func (s *jsonStorage) patchAdd(key, path, value string) error {
//		if !strings.HasPrefix(value, "{") && !strings.HasPrefix(value, "[") {
//			value = "\"" + value + "\""
//		}
//		return s.patch(key, fmt.Sprintf(`[{"op":"%s","path":"%s","value": %s}]`, add.String(), path, value))
//	}
//
//	func (s *jsonStorage) patchRemove(key, path string) error {
//		return s.patch(key, fmt.Sprintf(`[{"op":"%s","path":"%s"}]`, remove.String(), path))
//	}
//
//	func (s *jsonStorage) patchReplace(key, path, value string) error {
//		return s.patch(key, fmt.Sprintf(`[{"op":"%s","path":"%s","value": %s}]`, replace.String(), path, value))
//	}
//
//	func (s *jsonStorage) patch(key, patchJSON string) error {
//		s.logStr("patchJson=" + patchJSON)
//		patch, err := jsonpatch.DecodePatch([]byte(patchJSON))
//		if err != nil {
//			return err
//		}
//		storeJSON, err := s.getAsJSON(key)
//		if err != nil {
//			return err
//		}
//
//		s.logStr("storeJson=" + storeJSON)
//		modifiedStoreJSON, err := patch.Apply([]byte(storeJSON))
//		if err != nil {
//			return err
//		}
//		s.logStr("modifiedStoreJson=" + string(modifiedStoreJSON))
//		err = s.putAsJSON(key, string(modifiedStoreJSON))
//		if err != nil {
//			return err
//		}
//		return nil
//	}
//
//	func (s *jsonStorage) lookUp(key, jsonPath string) (interface{}, error) {
//		s.logStr("jsonpath=" + jsonPath)
//		value, err := s.get(key)
//		if err != nil {
//			return "", err
//		}
//		res, err := jsonpath.JsonPathLookup(value, jsonPath)
//		if err != nil {
//			return "", err
//		}
//		return res, nil
//	}
//
//	func (s *jsonStorage) lookUpJSON(key, jsonPath string) (string, error) {
//		res, err := s.lookUp(key, jsonPath)
//		if err != nil {
//			return "", err
//		}
//		resJSON, err := json.Marshal(res)
//		if err != nil {
//			return "", err
//		}
//		s.logStr(fmt.Sprintf("jsonPath jsonresult='%s'", resJSON))
//		return string(resJSON), nil
//	}
//
//	func (s *jsonStorage) logStr(message string) {
//		if s.log {
//			log.Print(message)
//		}
//	}
func NewKVStoreTemplateFuncMap(storage Storage) template.FuncMap {
	return template.FuncMap{
		"kvStoreGet": func(store, key string) interface{} {
			val, err := storage.Get(store, key)
			if err != nil {
				return err.Error()
			}
			return val
		},
		"kvStorePut": func(store, key string, val interface{}) string {
			if err := storage.Put(store, key, val); err != nil {
				return err.Error()
			}
			return ""
		},
		"kvStoreRemove": func(store, key string) string {
			if err := storage.Put(store, key, nil); err != nil {
				return err.Error()
			}
			return ""
		},
	}
}
