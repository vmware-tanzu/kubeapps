// Copyright 2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package resources

import (
	"errors"
	"github.com/google/go-cmp/cmp"
	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/plugins/pkg/clientgetter"
	authorizationv1 "k8s.io/api/authorization/v1"
	apiv1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
	typfake "k8s.io/client-go/kubernetes/fake"
	k8stesting "k8s.io/client-go/testing"
	"testing"
)

type ClientReaction struct {
	verb     string
	resource string
	reaction k8stesting.ReactionFunc
}

func TestFindAccessibleNamespaces(t *testing.T) {

	ns1 := apiv1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "ns-1",
		},
		Status: apiv1.NamespaceStatus{
			Phase: apiv1.NamespaceActive,
		},
	}

	ns2 := apiv1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "ns-2",
		},
		Status: apiv1.NamespaceStatus{
			Phase: apiv1.NamespaceActive,
		},
	}

	nsResource := schema.GroupResource{
		Group:    "v1",
		Resource: "namespaces",
	}

	testCases := []struct {
		name               string
		reactors           []*ClientReaction
		inClusterReactors  []*ClientReaction
		existingNamespaces []*apiv1.Namespace
		expectedNamespaces []apiv1.Namespace
		expectedErr        error
	}{
		{
			name: "returns all namespaces with listing permissions",
			existingNamespaces: []*apiv1.Namespace{
				&ns1,
				&ns2,
			},
			expectedNamespaces: []apiv1.Namespace{
				ns1,
				ns2,
			},
		},
		{
			name: "returns error when not failing but other than forbidden",
			reactors: []*ClientReaction{
				{
					verb:     "list",
					resource: "namespaces",
					reaction: func(action k8stesting.Action) (handled bool, ret k8sruntime.Object, err error) {
						return true, nil, k8sErrors.NewUnauthorized("bang")
					},
				},
			},
			expectedErr: k8sErrors.NewUnauthorized("bang"),
		},
		{
			name: "returns allowed namespaces filtered when can't list namespaces cluster-wide",
			existingNamespaces: []*apiv1.Namespace{
				&ns1,
				&ns2,
			},
			reactors: []*ClientReaction{
				{
					verb:     "list",
					resource: "namespaces",
					reaction: func(action k8stesting.Action) (handled bool, ret k8sruntime.Object, err error) {
						return true, nil, k8sErrors.NewForbidden(nsResource, "", errors.New("bang"))
					},
				},
				{
					verb:     "create",
					resource: "selfsubjectaccessreviews",
					reaction: func(action k8stesting.Action) (handled bool, ret k8sruntime.Object, err error) {
						createAction := action.(k8stesting.CreateActionImpl)
						accessReview := createAction.Object.(*authorizationv1.SelfSubjectAccessReview)
						switch accessReview.Spec.ResourceAttributes.Namespace {
						case "ns-2":
							return true, &authorizationv1.SelfSubjectAccessReview{
								Status: authorizationv1.SubjectAccessReviewStatus{Allowed: false},
							}, nil
						default:
							return true, &authorizationv1.SelfSubjectAccessReview{
								Status: authorizationv1.SubjectAccessReviewStatus{Allowed: true},
							}, nil
						}
					},
				},
			},
			expectedNamespaces: []apiv1.Namespace{
				ns1,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			var typedObjects []k8sruntime.Object
			if tc.existingNamespaces != nil {
				for _, ns := range tc.existingNamespaces {
					typedObjects = append(typedObjects, ns)
				}
			}

			clusterTypedClient, inClusterClient := newTypedClients(typedObjects, tc.reactors, tc.inClusterReactors)

			namespaces, err := FindAccessibleNamespaces(clusterTypedClient, inClusterClient, 1)

			if tc.expectedErr != nil && err != nil {
				if got, want := err.Error(), tc.expectedErr.Error(); !cmp.Equal(want, got) {
					t.Errorf("in %s: mismatch (-want +got):\n%s", tc.name, cmp.Diff(want, got))
				}
			} else if err != nil {
				t.Fatalf("in %s: %+v", tc.name, err)
			}

			// check
			if !cmp.Equal(namespaces, tc.expectedNamespaces) {
				t.Errorf("Unexpected result (-want +got):\n%s", cmp.Diff(namespaces, tc.expectedNamespaces))
			}

		})
	}
}

func newTypedClients(objects []k8sruntime.Object, clientReactions []*ClientReaction, inClusterReactions []*ClientReaction) (clientgetter.TypedClientFunc, clientgetter.TypedClientFunc) {
	clusterClient := typfake.NewSimpleClientset(objects...)
	for _, reaction := range clientReactions {
		clusterClient.PrependReactor(reaction.verb, reaction.resource, reaction.reaction)
	}

	inClusterClient := typfake.NewSimpleClientset(objects...)
	for _, reaction := range inClusterReactions {
		inClusterClient.PrependReactor(reaction.verb, reaction.resource, reaction.reaction)
	}

	return func() (kubernetes.Interface, error) {
			return clusterClient, nil
		}, func() (kubernetes.Interface, error) {
			return inClusterClient, nil
		}
}
