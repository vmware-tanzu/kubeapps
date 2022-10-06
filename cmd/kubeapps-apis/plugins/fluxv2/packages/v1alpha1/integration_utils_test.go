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
	sourcev1 "github.com/fluxcd/source-controller/api/v1beta2"
	"github.com/go-redis/redis/v8"
	"github.com/vmware-tanzu/kubeapps/cmd/apprepository-controller/pkg/client/clientset/versioned/scheme"
	plugins "github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/gen/core/plugins/v1alpha1"
	fluxplugin "github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/gen/plugins/fluxv2/packages/v1alpha1"
	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/plugins/fluxv2/packages/v1alpha1/common"
	"github.com/vmware-tanzu/kubeapps/pkg/chart/models"
	"github.com/vmware-tanzu/kubeapps/pkg/helm"
	httpclient "github.com/vmware-tanzu/kubeapps/pkg/http-client"
	"golang.org/x/sync/semaphore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	apiv1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
	"k8s.io/kubectl/pkg/cmd/cp"
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
	// This relies on fluxv2plugin-testdata-svc service stood up by testdata/integ-test-env.sh
	podinfo_repo_url = "http://fluxv2plugin-testdata-svc.default.svc.cluster.local:80/podinfo"

	// same as above but requires HTTP basic authentication: user: foo, password: bar
	podinfo_basic_auth_repo_url = "http://fluxv2plugin-testdata-svc.default.svc.cluster.local:80/podinfo-basic-auth"

	// same as above but requires TLS
	podinfo_tls_repo_url = "https://fluxv2plugin-testdata-ssl-svc.default.svc.cluster.local:443"

	// download bitnami index.yaml once, push it to the flux2testdata pod and use
	// that URL to avoid intermittent
	// "Failed: failed to fetch Helm repository index: failed to cache index to temporary file: unexpected EOF"
	// This is the URL of local copy of http://charts.bitnami.com/bitnami/index.yaml.
	// It gets set up at the time you build the docker image for fluxv2plugin-testdata-app.
	// Note this solution only avoids having to GET index.yaml,
	// all the chart .tgz files are still retrieved from bitnami.com
	in_cluster_bitnami_url = "http://fluxv2plugin-testdata-svc.default.svc.cluster.local:80/bitnami"

	// port forward is done programmatically
	outside_cluster_bitnami_url = "http://localhost:50057/bitnami"

	// an OCI registry with a single chart (podinfo)
	// a clone of "oci://ghcr.io/stefanprodan/charts"
	// gets setup by integ-test-env.sh
	github_stefanprodan_podinfo_oci_registry_url         = "oci://ghcr.io/gfichtenholt/stefanprodan-podinfo-clone"
	harbor_stefanprodan_podinfo_oci_registry_url         = "oci://demo.goharbor.io/stefanprodan-podinfo-clone"
	harbor_stefanprodan_podinfo_private_oci_registry_url = "oci://demo.goharbor.io/stefanprodan-podinfo-clone-private"
	gcp_stefanprodan_podinfo_oci_registry_url            = "oci://us-west1-docker.pkg.dev/vmware-kubeapps-ci/stefanprodan-podinfo-clone"

	// the URL of local in cluster helm registry. Gets deployed via ./integ-test-env.sh
	// in_cluster_oci_registry_url = "oci://registry-app-svc.default.svc.cluster.local:5000/helm-charts"

	github_gfichtenholt_podinfo_oci_registry_url = "oci://ghcr.io/gfichtenholt/helm-charts"

	// admin/Harbor12345 is a well known default login for harbor registries
	harbor_host       = "demo.goharbor.io"
	harbor_admin_user = "admin"
	harbor_admin_pwd  = "Harbor12345"
)

