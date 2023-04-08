package kvstore

import (
	"github.com/stretchr/testify/assert"
	"math/rand"
	"testing"
)

func TestStorage_PutGet(t *testing.T) {
	kvstore := NewInmemoryStorage()
	store := randString(10)
	key := randString(10)
	val, err := kvstore.Get(store, key)
	assert.NoError(t, err)
	assert.Empty(t, val)
	val = map[string]interface{}{randString(10): randString(10), randString(10): randString(10)}
	err = kvstore.Put(store, key, val)
	assert.NoError(t, err)
	getVal, err := kvstore.Get(store, key)
	assert.NoError(t, err)
	assert.Equal(t, val, getVal)
}

func TestStorage_PutGetAll(t *testing.T) {
	kvstore := NewInmemoryStorage()
	store := randString(10)
	key1 := randString(10)
	val, err := kvstore.Get(store, key1)
	assert.NoError(t, err)
	assert.Empty(t, val)
	key2 := randString(10)
	val, err = kvstore.Get(store, key2)
	assert.NoError(t, err)
	assert.Empty(t, val)
	val1 := map[string]interface{}{randString(10): randString(10), randString(10): randString(10)}
	err = kvstore.Put(store, key1, val1)
	assert.NoError(t, err)
	val2 := map[string]interface{}{randString(10): randString(10), randString(10): randString(10)}
	err = kvstore.Put(store, key2, val2)
	assert.NoError(t, err)
	getAll, err := kvstore.GetAll(store)
	assert.NoError(t, err)
	assert.Equal(t, map[string]interface{}{key1: val1, key2: val2}, getAll)
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func randString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Int63()%int64(len(letterBytes))]
	}
	return string(b)
}
