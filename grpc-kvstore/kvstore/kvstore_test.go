package kvstore

import (
	"log"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/alitari/mockgo-server/mockgo/logging"
	"github.com/alitari/mockgo-server/mockgo/util"
	"github.com/stretchr/testify/assert"
)

var clusterSize = 2
var startPort = 50151

var kvstores []*GrpcKVStore

func TestMain(m *testing.M) {
	startKVStoreCluster()
	time.Sleep(300 * time.Millisecond)
	code := util.RunAndCheckCoverage("main", m, 0.4)
	stopCluster()
	os.Exit(code)
}

func getClusterAdresses() []string {
	var clusterAddresses []string
	for i := 0; i < clusterSize; i++ {
		clusterAddresses = append(clusterAddresses, "localhost:"+strconv.Itoa(startPort+i))
	}
	return clusterAddresses
}

func startKVStoreCluster() {
	addresses := getClusterAdresses()
	for i := 0; i < clusterSize; i++ {
		kvStore, err := NewGrpcKVstore(addresses, startPort+i, logging.NewLoggerUtil(logging.Debug))
		if err != nil {
			log.Fatal(err)
		}
		kvstores = append(kvstores, kvStore)
	}
}

func stopCluster() {
	for _, kvstore := range kvstores {
		kvstore.StopServe()
	}
}
func TestKVStore_Put(t *testing.T) {
	storeKey := "mystore"
	val, err := kvstores[0].GetVal(storeKey)
	assert.NoError(t, err)
	assert.Nil(t, val)

	val, err = kvstores[1].GetVal(storeKey)
	assert.NoError(t, err)
	assert.Nil(t, val)

	storeVal := map[string]interface{}{"key1": []interface{}{"val11", "val22"}, "key2": []interface{}{"val21"}}
	kvstores[0].PutVal(storeKey, storeVal)
	val, err = kvstores[0].GetVal(storeKey)
	assert.NoError(t, err)
	assert.EqualValues(t, storeVal, val)

	val, err = kvstores[1].GetVal(storeKey)
	assert.NoError(t, err)
	assert.EqualValues(t, storeVal, val)

	storeVal2 := map[string]interface{}{"key1": "val1", "key2": "val2"}
	kvstores[1].PutVal(storeKey, storeVal2)
	val, err = kvstores[1].GetVal(storeKey)
	assert.NoError(t, err)
	assert.Equal(t, storeVal2, val)

	val, err = kvstores[0].GetVal(storeKey)
	assert.NoError(t, err)
	assert.Equal(t, storeVal2, val)
}