func checkEnv(t *testing.T) (fluxplugin.FluxV2PackagesServiceClient, fluxplugin.FluxV2RepositoriesServiceClient, error) {
	enableEnvVar := os.Getenv(envVarFluxIntegrationTests)
	runTests := false
	if enableEnvVar != "" {
		var err error
		runTests, err = strconv.ParseBool(enableEnvVar)
		if err != nil {
			return nil, nil, err
		}
	}

	if !runTests {
		t.Skipf("skipping flux plugin integration tests because environment variable [%q] not set to be true", envVarFluxIntegrationTests)
		return nil, nil, nil
	} else {
		if up, err := isLocalKindClusterUp(t); err != nil || !up {
			return nil, nil, fmt.Errorf("Failed to find local kind cluster due to: [%v]", err)
		}
		var fluxPluginPackagesClient fluxplugin.FluxV2PackagesServiceClient
		var fluxPluginReposClient fluxplugin.FluxV2RepositoriesServiceClient
		var err error
		if fluxPluginPackagesClient, fluxPluginReposClient, err = getFluxPluginClients(t); err != nil {
			return nil, nil, fmt.Errorf("Failed to get fluxv2 plugin due to: [%v]", err)
		}

		// check the fluxv2plugin-testdata-svc is deployed - without it,
		// one gets timeout errors when trying to index a repo, and it takes a really
		// long time
		typedClient, err := kubeGetTypedClient()
		if err != nil {
			return nil, nil, err
		}
		ctx, cancel := context.WithTimeout(context.Background(), defaultContextTimeout)
		defer cancel()
		_, err = typedClient.CoreV1().Services("default").Get(ctx, "fluxv2plugin-testdata-svc", metav1.GetOptions{})
		if err != nil {
			return nil, nil, fmt.Errorf("Failed to get service [default/fluxv2plugin-testdata-svc] due to: [%v]", err)
		}

		// Check for helmrepositories left over from manual testing. This has caused me a lot grief
		var l *sourcev1.HelmRepositoryList
		var names []string
		const maxWait = 25
		for i := 0; i <= maxWait; i++ {
			l, err = kubeListAllHelmRepositories(t)
			if err != nil {
				return nil, nil, fmt.Errorf("Failed to get list of HelmRepositories due to: [%v]", err)
			} else if len(l.Items) != 0 {
				names = []string{}
				for _, p := range l.Items {
					names = append(names, p.GetNamespace()+"/"+p.GetName())
				}
				t.Logf("Waiting 2s until HelmRepositories %s are gone...", names)
				time.Sleep(2 * time.Second)
			} else {
				break
			}
		}
		if len(l.Items) != 0 {
			t.Logf("The following existing HelmRepositories where found in the cluster: %s", names)
			t.Logf("You may use command [kubectl delete helmrepositories --all] to delete them")
			return nil, nil, fmt.Errorf("Failed due to existing HelmRepositories in the cluster")
		}
		rand.Seed(time.Now().UnixNano())
		return fluxPluginPackagesClient, fluxPluginReposClient, nil
	}
}

func isLocalKindClusterUp(t *testing.T) (up bool, err error) {
	t.Logf("+isLocalKindClusterUp")

	out, err := execCommand(t, "", "kind", []string{"get", "clusters"})
	if err != nil {
		return false, err
	}
	words := strings.Split(out, " \n")
	found := false
	for _, word := range words {
		if word == "kubeapps" {
			found = true
		}
	}
	if !found {
		return false, nil
	}

	// naively assume that if the api server reports nodes, the cluster is up
	typedClient, err := kubeGetTypedClient()
	if err != nil {
		t.Logf("Failed to get typed client due to: %+v", err)
		return false, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultContextTimeout)
	defer cancel()

	nodeList, err := typedClient.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		t.Logf("Failed to get list of nodes due to %+v", err)
		return false, err
	}

	if len(nodeList.Items) == 1 || nodeList.Items[0].Name == "node/kubeapps-control-plane" {
		return true, nil
	} else {
		return false, fmt.Errorf("unexpected cluster nodes: [%v]", nodeList)
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
		return nil, nil, fmt.Errorf("failed to dial [%s] due to: %v", target, err)
	}
	t.Cleanup(func() { conn.Close() })
	pluginsCli := plugins.NewPluginsServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), defaultContextTimeout)
	defer cancel()
	response, err := pluginsCli.GetConfiguredPlugins(ctx, &plugins.GetConfiguredPluginsRequest{})

	if err != nil {
		return nil, nil, fmt.Errorf("failed to GetConfiguredPlugins due to: %v", err)
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

// This creates a flux helm repository CRD
func kubeAddHelmRepository(t *testing.T, name types.NamespacedName, typ, url, secretName string, interval time.Duration) error {
	t.Logf("+kubeAddHelmRepository(%s,%s,%s)", name, typ, url)
	if interval <= 0 {
		interval = time.Duration(10 * time.Minute)
	}
	repo := sourcev1.HelmRepository{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name.Name,
			Namespace: name.Namespace,
		},
		Spec: sourcev1.HelmRepositorySpec{
			URL:      url,
			Interval: metav1.Duration{Duration: interval},
		},
	}

	if typ != "" {
		repo.Spec.Type = typ
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
		t.Logf("Creating HelmRepository: %s\n...", common.PrettyPrint(repo))
		return ifc.Create(ctx, &repo)
	}
}

func kubeAddHelmRepositoryAndCleanup(t *testing.T, name types.NamespacedName, typ, url, secretName string, interval time.Duration) error {
	t.Logf("+kubeAddHelmRepositoryAndCleanup(%s)", name)
	err := kubeAddHelmRepository(t, name, typ, url, secretName, interval)
	if err == nil {
		t.Cleanup(func() {
			err := kubeDeleteHelmRepository(t, name)
			if err != nil {
				t.Logf("Failed to delete helm repository [%s] due to [%v]", name, err)
			}
		})
	}
	return err
}

func kubeGetHelmRepository(t *testing.T, name types.NamespacedName) (*sourcev1.HelmRepository, error) {
	t.Logf("+kubeGetHelmRepository(%s)", name)

	ctx, cancel := context.WithTimeout(context.Background(), defaultContextTimeout)
	defer cancel()
	if ifc, err := kubeGetCtrlClient(); err != nil {
		return nil, err
	} else {
		var repo sourcev1.HelmRepository
		if err := ifc.Get(ctx, name, &repo); err != nil {
			return nil, err
		}
		return &repo, nil
	}
}

