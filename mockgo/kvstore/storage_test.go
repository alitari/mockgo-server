package kvstore

import (
	"bytes"
	"fmt"
	"math/rand"
	"testing"
	"text/template"

	"github.com/stretchr/testify/assert"
)

func createInMemoryStorage() *jsonStorage {
	kvstoreImpl := NewInmemoryStorage()
	return newJSONStorage(kvstoreImpl, true)
}

func TestKVStore_GetPut(t *testing.T) {
	kvstore := createInMemoryStorage()
	key := randString(10)
	store, err := kvstore.get(key)
	assert.NoError(t, err)
	assert.Empty(t, store)
	storeVal := map[string]interface{}{"store1": "storeval11", "store2": "storeval11"}
	err = kvstore.put(key, &storeVal)
	assert.NoError(t, err)
	val, err := kvstore.get(key)
	assert.NoError(t, err)
	assert.Equal(t, &storeVal, val)
}

func TestKVStore_PutGetJson(t *testing.T) {
	kvstore := createInMemoryStorage()
	key := randString(10)
	store, err := kvstore.get(key)
	assert.NoError(t, err)
	assert.Empty(t, store)
	storeVal := map[string]interface{}{"store1": "storeval11", "store2": "storeval11"}
	err = kvstore.put(key, &storeVal)
	assert.NoError(t, err)
	storeJSON, err := kvstore.getAsJSON(key)
	assert.NoError(t, err)
	assert.Equal(t, `{"store1":"storeval11","store2":"storeval11"}`, storeJSON)
}

func TestKVStore_PutGetJsonError(t *testing.T) {
	kvstore := createInMemoryStorage()
	key := randString(10)
	storeVal := make(chan int)
	err := kvstore.put(key, storeVal)
	assert.NoError(t, err)
	_, err = kvstore.getAsJSON(key)
	assert.ErrorContains(t, err, "json: unsupported type: chan int")
}

func TestKVStore_PutJsonGet(t *testing.T) {
	kvstore := createInMemoryStorage()
	key := randString(10)
	store, err := kvstore.get(key)
	assert.NoError(t, err)
	assert.Empty(t, store)
	storeVal := map[string]interface{}{"store1": "storeval11", "store2": "storeval11"}
	err = kvstore.putAsJSON(key, `{"store1":"storeval11","store2":"storeval11"}`)
	assert.NoError(t, err)
	val, err := kvstore.get(key)
	assert.NoError(t, err)
	assert.Equal(t, storeVal, val)
}

func TestKVStore_PutJsonGet_Error(t *testing.T) {
	kvstore := createInMemoryStorage()
	key := randString(10)
	err := kvstore.putAsJSON(key, `{ invalid": json}`)
	assert.ErrorContains(t, err, "invalid character 'i' looking for beginning of object key string")
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
	kvstore := createInMemoryStorage()
	key := "bookstore"
	store, err := kvstore.get(key)
	assert.NoError(t, err)
	assert.Empty(t, store)
	err = kvstore.putAsJSON(key, bookstore)
	assert.NoError(t, err)

	res, err := kvstore.lookUp(key, `$.expensive`)
	assert.NoError(t, err)
	assert.Equal(t, float64(10), res)

	res, err = kvstore.lookUp(key, `$.store.book[0].price`)
	assert.NoError(t, err)
	assert.Equal(t, float64(8.95), res)

	res, err = kvstore.lookUp(key, `$.store.book[-1].isbn`)
	assert.NoError(t, err)
	assert.Equal(t, "0-395-19395-8", res)

	res, err = kvstore.lookUp(key, `$.store.book[0,1].price`)
	assert.NoError(t, err)
	assert.Equal(t, []interface{}{float64(8.95), float64(12.99)}, res)

	res, err = kvstore.lookUp(key, `$.store.book[0:2].price`)
	assert.NoError(t, err)
	assert.Equal(t, []interface{}{float64(8.95), float64(12.99), float64(8.99)}, res)

	res, err = kvstore.lookUp(key, `$.store.book[?(@.isbn)].price`)
	assert.NoError(t, err)
	assert.Equal(t, []interface{}{float64(8.99), float64(22.99)}, res)

	res, err = kvstore.lookUp(key, `$.store.book[?(@.price > 10)].title`)
	assert.NoError(t, err)
	assert.Equal(t, []interface{}{"Sword of Honour", "The Lord of the Rings"}, res)

	jsonPath := `$.store.book[?(@.author =~ /(?i).*REES/)].author`
	res, err = kvstore.lookUp(key, jsonPath)
	assert.NoError(t, err)
	assert.Equal(t, []interface{}{"Nigel Rees"}, res)
	resJSON, err := kvstore.lookUpJSON(key, jsonPath)
	assert.NoError(t, err)
	assert.Equal(t, `["Nigel Rees"]`, resJSON)

}

