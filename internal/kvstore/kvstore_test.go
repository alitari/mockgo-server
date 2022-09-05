package kvstore

import (
	"testing"

	"github.com/alitari/mockgo-server/internal/utils"
	"github.com/stretchr/testify/assert"
)

func TestKVStore_GetPutAll(t *testing.T) {
	CreateTheStore()
	all := TheKVStore.GetAll()
	assert.Empty(t, all)
	storeVal := map[string]interface{}{"store1": "storekey11", "storekey12": "storeval2"}
	TheKVStore.PutAll(storeVal)
	assert.Equal(t, storeVal, TheKVStore.GetAll())
}

func TestKVStore_GetPutAllAsJson(t *testing.T) {
	storeJson := `{"store1":{"storekey1":"storeval1"},"store2":{"storekey2":"storeval2"}}`
	CreateTheStore()
	allJson, err := TheKVStore.GetAllJson()
	assert.NoError(t, err)
	assert.Equal(t, "{}", allJson)
	err = TheKVStore.PutAllJson(storeJson)
	assert.NoError(t, err)
	actualStoreJson, err := TheKVStore.GetAllJson()
	assert.NoError(t, err)
	assert.Equal(t, storeJson, actualStoreJson)
}

func TestKVStore_PutAllGetAsAllJson(t *testing.T) {
	CreateTheStore()
	all := TheKVStore.GetAll()
	assert.Empty(t, all)
	storeVal := map[string]interface{}{"store1": "storekey11"}
	TheKVStore.PutAll(storeVal)
	actualStoreJson, err := TheKVStore.GetAllJson()
	assert.NoError(t, err)
	assert.Equal(t, `{"store1":"storekey11"}`, actualStoreJson)
}

func TestKVStore_PutAllAsJsonGetAll(t *testing.T) {
	storeVal := map[string]interface{}{"store1": "storeval1", "storekey12": "storeval2"}
	CreateTheStore()
	allJson, err := TheKVStore.GetAllJson()
	assert.NoError(t, err)
	assert.Equal(t, "{}", allJson)
	err = TheKVStore.PutAllJson(`{"store1": "storeval1", "storekey12": "storeval2"}`)
	assert.NoError(t, err)
	assert.Equal(t, storeVal, TheKVStore.GetAll())
}

func TestKVStore_GetPut(t *testing.T) {
	CreateTheStore()
	key := utils.RandString(10)
	store := TheKVStore.Get(key)
	assert.Empty(t, store)
	storeVal := map[string]interface{}{"store1": "storeval11", "store2": "storeval11"}
	TheKVStore.Put(key, &storeVal)
	assert.Equal(t, &storeVal, TheKVStore.Get(key))
}

func TestKVStore_PutGetJson(t *testing.T) {
	CreateTheStore()
	key := utils.RandString(10)
	store := TheKVStore.Get(key)
	assert.Empty(t, store)
	storeVal := map[string]interface{}{"store1": "storeval11", "store2": "storeval11"}
	TheKVStore.Put(key, &storeVal)
	storeJson, err := TheKVStore.GetAsJson(key)
	assert.NoError(t, err)
	assert.Equal(t, `{"store1":"storeval11","store2":"storeval11"}`, storeJson)
}

func TestKVStore_PutJsonGet(t *testing.T) {
	CreateTheStore()
	key := utils.RandString(10)
	store := TheKVStore.Get(key)
	assert.Empty(t, store)
	storeVal := map[string]interface{}{"store1": "storeval11", "store2": "storeval11"}
	err := TheKVStore.PutAsJson(key, `{"store1":"storeval11","store2":"storeval11"}`)
	assert.NoError(t, err)
	assert.Equal(t, storeVal, TheKVStore.Get(key))

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
	CreateTheStoreWithLog()
	key := "bookstore"
	store := TheKVStore.Get(key)
	assert.Empty(t, store)
	err := TheKVStore.PutAsJson(key, bookstore)
	assert.NoError(t, err)

	res, err := TheKVStore.LookUp(key, `$.expensive`)
	assert.NoError(t, err)
	assert.Equal(t, float64(10), res)

	res, err = TheKVStore.LookUp(key, `$.store.book[0].price`)
	assert.NoError(t, err)
	assert.Equal(t, float64(8.95), res)

	res, err = TheKVStore.LookUp(key, `$.store.book[-1].isbn`)
	assert.NoError(t, err)
	assert.Equal(t, "0-395-19395-8", res)

	res, err = TheKVStore.LookUp(key, `$.store.book[0,1].price`)
	assert.NoError(t, err)
	assert.Equal(t, []interface{}{float64(8.95), float64(12.99)}, res)

	res, err = TheKVStore.LookUp(key, `$.store.book[0:2].price`)
	assert.NoError(t, err)
	assert.Equal(t, []interface{}{float64(8.95), float64(12.99), float64(8.99)}, res)

	res, err = TheKVStore.LookUp(key, `$.store.book[?(@.isbn)].price`)
	assert.NoError(t, err)
	assert.Equal(t, []interface{}{float64(8.99), float64(22.99)}, res)

	res, err = TheKVStore.LookUp(key, `$.store.book[?(@.price > 10)].title`)
	assert.NoError(t, err)
	assert.Equal(t, []interface{}{"Sword of Honour", "The Lord of the Rings"}, res)

	jsonPath := `$.store.book[?(@.author =~ /(?i).*REES/)].author`
	res, err = TheKVStore.LookUp(key, jsonPath)
	assert.NoError(t, err)
	assert.Equal(t, []interface{}{"Nigel Rees"}, res)
	resJson, err := TheKVStore.LookUpJson(key, jsonPath)
	assert.NoError(t, err)
	assert.Equal(t, `["Nigel Rees"]`, resJson)

}

