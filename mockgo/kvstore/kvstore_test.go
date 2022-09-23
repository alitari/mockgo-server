package kvstore

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

func createInMemoryStore() *KVStoreJSON {
	kvstoreImpl := NewKVStoreInMemory()
	return NewKVStoreJSON(&kvstoreImpl, true)
}

func TestKVStore_GetPutAll(t *testing.T) {
	kvstore := createInMemoryStore()
	all := kvstore.GetAll()
	assert.Empty(t, all)
	storeVal := map[string]interface{}{"store1": "storekey11", "storekey12": "storeval2"}
	kvstore.PutAll(storeVal)
	assert.Equal(t, storeVal, kvstore.GetAll())
}

func TestKVStore_GetPutAllAsJson(t *testing.T) {
	storeJson := `{"store1":{"storekey1":"storeval1"},"store2":{"storekey2":"storeval2"}}`
	kvstore := createInMemoryStore()
	allJson, err := kvstore.GetAllJson()
	assert.NoError(t, err)
	assert.Equal(t, "{}", allJson)
	err = kvstore.PutAllJson(storeJson)
	assert.NoError(t, err)
	actualStoreJson, err := kvstore.GetAllJson()
	assert.NoError(t, err)
	assert.Equal(t, storeJson, actualStoreJson)
}

func TestKVStore_PutAllGetAsAllJson(t *testing.T) {
	kvstore := createInMemoryStore()
	all := kvstore.GetAll()
	assert.Empty(t, all)
	storeVal := map[string]interface{}{"store1": "storekey11"}
	kvstore.PutAll(storeVal)
	actualStoreJson, err := kvstore.GetAllJson()
	assert.NoError(t, err)
	assert.Equal(t, `{"store1":"storekey11"}`, actualStoreJson)
}

func TestKVStore_PutAllAsJsonGetAll(t *testing.T) {
	storeVal := map[string]interface{}{"store1": "storeval1", "storekey12": "storeval2"}
	kvstore := createInMemoryStore()
	allJson, err := kvstore.GetAllJson()
	assert.NoError(t, err)
	assert.Equal(t, "{}", allJson)
	err = kvstore.PutAllJson(`{"store1": "storeval1", "storekey12": "storeval2"}`)
	assert.NoError(t, err)
	assert.Equal(t, storeVal, kvstore.GetAll())
}

func TestKVStore_GetPut(t *testing.T) {
	kvstore := createInMemoryStore()
	key := randString(10)
	store := kvstore.Get(key)
	assert.Empty(t, store)
	storeVal := map[string]interface{}{"store1": "storeval11", "store2": "storeval11"}
	kvstore.Put(key, &storeVal)
	assert.Equal(t, &storeVal, kvstore.Get(key))
}

func TestKVStore_PutGetJson(t *testing.T) {
	kvstore := createInMemoryStore()
	key := randString(10)
	store := kvstore.Get(key)
	assert.Empty(t, store)
	storeVal := map[string]interface{}{"store1": "storeval11", "store2": "storeval11"}
	kvstore.Put(key, &storeVal)
	storeJson, err := kvstore.GetAsJson(key)
	assert.NoError(t, err)
	assert.Equal(t, `{"store1":"storeval11","store2":"storeval11"}`, storeJson)
}

func TestKVStore_PutJsonGet(t *testing.T) {
	kvstore := createInMemoryStore()
	key := randString(10)
	store := kvstore.Get(key)
	assert.Empty(t, store)
	storeVal := map[string]interface{}{"store1": "storeval11", "store2": "storeval11"}
	err := kvstore.PutAsJson(key, `{"store1":"storeval11","store2":"storeval11"}`)
	assert.NoError(t, err)
	assert.Equal(t, storeVal, kvstore.Get(key))
}

const bookstore = `
{
    "store": {
        "book": [
            {
                "category": "reference",
                "author": "Nigel Rees",
                "title": "Sayings of the Century",
                "price": 8.95
            },
            {
                "category": "fiction",
                "author": "Evelyn Waugh",
                "title": "Sword of Honour",
                "price": 12.99
            },
            {
                "category": "fiction",
                "author": "Herman Melville",
                "title": "Moby Dick",
                "isbn": "0-553-21311-3",
                "price": 8.99
            },
            {
                "category": "fiction",
                "author": "J. R. R. Tolkien",
                "title": "The Lord of the Rings",
                "isbn": "0-395-19395-8",
                "price": 22.99
            }
        ],
        "bicycle": {
            "color": "red",
            "price": 19.95
        }
    },
    "expensive": 10
}
`

