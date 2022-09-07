// Copyright 2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0
package main

import (
	"context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	authorizationapi "k8s.io/api/authorization/v1"
	corev1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	log "k8s.io/klog/v2"
	"math"
	"regexp"
	"strings"
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

func (s *Server) MaxWorkers() int {
	return int(s.clientQPS)
}

// GetAccessibleNamespaces return the list of namespaces that the user has permission to access
func (s *Server) GetAccessibleNamespaces(ctx context.Context, cluster string, trustedNamespaces []corev1.Namespace) ([]corev1.Namespace, error) {
	var namespaceList []corev1.Namespace

	if len(trustedNamespaces) > 0 {
		namespaceList = append(namespaceList, trustedNamespaces...)
	} else {

		typedClient, _, err := s.clusterServiceAccountClientGetter(ctx, cluster)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "unable to get the k8s client: '%v'", err)
		}

		// Try to list namespaces with the user token, for backward compatibility
		backgroundCtx := context.Background()
		namespaces, err := typedClient.CoreV1().Namespaces().List(backgroundCtx, metav1.ListOptions{})
		if err != nil {
			if k8sErrors.IsForbidden(err) {
				// The user doesn't have permissions to list namespaces, use the current pod's service account
				userClient, err := s.localServiceAccountClientGetter.Typed(backgroundCtx)
				if err != nil {
					return nil, err
				}
				namespaces, err = userClient.CoreV1().Namespaces().List(backgroundCtx, metav1.ListOptions{})
				if err != nil && k8sErrors.IsForbidden(err) {
					// Not even the configured kubeapps-apis service account has permission
					return nil, err
				}
			} else {
				return nil, err
			}

			// Filter namespaces in which the user has permissions to write (secrets) only
			namespaceList, err = filterAllowedNamespaces(typedClient, s.MaxWorkers(), namespaces.Items)
			if err != nil {
				return nil, err
			}
		} else {
			// If the user can list namespaces, do not filter them
			namespaceList = namespaces.Items
		}
	}

	// Filter out namespaces in terminating state
	return filterActiveNamespaces(namespaceList), nil
}

// getTrustedNamespacesFromHeader returns a list of namespaces from the header request
// The name and the value of the header field is specified by 2 variables:
// - headerName is a name of the expected header field, e.g. X-Consumer-Groups
// - headerPattern is a regular expression, and it matches only single regex group, e.g. ^namespace:([\w-]+)$
func getTrustedNamespacesFromHeader(ctx context.Context, headerName, headerPattern string) ([]corev1.Namespace, error) {
	var namespaces []corev1.Namespace
	if headerName == "" || headerPattern == "" {
		return []corev1.Namespace{}, nil
	}
	r, err := regexp.Compile(headerPattern)
	if err != nil {
		log.Errorf("unable to compile regular expression: %v", err)
		return namespaces, err
	}

	// Get trusted namespaces from the request header
	md, _ := metadata.FromIncomingContext(ctx)
	if md != nil && len(md[strings.ToLower(headerName)]) > 0 { // metadata is always lowercase
		headerValue := md[strings.ToLower(headerName)][0]
		trustedNamespaces := strings.Split(headerValue, ",")
		for _, n := range trustedNamespaces {
			// Check if the namespace matches the regex
			rns := r.FindStringSubmatch(strings.TrimSpace(n))
			if rns == nil || len(rns) < 2 {
				continue
			}
			ns := corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{Name: rns[1]},
				Status:     corev1.NamespaceStatus{Phase: corev1.NamespaceActive},
			}
			namespaces = append(namespaces, ns)
		}
	}
	return namespaces, nil
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
		resultChan <- checkNSResult{j, res.Status.Allowed, err}
	}
}

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

func filterActiveNamespaces(namespaces []corev1.Namespace) []corev1.Namespace {
	var readyNamespaces []corev1.Namespace
	for _, namespace := range namespaces {
		if namespace.Status.Phase == corev1.NamespaceActive {
			readyNamespaces = append(readyNamespaces, namespace)
		}
	}
	return readyNamespaces
}