func TestKVStore_PatchAdd(t *testing.T) {
	CreateTheStoreWithLog()
	key := utils.RandString(10)
	store := TheKVStore.Get(key)
	assert.Empty(t, store)
	storeVal := map[string]interface{}{"store1": "val1", "store2": "val2"}
	TheKVStore.Put(key, &storeVal)

	err := TheKVStore.PatchAdd(key, "/store1", `"val1patched"`)
	assert.NoError(t, err)
	storeJson, err := TheKVStore.GetAsJson(key)
	assert.NoError(t, err)
	assert.Equal(t, `{"store1":"val1patched","store2":"val2"}`, storeJson)

	err = TheKVStore.PatchAdd(key, "/store3", `"val3patched"`)
	assert.NoError(t, err)
	storeJson, err = TheKVStore.GetAsJson(key)
	assert.NoError(t, err)
	assert.Equal(t, `{"store1":"val1patched","store2":"val2","store3":"val3patched"}`, storeJson)

	err = TheKVStore.PatchAdd(key, "/store4", `["val41"]`)
	assert.NoError(t, err)
	storeJson, err = TheKVStore.GetAsJson(key)
	assert.NoError(t, err)
	assert.Equal(t, `{"store1":"val1patched","store2":"val2","store3":"val3patched","store4":["val41"]}`, storeJson)

	err = TheKVStore.PatchAdd(key, "/store4/1", `"val42"`)
	assert.NoError(t, err)
	storeJson, err = TheKVStore.GetAsJson(key)
	assert.NoError(t, err)
	assert.Equal(t, `{"store1":"val1patched","store2":"val2","store3":"val3patched","store4":["val41","val42"]}`, storeJson)

	err = TheKVStore.PatchAdd(key, "/store4/2", `{ "key43" : "val43" }`)
	assert.NoError(t, err)
	storeJson, err = TheKVStore.GetAsJson(key)
	assert.NoError(t, err)
	assert.Equal(t, `{"store1":"val1patched","store2":"val2","store3":"val3patched","store4":["val41","val42",{"key43":"val43"}]}`, storeJson)

	err = TheKVStore.PatchAdd(key, "/store4/-", `"key44"`)
	assert.NoError(t, err)
	storeJson, err = TheKVStore.GetAsJson(key)
	assert.NoError(t, err)
	assert.Equal(t, `{"store1":"val1patched","store2":"val2","store3":"val3patched","store4":["val41","val42",{"key43":"val43"},"key44"]}`, storeJson)
}

func TestKVStore_PatchRemove(t *testing.T) {
	CreateTheStoreWithLog()
	key := utils.RandString(10)
	storeVal := map[string]interface{}{"store1": "val1", "store2": "val2"}
	TheKVStore.Put(key, &storeVal)
	storeJson, err := TheKVStore.GetAsJson(key)
	assert.NoError(t, err)
	assert.Equal(t, `{"store1":"val1","store2":"val2"}`, storeJson)

	err = TheKVStore.PatchRemove(key, "/store1")
	assert.NoError(t, err)
	storeJson, err = TheKVStore.GetAsJson(key)
	assert.NoError(t, err)
	assert.Equal(t, `{"store2":"val2"}`, storeJson)
}

func TestKVStore_PatchReplace(t *testing.T) {
	CreateTheStoreWithLog()
	key := utils.RandString(10)
	storeVal := map[string]interface{}{"store1": "val1", "store2": "val2"}
	TheKVStore.Put(key, &storeVal)
	storeJson, err := TheKVStore.GetAsJson(key)
	assert.NoError(t, err)
	assert.Equal(t, `{"store1":"val1","store2":"val2"}`, storeJson)

	err = TheKVStore.PatchReplace(key, "/store1", `"replacedvalue"`)
	assert.NoError(t, err)
	storeJson, err = TheKVStore.GetAsJson(key)
	assert.NoError(t, err)
	assert.Equal(t, `{"store1":"replacedvalue","store2":"val2"}`, storeJson)
}