func TestKVStore_Lookup(t *testing.T) {
	kvstore := createInMemoryStore()
	key := "bookstore"
	store := kvstore.Get(key)
	assert.Empty(t, store)
	err := kvstore.PutAsJson(key, bookstore)
	assert.NoError(t, err)

	res, err := kvstore.LookUp(key, `$.expensive`)
	assert.NoError(t, err)
	assert.Equal(t, float64(10), res)

	res, err = kvstore.LookUp(key, `$.store.book[0].price`)
	assert.NoError(t, err)
	assert.Equal(t, float64(8.95), res)

	res, err = kvstore.LookUp(key, `$.store.book[-1].isbn`)
	assert.NoError(t, err)
	assert.Equal(t, "0-395-19395-8", res)

	res, err = kvstore.LookUp(key, `$.store.book[0,1].price`)
	assert.NoError(t, err)
	assert.Equal(t, []interface{}{float64(8.95), float64(12.99)}, res)

	res, err = kvstore.LookUp(key, `$.store.book[0:2].price`)
	assert.NoError(t, err)
	assert.Equal(t, []interface{}{float64(8.95), float64(12.99), float64(8.99)}, res)

	res, err = kvstore.LookUp(key, `$.store.book[?(@.isbn)].price`)
	assert.NoError(t, err)
	assert.Equal(t, []interface{}{float64(8.99), float64(22.99)}, res)

	res, err = kvstore.LookUp(key, `$.store.book[?(@.price > 10)].title`)
	assert.NoError(t, err)
	assert.Equal(t, []interface{}{"Sword of Honour", "The Lord of the Rings"}, res)

	jsonPath := `$.store.book[?(@.author =~ /(?i).*REES/)].author`
	res, err = kvstore.LookUp(key, jsonPath)
	assert.NoError(t, err)
	assert.Equal(t, []interface{}{"Nigel Rees"}, res)
	resJson, err := kvstore.LookUpJson(key, jsonPath)
	assert.NoError(t, err)
	assert.Equal(t, `["Nigel Rees"]`, resJson)

}

func TestKVStore_PatchAdd(t *testing.T) {
	kvstore := createInMemoryStore()
	key := randString(10)
	store := kvstore.Get(key)
	assert.Empty(t, store)
	storeVal := map[string]interface{}{"store1": "val1", "store2": "val2"}
	kvstore.Put(key, &storeVal)

	err := kvstore.PatchAdd(key, "/store1", "val1patched")
	assert.NoError(t, err)
	storeJson, err := kvstore.GetAsJson(key)
	assert.NoError(t, err)
	assert.Equal(t, `{"store1":"val1patched","store2":"val2"}`, storeJson)

	err = kvstore.PatchAdd(key, "/store3", "val3patched")
	assert.NoError(t, err)
	storeJson, err = kvstore.GetAsJson(key)
	assert.NoError(t, err)
	assert.Equal(t, `{"store1":"val1patched","store2":"val2","store3":"val3patched"}`, storeJson)

	err = kvstore.PatchAdd(key, "/store4", `["val41"]`)
	assert.NoError(t, err)
	storeJson, err = kvstore.GetAsJson(key)
	assert.NoError(t, err)
	assert.Equal(t, `{"store1":"val1patched","store2":"val2","store3":"val3patched","store4":["val41"]}`, storeJson)

	err = kvstore.PatchAdd(key, "/store4/1", "val42")
	assert.NoError(t, err)
	storeJson, err = kvstore.GetAsJson(key)
	assert.NoError(t, err)
	assert.Equal(t, `{"store1":"val1patched","store2":"val2","store3":"val3patched","store4":["val41","val42"]}`, storeJson)

	err = kvstore.PatchAdd(key, "/store4/2", `{ "key43" : "val43" }`)
	assert.NoError(t, err)
	storeJson, err = kvstore.GetAsJson(key)
	assert.NoError(t, err)
	assert.Equal(t, `{"store1":"val1patched","store2":"val2","store3":"val3patched","store4":["val41","val42",{"key43":"val43"}]}`, storeJson)

	err = kvstore.PatchAdd(key, "/store4/-", "key44")
	assert.NoError(t, err)
	storeJson, err = kvstore.GetAsJson(key)
	assert.NoError(t, err)
	assert.Equal(t, `{"store1":"val1patched","store2":"val2","store3":"val3patched","store4":["val41","val42",{"key43":"val43"},"key44"]}`, storeJson)
}

func TestKVStore_PatchRemove(t *testing.T) {
	kvstore := createInMemoryStore()
	key := randString(10)
	storeVal := map[string]interface{}{"store1": "val1", "store2": "val2"}
	kvstore.Put(key, &storeVal)
	storeJson, err := kvstore.GetAsJson(key)
	assert.NoError(t, err)
	assert.Equal(t, `{"store1":"val1","store2":"val2"}`, storeJson)

	err = kvstore.PatchRemove(key, "/store1")
	assert.NoError(t, err)
	storeJson, err = kvstore.GetAsJson(key)
	assert.NoError(t, err)
	assert.Equal(t, `{"store2":"val2"}`, storeJson)
}

func TestKVStore_PatchReplace(t *testing.T) {
	kvstore := createInMemoryStore()
	key := randString(10)
	storeVal := map[string]interface{}{"store1": "val1", "store2": "val2"}
	kvstore.Put(key, &storeVal)
	storeJson, err := kvstore.GetAsJson(key)
	assert.NoError(t, err)
	assert.Equal(t, `{"store1":"val1","store2":"val2"}`, storeJson)

	err = kvstore.PatchReplace(key, "/store1", `"replacedvalue"`)
	assert.NoError(t, err)
	storeJson, err = kvstore.GetAsJson(key)
	assert.NoError(t, err)
	assert.Equal(t, `{"store1":"replacedvalue","store2":"val2"}`, storeJson)
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func randString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Int63()%int64(len(letterBytes))]
	}
	return string(b)
}