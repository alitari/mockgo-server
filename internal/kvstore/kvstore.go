package kvstore

import (
	"encoding/json"
	"fmt"
	"log"

	jsonpatch "github.com/evanphx/json-patch"
	jsonpath "github.com/oliveagle/jsonpath"
)

type KVStore struct {
	log   bool
	store map[string]interface{}
}

var TheKVStore *KVStore

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

func CreateTheStore() *KVStore {
	TheKVStore = NewStore()
	return TheKVStore
}

func CreateTheStoreWithLog() *KVStore {
	CreateTheStore()
	TheKVStore.log = true
	return TheKVStore
}

func NewStore() *KVStore {
	return &KVStore{store: map[string]interface{}{}}
}

func (s *KVStore) PutAsJson(key, jsonStr string) error {
	var storeVal interface{}
	err := json.Unmarshal([]byte(jsonStr), &storeVal)
	if err != nil {
		return err
	}
	s.Put(key, storeVal)
	return nil
}

func (s *KVStore) Get(key string) interface{} {
	return s.store[key]
}

func (s *KVStore) Put(key string, storeVal interface{}) {
	s.store[key] = storeVal
}

func (s *KVStore) GetAsJson(key string) (string, error) {
	storeVal := s.Get(key)
	storeJson, err := json.Marshal(storeVal)
	if err != nil {
		return "", err
	}
	return string(storeJson), nil
}

func (s *KVStore) GetAll() map[string]interface{} {
	return s.store
}

func (s *KVStore) PutAll(content map[string]interface{}) {
	s.store = content
}

func (s *KVStore) PutAllJson(allStoreJson string) error {
	allStoreVal := &map[string]interface{}{}
	err := json.Unmarshal([]byte(allStoreJson), allStoreVal)
	if err != nil {
		return err
	}
	s.PutAll(*allStoreVal)
	return nil
}

func (s *KVStore) GetAllJson() (string, error) {
	storeVal := s.GetAll()
	storeJson, err := json.Marshal(storeVal)
	if err != nil {
		return "", err
	}
	return string(storeJson), nil
}

func (s *KVStore) Patch(key string, op PatchOp, path, value string) error {
	patchJson := fmt.Sprintf(`[ {"op": "%s", "path": "%s", "value": %s}]`, op.String(), path, value)
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

func (s *KVStore) LookUp(key, jsonPath string) (interface{}, error) {
	s.logStr("jsonpath=" + jsonPath)
	res, err := jsonpath.JsonPathLookup(s.Get(key), jsonPath)
	if err != nil {
		return "", err
	}
	return res, nil
}

func (s *KVStore) LookUpJson(key, jsonPath string) (string, error) {
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

func (s *KVStore) logStr(message string) {
	if s.log {
		log.Print(message)
	}
}
