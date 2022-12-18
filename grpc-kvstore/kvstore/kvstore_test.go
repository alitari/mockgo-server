package kvstore

import (
	"log"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/alitari/mockgo-server/mockgo/logging"
	"github.com/alitari/mockgo-server/mockgo/testutil"
	"github.com/stretchr/testify/assert"
)

var clusterSize = 2
var startPort = 50151

var grpcstorages []*grpcStorage

func TestMain(m *testing.M) {
	startStorageCluster()
	time.Sleep(300 * time.Millisecond)
	code := testutil.RunAndCheckCoverage("main", m, 0.4)
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

func startStorageCluster() {
	addresses := getClusterAdresses()
	for i := 0; i < clusterSize; i++ {
		kvStore, err := NewGrpcStorage(addresses, startPort+i, logging.NewLoggerUtil(logging.Debug))
		if err != nil {
			log.Fatal(err)
		}
		grpcstorages = append(grpcstorages, kvStore.(*grpcStorage))
	}
}

func stopCluster() {
	for _, kvstore := range grpcstorages {
		kvstore.StopServe()
	}
}
func TestKVStore_Put(t *testing.T) {
	storeKey := "mystore"
	val, err := grpcstorages[0].GetVal(storeKey)
	assert.NoError(t, err)
	assert.Nil(t, val)

	val, err = grpcstorages[1].GetVal(storeKey)
	assert.NoError(t, err)
	assert.Nil(t, val)

	storeVal := map[string]interface{}{"key1": []interface{}{"val11", "val22"}, "key2": []interface{}{"val21"}}
	grpcstorages[0].PutVal(storeKey, storeVal)
	val, err = grpcstorages[0].GetVal(storeKey)
	assert.NoError(t, err)
	assert.EqualValues(t, storeVal, val)

	val, err = grpcstorages[1].GetVal(storeKey)
	assert.NoError(t, err)
	assert.EqualValues(t, storeVal, val)

	storeVal2 := map[string]interface{}{"key1": "val1", "key2": "val2"}
	grpcstorages[1].PutVal(storeKey, storeVal2)
	val, err = grpcstorages[1].GetVal(storeKey)
	assert.NoError(t, err)
	assert.Equal(t, storeVal2, val)

	val, err = grpcstorages[0].GetVal(storeKey)
	assert.NoError(t, err)
	assert.Equal(t, storeVal2, val)
}
