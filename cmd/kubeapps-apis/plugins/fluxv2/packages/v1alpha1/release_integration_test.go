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
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	corev1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
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

// This is an integration test: it tests the full integration of flux plugin with flux back-end
// pre-requisites for these tests to run:
// 1) kind cluster with flux deployed
// 2) kubeapps apis apiserver service running with fluxv2 plug-in enabled, port forwarded to 8080, e.g.
//      kubectl -n kubeapps port-forward svc/kubeapps-internal-kubeappsapis 8080:8080
//    Didn't want to spend cycles writing port-forwarding code programmatically like https://github.com/anthhub/forwarder
//    at this point.

// if one or more of the above pre-requisites is not satisfied, the tests are simply skipped

func TestKindClusterCreateInstalledPackage(t *testing.T) {
	testCases := []struct {
		testName          string
		repoUrl           string
		request           *corev1.CreateInstalledPackageRequest
		expectedDetail    *corev1.InstalledPackageDetail
		expectedPodPrefix string
	}{
		{
			testName: "create test (simplest case)",
			// TODO: (gfichtenholt) stand up a pod that serves podinfo-index.yaml within the cluster
			// instead of relying on github.io
			repoUrl:           "https://stefanprodan.github.io/podinfo",
			request:           create_request_basic,
			expectedDetail:    expected_detail_basic,
			expectedPodPrefix: "@TARGET_NS@-my-podinfo-",
		},
		{
			testName:          "create package (semver constraint)",
			repoUrl:           "https://stefanprodan.github.io/podinfo",
			request:           create_request_semver_constraint,
			expectedDetail:    expected_detail_semver_constraint,
			expectedPodPrefix: "@TARGET_NS@-my-podinfo-2-",
		},
		{
			testName:          "create package (reconcile options)",
			repoUrl:           "https://stefanprodan.github.io/podinfo",
			request:           create_request_reconcile_options,
			expectedDetail:    expected_detail_reconcile_options,
			expectedPodPrefix: "@TARGET_NS@-my-podinfo-3-",
		},
		{
			testName:          "create package (with values)",
			repoUrl:           "https://stefanprodan.github.io/podinfo",
			request:           create_request_with_values,
			expectedDetail:    expected_detail_with_values,
			expectedPodPrefix: "@TARGET_NS@-my-podinfo-4-",
		},
	}

	if up, err := isLocalKindClusterUp(t); err != nil || !up {
		t.Skipf("skipping tests because due to failure to find local kind cluster: [%v]", err)
	}
	var fluxPluginClient fluxplugin.FluxV2PackagesServiceClient
	var err error
	if fluxPluginClient, err = getFluxPluginClient(t); err != nil {
		t.Skipf("skipping tests due to failure to get fluxv2 plugin: [%v]", err)
	}

	rand.Seed(time.Now().UnixNano())

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			availablePackageRef := tc.request.AvailablePackageRef
			idParts := strings.Split(availablePackageRef.Identifier, "/")
			err = kubeCreateHelmRepository(t, idParts[0], tc.repoUrl, availablePackageRef.Context.Namespace)
			if err != nil {
				t.Fatalf("%+v", err)
			}
			t.Cleanup(func() {
				err = kubeDeleteHelmRepository(t, idParts[0], availablePackageRef.Context.Namespace)
				if err != nil {
					t.Logf("Failed to delete helm source due to [%v]", err)
				}
			})

			// need to wait until repo is index by flux plugin
			const maxWait = 25
			for i := 0; i < maxWait; i++ {
				_, err := fluxPluginClient.GetAvailablePackageDetail(
					context.TODO(),
					&corev1.GetAvailablePackageDetailRequest{AvailablePackageRef: availablePackageRef})
				if err == nil {
					break
				} else if i == maxWait-1 {
					t.Fatalf("Timed out waiting for available package [%s], last error: [%v]", availablePackageRef, err)
				} else {
					t.Logf("waiting 1s for repository [%s] to be indexed, attempt [%d/%d]...", idParts[0], i+1, maxWait)
					time.Sleep(1 * time.Second)
				}
			}

			if tc.request.ReconciliationOptions != nil && tc.request.ReconciliationOptions.ServiceAccountName != "" {
				err = kubeCreateServiceAccount(t, tc.request.ReconciliationOptions.ServiceAccountName, "kubeapps")
				if err != nil {
					t.Fatalf("%+v", err)
				}
				t.Cleanup(func() {
					err = kubeDeleteServiceAccount(t, tc.request.ReconciliationOptions.ServiceAccountName, "kubeapps")
					if err != nil {
						t.Logf("Failed to delete service account due to [%v]", err)
					}
				})
			}

			// generate a unique target namespace for each test to avoid situations when tests are
			// run multiple times in a row and they fail due to the fact that the specified namespace
			// in in 'Terminating' state
			if tc.request.TargetContext.Namespace != "" {
				tc.request.TargetContext.Namespace += "-" + randSeq(4)
			}

			resp, err := fluxPluginClient.CreateInstalledPackage(context.TODO(), tc.request)
			if err != nil {
				t.Fatalf("%+v", err)
			}

			installedPackageRef := resp.InstalledPackageRef
			opts := cmpopts.IgnoreUnexported(
				corev1.InstalledPackageDetail{},
				corev1.InstalledPackageReference{},
				plugins.Plugin{},
				corev1.Context{})
			if got, want := installedPackageRef, tc.expectedDetail.InstalledPackageRef; !cmp.Equal(want, got, opts) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, opts))
			}

			t.Cleanup(func() {
				err = kubeDeleteHelmRelease(t, installedPackageRef.Identifier, installedPackageRef.Context.Namespace)
				if err != nil {
					t.Logf("Failed to delete helm release due to [%v]", err)
				}
				err = kubeDeleteNamespace(t, tc.request.TargetContext.Namespace)
				if err != nil {
					t.Logf("Failed to delete namespace [%s] due to [%v]", tc.request.TargetContext.Namespace, err)
				}
			})

			var actualDetail *corev1.InstalledPackageDetail
			for i := 0; i < maxWait; i++ {
				resp2, err := fluxPluginClient.GetInstalledPackageDetail(
					context.TODO(),
					&corev1.GetInstalledPackageDetailRequest{InstalledPackageRef: installedPackageRef})
				if err != nil {
					t.Fatalf("%+v", err)
				}

				if resp2.InstalledPackageDetail.Status.Ready == true &&
					resp2.InstalledPackageDetail.Status.Reason == corev1.InstalledPackageStatus_STATUS_REASON_INSTALLED {
					actualDetail = resp2.InstalledPackageDetail
					break
				} else {
					t.Logf("waiting 500ms due to: [%s], userReason: [%s], attempt [%d/%d]...",
						resp2.InstalledPackageDetail.Status.Reason, resp2.InstalledPackageDetail.Status.UserReason, i+1, maxWait)
					time.Sleep(500 * time.Millisecond)
				}
			}
			tc.expectedDetail.PostInstallationNotes = strings.ReplaceAll(
				tc.expectedDetail.PostInstallationNotes, "@TARGET_NS@", tc.request.TargetContext.Namespace)

			opts = cmpopts.IgnoreUnexported(
				corev1.GetInstalledPackageDetailResponse{},
				corev1.InstalledPackageDetail{},
				corev1.InstalledPackageReference{},
				corev1.Context{},
				corev1.VersionReference{},
				corev1.InstalledPackageStatus{},
				corev1.PackageAppVersion{},
				plugins.Plugin{},
				corev1.ReconciliationOptions{},
				corev1.AvailablePackageReference{})
			if got, want := actualDetail, tc.expectedDetail; !cmp.Equal(want, got, opts) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, opts))
			}

			// check artifacts in target namespace:
			tc.expectedPodPrefix = strings.ReplaceAll(
				tc.expectedPodPrefix, "@TARGET_NS@", tc.request.TargetContext.Namespace)
			pods, err := kubeGetPodNames(t, tc.request.TargetContext.Namespace)
			if err != nil {
				t.Fatalf("%+v", err)
			}
			if len(pods) != 1 {
				t.Errorf("expected 1 pod, got: %s", pods)
			} else if !strings.HasPrefix(pods[0], tc.expectedPodPrefix) {
				t.Errorf("expected pod with prefix [%s] not found in namespace [%s], pods found: [%v]",
					tc.expectedPodPrefix, tc.request.TargetContext.Namespace, pods)
			}
		})
	}
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
var letters = []rune("abcdefghijklmnopqrstuvwxyz0123456789")

