// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/vmware-tanzu/kubeapps/cmd/apprepository-controller/pkg/apis/apprepository/v1alpha1"
	corev1 "github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	plugins "github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/gen/core/plugins/v1alpha1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/release"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestCreateInstalledPackage(t *testing.T) {
	testCases := []struct {
		name               string
		releaseStub        releaseStub
		request            *corev1.CreateInstalledPackageRequest
		expectedResponse   *corev1.CreateInstalledPackageResponse
		expectedStatusCode codes.Code
		expectedRelease    *release.Release
	}{
		{
			name: "creates the installed package from repo without credentials",
			// this is just for populating the mock database
			releaseStub: releaseStub{
				chartID:       "bitnami/apache",
				latestVersion: "1.18.3",
			},
			request: &corev1.CreateInstalledPackageRequest{
				AvailablePackageRef: &corev1.AvailablePackageReference{
					Context: &corev1.Context{
						Namespace: globalPackagingNamespace,
					},
					Identifier: "bitnami/apache",
				},
				TargetContext: &corev1.Context{
					Namespace: "default",
				},
				Name: "my-apache",
				PkgVersionReference: &corev1.VersionReference{
					Version: "1.18.3",
				},
				Values: "{\"foo\": \"bar\"}",
			},
			expectedResponse: &corev1.CreateInstalledPackageResponse{
				InstalledPackageRef: &corev1.InstalledPackageReference{
					Context: &corev1.Context{
						Cluster:   "default",
						Namespace: "default",
					},
					Identifier: "my-apache",
					Plugin:     GetPluginDetail(),
				},
			},
			expectedStatusCode: codes.OK,
			expectedRelease: &release.Release{
				Name: "my-apache",
				Info: &release.Info{
					Description: "Install complete",
					Status:      release.StatusDeployed,
				},
				Chart: &chart.Chart{
					Metadata: &chart.Metadata{
						Name:    "apache",
						Version: "1.18.3",
					},
					Values: map[string]interface{}{},
				},
				Config:    map[string]interface{}{"foo": "bar"},
				Version:   1,
				Namespace: "default",
			},
		},
		{
			name: "returns invalid if available package ref invalid",
			request: &corev1.CreateInstalledPackageRequest{
				AvailablePackageRef: &corev1.AvailablePackageReference{
					Context: &corev1.Context{
						Namespace: globalPackagingNamespace,
					},
					Identifier: "not-a-valid-identifier",
				},
			},
			expectedStatusCode: codes.InvalidArgument,
		},
	}

	ignoredUnexported := cmpopts.IgnoreUnexported(
		corev1.CreateInstalledPackageResponse{},
		corev1.InstalledPackageReference{},
		corev1.Context{},
		plugins.Plugin{},
		chart.Chart{},
	)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			authorized := true
			actionConfig := newActionConfigFixture(t, tc.request.GetTargetContext().GetNamespace(), nil, nil)
			server, mockDB, cleanup := makeServer(t, authorized, actionConfig, &v1alpha1.AppRepository{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "bitnami",
					Namespace: globalPackagingNamespace,
				},
			})
			defer cleanup()
			populateAssetDB(t, mockDB, []releaseStub{tc.releaseStub})

			response, err := server.CreateInstalledPackage(context.Background(), tc.request)

			if got, want := status.Code(err), tc.expectedStatusCode; got != want {
				t.Fatalf("got: %+v, want: %+v, err: %+v", got, want, err)
			}

			// Verify the expected response (our contract to the caller).
			if got, want := response, tc.expectedResponse; !cmp.Equal(got, want, ignoredUnexported) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, ignoredUnexported))
			}

			if tc.expectedRelease != nil {
				// Verify the expected request was made to Helm (our contract to the helm lib).
				releases, err := actionConfig.Releases.Driver.List(func(*release.Release) bool { return true })
				if err != nil {
					t.Fatalf("%+v", err)
				}
				if got, want := len(releases), 1; got != want {
					t.Fatalf("got: %d, want: %d", got, want)
				}

				ignoredFields := cmpopts.IgnoreFields(release.Info{}, "FirstDeployed", "LastDeployed")
				if got, want := releases[0], tc.expectedRelease; !cmp.Equal(got, want, ignoredUnexported, ignoredFields) {
					t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, ignoredUnexported, ignoredFields))
				}
			}
		})
	}
}

func TestTimeoutCreateInstalledPackage(t *testing.T) {
	testCases := []struct {
		name           string
		timeoutSeconds int32
	}{
		{
			name:           "Timeout for Helm is passed to the release creation function",
			timeoutSeconds: 33,
		},
		{
			name:           "No timeout for Helm is passed to the release creation function",
			timeoutSeconds: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			authorized := true
			request := &corev1.CreateInstalledPackageRequest{
				AvailablePackageRef: &corev1.AvailablePackageReference{
					Context: &corev1.Context{
						Namespace: globalPackagingNamespace,
					},
					Identifier: "bitnami/apache",
				},
				TargetContext: &corev1.Context{
					Namespace: "default",
				},
				Name: "my-apache",
				PkgVersionReference: &corev1.VersionReference{
					Version: "1.18.3",
				},
				Values: "{\"foo\": \"bar\"}",
			}
			actionConfig := newActionConfigFixture(t, request.GetTargetContext().GetNamespace(), nil, nil)
			server, mockDB, cleanup := makeServer(t, authorized, actionConfig, &v1alpha1.AppRepository{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "bitnami",
					Namespace: globalPackagingNamespace,
				},
			})
			server.pluginConfig.TimeoutSeconds = tc.timeoutSeconds

			var effectiveTimeout int32 = -1
			var effectiveConfig *action.Configuration
			var effectiveName string
			var effectiveNs string
			// stub createRelease function
			server.createReleaseFunc = func(config *action.Configuration, name string, namespace string, valueString string, ch *chart.Chart,
				registrySecrets map[string]string, timeout int32) (*release.Release, error) {
				effectiveConfig = config
				effectiveTimeout = timeout
				effectiveName = name
				effectiveNs = namespace
				return &release.Release{}, nil
			}

			defer cleanup()
			rStub := releaseStub{
				chartID:       "bitnami/apache",
				latestVersion: "1.18.3",
			}
			populateAssetDB(t, mockDB, []releaseStub{rStub})

			_, err := server.CreateInstalledPackage(context.Background(), request)

			if got, want := effectiveTimeout, tc.timeoutSeconds; got != want {
				t.Fatalf("got: %+v, want: %+v, err: %+v", got, want, err)
			}
			if got, want := effectiveConfig, actionConfig; got != want {
				t.Fatalf("got: %+v, want: %+v, err: %+v", got, want, err)
			}
			if got, want := effectiveName, request.Name; got != want {
				t.Fatalf("got: %+v, want: %+v, err: %+v", got, want, err)
			}
			if got, want := effectiveNs, request.TargetContext.Namespace; got != want {
				t.Fatalf("got: %+v, want: %+v, err: %+v", got, want, err)
			}
		})
	}
}
