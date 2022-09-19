package utils

import (
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestK8SUtils_ProxyGetPods(t *testing.T) {
	httpClient := CreateHttpClient(1 * time.Second)
	ips, err := ProxyK8sGetPodsIps(httpClient, 8090, "mockgo")
	assert.NoError(t, err)
	log.Printf("ips=%v", ips)
}

func TestK8SUtils_GetPods(t *testing.T) {
	err := newK8sOutClusterConfig("minikube", "mockgo")
	assert.NoError(t, err)
	pods, err := k8sGetClusterPods("mockgo")
	assert.NoError(t, err)
	log.Printf("podsCount=%d", len(pods))
	ips, err  := K8sGetPodsIps("mockgo")
	assert.NoError(t, err)
	log.Printf("podsIps=%v", ips)
}
