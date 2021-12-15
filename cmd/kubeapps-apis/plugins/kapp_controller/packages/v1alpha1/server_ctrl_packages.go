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
	"strings"
	"sync"
	"time"

	corev1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	packagingv1alpha1 "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/packaging/v1alpha1"
	datapackagingv1alpha1 "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apiserver/apis/datapackaging/v1alpha1"
	vendirVersions "github.com/vmware-tanzu/carvel-vendir/pkg/vendir/versions/v1alpha1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

	// TODO(agamez): DRY up this logic (cf GetInstalledPackageSummaries)
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
					// TODO(agamez): perhaps we better fetch all the packages and filter ourselves to reduce the k8s calls
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
					availablePackageSummary, err := s.buildAvailablePackageSummary(pkgMetadata, pkgVersionsMap, cluster)
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
	// However, should we just return an error if so? See https://github.com/kubeapps/kubeapps/pull/3784#discussion_r754836475
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
		Categories:                categories,
		NextPageToken:             nextPageToken,
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
		return nil, status.Errorf(codes.Internal, "unable to get the PkgVersionsMap: '%v'", err)

	}

	// TODO(minelson): support configurable version summary for kapp-controller pkgs
	// as already done for Helm (see #3588 for more info).
	versions := make([]*corev1.PackageAppVersion, len(pkgVersionsMap[identifier]))
	for i, v := range pkgVersionsMap[identifier] {
		// Currently, PkgVersion and AppVersion are the same
		// https://kubernetes.slack.com/archives/CH8KCCKA5/p1636386358322000?thread_ts=1636371493.320900&cid=CH8KCCKA5
		versions[i] = &corev1.PackageAppVersion{
			PkgVersion: v.version.String(),
			AppVersion: v.version.String(),
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
		return nil, status.Errorf(codes.Internal, "unable to get the PkgVersionsMap: '%v'", err)
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

	availablePackageDetail, err := s.buildAvailablePackageDetail(pkgMetadata, requestedPkgVersion, foundPkgSemver, cluster)
	if err != nil {
		return nil, errorByStatus("create", "AvailablePackageDetail", pkgMetadata.Name, err)

	}

	return &corev1.GetAvailablePackageDetailResponse{
		AvailablePackageDetail: availablePackageDetail,
	}, nil
}

// GetInstalledPackageSummaries returns the installed packages managed by the 'kapp_controller' plugin
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
	// TODO(agamez): we should be paginating this request rather than requesting everything every time
	pkgInstalls, err := s.getPkgInstalls(ctx, cluster, namespace)
	if err != nil {
		return nil, errorByStatus("get", "PackageInstall", "", err)
	}

	// paginate the list of results
	installedPkgSummaries := make([]*corev1.InstalledPackageSummary, len(pkgInstalls))

	// create the waiting group for processing each item aynchronously
	var wg sync.WaitGroup

	// TODO(agamez): DRY up this logic (cf GetAvailablePackageSummaries)
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
					installedPackageSummary, err := s.buildInstalledPackageSummary(pkgInstall, pkgMetadata, pkgVersionsMap, cluster)
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
	// However, should we just return an error if so? See https://github.com/kubeapps/kubeapps/pull/3784#discussion_r754836475
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

// GetInstalledPackageDetail returns the package metadata managed by the 'kapp_controller' plugin
func (s *Server) GetInstalledPackageDetail(ctx context.Context, request *corev1.GetInstalledPackageDetailRequest) (*corev1.GetInstalledPackageDetailResponse, error) {
	log.Infof("+kapp-controller GetInstalledPackageDetail")

	// Retrieve the proper parameters from the request
	cluster := request.GetInstalledPackageRef().GetContext().GetCluster()
	namespace := request.GetInstalledPackageRef().GetContext().GetNamespace()
	installedPackageRefId := request.GetInstalledPackageRef().GetIdentifier()

	if cluster == "" {
		cluster = s.globalPackagingCluster
	}

	typedClient, _, err := s.GetClients(ctx, cluster)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "unable to get the k8s client: '%v'", err)
	}

	// fetch the package install
	pkgInstall, err := s.getPkgInstall(ctx, cluster, namespace, installedPackageRefId)
	if err != nil {
		return nil, errorByStatus("get", "PackageInstall", installedPackageRefId, err)
	}

	// fetch the resulting deployed app after the installation
	app, err := s.getApp(ctx, cluster, namespace, installedPackageRefId)
	if err != nil {
		return nil, errorByStatus("get", "App", installedPackageRefId, err)
	}

	// retrieve the package metadata associated with this installed package
	pkgName := pkgInstall.Spec.PackageRef.RefName
	pkgMetadata, err := s.getPkgMetadata(ctx, cluster, namespace, pkgName)
	if err != nil {
		return nil, errorByStatus("get", "PackageMetadata", pkgName, err)
	}

	// Use the field selector to return only Package CRs that match on the spec.refName.
	fieldSelector := fmt.Sprintf("spec.refName=%s", pkgMetadata.Name)
	pkgs, err := s.getPkgsWithFieldSelector(ctx, cluster, namespace, fieldSelector)
	if err != nil {
		return nil, errorByStatus("get", "Package", "", err)
	}
	pkgVersionsMap, err := getPkgVersionsMap(pkgs)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "unable to get the PkgVersionsMap: '%v'", err)
	}

	// get the values applies i) get the secret name where it is stored; 2) get the values from the secret
	valuesApplied := ""
	// retrieve every value and build a single string containing all of them
	for _, pkgInstallValue := range pkgInstall.Spec.Values {
		secretRefName := pkgInstallValue.SecretRef.Name
		// if there is a secret containing the applied values of this installed package, get the them
		if secretRefName != "" {
			values, err := typedClient.CoreV1().Secrets(namespace).Get(ctx, secretRefName, metav1.GetOptions{})
			if err != nil {
				if errors.IsNotFound(err) {
					log.Warningf("The referenced secret does not exist: %s", errorByStatus("get", "Secret", secretRefName, err).Error())
				} else {
					return nil, errorByStatus("get", "Secret", secretRefName, err)
				}
			}
			if values != nil {
				for fileName, valuesContent := range values.Data {
					valuesApplied = fmt.Sprintf("%s\n# %s\n%s\n---", valuesApplied, fileName, valuesContent)
				}
			}
		}
	}
	// trim the new doc separator in the last element
	valuesApplied = strings.Trim(valuesApplied, "---")

	installedPackageDetail, err := s.buildInstalledPackageDetail(pkgInstall, pkgMetadata, pkgVersionsMap, app, valuesApplied, cluster)
	if err != nil {
		return nil, errorByStatus("create", "InstalledPackageDetail", pkgInstall.Name, err)

	}

	response := &corev1.GetInstalledPackageDetailResponse{
		InstalledPackageDetail: installedPackageDetail,
	}
	return response, nil
}

