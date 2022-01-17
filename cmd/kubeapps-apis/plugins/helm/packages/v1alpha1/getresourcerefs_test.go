// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/kubeapps/kubeapps/cmd/apprepository-controller/pkg/apis/apprepository/v1alpha1"
	corev1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	"github.com/kubeapps/kubeapps/cmd/kubeapps-apis/plugins/pkg/resourcerefs"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type TestCase struct {
	name               string
	existingReleases   []resourcerefs.TestReleaseStub
	request            *corev1.GetInstalledPackageResourceRefsRequest
	expectedResponse   *corev1.GetInstalledPackageResourceRefsResponse
	expectedStatusCode codes.Code
}

func TestGetInstalledPackageResourceRefs(t *testing.T) {

	// TODO (gfichtenholt) what's missing here is a call to
	// resourcerefs_test.go init() fuction. I spent quite some time but have not yet
	// figured out how to make that happen. So I am commenting this test out until I do

	if len(resourcerefs.TestCases1) == 0 {
		t.Logf("Expected non-empty array [resourcerefs.TestCases1]")
	}

	testCases := []TestCase{
		/*
				{
					name:             resourcerefs.TestCases1[0].Name,
					existingReleases: resourcerefs.TestCases1[0].ExistingReleases,
					request: &corev1.GetInstalledPackageResourceRefsRequest{
						InstalledPackageRef: &corev1.InstalledPackageReference{
							Context: &corev1.Context{
								Cluster:   "default",
								Namespace: "default",
							},
							Identifier: "my-apache",
						},
					},
					expectedResponse: &corev1.GetInstalledPackageResourceRefsResponse{
						Context: &corev1.Context{
							Cluster:   "default",
							Namespace: "default",
						},
						ResourceRefs: resourcerefs.TestCases1[0].ExpectedResourceRefs,
					},
				},
			}
					, TestCase{
						name:             tc.Name,
						existingReleases: tc.ExistingReleases,
						request: &corev1.GetInstalledPackageResourceRefsRequest{
							InstalledPackageRef: &corev1.InstalledPackageReference{
								Context: &corev1.Context{
									Cluster:   "default",
									Namespace: "default",
								},
								Identifier: "my-apache",
							},
						},
						expectedResponse: &corev1.GetInstalledPackageResourceRefsResponse{
							Context: &corev1.Context{
								Cluster:   "default",
								Namespace: "default",
							},
							ResourceRefs: tc.ExpectedResourceRefs,
						},
					}, TestCase{
						name:             tc.Name,
						existingReleases: tc.ExistingReleases,
						request: &corev1.GetInstalledPackageResourceRefsRequest{
							InstalledPackageRef: &corev1.InstalledPackageReference{
								Context: &corev1.Context{
									Cluster:   "default",
									Namespace: "default",
								},
								Identifier: "my-apache",
							},
						},
					}, TestCase{
						name:             tc.Name,
						existingReleases: tc.ExistingReleases,
						request: &corev1.GetInstalledPackageResourceRefsRequest{
							InstalledPackageRef: &corev1.InstalledPackageReference{
								Context: &corev1.Context{
									Cluster:   "default",
									Namespace: "default",
								},
								Identifier: "my-apache",
							},
						},
						expectedResponse: &corev1.GetInstalledPackageResourceRefsResponse{
							Context: &corev1.Context{
								Cluster:   "default",
								Namespace: "default",
							},
							ResourceRefs: tc.ExpectedResourceRefs,
						},
					}, TestCase{
						name:             tc.Name,
						existingReleases: tc.ExistingReleases,
						request: &corev1.GetInstalledPackageResourceRefsRequest{
							InstalledPackageRef: &corev1.InstalledPackageReference{
								Context: &corev1.Context{
									Cluster:   "default",
									Namespace: "default",
								},
								Identifier: "my-iis",
							},
						},
						expectedStatusCode: codes.NotFound,
					}, TestCase{
						name:             tc.Name,
						existingReleases: tc.ExistingReleases,
						request: &corev1.GetInstalledPackageResourceRefsRequest{
							InstalledPackageRef: &corev1.InstalledPackageReference{
								Context: &corev1.Context{
									Cluster:   "default",
									Namespace: "default",
								},
								Identifier: "my-apache",
							},
						},
						expectedStatusCode: codes.Internal,
					}, TestCase{
						name: tc.Name,
						// See https://github.com/kubeapps/kubeapps/issues/632
						existingReleases: tc.ExistingReleases,
						request: &corev1.GetInstalledPackageResourceRefsRequest{
							InstalledPackageRef: &corev1.InstalledPackageReference{
								Context: &corev1.Context{
									Cluster:   "default",
									Namespace: "default",
								},
								Identifier: "my-apache",
							},
						},
						expectedResponse: &corev1.GetInstalledPackageResourceRefsResponse{
							Context: &corev1.Context{
								Cluster:   "default",
								Namespace: "default",
							},
							ResourceRefs: tc.ExpectedResourceRefs,
						},
					}, TestCase{
						name:             tc.Name,
						existingReleases: tc.ExistingReleases,
						request: &corev1.GetInstalledPackageResourceRefsRequest{
							InstalledPackageRef: &corev1.InstalledPackageReference{
								Context: &corev1.Context{
									Cluster:   "default",
									Namespace: "default",
								},
								Identifier: "my-apache",
							},
						},
						expectedResponse: &corev1.GetInstalledPackageResourceRefsResponse{
							Context: &corev1.Context{
								Cluster:   "default",
								Namespace: "default",
							},
							ResourceRefs: tc.ExpectedResourceRefs,
						},
					}, TestCase{
						name:             tc.Name,
						existingReleases: tc.ExistingReleases,
						request: &corev1.GetInstalledPackageResourceRefsRequest{
							InstalledPackageRef: &corev1.InstalledPackageReference{
								Context: &corev1.Context{
									Cluster:   "default",
									Namespace: "default",
								},
								Identifier: "my-apache",
							},
						},
						expectedResponse: &corev1.GetInstalledPackageResourceRefsResponse{
							Context: &corev1.Context{
								Cluster:   "default",
								Namespace: "default",
							},
							ResourceRefs: tc.ExpectedResourceRefs,
						},
					}, TestCase{
						name: tc.Name,
						// See https://kubernetes.io/docs/reference/kubernetes-api/authorization-resources/role-v1/#RoleList
						existingReleases: tc.ExistingReleases,
						request: &corev1.GetInstalledPackageResourceRefsRequest{
							InstalledPackageRef: &corev1.InstalledPackageReference{
								Context: &corev1.Context{
									Cluster:   "default",
									Namespace: "default",
								},
								Identifier: "my-apache",
							},
						},
						expectedResponse: &corev1.GetInstalledPackageResourceRefsResponse{
							Context: &corev1.Context{
								Cluster:   "default",
								Namespace: "default",
							},
							ResourceRefs: tc.ExpectedResourceRefs,
						},
					}, TestCase{
						name: tc.Name,
						// See https://kubernetes.io/docs/reference/kubernetes-api/authorization-resources/cluster-role-v1/#ClusterRoleList
						existingReleases: tc.ExistingReleases,
						request: &corev1.GetInstalledPackageResourceRefsRequest{
							InstalledPackageRef: &corev1.InstalledPackageReference{
								Context: &corev1.Context{
									Cluster:   "default",
									Namespace: "default",
								},
								Identifier: "my-apache",
							},
						},
						expectedResponse: &corev1.GetInstalledPackageResourceRefsResponse{
							Context: &corev1.Context{
								Cluster:   "default",
								Namespace: "default",
							},
							ResourceRefs: tc.ExpectedResourceRefs,
						},
					})
				}
		*/
	}

	ignoredFields := cmpopts.IgnoreUnexported(
		corev1.GetInstalledPackageResourceRefsResponse{},
		corev1.ResourceRef{},
		corev1.Context{},
	)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			authorized := true
			actionConfig := newActionConfigFixture(
				t,
				tc.request.GetInstalledPackageRef().GetContext().GetNamespace(),
				toHelmReleaseStubs(tc.existingReleases),
				nil)

			server, _, cleanup := makeServer(t, authorized, actionConfig, &v1alpha1.AppRepository{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "bitnami",
					Namespace: globalPackagingNamespace,
				},
			})
			defer cleanup()

			response, err := server.GetInstalledPackageResourceRefs(context.Background(), tc.request)

			if got, want := status.Code(err), tc.expectedStatusCode; got != want {
				t.Fatalf("got: %+v, want: %+v, err: %+v", got, want, err)
			}

			if got, want := response, tc.expectedResponse; !cmp.Equal(want, got, ignoredFields) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, ignoredFields))
			}
		})
	}
}

func toHelmReleaseStubs(in []resourcerefs.TestReleaseStub) []releaseStub {
	out := []releaseStub{}
	for _, r := range in {
		out = append(out, releaseStub{name: r.Name, namespace: r.Namespace, manifest: r.Manifest})
	}
	return out
}
