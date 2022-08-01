package kvstore

import "encoding/json"

type KVStore struct {
	store map[string]*map[string]interface{}
}

var TheKVStore *KVStore

func CreateTheStore() *KVStore {
	TheKVStore = NewStore()
	return TheKVStore
}

func NewStore() *KVStore {
	return &KVStore{store: map[string]*map[string]interface{}{}}
}

func (s *KVStore) Put(key string, value string) error {
	jsonvalue := &map[string]interface{}{}
	err := json.Unmarshal([]byte(value), jsonvalue)
	if err != nil {
		return err
	}
	s.store[key] = jsonvalue
	return nil
}

func (s *KVStore) Get(key string) (*map[string]interface{}, error) {
	return s.store[key], nil
}

func (s *KVStore) GetAll() map[string]*map[string]interface{} {
	return s.store
}

func (s *KVStore) PutAll(content map[string]*map[string]interface{}) {
	s.store = content
}

func (s *KVStore) PutAllJson(content string) error {
	store := &map[string]*map[string]interface{}{}
	err := json.Unmarshal([]byte(content), store)
	if err != nil {
		return err
	}
	s.PutAll(*store)
	return nil
}
