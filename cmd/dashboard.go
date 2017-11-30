/*
Copyright (c) 2017 Bitnami

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cmd

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"syscall"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
)

const (
	selector         = "name=nginx-ingress-controller"
	ingressNamespace = "kubeapps"
)

var dashboardCmd = &cobra.Command{
	Use:   "dashboard FLAG",
	Short: "Opens the KubeApps Dashboard",
	Long:  "Opens the KubeApps Dashboard",
	RunE: func(cmd *cobra.Command, args []string) error {
		pool, disco, err := restClientPool()
		if err != nil {
			return err
		}

		gvk := schema.GroupVersionKind{Version: "v1", Kind: "Pod"}
		rc, err := clientForGroupVersionKind(pool, disco, gvk, ingressNamespace)
		if err != nil {
			return err
		}

		podList, err := rc.List(metav1.ListOptions{LabelSelector: selector})
		if err != nil {
			return err
		}

		pods := podList.(*unstructured.UnstructuredList).Items

		if len(pods) == 0 {
			return errors.New("nginx ingress controller pod not found, run kubeapps up first")
		}

		podName := pods[0].GetName()

		localPort, err := cmd.Flags().GetInt("port")
		if err != nil {
			return err
		}

		return runPortforward(podName, localPort)
	},
}

func runPortforward(podName string, localPort int) error {
	cmd, err := exec.LookPath("kubectl")
	if err != nil {
		return err
	}
	args := []string{"kubectl", "--namespace", ingressNamespace, "port-forward", podName, fmt.Sprintf("%d:80", localPort)}

	env := os.Environ()

	openInBrowser(fmt.Sprintf("http://localhost:%d", localPort))
	return syscall.Exec(cmd, args, env)
}

func openInBrowser(url string) error {
	args := []string{"xdg-open"}
	switch runtime.GOOS {
	case "darwin":
		args = []string{"open"}
	case "windows":
		args = []string{"cmd", "/c", "start"}
	}
	cmd := exec.Command(args[0], append(args[1:], url)...)
	return cmd.Start()
}

func clientForGroupVersionKind(pool dynamic.ClientPool, disco discovery.DiscoveryInterface, gvk schema.GroupVersionKind, namespace string) (*dynamic.ResourceClient, error) {
	client, err := pool.ClientForGroupVersionKind(gvk)
	if err != nil {
		return nil, err
	}

	resource, err := serverResourceForGroupVersionKind(disco, gvk)
	if err != nil {
		return nil, err
	}

	rc := client.Resource(resource, namespace)
	return rc, nil
}

// taken from https://github.com/ksonnet/kubecfg/blob/897a3db8a83ca195a2825b1fabe59ffca103e700/utils/client.go#L156
func serverResourceForGroupVersionKind(disco discovery.DiscoveryInterface, gvk schema.GroupVersionKind) (*metav1.APIResource, error) {
	resources, err := disco.ServerResourcesForGroupVersion(gvk.GroupVersion().String())
	if err != nil {
		return nil, err
	}

	for _, r := range resources.APIResources {
		if r.Kind == gvk.Kind {
			return &r, nil
		}
	}

	return nil, fmt.Errorf("Server is unable to handle %s", gvk)
}

func init() {
	RootCmd.AddCommand(dashboardCmd)
	dashboardCmd.Flags().Int("port", 8002, "local port")
}
