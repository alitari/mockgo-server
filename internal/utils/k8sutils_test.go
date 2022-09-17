package utils

import (
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestK8SUtils_GetPods(t *testing.T) {
	httpClient := CreateHttpClient(1 * time.Second)
	ips, err := K8sGetPodsIps(httpClient, 8090, "mockgo")
	assert.NoError(t, err)
	log.Printf("ips=%v", ips)
}
