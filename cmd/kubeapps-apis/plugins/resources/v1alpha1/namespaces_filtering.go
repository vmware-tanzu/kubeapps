// Copyright 2022-2023 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0
package main

import (
	"context"
	"net/http"
	"regexp"
	"strings"

	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/plugins/pkg/resources"
	"github.com/vmware-tanzu/kubeapps/pkg/kube"
	"google.golang.org/grpc/metadata"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	log "k8s.io/klog/v2"
)

func (s *Server) MaxWorkers() int {
	return int(s.clientQPS)
}

// GetAccessibleNamespaces return the list of namespaces that the user has permission to access
func (s *Server) GetAccessibleNamespaces(ctx context.Context, headers http.Header, cluster string, trustedNamespaces []corev1.Namespace) ([]corev1.Namespace, error) {
	var namespaceList []corev1.Namespace

	if len(trustedNamespaces) > 0 {
		namespaceList = append(namespaceList, trustedNamespaces...)
	} else {
		userTypedClientFunc := func() (kubernetes.Interface, error) {
			return s.clientGetter.Typed(headers, cluster)
		}

		// The service account client returned for fetching namespaces depends on whether
		// the target cluster is the same one Kubeapps is running on (in which case,
		// we use the pod's token which will have been configured for access) or the
		// token from the clusters config (if it exists).
		var serviceAccountTypedClientFunc func() (kubernetes.Interface, error)
		if kube.IsKubeappsClusterRef(cluster) {
			serviceAccountTypedClientFunc = func() (kubernetes.Interface, error) {
				// Not using ctx here so that we can't inadvertently send the user
				// creds.
				return s.localServiceAccountClientGetter.Typed(context.Background())
			}
		} else {
			serviceAccountTypedClientFunc = func() (kubernetes.Interface, error) {
				return s.clusterServiceAccountClientGetter.Typed(headers, cluster)
			}
		}

		var err error
		namespaceList, err = resources.FindAccessibleNamespaces(userTypedClientFunc, serviceAccountTypedClientFunc, s.MaxWorkers())
		if err != nil {
			return nil, err
		}
	}

	// Filter out namespaces in terminating state
	return resources.FilterActiveNamespaces(namespaceList), nil
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
		log.Errorf("Unable to compile regular expression: %v", err)
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