func kubeListAllHelmRepositories(t *testing.T) (*sourcev1.HelmRepositoryList, error) {
	t.Logf("+kubeListAllHelmRepositories()")

	ctx, cancel := context.WithTimeout(context.Background(), defaultContextTimeout)
	defer cancel()
	if ifc, err := kubeGetCtrlClient(); err != nil {
		return nil, err
	} else {
		var repoList sourcev1.HelmRepositoryList
		if err := ifc.List(ctx, &repoList); err != nil {
			return nil, err
		}
		return &repoList, nil
	}
}

func kubeWaitUntilHelmRepositoryIsReady(t *testing.T, name types.NamespacedName) error {
	t.Logf("+kubeWaitUntilHelmRepositoryIsReady(%s)", name)
	defer func() {
		t.Logf("-kubeWaitUntilHelmRepositoryIsReady(%s)", name)
	}()

	if ifc, err := kubeGetCtrlClient(); err != nil {
		return err
	} else {
		ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
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
						t.Logf("[%d:%d:%d] Got event: type: [%v], name: [%s/%s], complete: [%t], success: [%t], reason: [%s]",
							hour, minute, second, event.Type, repo.Namespace, repo.Name, complete, success, reason)
						if name.Name == repo.Name && name.Namespace == repo.Namespace {
							if complete && success {
								return nil
							} else if complete && !success {
								return fmt.Errorf("%v", reason)
							}
						}
					}
				}
			}
		}
	}
}

// this should eventually be replaced with flux plugin's DeleteRepository()
func kubeDeleteHelmRepository(t *testing.T, name types.NamespacedName) error {
	t.Logf("+kubeDeleteHelmRepository(%s)", name)
	repo := &sourcev1.HelmRepository{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name.Name,
			Namespace: name.Namespace,
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

func kubeExistsHelmRepository(t *testing.T, name types.NamespacedName) (bool, error) {
	t.Logf("+kubeExistsHelmRepository(%s)", name)
	ctx, cancel := context.WithTimeout(context.Background(), defaultContextTimeout)
	defer cancel()
	var repo sourcev1.HelmRepository
	if ifc, err := kubeGetCtrlClient(); err != nil {
		return false, err
	} else if err = ifc.Get(ctx, name, &repo); err == nil {
		return true, nil
	} else {
		return false, nil
	}
}

func kubeDeleteHelmRelease(t *testing.T, name types.NamespacedName) error {
	t.Logf("+kubeDeleteHelmRelease(%s)", name)
	ctx, cancel := context.WithTimeout(context.Background(), defaultContextTimeout)
	defer cancel()
	release := &helmv2.HelmRelease{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name.Name,
			Namespace: name.Namespace,
		},
	}
	if ifc, err := kubeGetCtrlClient(); err != nil {
		return err
	} else {
		return ifc.Delete(ctx, release)
	}
}

func kubeExistsHelmRelease(t *testing.T, name types.NamespacedName) (bool, error) {
	t.Logf("+kubeExistsHelmRelease(%s)", name)
	ctx, cancel := context.WithTimeout(context.Background(), defaultContextTimeout)
	defer cancel()
	var rel helmv2.HelmRelease
	if ifc, err := kubeGetCtrlClient(); err != nil {
		return false, err
	} else if err = ifc.Get(ctx, name, &rel); err == nil {
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

func kubeCreateClusterRole(t *testing.T, name string) error {
	t.Logf("+kubeCreateClusterRole(%s)", name)
	typedClient, err := kubeGetTypedClient()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), defaultContextTimeout)
	defer cancel()
	_, err = typedClient.RbacV1().ClusterRoles().Create(ctx, &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}, metav1.CreateOptions{})
	return err
}

func kubeDeleteClusterRole(t *testing.T, name string) error {
	t.Logf("+kubeDeleteClusterRole(%s)", name)
	typedClient, err := kubeGetTypedClient()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), defaultContextTimeout)
	defer cancel()
	return typedClient.RbacV1().ClusterRoles().Delete(ctx, name, metav1.DeleteOptions{})
}

func kubeCreateRole(t *testing.T, name types.NamespacedName, rules []rbacv1.PolicyRule) error {
	t.Logf("+kubeCreateRole(%s)", name)
	typedClient, err := kubeGetTypedClient()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), defaultContextTimeout)
	defer cancel()
	_, err = typedClient.RbacV1().Roles(name.Namespace).Create(ctx, &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name: name.Name,
		},
		Rules: rules,
	}, metav1.CreateOptions{})
	return err
}

func kubeDeleteRole(t *testing.T, name types.NamespacedName) error {
	t.Logf("+kubeDeleteRole(%s)", name)
	typedClient, err := kubeGetTypedClient()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), defaultContextTimeout)
	defer cancel()
	return typedClient.RbacV1().Roles(name.Name).Delete(ctx, name.Namespace, metav1.DeleteOptions{})
}

