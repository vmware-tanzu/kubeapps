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
	"bytes"
	"context"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	plugins "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/plugins/v1alpha1"
	fluxplugin "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/plugins/fluxv2/packages/v1alpha1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	kubecorev1 "k8s.io/api/core/v1"
	kuberbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
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

		// check the fluxv2plugin-testdata-svc is deployed - without it,
		// one gets timeout errors when trying to index a repo, and it takes a really
		// long time
		typedClient, err := kubeGetTypedClient()
		if err != nil {
			t.Fatalf("%+v", err)
		}
		_, err = typedClient.CoreV1().Services("default").Get(context.TODO(), "fluxv2plugin-testdata-svc", metav1.GetOptions{})
		if err != nil {
			t.Fatalf("Failed to get service [default/fluxv2plugin-testdata-svc] due to: [%v]", err)
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
	} else if _, err := ifc.Create(context.TODO(), &unstructuredRepo, metav1.CreateOptions{}); err != nil {
		return err
	} else {
		return nil
	}
}

func kubeWaitUntilHelmRepositoryIsReady(t *testing.T, name, namespace string) error {
	t.Logf("+kubeWaitUntilHelmRepositoryIsReady(%s,%s)", name, namespace)
	if ifc, err := kubeGetHelmRepositoryResourceInterface(namespace); err != nil {
		return err
	} else {
		if watcher, err := ifc.Watch(context.TODO(), metav1.ListOptions{}); err != nil {
			return err
		} else {
			ch := watcher.ResultChan()
			defer watcher.Stop()
			for {
				event, ok := <-ch
				if !ok {
					return errors.New("Channel was closed unexpectedly")
				}
				if event.Type == "" {
					// not quite sure why this happens (the docs don't say), but it seems to happen quite often
					continue
				}
				switch event.Type {
				case watch.Added, watch.Modified:
					if unstructuredRepo, ok := event.Object.(*unstructured.Unstructured); !ok {
						return errors.New("Could not cast to unstructured.Unstructured")
					} else {
						hour, minute, second := time.Now().Clock()
						complete, success, reason := isHelmRepositoryReady(unstructuredRepo.Object)
						t.Logf("[%d:%d:%d] Got event: type: [%v], reason [%s]", hour, minute, second, event.Type, reason)
						if complete && success {
							return nil
						}
					}
				}
			}
		}
	}
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

func kubeDeleteHelmRelease(t *testing.T, name, namespace string) error {
	t.Logf("+kubeDeleteHelmRelease(%s,%s)", name, namespace)
	if ifc, err := kubeGetHelmReleaseResourceInterface(namespace); err != nil {
		return err
	} else if err = ifc.Delete(context.TODO(), name, metav1.DeleteOptions{}); err != nil {
		return err
	}
	return nil
}

func kubeExistsHelmRelease(t *testing.T, name, namespace string) (bool, error) {
	t.Logf("+kubeExistsHelmRelease(%s,%s)", name, namespace)
	if ifc, err := kubeGetHelmReleaseResourceInterface(namespace); err != nil {
		return false, err
	} else if _, err = ifc.Get(context.TODO(), name, metav1.GetOptions{}); err == nil {
		return true, nil
	} else {
		return false, nil
	}
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

// will create a service account with cluster-admin privs and return the associated
// Bearer token (base64-encoded)
func kubeCreateAdminServiceAccount(t *testing.T, name, namespace string) (string, error) {
	t.Logf("+kubeCreateAdminServiceAccount(%s,%s)", name, namespace)
	typedClient, err := kubeGetTypedClient()
	if err != nil {
		return "", err
	}
	_, err = typedClient.CoreV1().ServiceAccounts(namespace).Create(
		context.TODO(),
		&kubecorev1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
		},
		metav1.CreateOptions{})
	if err != nil {
		return "", err
	}

	secretName := ""
	for i := 0; i < 10; i++ {
		svcAccount, err := typedClient.CoreV1().ServiceAccounts(namespace).Get(context.TODO(), name, metav1.GetOptions{})
		if err != nil {
			return "", err
		}
		if len(svcAccount.Secrets) >= 1 && svcAccount.Secrets[0].Name != "" {
			secretName = svcAccount.Secrets[0].Name
			break
		}
		t.Logf("Waiting 1s for service account [%s] secret to be set up... [%d/%d]", name, i+1, 10)
		time.Sleep(1 * time.Second)
	}
	if secretName == "" {
		return "", fmt.Errorf("Service account [%s] has no secrets", name)
	}

	secret, err := typedClient.CoreV1().Secrets(namespace).Get(
		context.TODO(),
		secretName,
		metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	token := secret.Data["token"]
	if token == nil {
		return "", err
	}
	_, err = typedClient.RbacV1().ClusterRoleBindings().Create(
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
		metav1.CreateOptions{})
	if err != nil {
		return "", err
	}
	return string(token), nil
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

func kubeCreateNamespace(t *testing.T, namespace string) error {
	t.Logf("+kubeCreateNamespace(%s)", namespace)
	if typedClient, err := kubeGetTypedClient(); err != nil {
		return err
	} else if _, err = typedClient.CoreV1().Namespaces().Create(
		context.TODO(),
		&kubecorev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: namespace,
			},
		},
		metav1.CreateOptions{}); err != nil {
		return err
	}
	return nil
}

func kubeDeleteNamespace(t *testing.T, namespace string) error {
	t.Logf("+kubeDeleteNamespace(%s)", namespace)
	if typedClient, err := kubeGetTypedClient(); err != nil {
		return err
	} else if err = typedClient.CoreV1().Namespaces().Delete(
		context.TODO(),
		namespace,
		metav1.DeleteOptions{}); err != nil {
		return err
	}
	return nil
}

func kubeGetSecret(t *testing.T, namespace, name, dataKey string) (string, error) {
	t.Logf("+kubeGetSecret(%s, %s, %s)", namespace, name, dataKey)
	if typedClient, err := kubeGetTypedClient(); err != nil {
		return "", err
	} else if secret, err := typedClient.CoreV1().Secrets(namespace).Get(
		context.TODO(),
		name,
		metav1.GetOptions{}); err != nil {
		return "", err
	} else {
		token := secret.Data[dataKey]
		if token == nil {
			return "", errors.New("No data found")
		}
		return string(token), nil
	}
}

func kubePortForwardToRedis(t *testing.T) error {
	t.Logf("+kubePortForwardToRedis")
	stopChan, readyChan := make(chan struct{}, 1), make(chan struct{}, 1)
	go func() {
		if err := func() error {
			// ref https://github.com/kubernetes/client-go/issues/51
			if config, err := restConfig(); err != nil {
				return err
			} else if roundTripper, upgrader, err := spdy.RoundTripperFor(config); err != nil {
				return err
			} else {
				path := fmt.Sprintf("/api/v1/namespaces/%s/pods/%s/portforward", "kubeapps", "kubeapps-redis-master-0")
				hostIP := strings.TrimLeft(config.Host, "htps:/")
				serverURL := url.URL{Scheme: "https", Path: path, Host: hostIP}
				dialer := spdy.NewDialer(upgrader, &http.Client{Transport: roundTripper}, http.MethodPost, &serverURL)
				out, errOut := new(bytes.Buffer), new(bytes.Buffer)
				if forwarder, err := portforward.New(dialer, []string{"6379"}, stopChan, readyChan, out, errOut); err != nil {
					return err
				} else {
					go func() {
						for range readyChan { // Kubernetes will close this channel when it has something to tell us.
						}
						if len(errOut.String()) != 0 {
							t.Errorf("kubePortForwardToRedis: %s", errOut.String())
						} else if len(out.String()) != 0 {
							t.Logf("kubePortForwardToRedis: %s", out.String())
						}
					}()
					if err = forwarder.ForwardPorts(); err != nil { // Locks until stopChan is closed.
						return err
					}
				}
			}
			return nil
		}(); err != nil {
			t.Errorf("%+v", err)
		}
	}()
	// this will stop the port forwarding
	t.Cleanup(func() { close(stopChan) })

	// this will wait until port-forwarding is set up
	select {
	case <-readyChan:
		return nil
	case <-time.After(10 * time.Second):
		return errors.New("failed to start portforward in 10s")
	}
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

func newGrpcContext(t *testing.T, name string) context.Context {
	token, err := kubeCreateAdminServiceAccount(t, name, "default")
	if err != nil {
		t.Fatalf("Failed to create service account due to: %+v", err)
	}
	t.Cleanup(func() {
		if err := kubeDeleteServiceAccount(t, name, "default"); err != nil {
			t.Logf("Failed to delete service account due to: %+v", err)
		}
	})
	return metadata.NewOutgoingContext(
		context.TODO(),
		metadata.Pairs("Authorization", "Bearer "+token))
}

func redisCheckTinyMaxMemory(t *testing.T, redisCli *redis.Client, expectedMaxMemory string) error {
	maxmemory, err := redisCli.ConfigGet(redisCli.Context(), "maxmemory").Result()
	if err != nil {
		return err
	} else {
		currentMaxMemory := fmt.Sprintf("%v", maxmemory[1])
		t.Logf("Current redis maxmemory = [%s]", currentMaxMemory)
		if currentMaxMemory != expectedMaxMemory {
			t.Fatalf("This test requires redis config maxmemory to be set to %s", expectedMaxMemory)
		}
	}
	maxmemoryPolicy, err := redisCli.ConfigGet(redisCli.Context(), "maxmemory-policy").Result()
	if err != nil {
		return err
	} else {
		currentMaxMemoryPolicy := fmt.Sprintf("%v", maxmemoryPolicy[1])
		t.Logf("Current maxmemory policy = [%s]", currentMaxMemoryPolicy)
		if currentMaxMemoryPolicy != "allkeys-lru" {
			t.Fatalf("This test requires redis config maxmemory-policy to be set to allkeys-lru")
		}
	}
	return nil
}

// global vars
var (
	dynamicClient dynamic.Interface
	typedClient   kubernetes.Interface
	letters       = []rune("abcdefghijklmnopqrstuvwxyz0123456789")
)
