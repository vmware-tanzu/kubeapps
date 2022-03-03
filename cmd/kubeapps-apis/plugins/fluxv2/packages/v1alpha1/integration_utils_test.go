// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

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

	helmv2 "github.com/fluxcd/helm-controller/api/v2beta1"
	"github.com/fluxcd/pkg/apis/meta"
	sourcev1 "github.com/fluxcd/source-controller/api/v1beta1"
	"github.com/go-redis/redis/v8"
	plugins "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/plugins/v1alpha1"
	fluxplugin "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/plugins/fluxv2/packages/v1alpha1"
	"github.com/kubeapps/kubeapps/pkg/chart/models"
	"github.com/kubeapps/kubeapps/pkg/helm"
	httpclient "github.com/kubeapps/kubeapps/pkg/http-client"
	"golang.org/x/sync/semaphore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	apiv1 "k8s.io/api/core/v1"
	kubecorev1 "k8s.io/api/core/v1"
	kuberbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	// EnvvarFluxIntegrationTests enables tests that run against a local kind cluster
	envVarFluxIntegrationTests = "ENABLE_FLUX_INTEGRATION_TESTS"
	defaultContextTimeout      = 30 * time.Second

	// This is local copy of the first few entries
	// on "https://stefanprodan.github.io/podinfo/index.yaml" as of Sept 10 2021 with the chart
	// urls modified to link to .tgz files also within the local cluster.
	// If we want other repos, we'll have add directories and tinker with ./Dockerfile and NGINX conf.
	// This relies on fluxv2plugin-testdata-svc service stood up by testdata/kind-cluster-setup.sh
	podinfo_repo_url = "http://fluxv2plugin-testdata-svc.default.svc.cluster.local:80/podinfo"

	// same as above but requires HTTP basic authentication: user: foo, password: bar
	podinfo_basic_auth_repo_url = "http://fluxv2plugin-testdata-svc.default.svc.cluster.local:80/podinfo-basic-auth"

	// same as above but requires TLS
	podinfo_tls_repo_url = "https://fluxv2plugin-testdata-ssl-svc.default.svc.cluster.local:443"
)

func checkEnv(t *testing.T) (fluxplugin.FluxV2PackagesServiceClient, fluxplugin.FluxV2RepositoriesServiceClient) {
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
		t.Skipf("skipping flux plugin integration tests because environment variable %q not set to be true", envVarFluxIntegrationTests)
	} else {
		if up, err := isLocalKindClusterUp(t); err != nil || !up {
			t.Fatalf("Failed to find local kind cluster due to: [%v]", err)
		}
		var fluxPluginPackagesClient fluxplugin.FluxV2PackagesServiceClient
		var fluxPluginReposClient fluxplugin.FluxV2RepositoriesServiceClient
		var err error
		if fluxPluginPackagesClient, fluxPluginReposClient, err = getFluxPluginClients(t); err != nil {
			t.Fatalf("Failed to get fluxv2 plugin due to: [%v]", err)
		}

		// check the fluxv2plugin-testdata-svc is deployed - without it,
		// one gets timeout errors when trying to index a repo, and it takes a really
		// long time
		typedClient, err := kubeGetTypedClient()
		if err != nil {
			t.Fatalf("%+v", err)
		}
		ctx, cancel := context.WithTimeout(context.Background(), defaultContextTimeout)
		defer cancel()
		_, err = typedClient.CoreV1().Services("default").Get(ctx, "fluxv2plugin-testdata-svc", metav1.GetOptions{})
		if err != nil {
			t.Fatalf("Failed to get service [default/fluxv2plugin-testdata-svc] due to: [%v]", err)
		}

		rand.Seed(time.Now().UnixNano())
		return fluxPluginPackagesClient, fluxPluginReposClient
	}
	return nil, nil
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

	ctx, cancel := context.WithTimeout(context.Background(), defaultContextTimeout)
	defer cancel()
	nodeList, err := typedClient.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
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

