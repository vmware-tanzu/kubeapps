// Copyright 2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package resources

import (
	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/plugins/pkg/clientgetter"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	authorizationapi "k8s.io/api/authorization/v1"
	corev1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	log "k8s.io/klog/v2"
	"math"
	"sync"
)

type checkNSJob struct {
	ns corev1.Namespace
}

type checkNSResult struct {
	checkNSJob
	allowed bool
	Error   error
}

// FindAccessibleNamespaces return the raw list of namespaces that the user has permission to access
// Not filtered by any status (e.g. Active), but actual access is checked.
func FindAccessibleNamespaces(clusterTypedClientGetter clientgetter.TypedClientFunc, inClusterClientGetter clientgetter.TypedClientFunc, maxWorkers int) ([]corev1.Namespace, error) {
	typedClient, err := clusterTypedClientGetter()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "unable to get the k8s client: '%v'", err)
	}

	backgroundCtx := context.Background()

	// Try to list namespaces first with the user token
	namespaces, err := typedClient.CoreV1().Namespaces().List(backgroundCtx, metav1.ListOptions{})
	if err != nil {
		if k8sErrors.IsForbidden(err) {
			// The user doesn't have permissions to list namespaces, then use the current pod's service account
			inClusterClient, err := inClusterClientGetter()
			if err != nil {
				return nil, status.Errorf(codes.Internal, "unable to get the in-cluster k8s client: '%v'", err)
			}
			namespaces, err = inClusterClient.CoreV1().Namespaces().List(backgroundCtx, metav1.ListOptions{})
			if err != nil && k8sErrors.IsForbidden(err) {
				// Not even the configured kubeapps-apis service account has permission
				return nil, err
			}
		}
		return nil, err

		// Filter namespaces in which the user has permissions to write (secrets) only
		if namespaceList, err := filterAllowedNamespaces(typedClient, maxWorkers, namespaces.Items); err != nil {
			return nil, err
		}
		return namespaceList, nil
	}
	// If the user can list namespaces, do not filter them
	return namespaces.Items, nil
}

func nsCheckerWorker(client kubernetes.Interface, nsJobs <-chan checkNSJob, resultChan chan checkNSResult) {
	for j := range nsJobs {
		res, err := client.AuthorizationV1().SelfSubjectAccessReviews().Create(context.TODO(), &authorizationapi.SelfSubjectAccessReview{
			Spec: authorizationapi.SelfSubjectAccessReviewSpec{
				ResourceAttributes: &authorizationapi.ResourceAttributes{
					Group:     "",
					Resource:  "secrets",
					Verb:      "get",
					Namespace: j.ns.Name,
				},
			},
		}, metav1.CreateOptions{})
		var allowed bool
		if err != nil {
			allowed = false
		} else {
			allowed = res.Status.Allowed
		}
		resultChan <- checkNSResult{j, allowed, err}
	}
}

// filterAllowedNamespaces check to which namespaces the user has access.
// By access is considered the role of getting secrets in the namespace.
func filterAllowedNamespaces(userClient kubernetes.Interface, maxWorkers int, namespaces []corev1.Namespace) ([]corev1.Namespace, error) {
	var allowedNamespaces []corev1.Namespace

	var wg sync.WaitGroup
	workers := int(math.Min(float64(len(namespaces)), float64(maxWorkers)))
	checkNSJobs := make(chan checkNSJob, workers)
	nsCheckRes := make(chan checkNSResult, workers)

	// Process maxReq ns at a time
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			nsCheckerWorker(userClient, checkNSJobs, nsCheckRes)
			wg.Done()
		}()
	}
	go func() {
		wg.Wait()
		close(nsCheckRes)
	}()

	go func() {
		for _, ns := range namespaces {
			checkNSJobs <- checkNSJob{ns}
		}
		close(checkNSJobs)
	}()

	// Start receiving results
	for res := range nsCheckRes {
		if res.Error == nil {
			if res.allowed {
				allowedNamespaces = append(allowedNamespaces, res.ns)
			}
		} else {
			log.Errorf("failed to check namespace permissions. Got %v", res.Error)
		}
	}
	return allowedNamespaces, nil
}

func FilterActiveNamespaces(namespaces []corev1.Namespace) []corev1.Namespace {
	var readyNamespaces []corev1.Namespace
	for _, namespace := range namespaces {
		if namespace.Status.Phase == corev1.NamespaceActive {
			readyNamespaces = append(readyNamespaces, namespace)
		}
	}
	return readyNamespaces
}