var typedClient kubernetes.Interface
var dynamicClient dynamic.Interface

var create_request_basic = &corev1.CreateInstalledPackageRequest{
	AvailablePackageRef: &corev1.AvailablePackageReference{
		Identifier: "podinfo-1/podinfo",
		Context: &corev1.Context{
			Namespace: "default",
		},
	},
	Name: "my-podinfo",
	TargetContext: &corev1.Context{
		Namespace: "test-1",
	},
}

var expected_detail_basic = &corev1.InstalledPackageDetail{
	InstalledPackageRef: &corev1.InstalledPackageReference{
		Context: &corev1.Context{
			Namespace: "kubeapps",
		},
		Identifier: "my-podinfo",
		Plugin:     fluxPlugin,
	},
	PkgVersionReference: &corev1.VersionReference{
		Version: "*",
	},
	Name: "my-podinfo",
	CurrentVersion: &corev1.PackageAppVersion{
		PkgVersion: "6.0.0",
		AppVersion: "6.0.0",
	},
	ReconciliationOptions: &corev1.ReconciliationOptions{
		Interval: 60,
	},
	Status: &corev1.InstalledPackageStatus{
		Ready:      true,
		Reason:     corev1.InstalledPackageStatus_STATUS_REASON_INSTALLED,
		UserReason: "ReconciliationSucceeded: Release reconciliation succeeded",
	},
	PostInstallationNotes: "1. Get the application URL by running these commands:\n  " +
		"echo \"Visit http://127.0.0.1:8080 to use your application\"\n  " +
		"kubectl -n @TARGET_NS@ port-forward deploy/@TARGET_NS@-my-podinfo 8080:9898\n",
	AvailablePackageRef: &corev1.AvailablePackageReference{
		Identifier: "podinfo-1/podinfo",
		Context: &corev1.Context{
			Namespace: "default",
		},
		Plugin: fluxPlugin,
	},
}

