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
	"sync"

	corev1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	datapackagingv1alpha1 "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apiserver/apis/datapackaging/v1alpha1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	log "k8s.io/klog/v2"
)

// GetAvailablePackageSummaries returns the available packages managed by the 'kapp_controller' plugin
func (s *Server) GetAvailablePackageSummaries(ctx context.Context, request *corev1.GetAvailablePackageSummariesRequest) (*corev1.GetAvailablePackageSummariesResponse, error) {
	log.Infof("+kapp-controller GetAvailablePackageSummaries")

	// Retrieve the proper parameters from the request
	namespace := request.GetContext().GetNamespace()
	cluster := request.GetContext().GetCluster()
	pageSize := request.GetPaginationOptions().GetPageSize()
	pageOffset, err := pageOffsetFromPageToken(request.GetPaginationOptions().GetPageToken())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "unable to intepret page token %q: %v", request.GetPaginationOptions().GetPageToken(), err)
	}
	// Assume the default cluster if none is specified
	if cluster == "" {
		cluster = s.globalPackagingCluster
	}
	// fetch all the package metadatas
	pkgMetadatas, err := s.getPkgMetadatas(ctx, cluster, namespace)
	if err != nil {
		return nil, errorByStatus("get", "PackageMetadata", "", err)
	}

	// paginate the list of results
	availablePackageSummaries := make([]*corev1.AvailablePackageSummary, len(pkgMetadatas))

	// create the waiting group for processing each item aynchronously
	var wg sync.WaitGroup

	if len(pkgMetadatas) > 0 {
		startAt := -1
		if pageSize > 0 {
			startAt = int(pageSize) * pageOffset
		}
		for i, pkgMetadata := range pkgMetadatas {
			wg.Add(1)
			if startAt <= i {
				go func(i int, pkgMetadata *datapackagingv1alpha1.PackageMetadata) error {
					defer wg.Done()
					// fetch the associated packages
					// Use the field selector to return only Package CRs that match on the spec.refName.
					fieldSelector := fmt.Sprintf("spec.refName=%s", pkgMetadata.Name)
					pkgs, err := s.getPkgsWithFieldSelector(ctx, cluster, namespace, fieldSelector)
					if err != nil {
						return errorByStatus("get", "Package", pkgMetadata.Name, err)
					}
					pkgVersionsMap, err := getPkgVersionsMap(pkgs)
					if err != nil {
						return err
					}

					// generate the availablePackageSummary from the fetched information
					availablePackageSummary, err := s.getAvailablePackageSummary(pkgMetadata, pkgVersionsMap, cluster)
					if err != nil {
						return status.Errorf(codes.Internal, fmt.Sprintf("unable to create the AvailablePackageSummary: %v", err))
					}

					// append the availablePackageSummary to the slice
					availablePackageSummaries[i] = availablePackageSummary
					return nil
				}(i, pkgMetadata)
			}
			// if we've reached the end of the page, stop iterating
			if pageSize > 0 && len(availablePackageSummaries) == int(pageSize) {
				break
			}
		}
	}
	wg.Wait() // Wait until each goroutine has finished

	// TODO(agamez): the slice with make is filled with <nil>, in case of an error in the
	// i goroutine, the i-th <nil> stub will remain. Check if 'errgroup' works here, but I haven't
	// been able so far.
	// An alternative is using channels to perform a fine-grained control... but not sure if it worths

	// filter out <nil> values
	availablePackageSummariesNilSafe := []*corev1.AvailablePackageSummary{}
	categories := []string{}
	for _, availablePackageSummary := range availablePackageSummaries {
		if availablePackageSummary != nil {
			availablePackageSummariesNilSafe = append(availablePackageSummariesNilSafe, availablePackageSummary)
			categories = append(categories, availablePackageSummary.Categories...)

		}
	}
	// if no results whatsoever, throw an error
	if len(availablePackageSummariesNilSafe) == 0 {
		return nil, status.Errorf(codes.NotFound, fmt.Sprintf("no available packages: %v", err))
	}

	// Only return a next page token if the request was for pagination and
	// the results are a full page.
	nextPageToken := ""
	if pageSize > 0 && len(availablePackageSummariesNilSafe) == int(pageSize) {
		nextPageToken = fmt.Sprintf("%d", pageOffset+1)
	}
	response := &corev1.GetAvailablePackageSummariesResponse{
		AvailablePackageSummaries: availablePackageSummariesNilSafe,
		// TODO(agamez): populate this field
		Categories:    categories,
		NextPageToken: nextPageToken,
	}
	return response, nil
}