// CreateInstalledPackage creates an installed package managed by the 'kapp_controller' plugin
func (s *Server) CreateInstalledPackage(ctx context.Context, request *corev1.CreateInstalledPackageRequest) (*corev1.CreateInstalledPackageResponse, error) {
	log.Infof("+kapp-controller CreateInstalledPackage")

	// Validate the request
	if request.GetAvailablePackageRef().GetContext().GetNamespace() == "" || request.GetAvailablePackageRef().GetIdentifier() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "required context or identifier not provided")
	}
	if request == nil || request.GetAvailablePackageRef() == nil {
		return nil, status.Errorf(codes.InvalidArgument, "no request AvailablePackageRef provided")
	}
	if request.GetName() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "no request Name provided")
	}
	if request.GetTargetContext() == nil || request.GetTargetContext().GetNamespace() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "no request TargetContext namespace provided")
	}
	if request.GetReconciliationOptions() == nil || request.GetReconciliationOptions().GetServiceAccountName() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "no request ReconciliationOptions serviceAccountName provided")
	}

	// Retrieve the proper parameters from the request
	packageRef := request.GetAvailablePackageRef()
	packageCluster := request.GetAvailablePackageRef().GetContext().GetCluster()
	packageNamespace := request.GetAvailablePackageRef().GetContext().GetNamespace()
	reconciliationOptions := request.GetReconciliationOptions()
	pkgVersion := request.GetPkgVersionReference().GetVersion()
	installedPackageName := request.GetName()
	values := request.GetValues()
	targetNamespace := request.GetTargetContext().GetNamespace()
	targetCluster := request.GetTargetContext().GetCluster()

	if targetCluster == "" {
		targetCluster = s.globalPackagingCluster
	}

	if targetCluster != packageCluster {
		return nil, status.Errorf(codes.InvalidArgument, "installing packages in other clusters in not supported yet")
	}

	typedClient, _, err := s.GetClients(ctx, targetCluster)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "unable to get the k8s client: '%v'", err)
	}

	// fetch the package metadata
	pkgMetadata, err := s.getPkgMetadata(ctx, packageCluster, packageNamespace, packageRef.Identifier)
	if err != nil {
		return nil, errorByStatus("get", "PackageMetadata", packageRef.Identifier, err)
	}

	// build a new secret object with the values
	secret, err := s.buildSecret(installedPackageName, values, targetNamespace)
	if err != nil {
		return nil, errorByStatus("create", "Secret", installedPackageName, err)

	}

	// build a new pkgInstall object
	newPkgInstall, err := s.buildPkgInstall(installedPackageName, targetCluster, targetNamespace, pkgMetadata.Name, pkgVersion, reconciliationOptions)
	if err != nil {
		return nil, errorByStatus("create", "PackageInstall", installedPackageName, err)
	}

	// create the Secret in the cluster
	// TODO(agamez): check when is the best moment to create this object.
	// See if we can delay the creation until the PackageInstall is successfully created.
	createdSecret, err := typedClient.CoreV1().Secrets(targetNamespace).Create(ctx, secret, metav1.CreateOptions{})
	if createdSecret == nil || err != nil {
		return nil, errorByStatus("create", "Secret", secret.Name, err)
	}

	// create the PackageInstall in the cluster
	createdPkgInstall, err := s.createPkgInstall(ctx, targetCluster, targetNamespace, newPkgInstall)
	if err != nil {
		// clean-up the secret if something fails
		err := typedClient.CoreV1().Secrets(targetNamespace).Delete(ctx, secret.Name, metav1.DeleteOptions{})
		if err != nil {
			return nil, errorByStatus("delete", "Secret", secret.Name, err)
		}
		return nil, errorByStatus("create", "PackageInstall", newPkgInstall.Name, err)
	}

	// generate the response
	installedRef := &corev1.InstalledPackageReference{
		Context: &corev1.Context{
			Namespace: createdPkgInstall.GetNamespace(),
			Cluster:   targetCluster,
		},
		Identifier: newPkgInstall.Name,
		Plugin:     GetPluginDetail(),
	}

	return &corev1.CreateInstalledPackageResponse{
		InstalledPackageRef: installedRef,
	}, nil
}

