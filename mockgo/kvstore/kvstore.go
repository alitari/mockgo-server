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

type Storage interface {
	GetVal(key string) (interface{}, error)
	PutVal(key string, storeVal interface{}) error
}

type JSONStorage struct {
	log   bool
	store Storage
}

type PatchOp int64

const (
	Add PatchOp = iota
	Replace
	Remove
)

func (pop PatchOp) String() string {
	switch pop {
	case Add:
		return "add"
	case Replace:
		return "replace"
	case Remove:
		return "remove"
	}
	return "unknown"
}

func NewJSONStorage(kvStore Storage, log bool) *JSONStorage {
	return &JSONStorage{store: kvStore, log: log}
}

func (s *JSONStorage) PutAsJSON(key, jsonStr string) error {
	var storeVal interface{}
	err := json.Unmarshal([]byte(jsonStr), &storeVal)
	if err != nil {
		return err
	}
	err = s.Put(key, storeVal)
	return err
}

func (s *JSONStorage) Get(key string) (interface{}, error) {
	return s.store.GetVal(key)
}

func (s *JSONStorage) Put(key string, storeVal interface{}) error {
	return s.store.PutVal(key, storeVal)
}

func (s *JSONStorage) GetAsJSON(key string) (string, error) {
	storeVal, err := s.Get(key)
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

func (s *JSONStorage) PatchAdd(key, path, value string) error {
	if !strings.HasPrefix(value, "{") && !strings.HasPrefix(value, "[") {
		value = "\"" + value + "\""
	}
	return s.patch(key, fmt.Sprintf(`[{"op":"%s","path":"%s","value": %s}]`, Add.String(), path, value))
}

func (s *JSONStorage) PatchRemove(key, path string) error {
	return s.patch(key, fmt.Sprintf(`[{"op":"%s","path":"%s"}]`, Remove.String(), path))
}

func (s *JSONStorage) PatchReplace(key, path, value string) error {
	return s.patch(key, fmt.Sprintf(`[{"op":"%s","path":"%s","value": %s}]`, Replace.String(), path, value))
}

func (s *JSONStorage) patch(key, patchJSON string) error {
	s.logStr("patchJson=" + patchJSON)
	patch, err := jsonpatch.DecodePatch([]byte(patchJSON))
	if err != nil {
		return err
	}
	storeJSON, err := s.GetAsJSON(key)
	if err != nil {
		return err
	}

	s.logStr("storeJson=" + storeJSON)
	modifiedStoreJSON, err := patch.Apply([]byte(storeJSON))
	if err != nil {
		return err
	}
	s.logStr("modifiedStoreJson=" + string(modifiedStoreJSON))
	err = s.PutAsJSON(key, string(modifiedStoreJSON))
	if err != nil {
		return err
	}
	return nil
}

func (s *JSONStorage) LookUp(key, jsonPath string) (interface{}, error) {
	s.logStr("jsonpath=" + jsonPath)
	value, err := s.Get(key)
	if err != nil {
		return "", err
	}
	res, err := jsonpath.JsonPathLookup(value, jsonPath)
	if err != nil {
		return "", err
	}
	return res, nil
}

func (s *JSONStorage) LookUpJSON(key, jsonPath string) (string, error) {
	res, err := s.LookUp(key, jsonPath)
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

func (s *JSONStorage) logStr(message string) {
	if s.log {
		log.Print(message)
	}
}

func (s *JSONStorage) TemplateFuncMap() template.FuncMap {
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
			if err := s.PatchAdd(key, path, value); err != nil {
				return err.Error()
			}
			return ""
		},
		"kvStoreRemove": func(key, path string) string {
			if err := s.PatchRemove(key, path); err != nil {
				return err.Error()
			}
			return ""
		},
		"kvStoreJsonPath": func(key, jsonPath string) interface{} {
			val, err := s.LookUp(key, jsonPath)
			if err != nil {
				return err.Error()
			}
			return val
		},
	}
}