func kubeCreateClusterRoleBinding(t *testing.T, name types.NamespacedName, role string) error {
	t.Logf("+kubeCreateClusterRoleBinding(%s,%s)", name, role)
	typedClient, err := kubeGetTypedClient()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultContextTimeout)
	defer cancel()

	_, err = typedClient.RbacV1().ClusterRoleBindings().Create(
		ctx,
		&rbacv1.ClusterRoleBinding{
			ObjectMeta: metav1.ObjectMeta{
				Name: name.Name + "-binding",
			},
			Subjects: []rbacv1.Subject{
				{
					Kind:      rbacv1.ServiceAccountKind,
					Name:      name.Name,
					Namespace: name.Namespace,
				},
			},
			RoleRef: rbacv1.RoleRef{
				Kind: "ClusterRole",
				Name: role,
			},
		},
		metav1.CreateOptions{})
	return err
}

func kubeCreateServiceAccount(t *testing.T, name types.NamespacedName) error {
	t.Logf("+kubeCreateServiceAccount(%s)", name)
	typedClient, err := kubeGetTypedClient()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultContextTimeout)
	defer cancel()

	_, err = typedClient.CoreV1().ServiceAccounts(name.Namespace).Create(
		ctx,
		&apiv1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name.Name,
				Namespace: name.Namespace,
			},
		},
		metav1.CreateOptions{})
	return err
}

func kubeCreateServiceAccountWithClusterRole(t *testing.T, name types.NamespacedName, role string) (string, error) {
	t.Logf("+kubeCreateServiceAccountWithClusterRole(%s,%s)", name, role)

	// https://itnext.io/big-change-in-k8s-1-24-about-serviceaccounts-and-their-secrets-4b909a4af4e0
	// and
	// https://github.com/vmware-tanzu/kubeapps/pull/4772
	// it used to be the case that creating service account would automatically create an
	// associated secret service account token
	// (per https://kubernetes.io/docs/tasks/configure-pod-container/configure-service-account/)
	// but starting with 1.24 it doesn't. So now I do it manually
	err := kubeCreateServiceAccount(t, name)
	if err != nil {
		return "", err
	}

	err = kubeCreateClusterRoleBinding(t, name, role)
	if err != nil {
		return "", err
	}

	secretName := types.NamespacedName{Name: name.Name + "-token", Namespace: name.Namespace}
	err = kubeCreateSecret(t, &apiv1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName.Name,
			Namespace: secretName.Namespace,
			Annotations: map[string]string{
				apiv1.ServiceAccountNameKey: name.Name,
			},
		},
		Type: apiv1.SecretTypeServiceAccountToken,
	})
	if err != nil {
		return "", err
	}

	var token string
	for i := 0; i < 10; i++ {
		token, err = kubeGetSecretToken(t, secretName, "token")
		if token != "" && err == nil {
			break
		}
		t.Logf("Waiting 1s for service account token in secret [%s] to be set up... [%d/%d]", secretName, i+1, 10)
		time.Sleep(1 * time.Second)
	}

	if token == "" {
		return "", fmt.Errorf("Failed to get token from secret: [%s]", secretName)
	}
	return token, nil
}

func kubeCreateServiceAccountWithRoles(t *testing.T, name types.NamespacedName, namespacesToRoles map[string]string) (string, error) {
	t.Logf("+kubeCreateServiceAccountWithRoles(%s,%s)", name, namespacesToRoles)
	err := kubeCreateServiceAccount(t, name)
	if err != nil {
		return "", err
	}

	typedClient, err := kubeGetTypedClient()
	if err != nil {
		return "", err
	}

	for ns, role := range namespacesToRoles {
		ctx, cancel := context.WithTimeout(context.Background(), defaultContextTimeout)
		defer cancel()

		_, err = typedClient.RbacV1().RoleBindings(ns).Create(
			ctx,
			&rbacv1.RoleBinding{
				ObjectMeta: metav1.ObjectMeta{
					Name:      name.Name + "-binding",
					Namespace: ns,
				},
				Subjects: []rbacv1.Subject{
					{
						Kind:      rbacv1.ServiceAccountKind,
						Name:      name.Name,
						Namespace: name.Namespace,
					},
				},
				RoleRef: rbacv1.RoleRef{
					Kind: "Role",
					Name: role,
				},
			},
			metav1.CreateOptions{})
		if err != nil {
			return "", err
		}
	}

	secretName := types.NamespacedName{Name: name.Name + "-token", Namespace: name.Namespace}
	err = kubeCreateSecret(t, &apiv1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName.Name,
			Namespace: secretName.Namespace,
			Annotations: map[string]string{
				apiv1.ServiceAccountNameKey: name.Name,
			},
		},
		Type: apiv1.SecretTypeServiceAccountToken,
	})
	if err != nil {
		return "", err
	}

	var token string
	for i := 0; i < 10; i++ {
		token, err = kubeGetSecretToken(t, secretName, "token")
		if token != "" && err == nil {
			break
		}
		t.Logf("Waiting 1s for service account token in secret [%s] to be set up... [%d/%d]", secretName, i+1, 10)
		time.Sleep(1 * time.Second)
	}

	if token == "" {
		return "", fmt.Errorf("Failed to get token from secret: [%s]", secretName)
	}
	return token, nil
}

// ref: https://kubernetes.io/docs/reference/access-authn-authz/rbac/#user-facing-roles
// will create a service account with cluster-admin privs and return the associated
// Bearer token (base64-encoded)
func kubeCreateAdminServiceAccount(t *testing.T, name types.NamespacedName) (string, error) {
	return kubeCreateServiceAccountWithClusterRole(t, name, "cluster-admin")
}

