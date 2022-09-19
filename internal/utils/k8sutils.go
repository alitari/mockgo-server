package utils

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/oliveagle/jsonpath"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

var clientset *kubernetes.Clientset
var clusterNamespace string

func NewK8sInClusterConfig() error {
	config, err := rest.InClusterConfig()
	if err != nil {
		return err
	}
	// creates the clientset
	clientset, err = kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}

	namespaceBytes, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
	if err != nil {
		return err
	}
	clusterNamespace = strings.TrimSpace(string(namespaceBytes))
	return nil
}

func newK8sOutClusterConfig(configFile, namespace string) error {
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", configFile), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		return err
	}

	clientset, err = kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}
	clusterNamespace = namespace
	return nil
}

func k8sGetPodInfos(fieldSelector, labelSelector string, podFilter func(pod *v1.Pod) bool, podMap func(pod *v1.Pod) string) ([]string, error) {
	pods, err := clientset.CoreV1().Pods(clusterNamespace).List(context.TODO(),
		metav1.ListOptions{FieldSelector: fieldSelector, LabelSelector: labelSelector})
	if err != nil {
		return nil, err
	}
	var filteredPods []string
	for _, pod := range pods.Items {
		if podFilter(&pod) {
			filteredPods = append(filteredPods, podMap(&pod))
		}
	}
	return filteredPods, nil
}

func K8sGetPodsIps(clusterPodLabelValue string) ([]string, error) {
	infos, err := k8sGetPodInfos("status.phase=Running", "cluster="+clusterPodLabelValue,
		func(pod *v1.Pod) bool {
			for _, condition := range pod.Status.Conditions {
				if condition.Type == v1.PodReady && condition.Status == v1.ConditionTrue {
					return true
				}
			}
			return false
		},
		func(pod *v1.Pod) string {
			return pod.Status.PodIP
		})
	if err != nil {
		return nil, err
	}
	return infos, nil
}

func k8sGetClusterPods(clusterPodLabelValue string) ([]*v1.Pod, error) {
	pods, err := clientset.CoreV1().Pods(clusterNamespace).List(context.TODO(),
		metav1.ListOptions{FieldSelector: "status.phase=Running", LabelSelector: "cluster=" + clusterPodLabelValue})
	if err != nil {
		return nil, err
	}
	var filteredPods []*v1.Pod
	for _, pod := range pods.Items {
		for _, condition := range pod.Status.Conditions {
			if condition.Type == v1.PodReady && condition.Status == v1.ConditionTrue {
				filteredPods = append(filteredPods, &pod)
			}
		}
	}
	return filteredPods, nil
}

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
	bodyBytes, err := io.ReadAll(response.Body)
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

func _k8sGetPods(httpClient http.Client, port int, clusterPodLabelValue string) (interface{}, error) {
	return proxycallK8s(httpClient, http.MethodGet, port, "api/v1/namespaces/mockgo/pods?fieldSelector=status.phase%3DRunning&labelSelector=cluster%3D"+clusterPodLabelValue, http.StatusOK)
}

func ProxyK8sGetPodsIps(httpClient http.Client, port int, clusterPodLabelValue string) ([]string, error) {
	pods, err := _k8sGetPods(httpClient, port, clusterPodLabelValue)
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
