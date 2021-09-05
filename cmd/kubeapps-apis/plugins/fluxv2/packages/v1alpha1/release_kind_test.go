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
	"io"
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
					UserReason: "ReconciliationSucceeded",
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

			var actualDetail *corev1.InstalledPackageDetail
			for i := 0; i < maxWait; i++ {
				resp2, err := fluxPluginClient.GetInstalledPackageDetail(
					context.TODO(),
					&corev1.GetInstalledPackageDetailRequest{InstalledPackageRef: installedPackageRef})
				if err != nil {
					t.Fatalf("%+v", err)
				}
				if resp2.InstalledPackageDetail.Status.Reason == corev1.InstalledPackageStatus_STATUS_REASON_PENDING && i < maxWait-1 {
					t.Logf("waiting 500ms for installation to complete...")
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

			err = fluxCliDeleteHelmRelease(t, installedPackageRef.Identifier, installedPackageRef.Context.Namespace)
			if err != nil {
				t.Fatalf("%+v", err)
			}

			err = fluxCliDeleteSource(t, idParts[0], availablePackageRef.Context.Namespace)
			if err != nil {
				t.Fatalf("%+v", err)
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
	if string(bytes) != "kubeapps\n" {
		return false, nil
	}

	// naively assume that if the api server reports nodes, the cluster is up
	cmd = exec.Command("kubectl", "get", "nodes", "-o=name")
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
func fluxCliCreateSource(t *testing.T, name, url, namespace string) (err error) {
	t.Logf("+fluxCliCreateSource(%s)", name)
	cmd := exec.Command("flux", "create", "source", "helm", name, "--url", url, "--namespace", namespace)
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
func fluxCliDeleteSource(t *testing.T, name, namespace string) (err error) {
	t.Logf("+fluxCliDeleteHelmRelease(%s)", name)
	cmd := exec.Command("flux", "delete", "source", "helm", name, "--namespace", namespace)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	go func() {
		defer stdin.Close()
		io.WriteString(stdin, "y\n")
	}()
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
func fluxCliDeleteHelmRelease(t *testing.T, name, namespace string) (err error) {
	t.Logf("+fluxCliDeleteHelmRelease(%s)", name)
	cmd := exec.Command("flux", "delete", "helmrelease", name, "--namespace", namespace)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	go func() {
		defer stdin.Close()
		io.WriteString(stdin, "y\n")
	}()
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
	cmd := exec.Command("kubectl", "get", "pods", "-n", namespace, "-o=name")
	bytes, err := cmd.CombinedOutput()
	if err != nil {
		t.Logf("%s", string(bytes))
		return nil, err
	}
	return strings.Split(string(bytes), " \n"), nil
}
