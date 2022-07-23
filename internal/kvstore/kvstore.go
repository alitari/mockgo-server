package kvstore

import "encoding/json"

type KVStore struct {
	store map[string]*map[string]interface{}
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

func (s *KVStore) Get(key string) (interface{}, error) {
	return s.store[key], nil
}
