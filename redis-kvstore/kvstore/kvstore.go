package kvstore

import (
	"context"
	"encoding/json"
	"github.com/redis/go-redis/v9"
	"time"
)

// RedisStorage is a kvstore.Storage implementation using redis as backend.
type RedisStorage struct {
	client *redis.Client
}

// NewRedisStorage creates a new kvstore.Storage using redis as backend.
func NewRedisStorage(address, password string, db int) (*RedisStorage, error) {
	storage := &RedisStorage{
		client: redis.NewClient(&redis.Options{
			Addr:            address,
			Password:        password,
			DB:              db,
			MaxRetries:      30,
			MinRetryBackoff: 500 * time.Millisecond,
			MaxRetryBackoff: 2 * time.Second,
		}),
	}
	if err := storage.checkConnectivity(); err != nil {
		return nil, err
	}
	return storage, nil
}

func (r *RedisStorage) checkConnectivity() error {
	var ctx = context.Background()
	status := r.client.Ping(ctx)
	return status.Err()
}

// Put stores a value in the kvstore.
func (r *RedisStorage) Put(store, key string, val interface{}) error {
	var ctx = context.Background()
	mval, err := json.Marshal(val)
	if err != nil {
		return err
	}
	status := r.client.HSet(ctx, store, key, string(mval))
	return status.Err()
}

// Get returns a value from the kvstore.
func (r *RedisStorage) Get(store, key string) (interface{}, error) {
	var ctx = context.Background()
	hGet := r.client.HGet(ctx, store, key)
	if hGet.Err() != nil {
		return nil, hGet.Err()
	}
	var result interface{}
	err := json.Unmarshal([]byte(hGet.Val()), &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// GetAll returns all values from the kvstore.
func (r *RedisStorage) GetAll(store string) (map[string]interface{}, error) {
	var ctx = context.Background()
	status := r.client.HGetAll(ctx, store)
	if status.Err() != nil {
		return nil, status.Err()
	}
	result := make(map[string]interface{})
	for k, v := range status.Val() {
		var resultVal interface{}
		err := json.Unmarshal([]byte(v), &resultVal)
		if err != nil {
			return nil, err
		}
		result[k] = resultVal
	}
	return result, nil
}