var create_request_semver_constraint = &corev1.CreateInstalledPackageRequest{
	AvailablePackageRef: &corev1.AvailablePackageReference{
		Identifier: "podinfo-2/podinfo",
		Context: &corev1.Context{
			Namespace: "default",
		},
	},
	Name: "my-podinfo-2",
	TargetContext: &corev1.Context{
		Namespace: "test-2",
	},
	PkgVersionReference: &corev1.VersionReference{
		Version: "> 5",
	},
}

var expected_detail_semver_constraint = &corev1.InstalledPackageDetail{
	InstalledPackageRef: &corev1.InstalledPackageReference{
		Context: &corev1.Context{
			Namespace: "kubeapps",
		},
		Identifier: "my-podinfo-2",
		Plugin:     fluxPlugin,
	},
	PkgVersionReference: &corev1.VersionReference{
		Version: "> 5",
	},
	Name: "my-podinfo-2",
	CurrentVersion: &corev1.PackageAppVersion{
		PkgVersion: "6.0.0",
		AppVersion: "6.0.0",
	},
	ReconciliationOptions: &corev1.ReconciliationOptions{
		Interval: 60,
	},
	Status: &corev1.InstalledPackageStatus{
		Ready:      true,
		Reason:     corev1.InstalledPackageStatus_STATUS_REASON_INSTALLED,
		UserReason: "ReconciliationSucceeded: Release reconciliation succeeded",
	},
	PostInstallationNotes: "1. Get the application URL by running these commands:\n  " +
		"echo \"Visit http://127.0.0.1:8080 to use your application\"\n  " +
		"kubectl -n @TARGET_NS@ port-forward deploy/@TARGET_NS@-my-podinfo-2 8080:9898\n",
	AvailablePackageRef: &corev1.AvailablePackageReference{
		Identifier: "podinfo-2/podinfo",
		Context: &corev1.Context{
			Namespace: "default",
		},
		Plugin: fluxPlugin,
	},
}

