package kvstore

type KVStoreInMemory struct {
	store map[string]interface{}
}

func (s *KVStoreInMemory) GetVal(key string) (interface{}, error) {
	return s.store[key],nil
}

func (s *KVStoreInMemory) PutVal(key string, storeVal interface{}) error {
	s.store[key] = storeVal
	return nil
}

func (s *KVStoreInMemory) GetAll() (map[string]interface{}, error) {
	return s.store,nil
}

func (s *KVStoreInMemory) PutAll(content map[string]interface{}) error{
	s.store = content
	return nil
}

func NewKVStoreInMemory() KVStore {
	return &KVStoreInMemory{store: map[string]interface{}{}}
}

