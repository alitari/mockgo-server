package kvstore

type InmemoryKVStore struct {
	store map[string]interface{}
}

func (s *InmemoryKVStore) GetVal(key string) (interface{}, error) {
	return s.store[key], nil
}

func (s *InmemoryKVStore) PutVal(key string, storeVal interface{}) error {
	s.store[key] = storeVal
	return nil
}

func (s *InmemoryKVStore) GetAll() (map[string]interface{}, error) {
	return s.store, nil
}

func (s *InmemoryKVStore) PutAll(content map[string]interface{}) error {
	s.store = content
	return nil
}

func NewInmemoryKVStore() *InmemoryKVStore {
	return &InmemoryKVStore{store: map[string]interface{}{}}
}
