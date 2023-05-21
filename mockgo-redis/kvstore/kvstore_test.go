package kvstore

import (
	"fmt"
	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redismock/v9"
	"github.com/stretchr/testify/assert"
	"testing"
)

var storage *RedisStorage
var clientmock redismock.ClientMock

func createRedismockStorage() {
	client, mock := redismock.NewClientMock()
	clientmock = mock
	storage = &RedisStorage{
		client: client,
	}
}
func createMiniRedisStorage() {
	miniredis := miniredis.NewMiniRedis()
	err := miniredis.Start()
	if err != nil {
		panic(err)
	}
	storage, err = NewRedisStorage(miniredis.Addr(), "", 0)
	if err != nil {
		panic(err)
	}
}

func TestRedisStorage_Get(t *testing.T) {
	createRedismockStorage()
	store := "mystore"
	key := "mystorekey"
	val := "{ \"key1\": [\"val11\", \"val22\"], \"key2\": [\"val21\"] }"
	clientmock.ExpectHGet(store, key).SetVal(val)
	getVal, err := storage.Get(store, key)
	assert.NoError(t, err)
	assert.EqualValues(t, map[string]interface{}{"key1": []interface{}{"val11", "val22"}, "key2": []interface{}{"val21"}}, getVal)
}

func TestRedisStorage_Get_error(t *testing.T) {
	createRedismockStorage()
	store := "mystore"
	key := "mystorekey"
	clientmock.ExpectHGet(store, key).SetErr(fmt.Errorf("my error"))
	getVal, err := storage.Get(store, key)
	assert.EqualError(t, err, "my error")
	assert.Nil(t, getVal)
}

func TestRedisStorage_GetAll(t *testing.T) {
	createRedismockStorage()
	store := "mystore"
	val := "{ \"key1\": [\"val11\", \"val22\"], \"key2\": [\"val21\"] }"
	clientmock.ExpectHGetAll(store).SetVal(map[string]string{"mystorekey": val})
	getAll, err := storage.GetAll(store)
	assert.NoError(t, err)
	assert.EqualValues(t, map[string]interface{}{"mystorekey": map[string]interface{}{"key1": []interface{}{"val11", "val22"}, "key2": []interface{}{"val21"}}}, getAll)
}

func TestRedisStorage_GetAll_error(t *testing.T) {
	createRedismockStorage()
	store := "mystore"
	clientmock.ExpectHGetAll(store).SetErr(fmt.Errorf("my error"))
	getAll, err := storage.GetAll(store)
	assert.EqualError(t, err, "my error")
	assert.Nil(t, getAll)
}

func TestRedisStorage_Put(t *testing.T) {
	createRedismockStorage()
	store := "mystore"
	key := "mystorekey"
	val := map[string]interface{}{"key1": []interface{}{"val11", "val22"}, "key2": []interface{}{"val21"}}
	clientmock.ExpectHSet(store, key, "{\"key1\":[\"val11\",\"val22\"],\"key2\":[\"val21\"]}").SetVal(1)
	err := storage.Put(store, key, val)
	assert.NoError(t, err)
}

func TestRedisStorage_Put_error(t *testing.T) {
	createRedismockStorage()
	store := "mystore"
	key := "mystorekey"
	val := map[string]interface{}{"key1": []interface{}{"val11", "val22"}, "key2": []interface{}{"val21"}}
	clientmock.ExpectHSet(store, key, "{\"key1\":[\"val11\",\"val22\"],\"key2\":[\"val21\"]}").
		SetErr(fmt.Errorf("my error"))
	err := storage.Put(store, key, val)
	assert.EqualError(t, err, "my error")
}

func TestRedisStorage_Put_Get(t *testing.T) {
	createMiniRedisStorage()
	store := "mystore"
	key := "mystorekey"
	val := map[string]interface{}{"key1": []interface{}{"val11", "val22"}, "key2": []interface{}{"val21"}}
	err := storage.Put(store, key, val)
	assert.NoError(t, err)
	getVal, err := storage.Get(store, key)
	assert.NoError(t, err)
	assert.EqualValues(t, val, getVal)
}