func kubeCreateFluxPluginServiceAccount(t *testing.T, name types.NamespacedName) (string, error) {
	return kubeCreateServiceAccountWithClusterRole(t, name, "kubeapps:controller:kubeapps-apis-fluxv2-plugin")
}

func kubeDeleteServiceAccountWithClusterRoleBinding(t *testing.T, name types.NamespacedName) error {
	t.Logf("+kubeDeleteServiceAccountWithClusterRoleBinding(%s)", name)
	typedClient, err := kubeGetTypedClient()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), defaultContextTimeout)
	defer cancel()
	err = typedClient.RbacV1().ClusterRoleBindings().Delete(
		ctx,
		name.Name+"-binding",
		metav1.DeleteOptions{})
	if err != nil {
		return err
	}
	ctx, cancel = context.WithTimeout(context.Background(), defaultContextTimeout)
	defer cancel()
	err = typedClient.CoreV1().ServiceAccounts(name.Namespace).Delete(
		ctx,
		name.Name,
		metav1.DeleteOptions{})
	if err != nil {
		return err
	}
	return nil
}

func kubeDeleteServiceAccountWithRoleBindings(t *testing.T, name types.NamespacedName, nsToRole map[string]string) error {
	t.Logf("+kubeDeleteServiceAccountWithRoleBindings(%s,%s)", name, nsToRole)
	typedClient, err := kubeGetTypedClient()
	if err != nil {
		return err
	}
	for ns := range nsToRole {
		ctx, cancel := context.WithTimeout(context.Background(), defaultContextTimeout)
		defer cancel()
		err = typedClient.RbacV1().RoleBindings(ns).Delete(
			ctx,
			name.Name+"-binding",
			metav1.DeleteOptions{})
		if err != nil {
			return err
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultContextTimeout)
	defer cancel()
	err = typedClient.CoreV1().ServiceAccounts(name.Namespace).Delete(
		ctx,
		name.Name,
		metav1.DeleteOptions{})
	if err != nil {
		return err
	}
	return nil
}

func kubeCreateNamespaceAndCleanup(t *testing.T, namespace string) error {
	t.Logf("+kubeCreateNamespace(%s)", namespace)
	typedClient, err := kubeGetTypedClient()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), defaultContextTimeout)
	defer cancel()
	if _, err = typedClient.CoreV1().Namespaces().Create(
		ctx,
		&apiv1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: namespace,
			},
		},
		metav1.CreateOptions{}); err == nil {
		t.Cleanup(func() {
			if err := kubeDeleteNamespace(t, namespace); err != nil {
				t.Logf("Failed to delete namespace [%s] due to [%v]", namespace, err)
			}
		})
	}
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

func kubeGetSecretToken(t *testing.T, name types.NamespacedName, dataKey string) (string, error) {
	t.Logf("+kubeGetSecretToken(%s, %s)", name, dataKey)
	if secret, err := kubeGetSecret(t, name); err == nil && secret != nil {
		token := secret.Data[dataKey]
		if token == nil {
			return "", errors.New("No data found")
		}
		return string(token), nil
	} else {
		return "", err
	}
}

func kubeCreateSecret(t *testing.T, secret *apiv1.Secret) error {
	t.Logf("+kubeCreateSecret(%s, %s)", secret.Namespace, secret.Name)
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

func kubeSetSecretOwnerRef(t *testing.T, secretName types.NamespacedName, ownerRepo *sourcev1.HelmRepository) error {
	t.Logf("+kubeSetSecretOwnerRef(%s, %s)", secretName, ownerRepo.Name)
	typedClient, err := kubeGetTypedClient()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultContextTimeout)
	defer cancel()

	secretsInterface := typedClient.CoreV1().Secrets(secretName.Namespace)
	secret, err := secretsInterface.Get(
		ctx,
		secretName.Name,
		metav1.GetOptions{})
	if err != nil {
		return err
	}

	secret.OwnerReferences = []metav1.OwnerReference{
		*metav1.NewControllerRef(
			ownerRepo,
			schema.GroupVersionKind{
				Group:   sourcev1.GroupVersion.Group,
				Version: sourcev1.GroupVersion.Version,
				Kind:    sourcev1.HelmRepositoryKind,
			}),
	}

	if _, err := secretsInterface.Update(ctx, secret, metav1.UpdateOptions{}); err != nil {
		return err
	} else {
		return nil
	}
}

func kubeCreateSecretAndCleanup(t *testing.T, secret *apiv1.Secret) error {
	secretName := types.NamespacedName{Name: secret.Name, Namespace: secret.Namespace}
	t.Logf("+kubeCreateSecretAndCleanup(%s)", secretName)
	err := kubeCreateSecret(t, secret)
	if err != nil {
		return err
	}
	t.Cleanup(func() {
		err := kubeDeleteSecret(t, secretName)
		if err != nil {
			t.Logf("Failed to delete secret [%s] due to [%v]", secretName, err)
		}
	})
	return nil
}