func getFluxPluginClients(t *testing.T) (fluxplugin.FluxV2PackagesServiceClient, fluxplugin.FluxV2RepositoriesServiceClient, error) {
	t.Logf("+getFluxPluginClients")

	var opts []grpc.DialOption
	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	opts = append(opts, grpc.WithBlock())
	target := "localhost:8080"
	conn, err := grpc.Dial(target, opts...)
	if err != nil {
		t.Fatalf("failed to dial [%s] due to: %v", target, err)
	}
	t.Cleanup(func() { conn.Close() })
	pluginsCli := plugins.NewPluginsServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), defaultContextTimeout)
	defer cancel()
	response, err := pluginsCli.GetConfiguredPlugins(ctx, &plugins.GetConfiguredPluginsRequest{})

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
		return nil, nil, fmt.Errorf("kubeapps Fluxv2 plugin is not registered, found these plugins: %v", response.Plugins)
	}
	return fluxplugin.NewFluxV2PackagesServiceClient(conn), fluxplugin.NewFluxV2RepositoriesServiceClient(conn), nil
}

// This creates a flux helm repository CRD. The usage of this func should be minimized as much as
// possible in favor of flux Plugin's AddPackageRepository() call
func kubeAddHelmRepository(t *testing.T, name, url, namespace, secretName string) error {
	t.Logf("+kubeCreateHelmRepository(%s,%s)", name, namespace)
	repo := sourcev1.HelmRepository{
		TypeMeta: metav1.TypeMeta{
			Kind:       sourcev1.HelmRepositoryKind,
			APIVersion: sourcev1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: sourcev1.HelmRepositorySpec{
			URL: url,
		},
	}

	if secretName != "" {
		repo.Spec.SecretRef = &meta.LocalObjectReference{
			Name: secretName,
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultContextTimeout)
	defer cancel()
	if ifc, err := kubeGetCtrlClient(); err != nil {
		return err
	} else {
		return ifc.Create(ctx, &repo)
	}
}

func kubeWaitUntilHelmRepositoryIsReady(t *testing.T, name, namespace string) error {
	t.Logf("+kubeWaitUntilHelmRepositoryIsReady(%s,%s)", name, namespace)

	if ifc, err := kubeGetCtrlClient(); err != nil {
		return err
	} else {
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()
		var repoList sourcev1.HelmRepositoryList
		if watcher, err := ifc.Watch(ctx, &repoList); err != nil {
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
					if repo, ok := event.Object.(*sourcev1.HelmRepository); !ok {
						return errors.New("Could not cast to *sourcev1.HelmRepository")
					} else {
						hour, minute, second := time.Now().Clock()
						complete, success, reason := isHelmRepositoryReady(*repo)
						t.Logf("[%d:%d:%d] Got event: type: [%v], reason [%s]", hour, minute, second, event.Type, reason)
						if complete && success {
							return nil
						} else if complete && !success {
							return errors.New(reason)
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
	repo := &sourcev1.HelmRepository{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
	ctx, cancel := context.WithTimeout(context.Background(), defaultContextTimeout)
	defer cancel()
	if ifc, err := kubeGetCtrlClient(); err != nil {
		return err
	} else {
		return ifc.Delete(ctx, repo)
	}
}

func kubeDeleteHelmRelease(t *testing.T, name, namespace string) error {
	t.Logf("+kubeDeleteHelmRelease(%s,%s)", name, namespace)
	ctx, cancel := context.WithTimeout(context.Background(), defaultContextTimeout)
	defer cancel()
	release := &helmv2.HelmRelease{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
	if ifc, err := kubeGetCtrlClient(); err != nil {
		return err
	} else {
		return ifc.Delete(ctx, release)
	}
}

func kubeExistsHelmRelease(t *testing.T, name, namespace string) (bool, error) {
	t.Logf("+kubeExistsHelmRelease(%s,%s)", name, namespace)
	ctx, cancel := context.WithTimeout(context.Background(), defaultContextTimeout)
	defer cancel()
	key := types.NamespacedName{Name: name, Namespace: namespace}
	var rel helmv2.HelmRelease
	if ifc, err := kubeGetCtrlClient(); err != nil {
		return false, err
	} else if err = ifc.Get(ctx, key, &rel); err == nil {
		return true, nil
	} else {
		return false, nil
	}
}

func kubeGetPodNames(t *testing.T, namespace string) (names []string, err error) {
	t.Logf("+kubeGetPodNames(%s)", namespace)
	ctx, cancel := context.WithTimeout(context.Background(), defaultContextTimeout)
	defer cancel()
	if typedClient, err := kubeGetTypedClient(); err != nil {
		return nil, err
	} else if podList, err := typedClient.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{}); err != nil {
		return nil, err
	} else {
		names := []string{}
		for _, p := range podList.Items {
			names = append(names, p.GetName())
		}
		return names, nil
	}
}

func kubeCreateServiceAccountWithClusterRole(t *testing.T, name, namespace, role string) (string, error) {
	t.Logf("+kubeCreateServiceAccountWithClusterRole(%s,%s,%s)", name, namespace, role)
	typedClient, err := kubeGetTypedClient()
	if err != nil {
		return "", err
	}
	ctx, cancel := context.WithTimeout(context.Background(), defaultContextTimeout)
	defer cancel()
	_, err = typedClient.CoreV1().ServiceAccounts(namespace).Create(
		ctx,
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
		ctx, cancel := context.WithTimeout(context.Background(), defaultContextTimeout)
		defer cancel()
		svcAccount, err := typedClient.CoreV1().ServiceAccounts(namespace).Get(ctx, name, metav1.GetOptions{})
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

	ctx, cancel = context.WithTimeout(context.Background(), defaultContextTimeout)
	defer cancel()
	secret, err := typedClient.CoreV1().Secrets(namespace).Get(
		ctx,
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
		ctx,
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
				Name: role,
			},
		},
		metav1.CreateOptions{})
	if err != nil {
		return "", err
	}
	return string(token), nil
}

// ref: https://kubernetes.io/docs/reference/access-authn-authz/rbac/#user-facing-roles
// will create a service account with cluster-admin privs and return the associated
// Bearer token (base64-encoded)
func kubeCreateAdminServiceAccount(t *testing.T, name, namespace string) (string, error) {
	return kubeCreateServiceAccountWithClusterRole(t, name, namespace, "cluster-admin")
}

func kubeCreateFluxPluginServiceAccount(t *testing.T, name, namespace string) (string, error) {
	return kubeCreateServiceAccountWithClusterRole(t, name, namespace, "kubeapps:controller:kubeapps-apis-fluxv2-plugin")
}

func kubeDeleteServiceAccount(t *testing.T, name, namespace string) error {
	t.Logf("+kubeDeleteServiceAccount(%s,%s)", name, namespace)
	typedClient, err := kubeGetTypedClient()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), defaultContextTimeout)
	defer cancel()
	err = typedClient.RbacV1().ClusterRoleBindings().Delete(
		ctx,
		name+"-binding",
		metav1.DeleteOptions{})
	if err != nil {
		return err
	}
	ctx, cancel = context.WithTimeout(context.Background(), defaultContextTimeout)
	defer cancel()
	err = typedClient.CoreV1().ServiceAccounts(namespace).Delete(
		ctx,
		name,
		metav1.DeleteOptions{})
	if err != nil {
		return err
	}
	return nil
}

func kubeCreateNamespace(t *testing.T, namespace string) error {
	t.Logf("+kubeCreateNamespace(%s)", namespace)
	typedClient, err := kubeGetTypedClient()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), defaultContextTimeout)
	defer cancel()
	_, err = typedClient.CoreV1().Namespaces().Create(
		ctx,
		&kubecorev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: namespace,
			},
		},
		metav1.CreateOptions{})
	return err
}

