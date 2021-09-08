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
)

const (
	k8s_context = "kind-kubeapps"
)

// pre-requisites for these tests to run:
// 1) kind cluster with flux deployed
// 2) kubeapps apis apiserver service running with fluxv2 plug-in enabled, port forwarded to 8080
// 3) kubectl CLI on client side
// 4) flux CLI on client side

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
			repoUrl: "https://stefanprodan.github.io/podinfo",
			request: &corev1.CreateInstalledPackageRequest{
				AvailablePackageRef: &corev1.AvailablePackageReference{
					Identifier: "podinfo/podinfo",
					Context: &corev1.Context{
						Namespace: "default",
					},
				},
				Name: "my-podinfo",
				TargetContext: &corev1.Context{
					Namespace: "test",
				},
			},
			expectedDetail: &corev1.InstalledPackageDetail{
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
				PostInstallationNotes: "1. Get the application URL by running these commands:\n  echo \"Visit http://127.0.0.1:8080 to use your application\"\n  kubectl -n test port-forward deploy/test-my-podinfo 8080:9898\n",
				AvailablePackageRef: &corev1.AvailablePackageReference{
					Identifier: "podinfo/podinfo",
					Context: &corev1.Context{
						Namespace: "default",
					},
					Plugin: fluxPlugin,
				},
			},
			expectedPodPrefix: "pod/test-my-podinfo-",
		},
		{
			testName: "create package (semver constraint)",
			repoUrl:  "https://stefanprodan.github.io/podinfo",
			request: &corev1.CreateInstalledPackageRequest{
				AvailablePackageRef: &corev1.AvailablePackageReference{
					Identifier: "podinfo/podinfo",
					Context: &corev1.Context{
						Namespace: "default",
					},
				},
				Name: "my-podinfo-2",
				TargetContext: &corev1.Context{
					Namespace: "test",
				},
				PkgVersionReference: &corev1.VersionReference{
					Version: "> 5",
				},
			},
			expectedDetail: &corev1.InstalledPackageDetail{
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
				PostInstallationNotes: "1. Get the application URL by running these commands:\n  echo \"Visit http://127.0.0.1:8080 to use your application\"\n  kubectl -n test port-forward deploy/test-my-podinfo-2 8080:9898\n",
				AvailablePackageRef: &corev1.AvailablePackageReference{
					Identifier: "podinfo/podinfo",
					Context: &corev1.Context{
						Namespace: "default",
					},
					Plugin: fluxPlugin,
				},
			},
			expectedPodPrefix: "pod/test-my-podinfo-2-",
		},
		{
			testName: "create package (reconcile options)",
			repoUrl:  "https://stefanprodan.github.io/podinfo",
			request: &corev1.CreateInstalledPackageRequest{
				AvailablePackageRef: &corev1.AvailablePackageReference{
					Identifier: "podinfo/podinfo",
					Context: &corev1.Context{
						Namespace: "default",
					},
				},
				Name: "my-podinfo-3",
				TargetContext: &corev1.Context{
					Namespace: "test",
				},
				ReconciliationOptions: &corev1.ReconciliationOptions{
					Interval:           60,
					Suspend:            false,
					ServiceAccountName: "foo",
				},
			},
			expectedDetail: &corev1.InstalledPackageDetail{
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
				PostInstallationNotes: "1. Get the application URL by running these commands:\n  echo \"Visit http://127.0.0.1:8080 to use your application\"\n  kubectl -n test port-forward deploy/test-my-podinfo-3 8080:9898\n",
				AvailablePackageRef: &corev1.AvailablePackageReference{
					Identifier: "podinfo/podinfo",
					Context: &corev1.Context{
						Namespace: "default",
					},
					Plugin: fluxPlugin,
				},
			},
			expectedPodPrefix: "pod/test-my-podinfo-3-",
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

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			availablePackageRef := tc.request.AvailablePackageRef
			idParts := strings.Split(availablePackageRef.Identifier, "/")
			err = fluxCliCreateSource(t, idParts[0], tc.repoUrl, availablePackageRef.Context.Namespace)
			if err != nil {
				t.Fatalf("%+v", err)
			}
			t.Cleanup(func() {
				err = fluxCliDeleteSource(t, idParts[0], availablePackageRef.Context.Namespace)
				if err != nil {
					t.Logf("Failed to delete helm source due to [%v]", err)
				}
			})

			// need to wait until repo is index by flux plugin
			const maxWait = 10
			for i := 0; i < maxWait; i++ {
				_, err := fluxPluginClient.GetAvailablePackageDetail(
					context.TODO(),
					&corev1.GetAvailablePackageDetailRequest{AvailablePackageRef: availablePackageRef})
				if err == nil {
					break
				} else if i == maxWait-1 {
					t.Fatalf("Timed out waiting for available package [%s], last error: [%v]", availablePackageRef, err)
				} else {
					t.Logf("waiting 500ms for repository to be indexed...")
					time.Sleep(500 * time.Millisecond)
				}
			}

			if tc.request.ReconciliationOptions != nil && tc.request.ReconciliationOptions.ServiceAccountName != "" {
				err = kubectlCreateServiceAccount(t, tc.request.ReconciliationOptions.ServiceAccountName, "kubeapps")
				if err != nil {
					t.Fatalf("%+v", err)
				}
				t.Cleanup(func() {
					err = kubectlDeleteServiceAccount(t, tc.request.ReconciliationOptions.ServiceAccountName, "kubeapps")
					if err != nil {
						t.Logf("Failed to delete service account due to [%v]", err)
					}
				})
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
				err = fluxCliDeleteHelmRelease(t, installedPackageRef.Identifier, installedPackageRef.Context.Namespace)
				if err != nil {
					t.Logf("Failed to delete helm release due to [%v]", err)
				}
				err = kubectlDeleteNamespace(t, tc.request.TargetContext.Namespace)
				if err != nil {
					t.Logf("Failed to delete namespace due to [%v]", err)
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
				if resp2.InstalledPackageDetail.Status.Reason == corev1.InstalledPackageStatus_STATUS_REASON_PENDING && i < maxWait-1 {
					t.Logf("current state: [%s], waiting 500ms for installation to complete...", resp2.InstalledPackageDetail.Status.UserReason)
					time.Sleep(500 * time.Millisecond)
				} else if resp2.InstalledPackageDetail.Status.Ready == true && resp2.InstalledPackageDetail.Status.Reason == corev1.InstalledPackageStatus_STATUS_REASON_INSTALLED {
					actualDetail = resp2.InstalledPackageDetail
					break
				} else {
					t.Fatalf("Unexpected response: [%v]", resp2)
				}
			}
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
			pods, err := kubectlGetPods(t, tc.request.TargetContext.Namespace)
			if err != nil {
				t.Fatalf("%+v", err)
			}
			if len(pods) != 1 {
				t.Errorf("expected 1 pod, got: %s", pods)
			} else if !strings.HasPrefix(pods[0], tc.expectedPodPrefix) {
				t.Errorf("expected pod with prefix [%s] not found in namespace [%s]",
					tc.expectedPodPrefix, tc.request.TargetContext.Namespace)
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
	cmd = exec.Command("kubectl", "get", "nodes", "-o=name", "--context", k8s_context)
	bytes, err = cmd.CombinedOutput()
	if err != nil {
		t.Logf("%s", string(bytes))
		return false, err
	}
	if string(bytes) == "node/kubeapps-control-plane\n" {
		return true, nil
	} else {
		return false, nil
	}
}

func getFluxPluginClient(t *testing.T) (fluxplugin.FluxV2PackagesServiceClient, error) {
	t.Logf("+getFluxPluginClient")

	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	opts = append(opts, grpc.WithBlock())
	conn, err := grpc.Dial("localhost:8080", opts...)
	if err != nil {
		t.Fatalf("fail to dial: %v", err)
	}
	t.Cleanup(func() { conn.Close() })
	pluginsCli := plugins.NewPluginsServiceClient(conn)
	response, err := pluginsCli.GetConfiguredPlugins(context.TODO(), &plugins.GetConfiguredPluginsRequest{})
	if err != nil {
		t.Fatalf("fail to dial: %v", err)
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
func fluxCliCreateSource(t *testing.T, name, url, namespace string) error {
	t.Logf("+fluxCliCreateSource(%s)", name)
	cmd := exec.Command("flux", "create", "source", "helm", name, "--url", url, "--namespace", namespace, "--context", k8s_context)
	bytes, err := cmd.CombinedOutput()
	if err != nil {
		t.Logf("%s", string(bytes))
		return err
	}
	if strings.Contains(string(bytes), "fetched revision: ") {
		return nil
	} else {
		return fmt.Errorf("Unexpected output from flux create: [%s]", string(bytes))
	}
}

// this should eventually be replaced with flux plugin's DeleteRepository()
func fluxCliDeleteSource(t *testing.T, name, namespace string) error {
	t.Logf("+fluxCliDeleteSource(%s)", name)
	cmd := exec.Command("flux", "delete", "source", "helm", name, "--namespace", namespace, "--context", k8s_context, "--silent")
	bytes, err := cmd.CombinedOutput()
	if err != nil {
		t.Logf("%s", string(bytes))
		return err
	}
	if strings.Contains(string(bytes), "source helm deleted") {
		return nil
	} else {
		return fmt.Errorf("Unexpected output from flux delete source: [%s]", string(bytes))
	}
}

// this should eventually be replaced with flux plugin's DeleteInstalledPackage()
func fluxCliDeleteHelmRelease(t *testing.T, name, namespace string) error {
	t.Logf("+fluxCliDeleteHelmRelease(%s,%s)", name, namespace)
	cmd := exec.Command("flux", "delete", "helmrelease", name, "--namespace", namespace, "--context", k8s_context, "--silent")
	bytes, err := cmd.CombinedOutput()
	if err != nil {
		t.Logf("%s", string(bytes))
		return err
	}
	if strings.Contains(string(bytes), "helmreleases deleted") {
		return nil
	} else {
		return fmt.Errorf("Unexpected output from flux delete helmrelease: [%s]", string(bytes))
	}
}

func kubectlGetPods(t *testing.T, namespace string) (names []string, err error) {
	t.Logf("+kubectlGetPods(%s)", namespace)
	cmd := exec.Command("kubectl", "get", "pods", "-n", namespace, "-o=name", "--context", k8s_context)
	bytes, err := cmd.CombinedOutput()
	if err != nil {
		t.Logf("%s", string(bytes))
		return nil, err
	}
	return strings.Split(string(bytes), " \n"), nil
}

// will create a service account with cluster-admin privs
func kubectlCreateServiceAccount(t *testing.T, name, namespace string) error {
	t.Logf("+kubectlCreateServiceAccount(%s,%s)", name, namespace)
	cmd := exec.Command("kubectl", "create", "serviceaccount", name, "-n", namespace, "--context", k8s_context)
	bytes, err := cmd.CombinedOutput()
	if err != nil {
		t.Logf("%s", string(bytes))
		return err
	}
	if !strings.Contains(string(bytes), "serviceaccount/"+name+" created") {
		return fmt.Errorf("Unexpected output from kubectl create serviceaccount: [%s]", string(bytes))
	}

	cmd = exec.Command("kubectl", "create", "clusterrolebinding", name+"-binding",
		"--clusterrole=cluster-admin", "--serviceaccount="+namespace+":"+name, "--context", k8s_context)
	bytes, err = cmd.CombinedOutput()
	if err != nil {
		t.Logf("%s", string(bytes))
		return err
	}
	if !strings.Contains(string(bytes), "clusterrolebinding.rbac.authorization.k8s.io/"+name+"-binding created") {
		return fmt.Errorf("Unexpected output from kubectl create clusterrolebinding: [%s]", string(bytes))
	}
	return nil
}

func kubectlDeleteServiceAccount(t *testing.T, name, namespace string) error {
	t.Logf("+kubectlDeleteServiceAccount(%s,%s)", name, namespace)
	cmd := exec.Command("kubectl", "delete", "clusterrolebinding", name+"-binding", "--context", k8s_context)
	bytes, err := cmd.CombinedOutput()
	if err != nil {
		t.Logf("%s", string(bytes))
		return err
	}
	if !strings.Contains(string(bytes), "clusterrolebinding.rbac.authorization.k8s.io \""+name+"-binding\" deleted") {
		return fmt.Errorf("Unexpected output from kubectl delete clusterrolebinding: [%s]", string(bytes))
	}

	cmd = exec.Command("kubectl", "delete", "serviceaccount", name, "-n", namespace, "--context", k8s_context)
	bytes, err = cmd.CombinedOutput()
	if err != nil {
		t.Logf("%s", string(bytes))
		return err
	}
	if strings.Contains(string(bytes), "serviceaccount \""+name+"\" deleted") {
		return nil
	} else {
		return fmt.Errorf("Unexpected output from kubectl delete serviceaccount: [%s]", string(bytes))
	}
}

func kubectlDeleteNamespace(t *testing.T, namespace string) error {
	t.Logf("+kubectlDeleteNamespace(%s)", namespace)
	cmd := exec.Command("kubectl", "delete", "namespace", namespace, "--context", k8s_context)
	bytes, err := cmd.CombinedOutput()
	if err != nil {
		t.Logf("%s", string(bytes))
		return err
	}
	if !strings.Contains(string(bytes), "namespace \""+namespace+"\" deleted") {
		return fmt.Errorf("Unexpected output from kubectl delete namespace: [%s]", string(bytes))
	}
	return nil
}
