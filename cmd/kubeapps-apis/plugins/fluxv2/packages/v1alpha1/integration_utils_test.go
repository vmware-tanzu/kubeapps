/*
Copyright © 2021 VMware
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
package main

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"testing"
	"time"

	plugins "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/plugins/v1alpha1"
	fluxplugin "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/plugins/fluxv2/packages/v1alpha1"
	"google.golang.org/grpc"
	kubecorev1 "k8s.io/api/core/v1"
	kuberbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	// EnvvarFluxIntegrationTests enables tests that run against a local kind cluster
	envVarFluxIntegrationTests = "ENABLE_FLUX_INTEGRATION_TESTS"
)

func checkEnv(t *testing.T) fluxplugin.FluxV2PackagesServiceClient {
	enableEnvVar := os.Getenv(envVarFluxIntegrationTests)
	runTests := false
	if enableEnvVar != "" {
		var err error
		runTests, err = strconv.ParseBool(enableEnvVar)
		if err != nil {
			t.Fatalf("%+v", err)
		}
	}

	if !runTests {
		t.Skipf("skipping flux plugin integration tests as %q not set to be true", envVarFluxIntegrationTests)
	} else {
		if up, err := isLocalKindClusterUp(t); err != nil || !up {
			t.Fatalf("Failed to find local kind cluster due to: [%v]", err)
		}
		var fluxPluginClient fluxplugin.FluxV2PackagesServiceClient
		var err error
		if fluxPluginClient, err = getFluxPluginClient(t); err != nil {
			t.Fatalf("Failed to get fluxv2 plugin due to: [%v]", err)
		}
		rand.Seed(time.Now().UnixNano())
		return fluxPluginClient
	}
	return nil
}

func isLocalKindClusterUp(t *testing.T) (up bool, err error) {
	t.Logf("+isLocalKindClusterUp")
	cmd := exec.Command("kind", "get", "clusters")
	bytes, err := cmd.CombinedOutput()
	if err != nil {
		t.Logf("%s", string(bytes))
		return false, err
	}
	if !strings.Contains(string(bytes), "kubeapps\n") {
		return false, nil
	}

	// naively assume that if the api server reports nodes, the cluster is up
	typedClient, err := kubeGetTypedClient()
	if err != nil {
		t.Logf("%s", string(bytes))
		return false, err
	}

	nodeList, err := typedClient.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		t.Logf("%s", string(bytes))
		return false, err
	}

	if len(nodeList.Items) == 1 || nodeList.Items[0].Name == "node/kubeapps-control-plane" {
		return true, nil
	} else {
		return false, fmt.Errorf("Unexpected cluster nodes: [%v]", nodeList)
	}
}

func getFluxPluginClient(t *testing.T) (fluxplugin.FluxV2PackagesServiceClient, error) {
	t.Logf("+getFluxPluginClient")

	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	opts = append(opts, grpc.WithBlock())
	target := "localhost:8080"
	conn, err := grpc.Dial(target, opts...)
	if err != nil {
		t.Fatalf("failed to dial [%s] due to: %v", target, err)
	}
	t.Cleanup(func() { conn.Close() })
	pluginsCli := plugins.NewPluginsServiceClient(conn)
	response, err := pluginsCli.GetConfiguredPlugins(context.TODO(), &plugins.GetConfiguredPluginsRequest{})
	if err != nil {
		t.Fatalf("failed to GetConfiguredPlugins due to: %v", err)
	}
	found := false
	for _, p := range response.Plugins {
		if p.Name == "fluxv2.packages" && p.Version == "v1alpha1" {
			found = true
			break
		}
	}
	if !found {
		return nil, fmt.Errorf("kubeapps Flux v2 plugin is not registered")
	}
	return fluxplugin.NewFluxV2PackagesServiceClient(conn), nil
}

// This should eventually be replaced with fluxPlugin CreateRepository() call as soon as we finalize
// the design
func kubeCreateHelmRepository(t *testing.T, name, url, namespace string) error {
	t.Logf("+kubeCreateHelmRepository(%s,%s)", name, namespace)
	unstructuredRepo := unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": fmt.Sprintf("%s/%s", fluxGroup, fluxVersion),
			"kind":       fluxHelmRepository,
			"metadata": map[string]interface{}{
				"name":      name,
				"namespace": namespace,
			},
			"spec": map[string]interface{}{
				"url":      url,
				"interval": "1m",
			},
		},
	}

	if ifc, err := kubeGetHelmRepositoryResourceInterface(namespace); err != nil {
		return err
	} else if _, err = ifc.Create(context.TODO(), &unstructuredRepo, metav1.CreateOptions{}); err != nil {
		return err
	}
	return nil
}

// this should eventually be replaced with flux plugin's DeleteRepository()
func kubeDeleteHelmRepository(t *testing.T, name, namespace string) error {
	t.Logf("+kubeDeleteHelmRepository(%s,%s)", name, namespace)
	if ifc, err := kubeGetHelmRepositoryResourceInterface(namespace); err != nil {
		return err
	} else if err = ifc.Delete(context.TODO(), name, metav1.DeleteOptions{}); err != nil {
		return err
	}
	return nil
}

// this should eventually be replaced with flux plugin's DeleteInstalledPackage()
func kubeDeleteHelmRelease(t *testing.T, name, namespace string) error {
	t.Logf("+kubeDeleteHelmRelease(%s,%s)", name, namespace)
	if ifc, err := kubeGetHelmReleaseResourceInterface(namespace); err != nil {
		return err
		// remove finalizer on HelmRelease cuz sometimes it gets stuck indefinitely
	} else if _, err = ifc.Patch(context.TODO(), name, types.JSONPatchType,
		[]byte("[ { \"op\": \"remove\", \"path\": \"/metadata/finalizers\" } ]"), metav1.PatchOptions{}); err != nil {
		return err
	} else if err = ifc.Delete(context.TODO(), name, metav1.DeleteOptions{}); err != nil {
		return err
	}
	return nil
}

func kubeGetPodNames(t *testing.T, namespace string) (names []string, err error) {
	t.Logf("+kubeGetPodNames(%s)", namespace)
	if typedClient, err := kubeGetTypedClient(); err != nil {
		return nil, err
	} else if podList, err := typedClient.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{}); err != nil {
		return nil, err
	} else {
		names := []string{}
		for _, p := range podList.Items {
			names = append(names, p.GetName())
		}
		return names, nil
	}
}

// will create a service account with cluster-admin privs
func kubeCreateServiceAccount(t *testing.T, name, namespace string) error {
	t.Logf("+kubeCreateServiceAccount(%s,%s)", name, namespace)
	if typedClient, err := kubeGetTypedClient(); err != nil {
		return err
	} else if _, err = typedClient.CoreV1().ServiceAccounts(namespace).Create(
		context.TODO(),
		&kubecorev1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
		},
		metav1.CreateOptions{}); err != nil {
		return err
	} else if _, err = typedClient.RbacV1().ClusterRoleBindings().Create(
		context.TODO(),
		&kuberbacv1.ClusterRoleBinding{
			ObjectMeta: metav1.ObjectMeta{
				Name: name + "-binding",
			},
			Subjects: []kuberbacv1.Subject{
				{
					Kind:      kuberbacv1.ServiceAccountKind,
					Name:      name,
					Namespace: namespace,
				},
			},
			RoleRef: kuberbacv1.RoleRef{
				Kind: "ClusterRole",
				Name: "cluster-admin",
			},
		},
		metav1.CreateOptions{}); err != nil {
		return err
	}
	return nil
}

func kubeDeleteServiceAccount(t *testing.T, name, namespace string) error {
	t.Logf("+kubeDeleteServiceAccount(%s,%s)", name, namespace)
	if typedClient, err := kubeGetTypedClient(); err != nil {
		return err
	} else if err = typedClient.RbacV1().ClusterRoleBindings().Delete(
		context.TODO(),
		name+"-binding",
		metav1.DeleteOptions{}); err != nil {
		return err
	} else if err = typedClient.CoreV1().ServiceAccounts(namespace).Delete(
		context.TODO(),
		name,
		metav1.DeleteOptions{}); err != nil {
		return err
	}
	return nil
}

func kubeDeleteNamespace(t *testing.T, namespace string) error {
	t.Logf("+kubeDeleteNamespace(%s)", namespace)
	if typedClient, err := kubeGetTypedClient(); err != nil {
		return err
	} else if typedClient.CoreV1().Namespaces().Delete(
		context.TODO(),
		namespace,
		metav1.DeleteOptions{}); err != nil {
		return err
	}
	return nil
}

func kubeGetHelmReleaseResourceInterface(namespace string) (dynamic.ResourceInterface, error) {
	clientset, err := kubeGetDynamicClient()
	if err != nil {
		return nil, err
	}
	relResource := schema.GroupVersionResource{
		Group:    fluxHelmReleaseGroup,
		Version:  fluxHelmReleaseVersion,
		Resource: fluxHelmReleases,
	}
	return clientset.Resource(relResource).Namespace(namespace), nil
}

func kubeGetHelmRepositoryResourceInterface(namespace string) (dynamic.ResourceInterface, error) {
	clientset, err := kubeGetDynamicClient()
	if err != nil {
		return nil, err
	}
	repoResource := schema.GroupVersionResource{
		Group:    fluxGroup,
		Version:  fluxVersion,
		Resource: fluxHelmRepositories,
	}
	return clientset.Resource(repoResource).Namespace(namespace), nil
}

func kubeGetDynamicClient() (dynamic.Interface, error) {
	if dynamicClient != nil {
		return dynamicClient, nil
	} else {
		if config, err := restConfig(); err != nil {
			return nil, err
		} else {
			dynamicClient, err = dynamic.NewForConfig(config)
			return dynamicClient, err
		}
	}
}

func kubeGetTypedClient() (kubernetes.Interface, error) {
	if typedClient != nil {
		return typedClient, nil
	} else {
		if config, err := restConfig(); err != nil {
			return nil, err
		} else {
			typedClient, err = kubernetes.NewForConfig(config)
			return typedClient, err
		}
	}
}

func restConfig() (*rest.Config, error) {
	kubeconfig := os.Getenv("KUBECONFIG")
	return clientcmd.BuildConfigFromFlags("", kubeconfig)
}

func randSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

// global vars
var (
	dynamicClient dynamic.Interface
	typedClient   kubernetes.Interface
	letters       = []rune("abcdefghijklmnopqrstuvwxyz0123456789")
)