func TestKVStore_PatchAdd(t *testing.T) {
	kvstore := createInMemoryStorage()
	key := randString(10)
	store, err := kvstore.get(key)
	assert.NoError(t, err)
	assert.Empty(t, store)
	storeVal := map[string]interface{}{"store1": "val1", "store2": "val2"}
	kvstore.put(key, &storeVal)

	err = kvstore.patchAdd(key, "/store1", "val1patched")
	assert.NoError(t, err)
	storeJSON, err := kvstore.getAsJSON(key)
	assert.NoError(t, err)
	assert.Equal(t, `{"store1":"val1patched","store2":"val2"}`, storeJSON)

	err = kvstore.patchAdd(key, "/store3", "val3patched")
	assert.NoError(t, err)
	storeJSON, err = kvstore.getAsJSON(key)
	assert.NoError(t, err)
	assert.Equal(t, `{"store1":"val1patched","store2":"val2","store3":"val3patched"}`, storeJSON)

	err = kvstore.patchAdd(key, "/store4", `["val41"]`)
	assert.NoError(t, err)
	storeJSON, err = kvstore.getAsJSON(key)
	assert.NoError(t, err)
	assert.Equal(t, `{"store1":"val1patched","store2":"val2","store3":"val3patched","store4":["val41"]}`, storeJSON)

	err = kvstore.patchAdd(key, "/store4/1", "val42")
	assert.NoError(t, err)
	storeJSON, err = kvstore.getAsJSON(key)
	assert.NoError(t, err)
	assert.Equal(t, `{"store1":"val1patched","store2":"val2","store3":"val3patched","store4":["val41","val42"]}`, storeJSON)

	err = kvstore.patchAdd(key, "/store4/2", `{ "key43" : "val43" }`)
	assert.NoError(t, err)
	storeJSON, err = kvstore.getAsJSON(key)
	assert.NoError(t, err)
	assert.Equal(t, `{"store1":"val1patched","store2":"val2","store3":"val3patched","store4":["val41","val42",{"key43":"val43"}]}`, storeJSON)

	err = kvstore.patchAdd(key, "/store4/-", "key44")
	assert.NoError(t, err)
	storeJSON, err = kvstore.getAsJSON(key)
	assert.NoError(t, err)
	assert.Equal(t, `{"store1":"val1patched","store2":"val2","store3":"val3patched","store4":["val41","val42",{"key43":"val43"},"key44"]}`, storeJSON)

	err = kvstore.patchAdd(key, "invalidPath", "invalid")
	assert.ErrorContains(t, err, "add operation does not apply: doc is missing path: \"invalidPath\": missing value")

}

func TestKVStore_PatchRemove(t *testing.T) {
	kvstore := createInMemoryStorage()
	key := randString(10)
	storeVal := map[string]interface{}{"store1": "val1", "store2": "val2"}
	err := kvstore.put(key, &storeVal)
	assert.NoError(t, err)
	storeJSON, err := kvstore.getAsJSON(key)
	assert.NoError(t, err)
	assert.Equal(t, `{"store1":"val1","store2":"val2"}`, storeJSON)

	err = kvstore.patchRemove(key, "/store1")
	assert.NoError(t, err)
	storeJSON, err = kvstore.getAsJSON(key)
	assert.NoError(t, err)
	assert.Equal(t, `{"store2":"val2"}`, storeJSON)
}

