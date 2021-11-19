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
	packagingv1alpha1 "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/packaging/v1alpha1"
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

// GetAvailablePackageVersions returns the package versions managed by the 'kapp_controller' plugin
func (s *Server) GetAvailablePackageVersions(ctx context.Context, request *corev1.GetAvailablePackageVersionsRequest) (*corev1.GetAvailablePackageVersionsResponse, error) {
	log.Infof("+kapp-controller GetAvailablePackageVersions")

	// Retrieve the proper parameters from the request
	namespace := request.GetAvailablePackageRef().GetContext().GetNamespace()
	cluster := request.GetAvailablePackageRef().GetContext().GetCluster()
	identifier := request.GetAvailablePackageRef().GetIdentifier()

	// Validate the request
	if namespace == "" || identifier == "" {
		return nil, status.Errorf(codes.InvalidArgument, "Required context or identifier not provided")
	}

	if cluster == "" {
		cluster = s.globalPackagingCluster
	}

	// Use the field selector to return only Package CRs that match on the spec.refName.
	fieldSelector := fmt.Sprintf("spec.refName=%s", identifier)
	pkgs, err := s.getPkgsWithFieldSelector(ctx, cluster, namespace, fieldSelector)
	if err != nil {
		return nil, errorByStatus("get", "Package", "", err)
	}
	pkgVersionsMap, err := getPkgVersionsMap(pkgs)
	if err != nil {
		return nil, err
	}

	// TODO(minelson): support configurable version summary for kapp-controller pkgs
	// as already done for Helm (see #3588 for more info).
	versions := make([]*corev1.PackageAppVersion, len(pkgVersionsMap[identifier]))
	for i, v := range pkgVersionsMap[identifier] {
		versions[i] = &corev1.PackageAppVersion{
			PkgVersion: v.version.String(),
		}
	}

	return &corev1.GetAvailablePackageVersionsResponse{
		PackageAppVersions: versions,
	}, nil
}

// GetAvailablePackageDetail returns the package metadata managed by the 'kapp_controller' plugin
func (s *Server) GetAvailablePackageDetail(ctx context.Context, request *corev1.GetAvailablePackageDetailRequest) (*corev1.GetAvailablePackageDetailResponse, error) {
	log.Infof("+kapp-controller GetAvailablePackageDetail")

	// Validate the request
	if request.GetAvailablePackageRef().GetContext() == nil {
		return nil, status.Errorf(codes.InvalidArgument, "no request AvailablePackageRef.Context provided")
	}

	// Retrieve the proper parameters from the request
	namespace := request.GetAvailablePackageRef().GetContext().GetNamespace()
	cluster := request.GetAvailablePackageRef().GetContext().GetCluster()
	identifier := request.GetAvailablePackageRef().GetIdentifier()
	requestedPkgVersion := request.GetPkgVersion()

	if cluster == "" {
		cluster = s.globalPackagingCluster
	}

	// fetch the package metadata
	pkgMetadata, err := s.getPkgMetadata(ctx, cluster, namespace, identifier)
	if err != nil {
		return nil, errorByStatus("get", "PackageMetadata", identifier, err)
	}

	// Use the field selector to return only Package CRs that match on the spec.refName.
	fieldSelector := fmt.Sprintf("spec.refName=%s", identifier)
	pkgs, err := s.getPkgsWithFieldSelector(ctx, cluster, namespace, fieldSelector)
	if err != nil {
		return nil, errorByStatus("get", "Package", identifier, err)
	}
	pkgVersionsMap, err := getPkgVersionsMap(pkgs)
	if err != nil {
		return nil, err
	}

	var foundPkgSemver = &pkgSemver{}
	if requestedPkgVersion != "" {
		// Ensure the version is available.
		for _, v := range pkgVersionsMap[identifier] {
			if v.version.String() == requestedPkgVersion {
				foundPkgSemver = &v
				break
			}
		}
		if foundPkgSemver.version == nil {
			return nil, status.Errorf(codes.NotFound, fmt.Sprintf("unable to find %q package with version %q", identifier, requestedPkgVersion))
		}
	} else {
		// If the pkgVersion wasn't specified, grab the packages to find the latest.
		if len(pkgVersionsMap[identifier]) > 0 {
			foundPkgSemver = &pkgVersionsMap[identifier][0]
			requestedPkgVersion = foundPkgSemver.version.String()
		} else {
			return nil, status.Errorf(codes.NotFound, fmt.Sprintf("unable to find any versions for the package %q", identifier))
		}
	}

	availablePackageDetail, err := s.getAvailablePackageDetail(pkgMetadata, requestedPkgVersion, foundPkgSemver, cluster)
	if err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("unable to create the AvailablePackageDetail: %v", err))
	}

	return &corev1.GetAvailablePackageDetailResponse{
		AvailablePackageDetail: availablePackageDetail,
	}, nil
}

