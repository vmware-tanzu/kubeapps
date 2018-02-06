package cmd

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	api "k8s.io/client-go/pkg/api/v1"
	"net/http"
)

func CheckPods(namespace string, client *kubernetes.Clientset) {
	options := v1.ListOptions{}
	pods, err := client.CoreV1().Pods(namespace).List(options)
	if err != nil {
		KFail(err)
		return
	}
	KInfo("checking", namespace, ":", len(pods.Items), "pods found")
	for _, pod := range pods.Items {
		if pod.Status.Phase != api.PodRunning && pod.Status.Phase != api.PodSucceeded {
			KFail(pod.Name, "not working")
			return
		}
	}
	KPass(namespace, "pod status: OK")
}

func CheckEndpoints(namespace string, client *kubernetes.Clientset) {
	options := v1.ListOptions{}
	svcs, err := client.CoreV1().Endpoints(namespace).List(options)
	if err != nil {
		KFail(err)
		return
	}
	KInfo("checking", namespace, ":", len(svcs.Items), "endpoints found")

	for _, svc := range svcs.Items {
		for _, ep := range svc.Subsets {
			if len(ep.NotReadyAddresses) > 0 {
				KFail(svc.Name, "has endpoint not ready")
				return
			}
		}
	}
	KPass(namespace, "endpoints status: OK")
}

func PingPath(path string, uri string) {
	var client http.Client
	req, _ := http.NewRequest("GET", uri+path, nil)
	resp, err := client.Do(req)
	if err != nil {
		KFail("trying to reach", uri+path, err)
		return
	}
	if resp.StatusCode > 399 {
		KFail(uri+path, "fail with response", resp.StatusCode)
	} else {
		KPass(uri+path, "pass with response", resp.StatusCode)
	}

}