func kubeDeleteSecret(t *testing.T, name types.NamespacedName) error {
	t.Logf("+kubeDeleteSecret(%s)", name)
	typedClient, err := kubeGetTypedClient()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), defaultContextTimeout)
	defer cancel()
	return typedClient.CoreV1().Secrets(name.Namespace).Delete(
		ctx,
		name.Name,
		metav1.DeleteOptions{})
}

func kubeGetSecret(t *testing.T, name types.NamespacedName) (*apiv1.Secret, error) {
	t.Logf("+kubeGetSecret(%s)", name)
	typedClient, err := kubeGetTypedClient()
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), defaultContextTimeout)
	defer cancel()
	secret, err := typedClient.CoreV1().Secrets(name.Namespace).Get(
		ctx,
		name.Name,
		metav1.GetOptions{})
	if err != nil {
		return nil, err
	} else {
		return secret, nil
	}
}

func kubeExistsSecret(t *testing.T, name types.NamespacedName) (bool, error) {
	t.Logf("+kubeExistsSecret(%s)", name)
	secret, err := kubeGetSecret(t, name)
	return err == nil && secret != nil, nil
}

func kubePortForwardToPod(t *testing.T, name types.NamespacedName, ports string) error {
	t.Logf("+kubePortForwardToPod(%s,%s)", name, ports)
	defer t.Logf("-kubePortForwardToPod")
	stopChan, readyChan := make(chan struct{}, 1), make(chan struct{}, 1)
	go func() {
		if err := func() error {
			// ref https://github.com/kubernetes/client-go/issues/51
			if config, err := restConfig(); err != nil {
				return err
			} else if roundTripper, upgrader, err := spdy.RoundTripperFor(config); err != nil {
				return err
			} else {
				path := fmt.Sprintf("/api/v1/namespaces/%s/pods/%s/portforward", name.Namespace, name.Name)
				hostIP := strings.TrimLeft(config.Host, "htps:/")
				serverURL := url.URL{Scheme: "https", Path: path, Host: hostIP}
				dialer := spdy.NewDialer(upgrader, &http.Client{Transport: roundTripper}, http.MethodPost, &serverURL)
				out, errOut := new(bytes.Buffer), new(bytes.Buffer)
				if forwarder, err := portforward.New(dialer, []string{ports}, stopChan, readyChan, out, errOut); err != nil {
					return err
				} else {
					go func() {
						for range readyChan { // Kubernetes will close this channel when it has something to tell us.
						}
						if len(errOut.String()) != 0 {
							t.Errorf("kubePortForwardToPod:\n%s", errOut.String())
						} else if len(out.String()) != 0 {
							t.Logf("kubePortForwardToPod:\n%s", out.String())
						}
					}()
					if err = forwarder.ForwardPorts(); err != nil { // Locks until stopChan is closed.
						return err
					}
				}
			}
			return nil
		}(); err != nil {
			t.Error(err)
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

func kubePortForwardToRedis(t *testing.T) error {
	t.Logf("+kubePortForwardToRedis")
	return kubePortForwardToPod(t, types.NamespacedName{
		Name:      "kubeapps-redis-master-0",
		Namespace: "kubeapps"},
		"6379")
}

func kubePortForwardToFluxTestdataApp(t *testing.T) error {
	t.Logf("+kubePortForwardToFluxTestdataApp")
	podName, err := getFluxPluginTestdataPodName()
	if err != nil {
		t.Fatal(err)
	}
	return kubePortForwardToPod(t, *podName, "50057:80")
}

// ref https://stackoverflow.com/questions/51686986/how-to-copy-file-to-container-with-kubernetes-client-go
// example kubectl cp /tmp/foo.txt default/fluxv2plugin-testdata-app-7f7dd58796-w2qbg:/
func kubeCopyFileToPod(t *testing.T, srcFile string, podName types.NamespacedName, destFile string) error {
	t.Logf("+kubeCopyFileToPod(%s, %s, %s)", srcFile, podName, destFile)
	ioStreams, _, _, _ := genericclioptions.NewTestIOStreams()
	copyOptions := cp.NewCopyOptions(ioStreams)
	restcfg, err := restConfig()
	if err != nil {
		return err
	}
	restcfg.APIPath = "/api"                                   // Make sure we target /api and not just /
	restcfg.GroupVersion = &schema.GroupVersion{Version: "v1"} // this targets the core api groups so the url path will be /api/v1
	restcfg.NegotiatedSerializer = serializer.WithoutConversionCodecFactory{CodecFactory: scheme.Codecs}
	copyOptions.ClientConfig = restcfg
	typedcli, err := kubeGetTypedClient()
	if err != nil {
		return err
	}
	copyOptions.Clientset = typedcli
	destSpec := fmt.Sprintf("%s/%s:%s", podName.Namespace, podName.Name, destFile)
	err = copyOptions.Run([]string{srcFile, destSpec})
	if err != nil {
		return fmt.Errorf("Could not run copy operation: %v", err)
	}
	return nil
}

func kubeGetCtrlClient() (ctrlclient.WithWatch, error) {
	if ctrlClient != nil {
		return ctrlClient, nil
	} else {
		if config, err := restConfig(); err != nil {
			return nil, err
		} else {
			scheme := runtime.NewScheme()
			err = sourcev1.AddToScheme(scheme)
			if err != nil {
				return nil, err
			}
			err = helmv2.AddToScheme(scheme)
			if err != nil {
				return nil, err
			}

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

func newGrpcAdminContext(t *testing.T, name types.NamespacedName) (context.Context, error) {
	token, err := kubeCreateAdminServiceAccount(t, name)
	if err != nil {
		return nil, fmt.Errorf("Failed to create service account due to: %+v", err)
	}
	t.Cleanup(func() {
		if err := kubeDeleteServiceAccountWithClusterRoleBinding(t, name); err != nil {
			t.Logf("Failed to delete service account due to: %+v", err)
		}
	})
	return newGrpcContext(t, token), nil
}

func newGrpcFluxPluginContext(t *testing.T, name types.NamespacedName) (context.Context, error) {
	token, err := kubeCreateFluxPluginServiceAccount(t, name)
	if err != nil {
		return nil, fmt.Errorf("Failed to create service account [%s] due to: %+v", name, err)
	}
	t.Cleanup(func() {
		if err := kubeDeleteServiceAccountWithClusterRoleBinding(t, name); err != nil {
			t.Logf("Failed to delete service account [%s] due to: %+v", name, err)
		}
	})
	return newGrpcContext(t, token), nil
}

func kubectlCanI(t *testing.T, name types.NamespacedName, verb, resource, checkThisNamespace string) string {
	args := []string{
		"auth",
		"can-i",
		verb,
		resource,
		"--namespace",
		checkThisNamespace,
		"--as",
		"system:serviceaccount:" + name.Namespace + ":" + name.Name,
	}

	out, _ := execCommand(t, "", "kubectl", args)
	return out
}

func newGrpcContextForServiceAccountWithoutAccessToAnyNamespace(t *testing.T, name types.NamespacedName) (context.Context, error) {
	role := name.Name + "-cluster-role"
	if err := kubeCreateClusterRole(t, role); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := kubeDeleteClusterRole(t, role); err != nil {
			t.Logf("Failed to delete cluster role [%s] due to: %+v", role, err)
		}
	})

	token, err := kubeCreateServiceAccountWithClusterRole(t, name, role)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := kubeDeleteServiceAccountWithClusterRoleBinding(t, name); err != nil {
			t.Logf("Failed to delete service account [%s] due to: %+v", name, err)
		}
	})
	return newGrpcContext(t, token), nil
}

func newGrpcContextForServiceAccountWithRules(t *testing.T, name types.NamespacedName, namespaceToRules map[string][]rbacv1.PolicyRule) (context.Context, error) {
	nsToRole := make(map[string]string)
	for ns, rules := range namespaceToRules {
		role := types.NamespacedName{
			Name:      name.Name + "-" + ns + "-role",
			Namespace: ns,
		}
		if err := kubeCreateRole(t, role, rules); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() {
			if err := kubeDeleteRole(t, role); err != nil {
				t.Logf("Failed to delete role [%s] due to: %+v", role, err)
			}
		})
		nsToRole[ns] = role.Name
	}

	token, err := kubeCreateServiceAccountWithRoles(t, name, nsToRole)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := kubeDeleteServiceAccountWithRoleBindings(t, name, nsToRole); err != nil {
			t.Logf("Failed to delete service account [%s] due to: %+v", name, err)
		}
	})

	return newGrpcContext(t, token), nil
}

