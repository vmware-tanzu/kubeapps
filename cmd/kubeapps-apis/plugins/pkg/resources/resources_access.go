// Copyright 2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package resources

import (
	"fmt"
	"sync"

	"golang.org/x/net/context"
	authorizationapi "k8s.io/api/authorization/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"

	log "k8s.io/klog/v2"
)

var AccessVerbs = []string{
	"create",
	"update",
	"delete",
	"get",
	"list",
	"watch",
}

func GetPermissionsOnResource(ctx context.Context, client kubernetes.Interface, gr schema.GroupResource, namespace string) (map[string]bool, error) {
	var wg sync.WaitGroup
	type accessReviewResult struct {
		verb    string
		allowed bool
	}
	accessReviewChan := make(chan accessReviewResult, len(AccessVerbs))

	// Each access review is requested in a go routine
	for _, v := range AccessVerbs {
		wg.Add(1)
		go func(verb string) {
			defer wg.Done()

			response, err := doResourceAccessReview(ctx, client, gr, verb, namespace)
			if err != nil {
				log.Errorf("Error finding permissions for %s/%s - %s in namespace %s: %v", gr.Group, gr.Resource, verb, namespace, err)
				return
			}
			accessReviewChan <- accessReviewResult{
				verb:    verb,
				allowed: response.Status.Allowed,
			}
		}(v)
	}
	go func() {
		wg.Wait()
		close(accessReviewChan)
	}()

	m := make(map[string]bool)

	// Convert the private struct to a regular map
	// The following loop will only terminate when the channel is closed
	for r := range accessReviewChan {
		m[fmt.Sprint(r.verb)] = r.allowed
	}
	return m, nil
}

func doResourceAccessReview(ctx context.Context, client kubernetes.Interface, gr schema.GroupResource, verb, namespace string) (*authorizationapi.SelfSubjectAccessReview, error) {
	return client.AuthorizationV1().SelfSubjectAccessReviews().Create(ctx, &authorizationapi.SelfSubjectAccessReview{
		Spec: authorizationapi.SelfSubjectAccessReviewSpec{
			ResourceAttributes: &authorizationapi.ResourceAttributes{
				Group:     gr.Group,
				Resource:  gr.Resource,
				Verb:      verb,
				Namespace: namespace,
			},
		},
	}, metav1.CreateOptions{})
}
