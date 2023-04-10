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

/*
NewKVStoreTemplateFuncMap creates a template.FuncMap for accessing a Storage
*/
func NewKVStoreTemplateFuncMap(storage Storage) template.FuncMap {
	return template.FuncMap{
		"kvStoreGet": func(store, key string) interface{} {
			val, err := storage.Get(store, key)
			if err != nil {
				return err.Error()
			}
			return val
		},
		"kvStoreGetAll": func(store string) interface{} {
			val, err := storage.GetAll(store)
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