// GetInstalledPackageSummaries returns the installed packagesmanaged by the 'kapp_controller' plugin
func (s *Server) GetInstalledPackageSummaries(ctx context.Context, request *corev1.GetInstalledPackageSummariesRequest) (*corev1.GetInstalledPackageSummariesResponse, error) {
	log.Infof("+kapp-controller GetInstalledPackageSummaries")
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

	// retrieve the list of installed packages
	pkgInstalls, err := s.getPkgInstalls(ctx, cluster, namespace)
	if err != nil {
		return nil, errorByStatus("get", "PackageInstall", "", err)
	}

	// paginate the list of results
	installedPkgSummaries := make([]*corev1.InstalledPackageSummary, len(pkgInstalls))

	// create the waiting group for processing each item aynchronously
	var wg sync.WaitGroup
	if len(pkgInstalls) > 0 {
		startAt := -1
		if pageSize > 0 {
			startAt = int(pageSize) * pageOffset
		}
		for i, pkgInstall := range pkgInstalls {
			wg.Add(1)
			if startAt <= i {
				go func(i int, pkgInstall *packagingv1alpha1.PackageInstall) error {
					defer wg.Done()
					// fetch additional information from the package metadata
					// TODO(agamez): if the repository where the package belongs to is not installed, it will throw an error;
					// decide whether we should ignore it or it is ok with returning an error
					pkgMetadata, err := s.getPkgMetadata(ctx, cluster, namespace, pkgInstall.Spec.PackageRef.RefName)
					if err != nil {
						return errorByStatus("get", "PackageMetadata", pkgInstall.Spec.PackageRef.RefName, err)
					}

					// Use the field selector to return only Package CRs that match on the spec.refName.
					fieldSelector := fmt.Sprintf("spec.refName=%s", pkgInstall.Spec.PackageRef.RefName)
					pkgs, err := s.getPkgsWithFieldSelector(ctx, cluster, namespace, fieldSelector)
					if err != nil {
						return errorByStatus("get", "Package", "", err)
					}
					pkgVersionsMap, err := getPkgVersionsMap(pkgs)
					if err != nil {
						return err
					}

					// generate the installedPackageSummary from the fetched information
					installedPackageSummary, err := s.getInstalledPackageSummary(pkgInstall, pkgMetadata, pkgVersionsMap, cluster)
					if err != nil {
						return status.Errorf(codes.Internal, fmt.Sprintf("unable to create the InstalledPackageSummary: %v", err))
					}

					// append the availablePackageSummary to the slice
					installedPkgSummaries[i] = installedPackageSummary

					return nil
				}(i, pkgInstall)
			}
			// if we've reached the end of the page, stop iterating
			if pageSize > 0 && len(installedPkgSummaries) == int(pageSize) {
				break
			}
		}
	}
	// Wait until each goroutine has finished
	wg.Wait()

	// TODO(agamez): the slice with make is filled with <nil>, in case of an error in the
	// i goroutine, the i-th <nil> stub will remain. Check if 'errgroup' works here, but I haven't
	// been able so far.
	// An alternative is using channels to perform a fine-grained control... but not sure if it worths

	// filter out <nil> values
	installedPkgSummariesNilSafe := []*corev1.InstalledPackageSummary{}
	for _, installedPkgSummary := range installedPkgSummaries {
		if installedPkgSummary != nil {
			installedPkgSummariesNilSafe = append(installedPkgSummariesNilSafe, installedPkgSummary)
		}
	}

	// Only return a next page token if the request was for pagination and
	// the results are a full page.
	nextPageToken := ""
	if pageSize > 0 && len(installedPkgSummariesNilSafe) == int(pageSize) {
		nextPageToken = fmt.Sprintf("%d", pageOffset+1)
	}
	response := &corev1.GetInstalledPackageSummariesResponse{
		InstalledPackageSummaries: installedPkgSummariesNilSafe,
		NextPageToken:             nextPageToken,
	}
	return response, nil
}
