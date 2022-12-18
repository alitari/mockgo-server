package kvstore

/*
InmemoryStorage is a implementation of Storage using local in-memory storage
*/
type InmemoryStorage struct {
	store map[string]interface{}
}

/*
GetVal returns a value for a given key
*/
func (s *InmemoryStorage) GetVal(key string) (interface{}, error) {
	return s.store[key], nil
}

/*
PutVal stores a value under a given key
*/
func (s *InmemoryStorage) PutVal(key string, storeVal interface{}) error {
	s.store[key] = storeVal
	return nil
}

/*
NewInmemoryStorage creates a new instance of InmemoryStorage
*/
func NewInmemoryStorage() *InmemoryStorage {
	return &InmemoryStorage{store: map[string]interface{}{}}
}