// UpdateInstalledPackage Updates an installed package managed by the 'kapp_controller' plugin
func (s *Server) UpdateInstalledPackage(ctx context.Context, request *corev1.UpdateInstalledPackageRequest) (*corev1.UpdateInstalledPackageResponse, error) {
	log.Infof("+kapp-controller UpdateInstalledPackage")

	// Validate the request
	if request == nil || request.GetInstalledPackageRef() == nil {
		return nil, status.Errorf(codes.InvalidArgument, "no request AvailablePackageRef provided")
	}
	if request.GetInstalledPackageRef().GetContext().GetNamespace() == "" || request.GetInstalledPackageRef().GetIdentifier() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "required context or identifier not provided")
	}
	if request.GetInstalledPackageRef().GetIdentifier() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "no request Name provided")
	}

	// Retrieve the proper parameters from the request
	packageCluster := request.GetInstalledPackageRef().GetContext().GetCluster()
	packageNamespace := request.GetInstalledPackageRef().GetContext().GetNamespace()
	reconciliationOptions := request.GetReconciliationOptions()
	pkgVersion := request.GetPkgVersionReference().GetVersion()
	installedPackageName := request.GetInstalledPackageRef().GetIdentifier()
	values := request.GetValues()

	if packageCluster == "" {
		packageCluster = s.globalPackagingCluster
	}

	typedClient, _, err := s.GetClients(ctx, packageCluster)
	if err != nil {
		return nil, err
	}

	// fetch the package install
	pkgInstall, err := s.getPkgInstall(ctx, packageCluster, packageNamespace, installedPackageName)
	if err != nil {
		return nil, errorByStatus("get", "PackageInstall", installedPackageName, err)
	}

	// Update the rest of the fields
	pkgInstall.Spec.PackageRef.VersionSelection = &vendirVersions.VersionSelectionSemver{Constraints: pkgVersion}
	if reconciliationOptions != nil {
		if reconciliationOptions.Interval > 0 {
			pkgInstall.Spec.SyncPeriod = &metav1.Duration{Duration: time.Duration(reconciliationOptions.Interval) * time.Second}
		}
		if reconciliationOptions.ServiceAccountName != "" {
			pkgInstall.Spec.ServiceAccountName = reconciliationOptions.ServiceAccountName
		}
		pkgInstall.Spec.Paused = reconciliationOptions.Suspend
	}

	// update the pkgInstall in the server
	updatedPkgInstall, err := s.updatePkgInstall(ctx, packageCluster, packageNamespace, pkgInstall)
	if err != nil {
		return nil, errorByStatus("get", "PackageInstall", installedPackageName, err)
	}

	// Update the values.yaml values file if any is passed, otherwise, delete the values
	if values != "" {
		secret, err := s.buildSecret(installedPackageName, values, packageNamespace)
		if err != nil {
			return nil, errorByStatus("upsate", "Secret", secret.Name, err)
		}
		updatedSecret, err := typedClient.CoreV1().Secrets(packageNamespace).Update(ctx, secret, metav1.UpdateOptions{})
		if updatedSecret == nil || err != nil {
			return nil, errorByStatus("update", "Secret", secret.Name, err)
		}
	} else {
		// Delete all the associated secrets
		// TODO(agamez): maybe it's too aggresive and we should be deleting only those secrets created by this plugin
		// See https://github.com/kubeapps/kubeapps/pull/3790#discussion_r754797195
		for _, packageInstallValue := range pkgInstall.Spec.Values {
			secretId := packageInstallValue.SecretRef.Name
			err := typedClient.CoreV1().Secrets(packageNamespace).Delete(ctx, secretId, metav1.DeleteOptions{})
			if errors.IsNotFound(err) {
				log.Warningf("The referenced secret does not exist: %s", errorByStatus("get", "Secret", secretId, err).Error())
			} else {
				return nil, errorByStatus("delete", "Secret", secretId, err)
			}
		}
	}

	// generate the response
	updatedRef := &corev1.InstalledPackageReference{
		Context: &corev1.Context{
			Namespace: updatedPkgInstall.GetNamespace(),
			Cluster:   packageCluster,
		},
		Identifier: updatedPkgInstall.Name,
		Plugin:     GetPluginDetail(),
	}
	return &corev1.UpdateInstalledPackageResponse{
		InstalledPackageRef: updatedRef,
	}, nil
}