var create_request_reconcile_options = &corev1.CreateInstalledPackageRequest{
	AvailablePackageRef: &corev1.AvailablePackageReference{
		Identifier: "podinfo-3/podinfo",
		Context: &corev1.Context{
			Namespace: "default",
		},
	},
	Name: "my-podinfo-3",
	TargetContext: &corev1.Context{
		Namespace: "test-3",
	},
	ReconciliationOptions: &corev1.ReconciliationOptions{
		Interval:           60,
		Suspend:            false,
		ServiceAccountName: "foo",
	},
}

var expected_detail_reconcile_options = &corev1.InstalledPackageDetail{
	InstalledPackageRef: &corev1.InstalledPackageReference{
		Context: &corev1.Context{
			Namespace: "kubeapps",
		},
		Identifier: "my-podinfo-3",
		Plugin:     fluxPlugin,
	},
	PkgVersionReference: &corev1.VersionReference{
		Version: "*",
	},
	Name: "my-podinfo-3",
	CurrentVersion: &corev1.PackageAppVersion{
		PkgVersion: "6.0.0",
		AppVersion: "6.0.0",
	},
	ReconciliationOptions: &corev1.ReconciliationOptions{
		Interval:           60,
		Suspend:            false,
		ServiceAccountName: "foo",
	},
	Status: &corev1.InstalledPackageStatus{
		Ready:      true,
		Reason:     corev1.InstalledPackageStatus_STATUS_REASON_INSTALLED,
		UserReason: "ReconciliationSucceeded: Release reconciliation succeeded",
	},
	PostInstallationNotes: "1. Get the application URL by running these commands:\n  " +
		"echo \"Visit http://127.0.0.1:8080 to use your application\"\n  " +
		"kubectl -n @TARGET_NS@ port-forward deploy/@TARGET_NS@-my-podinfo-3 8080:9898\n",
	AvailablePackageRef: &corev1.AvailablePackageReference{
		Identifier: "podinfo-3/podinfo",
		Context: &corev1.Context{
			Namespace: "default",
		},
		Plugin: fluxPlugin,
	},
}

var create_request_with_values = &corev1.CreateInstalledPackageRequest{
	AvailablePackageRef: &corev1.AvailablePackageReference{
		Identifier: "podinfo-4/podinfo",
		Context: &corev1.Context{
			Namespace: "default",
		},
	},
	Name: "my-podinfo-4",
	TargetContext: &corev1.Context{
		Namespace: "test-4",
	},
	Values: "{\"ui\": { \"message\": \"what we do in the shadows\" } }",
}

var expected_detail_with_values = &corev1.InstalledPackageDetail{
	InstalledPackageRef: &corev1.InstalledPackageReference{
		Context: &corev1.Context{
			Namespace: "kubeapps",
		},
		Identifier: "my-podinfo-4",
		Plugin:     fluxPlugin,
	},
	Name: "my-podinfo-4",
	CurrentVersion: &corev1.PackageAppVersion{
		PkgVersion: "6.0.0",
		AppVersion: "6.0.0",
	},
	PkgVersionReference: &corev1.VersionReference{
		Version: "*",
	},
	ReconciliationOptions: &corev1.ReconciliationOptions{
		Interval: 60,
	},
	Status: &corev1.InstalledPackageStatus{
		Ready:      true,
		Reason:     corev1.InstalledPackageStatus_STATUS_REASON_INSTALLED,
		UserReason: "ReconciliationSucceeded: Release reconciliation succeeded",
	},
	PostInstallationNotes: "1. Get the application URL by running these commands:\n  " +
		"echo \"Visit http://127.0.0.1:8080 to use your application\"\n  " +
		"kubectl -n @TARGET_NS@ port-forward deploy/@TARGET_NS@-my-podinfo-4 8080:9898\n",
	AvailablePackageRef: &corev1.AvailablePackageReference{
		Identifier: "podinfo-4/podinfo",
		Context: &corev1.Context{
			Namespace: "default",
		},
		Plugin: fluxPlugin,
	},
	ValuesApplied: "{\"ui\":{\"message\":\"what we do in the shadows\"}}",
}