func redisCheckTinyMaxMemory(t *testing.T, redisCli *redis.Client, expectedMaxMemory string) error {
	maxmemory, err := redisCli.ConfigGet(redisCli.Context(), "maxmemory").Result()
	if err != nil {
		return err
	} else {
		currentMaxMemory := fmt.Sprintf("%v", maxmemory[1])
		t.Logf("Current redis maxmemory = [%s]", currentMaxMemory)
		if currentMaxMemory != expectedMaxMemory {
			return fmt.Errorf("This test requires redis config maxmemory to be set to %s", expectedMaxMemory)
		}
	}
	maxmemoryPolicy, err := redisCli.ConfigGet(redisCli.Context(), "maxmemory-policy").Result()
	if err != nil {
		return err
	} else {
		currentMaxMemoryPolicy := fmt.Sprintf("%v", maxmemoryPolicy[1])
		t.Logf("Current maxmemory policy = [%s]", currentMaxMemoryPolicy)
		if currentMaxMemoryPolicy != "allkeys-lfu" {
			return fmt.Errorf("This test requires redis config maxmemory-policy to be set to allkeys-lfu")
		}
	}
	return nil
}

func newRedisClientForIntegrationTest(t *testing.T) (*redis.Client, error) {
	if err := kubePortForwardToRedis(t); err != nil {
		return nil, fmt.Errorf("kubePortForwardToRedis failed due to %+v", err)
	}
	name := types.NamespacedName{
		Name:      "kubeapps-redis",
		Namespace: "kubeapps",
	}
	redisPwd, err := kubeGetSecretToken(t, name, "redis-password")
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
		return nil, err
	} else {
		if len(keys) != 0 {
			return nil, fmt.Errorf("Failing due to unexpected state of the cache. Current keys: %s", keys)
		}
	}
	return redisCli, nil
}