// DeleteInstalledPackage Deletes an installed package managed by the 'kapp_controller' plugin
func (s *Server) DeleteInstalledPackage(ctx context.Context, request *corev1.DeleteInstalledPackageRequest) (*corev1.DeleteInstalledPackageResponse, error) {
	log.Infof("+kapp-controller DeleteInstalledPackage")

	// Validate the request
	if request == nil || request.GetInstalledPackageRef() == nil {
		return nil, status.Errorf(codes.InvalidArgument, "no request InstalledPackageRef provided")
	}
	if request.GetInstalledPackageRef().GetContext().GetNamespace() == "" || request.GetInstalledPackageRef().GetIdentifier() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "required context or identifier not provided")
	}

	// Retrieve the proper parameters from the request
	packageRef := request.GetInstalledPackageRef()
	identifier := request.GetInstalledPackageRef().GetIdentifier()
	namespace := packageRef.GetContext().GetNamespace()
	cluster := packageRef.GetContext().GetCluster()

	if cluster == "" {
		cluster = s.globalPackagingCluster
	}

	typedClient, _, err := s.GetClients(ctx, cluster)
	if err != nil {
		return nil, err
	}

	pkgInstall, err := s.getPkgInstall(ctx, cluster, namespace, identifier)
	if err != nil {
		return nil, errorByStatus("get", "PackageInstall", identifier, err)
	}

	// Delete the package install
	err = s.deletePkgInstall(ctx, cluster, namespace, identifier)
	if err != nil {
		return nil, errorByStatus("delete", "PackageInstall", identifier, err)
	}

	// Delete all the associated secrets
	// TODO(agamez): maybe it's too aggresive and we should be deleting only those secrets created by this plugin
	// See https://github.com/kubeapps/kubeapps/pull/3790#discussion_r754797195
	for _, packageInstallValue := range pkgInstall.Spec.Values {
		secretId := packageInstallValue.SecretRef.Name
		err := typedClient.CoreV1().Secrets(namespace).Delete(ctx, secretId, metav1.DeleteOptions{})
		if err != nil {
			if errors.IsNotFound(err) {
				log.Warningf("The referenced secret does not exist: %s", errorByStatus("get", "Secret", secretId, err).Error())
			} else {
				return nil, errorByStatus("delete", "Secret", secretId, err)
			}
		}
	}
	return &corev1.DeleteInstalledPackageResponse{}, nil
}

