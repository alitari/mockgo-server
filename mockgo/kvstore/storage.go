package kvstore

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"text/template"

	jsonpatch "github.com/evanphx/json-patch"
	jsonpath "github.com/oliveagle/jsonpath"
)

/*
Storage is *the* interface for the key value store

There can be implementations using:
- a local in-memory storage ,
- multiple servers for a distributed
- a database like redis
*/
type Storage interface {
	GetVal(key string) (interface{}, error)
	PutVal(key string, storeVal interface{}) error
}

type jsonStorage struct {
	log   bool
	store Storage
}

type patchOp int64

const (
	add patchOp = iota
	replace
	remove
)

func (pop patchOp) String() string {
	switch pop {
	case add:
		return "add"
	case replace:
		return "replace"
	case remove:
		return "remove"
	}
	return "unknown"
}

func newJSONStorage(kvStore Storage, log bool) *jsonStorage {
	return &jsonStorage{store: kvStore, log: log}
}

func (s *jsonStorage) putAsJSON(key, jsonStr string) error {
	var storeVal interface{}
	err := json.Unmarshal([]byte(jsonStr), &storeVal)
	if err != nil {
		return err
	}
	err = s.put(key, storeVal)
	return err
}

func (s *jsonStorage) get(key string) (interface{}, error) {
	return s.store.GetVal(key)
}

func (s *jsonStorage) put(key string, storeVal interface{}) error {
	return s.store.PutVal(key, storeVal)
}

func (s *jsonStorage) getAsJSON(key string) (string, error) {
	storeVal, err := s.get(key)
	if err != nil {
		return "", err
	}
	if storeVal == nil {
		return "{}", nil
	}
	storeJSON, err := json.Marshal(storeVal)
	if err != nil {
		return "", err
	}
	return string(storeJSON), nil
}

func (s *jsonStorage) patchAdd(key, path, value string) error {
	if !strings.HasPrefix(value, "{") && !strings.HasPrefix(value, "[") {
		value = "\"" + value + "\""
	}
	return s.patch(key, fmt.Sprintf(`[{"op":"%s","path":"%s","value": %s}]`, add.String(), path, value))
}

func (s *jsonStorage) patchRemove(key, path string) error {
	return s.patch(key, fmt.Sprintf(`[{"op":"%s","path":"%s"}]`, remove.String(), path))
}

func (s *jsonStorage) patchReplace(key, path, value string) error {
	return s.patch(key, fmt.Sprintf(`[{"op":"%s","path":"%s","value": %s}]`, replace.String(), path, value))
}

func (s *jsonStorage) patch(key, patchJSON string) error {
	s.logStr("patchJson=" + patchJSON)
	patch, err := jsonpatch.DecodePatch([]byte(patchJSON))
	if err != nil {
		return err
	}
	storeJSON, err := s.getAsJSON(key)
	if err != nil {
		return err
	}

	s.logStr("storeJson=" + storeJSON)
	modifiedStoreJSON, err := patch.Apply([]byte(storeJSON))
	if err != nil {
		return err
	}
	s.logStr("modifiedStoreJson=" + string(modifiedStoreJSON))
	err = s.putAsJSON(key, string(modifiedStoreJSON))
	if err != nil {
		return err
	}
	return nil
}

func (s *jsonStorage) lookUp(key, jsonPath string) (interface{}, error) {
	s.logStr("jsonpath=" + jsonPath)
	value, err := s.get(key)
	if err != nil {
		return "", err
	}
	res, err := jsonpath.JsonPathLookup(value, jsonPath)
	if err != nil {
		return "", err
	}
	return res, nil
}

func (s *jsonStorage) lookUpJSON(key, jsonPath string) (string, error) {
	res, err := s.lookUp(key, jsonPath)
	if err != nil {
		return "", err
	}
	resJSON, err := json.Marshal(res)
	if err != nil {
		return "", err
	}
	s.logStr(fmt.Sprintf("jsonPath jsonresult='%s'", resJSON))
	return string(resJSON), nil
}

func (s *jsonStorage) logStr(message string) {
	if s.log {
		log.Print(message)
	}
}

func (s *jsonStorage) templateFuncMap() template.FuncMap {
	return template.FuncMap{
		"kvStoreGet": func(key string) interface{} {
			val, err := s.store.GetVal(key)
			if err != nil {
				return ""
			}
			return val
		},
		"kvStorePut": func(key string, val interface{}) string {
			if err := s.store.PutVal(key, val); err != nil {
				return err.Error()
			}
			return ""
		},
		"kvStoreAdd": func(key, path, value string) string {
			if err := s.patchAdd(key, path, value); err != nil {
				return err.Error()
			}
			return ""
		},
		"kvStoreRemove": func(key, path string) string {
			if err := s.patchRemove(key, path); err != nil {
				return err.Error()
			}
			return ""
		},
		"kvStoreJsonPath": func(key, jsonPath string) interface{} {
			val, err := s.lookUp(key, jsonPath)
			if err != nil {
				return err.Error()
			}
			return val
		},
	}
}
