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
	storeVal := map[string]*map[string]interface{}{"store1": {"storekey11": "storeval1"}, "storekey12": {"storekey12": "storeval2"}}
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
	storeVal := map[string]*map[string]interface{}{"store1": {"storekey11": "storeval1"}, "storekey12": {"storekey12": "storeval2"}}
	TheKVStore.PutAll(storeVal)
	actualStoreJson, err := TheKVStore.GetAllJson()
	assert.NoError(t, err)
	assert.Equal(t, `{"store1":{"storekey11":"storeval1"},"storekey12":{"storekey12":"storeval2"}}`, actualStoreJson)
}

func TestKVStore_PutAllAsJsonGetAll(t *testing.T) {
	storeVal := map[string]*map[string]interface{}{"store1": {"storekey11": "storeval1"}, "storekey12": {"storekey12": "storeval2"}}
	CreateTheStore()
	allJson, err := TheKVStore.GetAllJson()
	assert.NoError(t, err)
	assert.Equal(t, "{}", allJson)
	err = TheKVStore.PutAllJson(`{"store1":{"storekey11":"storeval1"},"storekey12":{"storekey12":"storeval2"}}`)
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
	assert.Equal(t, storeVal, *TheKVStore.Get(key))
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
	assert.Equal(t, &storeVal, TheKVStore.Get(key))

}

func TestKVStore_PatchAdd(t *testing.T) {
	CreateTheStoreWithLog()
	key := utils.RandString(10)
	store := TheKVStore.Get(key)
	assert.Empty(t, store)
	storeVal := map[string]interface{}{"store1": "val1", "store2": "val2"}
	TheKVStore.Put(key, &storeVal)

	err := TheKVStore.Patch(key, Add, "/store1", `"val1patched"`)
	assert.NoError(t, err)
	storeJson, err := TheKVStore.GetAsJson(key)
	assert.NoError(t, err)
	assert.Equal(t, `{"store1":"val1patched","store2":"val2"}`, storeJson)

	err = TheKVStore.Patch(key, Add, "/store3", `"val3patched"`)
	assert.NoError(t, err)
	storeJson, err = TheKVStore.GetAsJson(key)
	assert.NoError(t, err)
	assert.Equal(t, `{"store1":"val1patched","store2":"val2","store3":"val3patched"}`, storeJson)

	err = TheKVStore.Patch(key, Add, "/store4", `["val41"]`)
	assert.NoError(t, err)
	storeJson, err = TheKVStore.GetAsJson(key)
	assert.NoError(t, err)
	assert.Equal(t, `{"store1":"val1patched","store2":"val2","store3":"val3patched","store4":["val41"]}`, storeJson)

	err = TheKVStore.Patch(key, Add, "/store4/1", `"val42"`)
	assert.NoError(t, err)
	storeJson, err = TheKVStore.GetAsJson(key)
	assert.NoError(t, err)
	assert.Equal(t, `{"store1":"val1patched","store2":"val2","store3":"val3patched","store4":["val41","val42"]}`, storeJson)

	err = TheKVStore.Patch(key, Add, "/store4/2", `{ "key43" : "val43" }`)
	assert.NoError(t, err)
	storeJson, err = TheKVStore.GetAsJson(key)
	assert.NoError(t, err)
	assert.Equal(t, `{"store1":"val1patched","store2":"val2","store3":"val3patched","store4":["val41","val42",{"key43":"val43"}]}`, storeJson)

	err = TheKVStore.Patch(key, Add, "/store4/-", `"key44"`)
	assert.NoError(t, err)
	storeJson, err = TheKVStore.GetAsJson(key)
	assert.NoError(t, err)
	assert.Equal(t, `{"store1":"val1patched","store2":"val2","store3":"val3patched","store4":["val41","val42",{"key43":"val43"},"key44"]}`, storeJson)
}