func kubeDeleteNamespace(t *testing.T, namespace string) error {
	t.Logf("+kubeDeleteNamespace(%s)", namespace)
	typedClient, err := kubeGetTypedClient()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), defaultContextTimeout)
	defer cancel()
	err = typedClient.CoreV1().Namespaces().Delete(
		ctx,
		namespace,
		metav1.DeleteOptions{})
	return err
}

func kubeGetSecret(t *testing.T, namespace, name, dataKey string) (string, error) {
	t.Logf("+kubeGetSecret(%s, %s, %s)", namespace, name, dataKey)
	typedClient, err := kubeGetTypedClient()
	if err != nil {
		return "", err
	}
	ctx, cancel := context.WithTimeout(context.Background(), defaultContextTimeout)
	defer cancel()
	secret, err := typedClient.CoreV1().Secrets(namespace).Get(
		ctx,
		name,
		metav1.GetOptions{})
	if err != nil {
		return "", err
	} else {
		token := secret.Data[dataKey]
		if token == nil {
			return "", errors.New("No data found")
		}
		return string(token), nil
	}
}

func kubeCreateSecret(t *testing.T, secret *apiv1.Secret) error {
	t.Logf("+kubeCreateSecret(%s, %s", secret.Namespace, secret.Name)
	typedClient, err := kubeGetTypedClient()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), defaultContextTimeout)
	defer cancel()
	_, err = typedClient.CoreV1().Secrets(secret.Namespace).Create(
		ctx,
		secret,
		metav1.CreateOptions{})
	return err
}