// GetInstalledPackageResourceRefs returns the references for the k8s resources of an installed package managed by the 'kapp_controller' plugin
func (s *Server) GetInstalledPackageResourceRefs(ctx context.Context, request *corev1.GetInstalledPackageResourceRefsRequest) (*corev1.GetInstalledPackageResourceRefsResponse, error) {
	log.Infof("+kapp-controller GetInstalledPackageResourceRefs")

	// Retrieve the proper parameters from the request
	cluster := request.GetInstalledPackageRef().GetContext().GetCluster()
	namespace := request.GetInstalledPackageRef().GetContext().GetNamespace()
	installedPackageRefId := request.GetInstalledPackageRef().GetIdentifier()

	if cluster == "" {
		cluster = s.globalPackagingCluster
	}

	typedClient, _, err := s.GetClients(ctx, cluster)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "unable to get the k8s client: '%v'", err)
	}

	refs := []*corev1.ResourceRef{}

	// TODO(agamez): apparently, App CRs not being created by a "kapp deploy"
	// don't have the proper annotations. So, in order to retrieve the annotation value,
	// we have to get the ConfigMap <AppName>-ctrl and, then, fetch the
	// vaulue of the key "labelValue" in data.spec.
	// See https://kubernetes.slack.com/archives/CH8KCCKA5/p1637842398026700

	// the ConfigMap name is, by convention, "<appname>-ctrl", but it will change in the near future
	cmName := fmt.Sprintf("%s-ctrl", installedPackageRefId)
	cm, err := typedClient.CoreV1().ConfigMaps(namespace).Get(ctx, cmName, metav1.GetOptions{})
	if err == nil && cm.Data["spec"] != "" {

		appLabelValue := extractValue(cm.Data["spec"], "labelValue")
		appLabelSelector := fmt.Sprintf("%s=%s", appLabelKey, appLabelValue)
		listOptions := metav1.ListOptions{LabelSelector: appLabelSelector}

		// TODO(agamez): perform an actual query over all the resources available in the cluster
		// this is currently just a PoC getting the bare minimum: pods, deployments, services and secrets.
		// Also, the xxx.Items[i] are not populating the Kind and APIVersion fields. Check why.

		// Fetching all the matching pods
		pods, err := typedClient.CoreV1().Pods(namespace).List(ctx, listOptions)
		if err != nil {
			return nil, errorByStatus("get", "Pods", "", err)
		}
		for _, resource := range pods.Items {
			refs = append(refs, &corev1.ResourceRef{
				ApiVersion: "core/v1",
				Kind:       "Pod",
				Name:       resource.Name,
				Namespace:  resource.Namespace,
			})
		}

		// Fetching all the matching deployments
		deployments, err := typedClient.AppsV1().Deployments(namespace).List(ctx, listOptions)
		if err != nil {
			return nil, errorByStatus("get", "Deployments", "", err)
		}
		for _, resource := range deployments.Items {
			refs = append(refs, &corev1.ResourceRef{
				ApiVersion: "apps/v1",
				Kind:       "Deployment",
				Name:       resource.ObjectMeta.Name,
				Namespace:  resource.ObjectMeta.Namespace,
			})
		}

		// Fetching all the matching services
		services, err := typedClient.CoreV1().Services(namespace).List(ctx, listOptions)
		if err != nil {
			return nil, errorByStatus("get", "services", "", err)
		}
		for _, resource := range services.Items {
			refs = append(refs, &corev1.ResourceRef{
				ApiVersion: "core/v1",
				Kind:       "Service",
				Name:       resource.ObjectMeta.Name,
				Namespace:  resource.ObjectMeta.Namespace,
			})
		}

		// Fetching all the matching secrets
		secrets, err := typedClient.CoreV1().Secrets(namespace).List(ctx, listOptions)
		if err != nil {
			return nil, errorByStatus("get", "Secrets", "", err)
		}
		for _, resource := range secrets.Items {
			refs = append(refs, &corev1.ResourceRef{
				ApiVersion: "core/v1",
				Kind:       "Secret",
				Name:       resource.ObjectMeta.Name,
				Namespace:  resource.ObjectMeta.Namespace,
			})
		}
	} else {
		log.Warning(errorByStatus("get", "ConfigMap", cmName, err))
	}

	return &corev1.GetInstalledPackageResourceRefsResponse{
		Context:      request.GetInstalledPackageRef().GetContext(),
		ResourceRefs: refs,
	}, nil
}
