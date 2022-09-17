package utils

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/oliveagle/jsonpath"
)

func proxycallK8s(httpClient http.Client, method string, port int, path string, expectedResponseStatus int) (interface{}, error) {
	request, err := http.NewRequest(method, fmt.Sprintf("http://localhost:%d/%s", port, path), nil)
	if err != nil {
		return nil, err
	}
	log.Printf("calling K8s: %s|%s", request.Method, request.URL.String())
	response, err := httpClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode != expectedResponseStatus {
		return nil, fmt.Errorf("unexpeced response status from K8s api, expected: %d, but is %d", expectedResponseStatus, response.StatusCode)
	}
	bodyBytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	var pods interface{}
	err = json.Unmarshal(bodyBytes, &pods)
	if err != nil {
		return "", err
	}
	return pods, nil
}

func k8sGetPods(httpClient http.Client, port int, clusterPodLabelValue string) (interface{}, error) {
	return proxycallK8s(httpClient, http.MethodGet, port, "api/v1/namespaces/mockgo/pods?fieldSelector=status.phase%3DRunning&labelSelector=cluster%3D"+clusterPodLabelValue, http.StatusOK)
}

func K8sGetPodsIps(httpClient http.Client, port int, clusterPodLabelValue string) ([]string, error) {
	pods, err := k8sGetPods(httpClient, port, clusterPodLabelValue)
	if err != nil {
		return nil, err
	}
	ips, err := jsonpath.JsonPathLookup(pods, `$.items.status.podIP`)
	if err != nil {
		return nil, err
	}
	var ipsStrings []string
	for _, ipss := range ips.([]interface{}) {
		ipsStrings = append(ipsStrings, fmt.Sprintf("%v", ipss))
	}
	return ipsStrings, nil
}
