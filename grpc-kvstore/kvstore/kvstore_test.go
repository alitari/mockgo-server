package kvstore

import (
	"go.uber.org/zap/zapcore"
	"log"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/alitari/mockgo-server/mockgo/testutil"
	"github.com/stretchr/testify/assert"
)

var clusterSize = 2
var startPort = 50151

var grpcstorages []*grpcStorage

func TestMain(m *testing.M) {
	startStorageCluster()
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
		kvStore, err := NewGrpcStorage(addresses, startPort+i, int(zapcore.DebugLevel))
		if err != nil {
			log.Fatal(err)
		}
		grpcstorages = append(grpcstorages, kvStore.(*grpcStorage))
	}
	time.Sleep(2000 * time.Millisecond)
}

func stopCluster() {
	for _, kvstore := range grpcstorages {
		kvstore.StopServe()
	}
}

func TestKVStore_Put(t *testing.T) {
	store := "mystore"
	key := "mystorekey"
	val, err := grpcstorages[0].Get(store, key)
	assert.NoError(t, err)
	assert.Nil(t, val)

	val, err = grpcstorages[1].Get(store, key)
	assert.NoError(t, err)
	assert.Nil(t, val)

	val = map[string]interface{}{"key1": []interface{}{"val11", "val22"}, "key2": []interface{}{"val21"}}
	grpcstorages[0].Put(store, key, val)
	time.Sleep(200 * time.Millisecond)
	getval, err := grpcstorages[0].Get(store, key)
	assert.NoError(t, err)
	assert.EqualValues(t, val, getval)

	getval, err = grpcstorages[1].Get(store, key)
	assert.NoError(t, err)
	assert.EqualValues(t, val, getval)

	val2 := map[string]interface{}{"key1": "val1", "key2": "val2"}
	grpcstorages[1].Put(store, key, val2)
	time.Sleep(200 * time.Millisecond)
	getVal2, err := grpcstorages[1].Get(store, key)
	assert.NoError(t, err)
	assert.Equal(t, val2, getVal2)

	getVal2, err = grpcstorages[0].Get(store, key)
	assert.NoError(t, err)
	assert.Equal(t, val2, getVal2)

	val, err = grpcstorages[0].Get(store+"changed", key)
	assert.NoError(t, err)
	assert.Nil(t, val)

	val, err = grpcstorages[1].Get(store+"changed", key)
	assert.NoError(t, err)
	assert.Nil(t, val)
}
