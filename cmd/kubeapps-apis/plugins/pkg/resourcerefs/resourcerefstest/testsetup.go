// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package resourcerefstest

import (
	corev1 "github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	"google.golang.org/grpc/codes"
)

// this is done so that test scenarios can be re-used in another package (helm and flux plug-ins)
// ref: https://stackoverflow.com/questions/28476307/how-to-get-test-environment-at-run-time
type TestReleaseStub struct {
	Name      string
	Namespace string
	Manifest  string
}

type TestCase struct {
	Name                  string
	ExistingReleases      []TestReleaseStub
	ExpectedResourceRefs  []*corev1.ResourceRef
	ExpectedErrStatusCode codes.Code
}

var (
	// will be properly initialized in resourcerefs_test.go init()
	TestCases1, TestCases2 = []TestCase(nil), []TestCase(nil)
)

func init() {
	TestCases1 = []TestCase{
		{
			Name: "returns resource references for helm installation",
			ExistingReleases: []TestReleaseStub{
				{
					Name:      "my-apache",
					Namespace: "default",
					Manifest: `
---
# Source: apache/templates/svc.yaml
apiVersion: v1
kind: Service
metadata:
  name: apache-test
  namespace: default
---
# Source: apache/templates/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: apache-test
  namespace: default
`,
				},
			},
			ExpectedResourceRefs: []*corev1.ResourceRef{
				{
					ApiVersion: "v1",
					Name:       "apache-test",
					Namespace:  "default",
					Kind:       "Service",
				},
				{
					ApiVersion: "apps/v1",
					Name:       "apache-test",
					Namespace:  "default",
					Kind:       "Deployment",
				},
			},
		},
		{
			Name: "returns resource references with explicit namespace when not present in helm manifest",
			ExistingReleases: []TestReleaseStub{
				{
					Name:      "my-apache",
					Namespace: "default",
					Manifest: `
---
# Source: apache/templates/svc.yaml
apiVersion: v1
kind: Service
metadata:
  name: apache-test
`,
				},
			},
			ExpectedResourceRefs: []*corev1.ResourceRef{
				{
					ApiVersion: "v1",
					Name:       "apache-test",
					Namespace:  "default",
					Kind:       "Service",
				},
			},
		},
		{
			Name: "returns resource references for resources in other namespaces",
			ExistingReleases: []TestReleaseStub{
				{
					Name:      "my-apache",
					Namespace: "default",
					Manifest: `
---
apiVersion: v1
kind: ClusterRole
metadata:
  name: test-cluster-role
---
# Source: apache/templates/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-other-namespace
  namespace: some-other-namespace
`,
				},
			},
			ExpectedResourceRefs: []*corev1.ResourceRef{
				{
					ApiVersion: "v1",
					Name:       "test-cluster-role",
					Kind:       "ClusterRole",
					Namespace:  "default",
				},
				{
					ApiVersion: "apps/v1",
					Name:       "test-other-namespace",
					Namespace:  "some-other-namespace",
					Kind:       "Deployment",
				},
			},
		},
		{
			Name: "skips resources that do not have a kind",
			ExistingReleases: []TestReleaseStub{
				{
					Name:      "my-apache",
					Namespace: "default",
					Manifest: `
---
# Source: apache/templates/svc.yaml
apiVersion: v1
otherstuff: ignored
metadata:
  name: apache-test
---
# Source: apache/templates/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: apache-test
  namespace: default
`,
				},
			},
			ExpectedResourceRefs: []*corev1.ResourceRef{
				{
					ApiVersion: "apps/v1",
					Name:       "apache-test",
					Namespace:  "default",
					Kind:       "Deployment",
				},
			},
		},
		{
			Name: "returns a not found error if the release is not found",
			ExistingReleases: []TestReleaseStub{
				{
					Name:      "my-apache",
					Namespace: "default",
				},
			},
			ExpectedResourceRefs: []*corev1.ResourceRef{},
		},
		{
			Name: "returns internal error if the yaml manifest cannot be parsed",
			ExistingReleases: []TestReleaseStub{
				{
					Name:      "my-apache",
					Namespace: "default",
					Manifest: `
---
apiVersion: v1
should not be :! parsed as yaml$
`,
				},
			},
			ExpectedErrStatusCode: codes.Internal,
		},
		{
			Name: "handles duplicate labels as helm does",
			// See https://github.com/vmware-tanzu/kubeapps/issues/632
			ExistingReleases: []TestReleaseStub{
				{
					Name:      "my-apache",
					Namespace: "default",
					Manifest: `
---
# Source: apache/templates/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: apache-test
  namespace: default
  label:
    chart: "gitea-0.2.0"
    chart: "gitea-0.2.0"
`,
				},
			},
			ExpectedResourceRefs: []*corev1.ResourceRef{
				{
					ApiVersion: "apps/v1",
					Name:       "apache-test",
					Namespace:  "default",
					Kind:       "Deployment",
				},
			},
		},
		{
			Name: "supports manifests with YAML type casting",
			ExistingReleases: []TestReleaseStub{
				{
					Name:      "my-apache",
					Namespace: "default",
					Manifest: `
---
# Source: apache/templates/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: !!string apache-test
  namespace: default
`,
				},
			},
			ExpectedResourceRefs: []*corev1.ResourceRef{
				{
					ApiVersion: "apps/v1",
					Name:       "apache-test",
					Namespace:  "default",
					Kind:       "Deployment",
				},
			},
		},
		{
			Name: "renders a list of items",
			ExistingReleases: []TestReleaseStub{
				{
					Name:      "my-apache",
					Namespace: "default",
					Manifest: `
---
apiVersion: v1
kind: List
items:
- apiVersion: apps/v1
  kind: Deployment
  metadata:
    name: apache-test
    namespace: default
- apiVersion: v1
  kind: Service
  metadata:
    name: apache-test
    namespace: default
`,
				},
			},
			ExpectedResourceRefs: []*corev1.ResourceRef{
				{
					ApiVersion: "apps/v1",
					Name:       "apache-test",
					Namespace:  "default",
					Kind:       "Deployment",
				},
				{
					ApiVersion: "v1",
					Name:       "apache-test",
					Namespace:  "default",
					Kind:       "Service",
				},
			},
		},
		{
			Name: "renders a rolelist of items",
			// See https://kubernetes.io/docs/reference/kubernetes-api/authorization-resources/role-v1/#RoleList
			ExistingReleases: []TestReleaseStub{
				{
					Name:      "my-apache",
					Namespace: "default",
					Manifest: `
---
apiVersion: v1
kind: RoleList
items:
- apiVersion: rbac.authorization.k8s.io/v1
  kind: Role
  metadata:
    name: role-1
    namespace: default
- apiVersion: rbac.authorization.k8s.io/v1
  kind: Role
  metadata:
    name: role-2
    namespace: default
`,
				},
			},
			ExpectedResourceRefs: []*corev1.ResourceRef{
				{
					ApiVersion: "rbac.authorization.k8s.io/v1",
					Name:       "role-1",
					Namespace:  "default",
					Kind:       "Role",
				},
				{
					ApiVersion: "rbac.authorization.k8s.io/v1",
					Name:       "role-2",
					Namespace:  "default",
					Kind:       "Role",
				},
			},
		},
		{
			Name: "renders a ClusterRoleList of items",
			// See https://kubernetes.io/docs/reference/kubernetes-api/authorization-resources/cluster-role-v1/#ClusterRoleList
			ExistingReleases: []TestReleaseStub{
				{
					Name:      "my-apache",
					Namespace: "default",
					Manifest: `
---
apiVersion: v1
kind: ClusterRoleList
items:
- apiVersion: rbac.authorization.k8s.io/v1
  kind: ClusterRole
  metadata:
    name: clusterrole-1
- apiVersion: rbac.authorization.k8s.io/v1
  kind: ClusterRole
  metadata:
    name: clusterrole-2
`,
				},
			},
			ExpectedResourceRefs: []*corev1.ResourceRef{
				{
					ApiVersion: "rbac.authorization.k8s.io/v1",
					Name:       "clusterrole-1",
					Namespace:  "default",
					Kind:       "ClusterRole",
				},
				{
					ApiVersion: "rbac.authorization.k8s.io/v1",
					Name:       "clusterrole-2",
					Namespace:  "default",
					Kind:       "ClusterRole",
				},
			},
		},
	}

	// Using the redis_existing_stub_completed data with
	// different manifests for each test.
	releaseNamespace := "test"
	releaseName := "my-redis"

	TestCases2 = []TestCase{
		{
			Name: "returns resource references for helm installation (2)",
			ExistingReleases: []TestReleaseStub{
				{
					Name:      releaseName,
					Namespace: releaseNamespace,
					Manifest: `
# Source: redis/templates/svc.yaml
apiVersion: v1
kind: Service
metadata:
  name: redis-test
  namespace: test
---
# Source: redis/templates/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: redis-test
  namespace: test
`,
				},
			},
			ExpectedResourceRefs: []*corev1.ResourceRef{
				{
					ApiVersion: "v1",
					Name:       "redis-test",
					Namespace:  "test",
					Kind:       "Service",
				},
				{
					ApiVersion: "apps/v1",
					Name:       "redis-test",
					Namespace:  "test",
					Kind:       "Deployment",
				},
			},
		},
		{
			Name: "returns resource references with explicit namespace when not present in helm manifest (2)",
			ExistingReleases: []TestReleaseStub{
				{
					Name:      releaseName,
					Namespace: releaseNamespace,
					Manifest: `
---
# Source: redis/templates/svc.yaml
apiVersion: v1
kind: Service
metadata:
  name: redis-test
`,
				},
			},
			ExpectedResourceRefs: []*corev1.ResourceRef{
				{
					ApiVersion: "v1",
					Name:       "redis-test",
					Namespace:  "test",
					Kind:       "Service",
				},
			},
		},
		{
			Name: "returns resource references for resources in other namespaces (2)",
			ExistingReleases: []TestReleaseStub{
				{
					Name:      releaseName,
					Namespace: releaseNamespace,
					Manifest: `
---
apiVersion: v1
kind: ClusterRole
metadata:
  name: test-cluster-role
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-other-namespace
  namespace: some-other-namespace
`,
				},
			},
			ExpectedResourceRefs: []*corev1.ResourceRef{
				{
					ApiVersion: "v1",
					Name:       "test-cluster-role",
					Namespace:  "test",
					Kind:       "ClusterRole",
				},
				{
					ApiVersion: "apps/v1",
					Name:       "test-other-namespace",
					Namespace:  "some-other-namespace",
					Kind:       "Deployment",
				},
			},
		},
		{
			Name: "skips resources that do not have a kind (2)",
			ExistingReleases: []TestReleaseStub{
				{
					Name:      releaseName,
					Namespace: releaseNamespace,
					Manifest: `
---
apiVersion: v1
otherstuff: ignored
metadata:
  name: redis-test
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: redis-test
`,
				},
			},
			ExpectedResourceRefs: []*corev1.ResourceRef{
				{
					ApiVersion: "apps/v1",
					Name:       "redis-test",
					Namespace:  "test",
					Kind:       "Deployment",
				},
			},
		},
		{
			Name:                  "returns a not found error if the helm release is not found (2)",
			ExpectedErrStatusCode: codes.NotFound,
		},
		{
			Name: "returns internal error if the yaml manifest cannot be parsed (2)",
			ExistingReleases: []TestReleaseStub{
				{
					Name:      releaseName,
					Namespace: releaseNamespace,
					Manifest: `
---
apiVersion: v1
should not be :! parsed as yaml$
`,
				},
			},
			ExpectedErrStatusCode: codes.Internal,
		},
		{
			Name: "handles duplicate labels in the manifest as helm does (2)",
			// See https://github.com/vmware-tanzu/kubeapps/issues/632
			ExistingReleases: []TestReleaseStub{
				{
					Name:      releaseName,
					Namespace: releaseNamespace,
					Manifest: `
---
# Source: apache/templates/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: redis-test
  label:
    chart: "redis-0.2.0"
    chart: "redis-0.2.0"
`,
				},
			},
			ExpectedResourceRefs: []*corev1.ResourceRef{
				{
					ApiVersion: "apps/v1",
					Name:       "redis-test",
					Namespace:  "test",
					Kind:       "Deployment",
				},
			},
		},
		{
			Name: "supports manifests with YAML type casting (2)",
			ExistingReleases: []TestReleaseStub{
				{
					Name:      releaseName,
					Namespace: releaseNamespace,
					Manifest: `
---
# Source: redis/templates/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: !!string redis-test
`,
				},
			},
			ExpectedResourceRefs: []*corev1.ResourceRef{
				{
					ApiVersion: "apps/v1",
					Name:       "redis-test",
					Namespace:  "test",
					Kind:       "Deployment",
				},
			},
		},
		{
			Name: "renders a list of items (2)",
			ExistingReleases: []TestReleaseStub{
				{
					Name:      releaseName,
					Namespace: releaseNamespace,
					Manifest: `
---
apiVersion: v1
kind: List
items:
- apiVersion: apps/v1
  kind: Deployment
  metadata:
    name: redis-test
    namespace: default
- apiVersion: v1
  kind: Service
  metadata:
    name: redis-test
    namespace: default
`,
				},
			},
			ExpectedResourceRefs: []*corev1.ResourceRef{
				{
					ApiVersion: "apps/v1",
					Name:       "redis-test",
					Namespace:  "default",
					Kind:       "Deployment",
				},
				{
					ApiVersion: "v1",
					Name:       "redis-test",
					Namespace:  "default",
					Kind:       "Service",
				},
			},
		},
		{
			Name: "renders a rolelist of items (2)",
			// See https://kubernetes.io/docs/reference/kubernetes-api/authorization-resources/role-v1/#RoleList
			ExistingReleases: []TestReleaseStub{
				{
					Name:      releaseName,
					Namespace: releaseNamespace,
					Manifest: `
---
apiVersion: v1
kind: RoleList
items:
- apiVersion: rbac.authorization.k8s.io/v1
  kind: Role
  metadata:
    name: role-1
    namespace: default
- apiVersion: rbac.authorization.k8s.io/v1
  kind: Role
  metadata:
    name: role-2
    namespace: default
`,
				},
			},
			ExpectedResourceRefs: []*corev1.ResourceRef{
				{
					ApiVersion: "rbac.authorization.k8s.io/v1",
					Name:       "role-1",
					Namespace:  "default",
					Kind:       "Role",
				},
				{
					ApiVersion: "rbac.authorization.k8s.io/v1",
					Name:       "role-2",
					Namespace:  "default",
					Kind:       "Role",
				},
			},
		},
		{
			Name: "renders a ClusterRoleList of items (2)",
			// See https://kubernetes.io/docs/reference/kubernetes-api/authorization-resources/cluster-role-v1/#ClusterRoleList
			ExistingReleases: []TestReleaseStub{
				{
					Name:      releaseName,
					Namespace: releaseNamespace,
					Manifest: `
---
apiVersion: v1
kind: ClusterRoleList
items:
- apiVersion: rbac.authorization.k8s.io/v1
  kind: ClusterRole
  metadata:
    name: clusterrole-1
- apiVersion: rbac.authorization.k8s.io/v1
  kind: ClusterRole
  metadata:
    name: clusterrole-2
`,
				},
			},
			ExpectedResourceRefs: []*corev1.ResourceRef{
				{
					ApiVersion: "rbac.authorization.k8s.io/v1",
					Name:       "clusterrole-1",
					Namespace:  "test",
					Kind:       "ClusterRole",
				},
				{
					ApiVersion: "rbac.authorization.k8s.io/v1",
					Name:       "clusterrole-2",
					Namespace:  "test",
					Kind:       "ClusterRole",
				},
			},
		},
		{
			Name: "returns resource refs for helm installation in non default ns (flux HelmRelease targetNamespace is set)",
			ExistingReleases: []TestReleaseStub{
				{
					Name:      "test2-my-redis",
					Namespace: "test2",
					Manifest: `
# Source: redis/templates/svc.yaml
apiVersion: v1
kind: Service
metadata:
  name: redis-test
  namespace: test2
---
# Source: redis/templates/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: redis-test
  namespace: test2
`,
				},
			},
			ExpectedResourceRefs: []*corev1.ResourceRef{
				{
					ApiVersion: "v1",
					Name:       "redis-test",
					Namespace:  "test2",
					Kind:       "Service",
				},
				{
					ApiVersion: "apps/v1",
					Name:       "redis-test",
					Namespace:  "test2",
					Kind:       "Deployment",
				},
			},
		},
	}
}
