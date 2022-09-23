package kvstore

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	jsonpatch "github.com/evanphx/json-patch"
	jsonpath "github.com/oliveagle/jsonpath"
)

type KVStore interface {
	GetVal(key string) interface{}
	PutVal(key string, storeVal interface{})
	GetAll() map[string]interface{}
	PutAll(content map[string]interface{})
}



type KVStoreJSON struct {
	log   bool
	store KVStore
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

type KVStoreInMemory struct {
	store map[string]interface{}
}

func (s *KVStoreInMemory) GetVal(key string) interface{} {
	return s.store[key]
}

func (s *KVStoreInMemory) PutVal(key string, storeVal interface{}) {
	s.store[key] = storeVal
}

func (s *KVStoreInMemory) GetAll() map[string]interface{} {
	return s.store
}

func (s *KVStoreInMemory) PutAll(content map[string]interface{}) {
	s.store = content
}

func NewKVStoreInMemory() KVStoreInMemory {
	return KVStoreInMemory{store: map[string]interface{}{}}
}

func NewKVStoreJSON(kvStore KVStore, log bool) *KVStoreJSON {
	return &KVStoreJSON{store: kvStore, log: log}
}

func (s *KVStoreJSON) PutAsJson(key, jsonStr string) error {
	var storeVal interface{}
	err := json.Unmarshal([]byte(jsonStr), &storeVal)
	if err != nil {
		return err
	}
	s.Put(key, storeVal)
	return nil
}

func (s *KVStoreJSON) Get(key string) interface{} {
	return s.store.GetVal(key)
}

func (s *KVStoreJSON) Put(key string, storeVal interface{}) {
	s.store.PutVal(key, storeVal)
}

func (s *KVStoreJSON) GetAsJson(key string) (string, error) {
	storeVal := s.Get(key)
	storeJson, err := json.Marshal(storeVal)
	if err != nil {
		return "", err
	}
	return string(storeJson), nil
}

func (s *KVStoreJSON) GetAll() map[string]interface{} {
	return s.store.GetAll()
}

func (s *KVStoreJSON) PutAll(content map[string]interface{}) {
	s.store.PutAll(content)
}

func (s *KVStoreJSON) PutAllJson(allStoreJson string) error {
	allStoreVal := &map[string]interface{}{}
	err := json.Unmarshal([]byte(allStoreJson), allStoreVal)
	if err != nil {
		return err
	}
	s.PutAll(*allStoreVal)
	return nil
}

func (s *KVStoreJSON) GetAllJson() (string, error) {
	storeVal := s.GetAll()
	storeJson, err := json.Marshal(storeVal)
	if err != nil {
		return "", err
	}
	return string(storeJson), nil
}

func (s *KVStoreJSON) PatchAdd(key, path, value string) error {
	if !strings.HasPrefix(value, "{") && !strings.HasPrefix(value, "[") {
		value = "\"" + value + "\""
	}
	return s.patch(key, fmt.Sprintf(`[{"op":"%s","path":"%s","value": %s}]`, Add.String(), path, value))
}

func (s *KVStoreJSON) PatchRemove(key, path string) error {
	return s.patch(key, fmt.Sprintf(`[{"op":"%s","path":"%s"}]`, Remove.String(), path))
}

func (s *KVStoreJSON) PatchReplace(key, path, value string) error {
	return s.patch(key, fmt.Sprintf(`[{"op":"%s","path":"%s","value": %s}]`, Replace.String(), path, value))
}

func (s *KVStoreJSON) patch(key, patchJson string) error {
	s.logStr("patchJson=" + patchJson)
	patch, err := jsonpatch.DecodePatch([]byte(patchJson))
	if err != nil {
		return err
	}
	storeJson, err := s.GetAsJson(key)
	if err != nil {
		return err
	}
	s.logStr("storeJson=" + storeJson)
	modifiedStoreJson, err := patch.Apply([]byte(storeJson))
	if err != nil {
		return err
	}
	s.logStr("modifiedStoreJson=" + string(modifiedStoreJson))
	err = s.PutAsJson(key, string(modifiedStoreJson))
	if err != nil {
		return err
	}
	return nil
}

func (s *KVStoreJSON) LookUp(key, jsonPath string) (interface{}, error) {
	s.logStr("jsonpath=" + jsonPath)
	res, err := jsonpath.JsonPathLookup(s.Get(key), jsonPath)
	if err != nil {
		return "", err
	}
	return res, nil
}

func (s *KVStoreJSON) LookUpJson(key, jsonPath string) (string, error) {
	res, err := s.LookUp(key, jsonPath)
	if err != nil {
		return "", err
	}
	resJson, err := json.Marshal(res)
	if err != nil {
		return "", err
	}
	s.logStr(fmt.Sprintf("jsonPath jsonresult='%s'", resJson))
	return string(resJson), nil
}

func (s *KVStoreJSON) logStr(message string) {
	if s.log {
		log.Print(message)
	}
}