func redisReceiveNotificationsLoop(t *testing.T, ch <-chan *redis.Message, sem *semaphore.Weighted, evictedRepos *sets.String) {
	if totalBitnamiCharts == -1 {
		t.Errorf("Error: unexpected state: number of charts in bitnami catalog is not initialized")
		t.Fail()
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

func usesBitnamiCatalog(t *testing.T) error {
	t.Logf("+usesBitnamiCatalog")

	if totalBitnamiCharts == -1 {
		// just need to do this once
		err := kubePortForwardToFluxTestdataApp(t)
		if err != nil {
			return err
		}

		byteArray, err := httpclient.Get(outside_cluster_bitnami_url+"/index.yaml", httpclient.New(), nil)
		if err != nil {
			return err
		}

		modelRepo := &models.Repo{
			Namespace: "default",
			Name:      "bitnami",
			URL:       outside_cluster_bitnami_url,
			Type:      "helm",
		}

		charts, err := helm.ChartsFromIndex(byteArray, modelRepo, true)
		if err != nil {
			return err
		}
		totalBitnamiCharts = len(charts)
		t.Logf("-usesBitnamiCatalog: total [%d] charts", totalBitnamiCharts)
	}
	return nil
}

func getFluxPluginTestdataPodName() (*types.NamespacedName, error) {
	cli, err := kubeGetTypedClient()
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), defaultContextTimeout)
	defer cancel()
	podList, err := cli.CoreV1().Pods("default").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	for _, p := range podList.Items {
		if strings.HasPrefix(p.Name, "fluxv2plugin-testdata-app-") {
			return &types.NamespacedName{
				Name:      p.Name,
				Namespace: p.Namespace}, nil
		}
	}
	return nil, fmt.Errorf("fluxplugin testdata pod not found")
}

func helmPushChartToMyGithubRegistry(t *testing.T, version string) error {
	t.Logf("+helmPushChartToMyGithubRegistry(%s)", version)
	defer t.Logf("-helmPushChartToMyGithubRegistry(%s)", version)

	args := []string{
		"pushChartToMyGithub",
		version,
	}

	// use the CLI for now
	_, err := execCommand(t, "./testdata", "./integ-test-env.sh", args)
	return err
}

func deleteChartFromMyGithubRegistry(t *testing.T, version string) error {
	t.Logf("+deleteChartFromMyGithubRegistry(%s)", version)
	defer t.Logf("-deleteChartFromMyGithubRegistry(%s)", version)

	args := []string{
		"deleteChartVersionFromMyGitHub",
		"6.1.6",
	}

	// use the CLI for now
	_, err := execCommand(t, "./testdata", "./integ-test-env.sh", args)
	return err
}

func setupHarborStefanProdanClone(t *testing.T) error {
	t.Logf("+setupHarborStefanProdanClone()")
	defer t.Logf("-setupHarborStefanProdanClone()")

	args := []string{
		"setupHarborStefanProdanClone",
		"--quick",
	}

	// use the CLI for now
	_, err := execCommand(t, "./testdata", "./integ-test-env.sh", args)
	return err
}

func setupHarborRobotAccount(t *testing.T) (string, string, error) {
	t.Logf("+setupHarborRobotAccount()")
	defer t.Logf("-setupHarborRobotAccount()")

	args := []string{
		"setupHarborRobotAccount",
	}

	// use the CLI for now
	out, err := execCommand(t, "./testdata", "./integ-test-env.sh", args)
	if err != nil {
		return "", "", err
	} else {
		i := strings.Index(out, "Robot account successfully created: [")
		if i >= 0 {
			out2 := out[i+37:]
			j := strings.Index(out2, "]")
			if j >= 0 {
				out3 := out2[:j]
				strs := strings.SplitN(out3, " ", 2)
				if len(strs) == 2 {
					return strs[0], strs[1], nil
				}
			}
		}
		return "", "", fmt.Errorf("unexpected response: %s", out)
	}
}

// ref https://cloud.google.com/artifact-registry/docs/helm/store-helm-charts#auth-token
// this token lasts 60 mins
func gcloudPrintAccessToken(t *testing.T) (string, error) {
	credFile := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	if credFile == "" {
		t.Fatalf("Environment variable [GOOGLE_APPLICATION_CREDENTIALS] needs to be set to run this test")
	}
	args := []string{
		"auth",
		"application-default",
		"print-access-token",
	}
	return execCommand(t, ".", "gcloud", args)
}

func execCommand(t *testing.T, dir, name string, args []string) (string, error) {
	t.Logf("About to execute command: [%s] with args %s...", name, args)
	// TODO (gfichtenholt) it'd be nice to have real-time updates
	cmd := exec.Command(name, args...)
	if dir != "" {
		cmd.Dir = dir
	}
	byteArray, err := cmd.CombinedOutput()
	out := strings.Trim(string(byteArray), "\n")
	t.Logf("Executed command: [%s], err: [%v], output: [\n%s\n]", cmd.String(), err, out)
	return out, err
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