func kubeDeleteSecret(t *testing.T, namespace, name string) error {
	t.Logf("+kubeDeleteSecret(%s, %s)", namespace, name)
	typedClient, err := kubeGetTypedClient()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), defaultContextTimeout)
	defer cancel()
	return typedClient.CoreV1().Secrets(namespace).Delete(
		ctx,
		name,
		metav1.DeleteOptions{})
}

func kubePortForwardToRedis(t *testing.T) error {
	t.Logf("+kubePortForwardToRedis")
	defer t.Logf("-kubePortForwardToRedis")
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

func kubeGetCtrlClient() (ctrlclient.WithWatch, error) {
	if ctrlClient != nil {
		return ctrlClient, nil
	} else {
		if config, err := restConfig(); err != nil {
			return nil, err
		} else {
			scheme := runtime.NewScheme()
			_ = sourcev1.AddToScheme(scheme)
			_ = helmv2.AddToScheme(scheme)

			return ctrlclient.NewWithWatch(config, ctrlclient.Options{Scheme: scheme})
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

func newGrpcContext(t *testing.T, token string) context.Context {
	return metadata.NewOutgoingContext(
		context.TODO(),
		metadata.Pairs("Authorization", "Bearer "+token))
}

func newGrpcAdminContext(t *testing.T, name string) context.Context {
	token, err := kubeCreateAdminServiceAccount(t, name, "default")
	if err != nil {
		t.Fatalf("Failed to create service account due to: %+v", err)
	}
	t.Cleanup(func() {
		if err := kubeDeleteServiceAccount(t, name, "default"); err != nil {
			t.Logf("Failed to delete service account due to: %+v", err)
		}
	})
	return newGrpcContext(t, token)
}

func newGrpcFluxPluginContext(t *testing.T, name string) context.Context {
	token, err := kubeCreateFluxPluginServiceAccount(t, name, "default")
	if err != nil {
		t.Fatalf("Failed to create service account due to: %+v", err)
	}
	t.Cleanup(func() {
		if err := kubeDeleteServiceAccount(t, name, "default"); err != nil {
			t.Logf("Failed to delete service account due to: %+v", err)
		}
	})
	return newGrpcContext(t, token)
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
		if currentMaxMemoryPolicy != "allkeys-lfu" {
			t.Fatalf("This test requires redis config maxmemory-policy to be set to allkeys-lfu")
		}
	}
	return nil
}

func newRedisClientForIntegrationTest(t *testing.T) (*redis.Client, error) {
	if err := kubePortForwardToRedis(t); err != nil {
		return nil, fmt.Errorf("kubePortForwardToRedis failed due to %+v", err)
	}
	redisPwd, err := kubeGetSecret(t, "kubeapps", "kubeapps-redis", "redis-password")
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	redisCli := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: redisPwd,
		DB:       0,
	})
	t.Cleanup(func() {
		// we want to make sure at the end of the test the cache is empty just as it was when
		// we started
		const maxWait = 60
		for i := 0; ; i++ {
			if keys, err := redisCli.Keys(redisCli.Context(), "*").Result(); err != nil {
				t.Errorf("redisCli.Keys() failed due to: %+v", err)
			} else {
				if len(keys) == 0 {
					break
				}
				if i < maxWait {
					t.Logf("Waiting 2s until cache is empty. Current number of keys: [%d]", len(keys))
					time.Sleep(2 * time.Second)
				} else {
					t.Errorf("Failed because there are still [%d] keys left in the cache", len(keys))
					break
				}
			}
		}
		redisCli.Close()
	})
	t.Logf("redisCli: %s", redisCli)

	// confidence test, we expect the cache to be empty at this point
	// if it's not, it's likely that some cleanup didn't happen due to earlier an stopped test
	// and you should be able to clean up manually
	// $ kubectl delete helmrepositories --all
	if keys, err := redisCli.Keys(redisCli.Context(), "*").Result(); err != nil {
		return nil, fmt.Errorf("%v", err)
	} else {
		if len(keys) != 0 {
			t.Fatalf("Failing due to unexpected state of the cache. Current keys: %s", keys)
		}
	}
	return redisCli, nil
}

func redisReceiveNotificationsLoop(t *testing.T, ch <-chan *redis.Message, sem *semaphore.Weighted, evictedRepos *sets.String) {
	if totalBitnamiCharts == -1 {
		t.Errorf("Error: unexpected state: number of charts in bitnami catalog is not initialized")
		return
	}

	// this for loop running in the background will signal to the main goroutine
	// when it is okay to proceed to load the next repo
	t.Logf("Listening for events from redis in the background...")
	reposAdded := sets.String{}
	var chartsLeftToSync = 0
	for {
		event, ok := <-ch
		if !ok {
			t.Logf("Redis publish channel was closed")
			break
		}
		t.Logf("Redis event: [%v]: [%v]", event.Channel, event.Payload)
		if event.Channel == "__keyevent@0__:set" {
			if strings.HasPrefix(event.Payload, "helmrepositories:default:bitnami-") {
				reposAdded.Insert(event.Payload)
				// I am keeping track of charts being synced in the cache so that I only
				// start to load repository N+1 after completely done with N, meaning waiting until
				// the model for the repo and all its (latest) charts are in the cache. Thinking
				// about it now, I am not sure it's actually critical for this test to enforce
				// that a repo AND its charts are completely synced before proceeding. To be
				// continued...
				chartsLeftToSync += totalBitnamiCharts
			} else if strings.HasPrefix(event.Payload, "helmcharts:default:bitnami-") {
				chartID := strings.Split(event.Payload, ":")[2]
				repoKey := "helmrepositories:default:" + strings.Split(chartID, "/")[0]
				if reposAdded.Has(repoKey) {
					chartsLeftToSync--
				}
				t.Logf("Charts left to sync: [%d]", chartsLeftToSync)
			}
			if reposAdded.Len() > 0 && chartsLeftToSync == 0 && sem != nil {
				// signal to the main goroutine it's okay to proceed to load the next copy
				sem.Release(1)
			}
		} else if event.Channel == "__keyevent@0__:evicted" &&
			strings.HasPrefix(event.Payload, "helmrepositories:default:bitnami-") {
			evictedRepos.Insert(event.Payload)
			if reposAdded.Len() > 0 && sem != nil {
				// signal to the main goroutine it's okay to proceed to load the next copy
				sem.Release(1)
			}
		}
	}
}

func initNumberOfChartsInBitnamiCatalog(t *testing.T) error {
	t.Logf("+initNumberOfChartsInBitnamiCatalog")

	bitnamiUrl := "https://charts.bitnami.com/bitnami"

	byteArray, err := httpclient.Get(bitnamiUrl+"/index.yaml", httpclient.New(), nil)
	if err != nil {
		return err
	}

	modelRepo := &models.Repo{
		Namespace: "default",
		Name:      "bitnami",
		URL:       bitnamiUrl,
		Type:      "helm",
	}

	charts, err := helm.ChartsFromIndex(byteArray, modelRepo, true)
	if err != nil {
		return err
	}

	totalBitnamiCharts = len(charts)
	t.Logf("+initNumberOfChartsInBitnamiCatalog: total [%d] charts", totalBitnamiCharts)
	return nil
}

// global vars
var (
	typedClient kubernetes.Interface
	ctrlClient  ctrlclient.WithWatch
	letters     = []rune("abcdefghijklmnopqrstuvwxyz0123456789")
	// total number of unique packages in bitnami repo,
	// initialized during running of the integration test
	totalBitnamiCharts = -1
)