func TestKVStore_PatchReplace(t *testing.T) {
	kvstore := createInMemoryStorage()
	key := randString(10)
	storeVal := map[string]interface{}{"store1": "val1", "store2": "val2"}
	err := kvstore.put(key, &storeVal)
	assert.NoError(t, err)
	storeJSON, err := kvstore.getAsJSON(key)
	assert.NoError(t, err)
	assert.Equal(t, `{"store1":"val1","store2":"val2"}`, storeJSON)

	err = kvstore.patchReplace(key, "/store1", `"replacedvalue"`)
	assert.NoError(t, err)
	storeJSON, err = kvstore.getAsJSON(key)
	assert.NoError(t, err)
	assert.Equal(t, `{"store1":"replacedvalue","store2":"val2"}`, storeJSON)
}

func TestKVStore_patch_invalidJsonError(t *testing.T) {
	kvstore := createInMemoryStorage()
	key := randString(10)
	err := kvstore.patch(key, `{ invalid json`)
	assert.ErrorContains(t, err, "invalid character 'i' looking for beginning of object key string")
}

func TestKVStore_patch_invalidPatchError(t *testing.T) {
	kvstore := createInMemoryStorage()
	key := randString(10)
	kvstore.put(key, `{ "foo":"bar" }`)
	err := kvstore.patch(key, `[{"op":"add","path":"/foo/no","value": "val"}]`)
	assert.ErrorContains(t, err, "json: cannot unmarshal string into Go value of type jsonpatch.partialDoc")
}

func TestKVStore_TemplateFuncMap(t *testing.T) {
	kvstore := createInMemoryStorage()
	key := randString(10)
	tpltCode := fmt.Sprintf(`{{ $kvs := kvStoreGet "%s" }}{{ printf "%%v" $kvs }}
{{ $kvs2 := kvStoreJsonPath "%s" "$" }}{{ printf "%%s" $kvs2 }}
{{ kvStoreRemove "%s" "/store1" -}}
{{ $kvs := kvStoreGet "%s" }}{{ printf "%%v" $kvs }}
{{ kvStoreAdd "%s" "/store3" "val3" -}}
{{ $kvs := kvStoreGet "%s" }}{{ printf "%%v" $kvs }}
{{ kvStorePut "%s" "allnew" -}}
{{ $kvs := kvStoreGet "%s" }}{{ printf "%%v" $kvs }}
{{ $kvs := kvStoreGet "notexisting" }}{{ printf "%%v" $kvs }}
{{ kvStoreAdd "%s" "wrongpath" "val3" }}
{{ kvStoreRemove "%s" "wrongPath" }}
{{ kvStoreJsonPath "%s" "wrongPath" }}
`, key, key, key, key, key, key, key, key, key, key, key)
	storeVal := map[string]interface{}{"store1": "val1", "store2": "val2"}
	err := kvstore.put(key, &storeVal)
	assert.NoError(t, err)
	templ, err := template.New("Test_TemplateFuncMap").Funcs(kvstore.templateFuncMap()).Parse(tpltCode)
	assert.NoError(t, err)
	var result bytes.Buffer
	err = templ.Execute(&result, nil)
	assert.NoError(t, err)
	assert.Equal(t, `&map[store1:val1 store2:val2]
&map[store1:val1 store2:val2]
map[store2:val2]
map[store2:val2 store3:val3]
allnew
<nil>
json: cannot unmarshal string into Go value of type jsonpatch.partialDoc
json: cannot unmarshal string into Go value of type jsonpatch.partialDoc
should start with '$'
`, result.String())

}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func randString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Int63()%int64(len(letterBytes))]
	}
	return string(b)
}
