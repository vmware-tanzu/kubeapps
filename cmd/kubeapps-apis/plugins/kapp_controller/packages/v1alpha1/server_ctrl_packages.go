// Copyright 2021-2024 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"carvel.dev/vendir/pkg/vendir/versions"
	vendirversions "carvel.dev/vendir/pkg/vendir/versions/v1alpha1"
	"github.com/bufbuild/connect-go"
	packagingv1alpha1 "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/packaging/v1alpha1"
	datapackagingv1alpha1 "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apiserver/apis/datapackaging/v1alpha1"
	kappctrlpackageinstall "github.com/vmware-tanzu/carvel-kapp-controller/pkg/packageinstall"
	corev1 "github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/plugins/pkg/connecterror"
	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/plugins/pkg/k8sutils"
	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/plugins/pkg/paginate"
	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/plugins/pkg/pkgutils"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	log "k8s.io/klog/v2"
)

const PACKAGES_CHANNEL_BUFFER_SIZE = 20

// GetAvailablePackageSummaries returns the available packages managed by the 'kapp_controller' plugin
func (s *Server) GetAvailablePackageSummaries(ctx context.Context, request *connect.Request[corev1.GetAvailablePackageSummariesRequest]) (*connect.Response[corev1.GetAvailablePackageSummariesResponse], error) {
	// Retrieve parameters from the request
	namespace := request.Msg.GetContext().GetNamespace()
	cluster := request.Msg.GetContext().GetCluster()
	log.InfoS("+kapp-controller GetAvailablePackageSummaries", "cluster", cluster, "namespace", namespace)

	// Retrieve additional parameters from the request
	pageSize := request.Msg.GetPaginationOptions().GetPageSize()
	itemOffset, err := paginate.ItemOffsetFromPageToken(request.Msg.GetPaginationOptions().GetPageToken())
	if err != nil {
		return nil, err
	}
	// Assume the default cluster if none is specified
	if cluster == "" {
		cluster = s.globalPackagingCluster
	}
	// fetch all the package metadatas
	pkgMetadatas, err := s.getPkgMetadatas(ctx, request.Header(), cluster, namespace)
	if err != nil {
		return nil, connecterror.FromK8sError("get", "PackageMetadata", "", err)
	}

	// Filter the package metadatas using any specified filter.
	pkgMetadatas = FilterMetadatas(pkgMetadatas, request.Msg.GetFilterOptions())

	availablePackageSummaries := []*corev1.AvailablePackageSummary{}
	categories := []string{}

	if len(pkgMetadatas) > 0 {
		// Update the slice to be the correct page of results.
		startAt := 0
		if pageSize > 0 {
			startAt = itemOffset
			if startAt > len(pkgMetadatas) {
				return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("Invalid pagination arguments %v", request.Msg.GetPaginationOptions()))
			}
			pkgMetadatas = pkgMetadatas[startAt:]
			if len(pkgMetadatas) > int(pageSize) {
				pkgMetadatas = pkgMetadatas[:pageSize]
			}
		}

		// Create a channel to receive all packages available in the namespace.
		// Using a buffered channel so that we don't block the network request if we
		// can't process fast enough.
		getPkgsChannel := make(chan *datapackagingv1alpha1.Package, PACKAGES_CHANNEL_BUFFER_SIZE)
		var getPkgsError error
		go func() {
			getPkgsError = s.getPkgs(ctx, request.Header(), cluster, namespace, getPkgsChannel)
		}()

		// Skip through the packages until we get to the first item in our
		// paginated results.
		currentPkg := <-getPkgsChannel
		for currentPkg != nil && len(pkgMetadatas) > 0 && currentPkg.Spec.RefName != pkgMetadatas[0].Name {
			currentPkg = <-getPkgsChannel
		}

		availablePackageSummaries = make([]*corev1.AvailablePackageSummary, len(pkgMetadatas))
		pkgsForMeta := []*datapackagingv1alpha1.Package{}
		for i, pkgMetadata := range pkgMetadatas {
			// currentPkg will be nil if the channel is closed and there's no
			// more items to consume.
			if currentPkg == nil {
				return nil, connecterror.FromK8sError("get", "Package", pkgMetadata.Name, fmt.Errorf("no package versions for the package %q", pkgMetadata.Name))
			}
			// The kapp-controller returns both packages and package metadata
			// in order. But some repositories have invalid data (TAP 1.0.2)
			// where a package is present *without* corresponding metadata.
			for currentPkg.Spec.RefName != pkgMetadata.Name {
				if currentPkg.Spec.RefName > pkgMetadata.Name {
					return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("Unexpected order for kapp-controller packages, expected %q, found %q", pkgMetadata.Name, currentPkg.Spec.RefName))
				}
				log.Errorf("Package %q did not have a corresponding metadata (want %q)", currentPkg.Spec.RefName, pkgMetadata.Name)
				currentPkg = <-getPkgsChannel
			}
			// Collect the packages for a particular refName to be able to send the
			// latest semver version. For the moment, kapp-controller just returns
			// CRs with the default alpha sorting of the CR name.
			// Ref https://kubernetes.slack.com/archives/CH8KCCKA5/p1646285201181119
			pkgsForMeta = append(pkgsForMeta, currentPkg)
			currentPkg = <-getPkgsChannel
			for currentPkg != nil && currentPkg.Spec.RefName == pkgMetadata.Name {
				pkgsForMeta = append(pkgsForMeta, currentPkg)
				currentPkg = <-getPkgsChannel
			}
			// At this point, we have all the packages collected that match
			// this ref name, and currentPkg is for the next meta name.
			pkgVersionMap, err := getPkgVersionsMap(pkgsForMeta)
			if err != nil || len(pkgVersionMap[pkgMetadata.Name]) == 0 {
				return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("Unable to calculate package versions map for packages: %v, err: %v", pkgsForMeta, err))
			}
			latestVersion := pkgVersionMap[pkgMetadata.Name][0].version.String()
			availablePackageSummary := s.buildAvailablePackageSummary(pkgMetadata, latestVersion, cluster)
			availablePackageSummaries[i] = availablePackageSummary
			categories = append(categories, availablePackageSummary.Categories...)

			// Reset the packages for the current meta name.
			pkgsForMeta = pkgsForMeta[:0]
		}

		// Verify no error during go routine.
		if getPkgsError != nil {
			return nil, connecterror.FromK8sError("get", "Package", "", err)
		}
	}

	// Only return a next page token if the request was for pagination and
	// the results are a full page.
	nextPageToken := ""
	if pageSize > 0 && len(availablePackageSummaries) == int(pageSize) {
		nextPageToken = fmt.Sprintf("%d", itemOffset+int(pageSize))
	}
	response := &corev1.GetAvailablePackageSummariesResponse{
		AvailablePackageSummaries: availablePackageSummaries,
		Categories:                categories,
		NextPageToken:             nextPageToken,
	}
	return connect.NewResponse(response), nil
}

// GetAvailablePackageVersions returns the package versions managed by the 'kapp_controller' plugin
func (s *Server) GetAvailablePackageVersions(ctx context.Context, request *connect.Request[corev1.GetAvailablePackageVersionsRequest]) (*connect.Response[corev1.GetAvailablePackageVersionsResponse], error) {
	// Retrieve parameters from the request
	namespace := request.Msg.GetAvailablePackageRef().GetContext().GetNamespace()
	cluster := request.Msg.GetAvailablePackageRef().GetContext().GetCluster()
	identifier := request.Msg.GetAvailablePackageRef().GetIdentifier()
	log.InfoS("+kapp-controller GetAvailablePackageVersions", "cluster", cluster, "namespace", namespace, "id", identifier)

	// Validate the request
	if namespace == "" || identifier == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("Required context or identifier not provided"))
	}

	if cluster == "" {
		cluster = s.globalPackagingCluster
	}

	_, pkgName, err := pkgutils.SplitPackageIdentifier(identifier)
	if err != nil {
		return nil, err
	}

	// Use the field selector to return only Package CRs that match on the spec.refName.
	fieldSelector := fmt.Sprintf("spec.refName=%s", pkgName)
	pkgs, err := s.getPkgsWithFieldSelector(ctx, request.Header(), cluster, namespace, fieldSelector)
	if err != nil {
		return nil, connecterror.FromK8sError("get", "Package", "", err)
	}
	pkgVersionsMap, err := getPkgVersionsMap(pkgs)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("Unable to get the PkgVersionsMap: '%w'", err))

	}

	// TODO(minelson): support configurable version summary for kapp-controller pkgs
	// as already done for Helm (see #3588 for more info).
	versions := make([]*corev1.PackageAppVersion, len(pkgVersionsMap[pkgName]))
	for i, v := range pkgVersionsMap[pkgName] {
		// Currently, PkgVersion and AppVersion are the same
		// https://kubernetes.slack.com/archives/CH8KCCKA5/p1636386358322000?thread_ts=1636371493.320900&cid=CH8KCCKA5
		versions[i] = &corev1.PackageAppVersion{
			PkgVersion: v.version.String(),
			AppVersion: v.version.String(),
		}
	}

	return connect.NewResponse(&corev1.GetAvailablePackageVersionsResponse{
		PackageAppVersions: versions,
	}), nil
}

// GetAvailablePackageDetail returns the package metadata managed by the 'kapp_controller' plugin
func (s *Server) GetAvailablePackageDetail(ctx context.Context, request *connect.Request[corev1.GetAvailablePackageDetailRequest]) (*connect.Response[corev1.GetAvailablePackageDetailResponse], error) {
	// Retrieve parameters from the request
	namespace := request.Msg.GetAvailablePackageRef().GetContext().GetNamespace()
	cluster := request.Msg.GetAvailablePackageRef().GetContext().GetCluster()
	identifier := request.Msg.GetAvailablePackageRef().GetIdentifier()
	log.InfoS("+kapp-controller GetAvailablePackageDetail", "cluster", cluster, "namespace", namespace, "identifier", identifier)

	// Retrieve additional parameters from the request
	requestedPkgVersion := request.Msg.GetPkgVersion()

	// Validate the request
	if request.Msg.GetAvailablePackageRef().GetContext() == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("No request AvailablePackageRef.Context provided"))
	}

	if cluster == "" {
		cluster = s.globalPackagingCluster
	}

	_, pkgName, err := pkgutils.SplitPackageIdentifier(identifier)
	if err != nil {
		return nil, err
	}

	// fetch the package metadata
	pkgMetadata, err := s.getPkgMetadata(ctx, request.Header(), cluster, namespace, pkgName)
	if err != nil {
		return nil, connecterror.FromK8sError("get", "PackageMetadata", pkgName, err)
	}

	// Use the field selector to return only Package CRs that match on the spec.refName.
	fieldSelector := fmt.Sprintf("spec.refName=%s", pkgName)
	pkgs, err := s.getPkgsWithFieldSelector(ctx, request.Header(), cluster, namespace, fieldSelector)
	if err != nil {
		return nil, connecterror.FromK8sError("get", "Package", pkgName, err)
	}
	pkgVersionsMap, err := getPkgVersionsMap(pkgs)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("Unable to get the PkgVersionsMap: '%w'", err))
	}

	var foundPkgSemver = &pkgSemver{}
	if requestedPkgVersion != "" {
		// Ensure the version is available.
		for i := range pkgVersionsMap[pkgName] {
			v := pkgVersionsMap[pkgName][i] // avoid implicit memory aliasing
			if v.version.String() == requestedPkgVersion {
				foundPkgSemver = &v
				break
			}
		}
		if foundPkgSemver.version == nil {
			return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("Unable to find %q package with version %q", pkgName, requestedPkgVersion))
		}
	} else {
		// If the pkgVersion wasn't specified, grab the packages to find the latest.
		if len(pkgVersionsMap[pkgName]) > 0 {
			foundPkgSemver = &pkgVersionsMap[pkgName][0]
			requestedPkgVersion = foundPkgSemver.version.String()
		} else {
			return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("Unable to find any versions for the package %q", pkgName))
		}
	}

	availablePackageDetail, err := s.buildAvailablePackageDetail(pkgMetadata, requestedPkgVersion, foundPkgSemver, cluster)
	if err != nil {
		return nil, connecterror.FromK8sError("create", "AvailablePackageDetail", pkgMetadata.Name, err)

	}

	return connect.NewResponse(&corev1.GetAvailablePackageDetailResponse{
		AvailablePackageDetail: availablePackageDetail,
	}), nil
}

// GetInstalledPackageSummaries returns the installed packages managed by the 'kapp_controller' plugin
func (s *Server) GetInstalledPackageSummaries(ctx context.Context, request *connect.Request[corev1.GetInstalledPackageSummariesRequest]) (*connect.Response[corev1.GetInstalledPackageSummariesResponse], error) {
	// Retrieve parameters from the request
	namespace := request.Msg.GetContext().GetNamespace()
	cluster := request.Msg.GetContext().GetCluster()
	log.Info("+kapp-controller GetInstalledPackageSummaries", "cluster", cluster, "namespace", namespace)

	// Retrieve additional parameters from the request
	pageSize := request.Msg.GetPaginationOptions().GetPageSize()
	itemOffset, err := paginate.ItemOffsetFromPageToken(request.Msg.GetPaginationOptions().GetPageToken())
	if err != nil {
		return nil, err
	}

	// Assume the default cluster if none is specified
	if cluster == "" {
		cluster = s.globalPackagingCluster
	}

	// retrieve the paginated list of installed packages
	// TODO(agamez): we should be paginating this request rather than requesting everything every time
	pkgInstalls, err := s.getPkgInstalls(ctx, request.Header(), cluster, namespace)
	if err != nil {
		return nil, connecterror.FromK8sError("get", "PackageInstall", "", err)
	}

	// paginate the list of results
	installedPkgSummaries := []*corev1.InstalledPackageSummary{}

	if len(pkgInstalls) > 0 {
		//nolint:ineffassign
		startAt := -1
		if pageSize > 0 {
			startAt = itemOffset
			if startAt > len(pkgInstalls) {
				return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("Invalid pagination arguments %v", request.Msg.GetPaginationOptions()))
			}
			pkgInstalls = pkgInstalls[startAt:]
			if len(pkgInstalls) > int(pageSize) {
				pkgInstalls = pkgInstalls[:pageSize]
			}
		}

		// Collect a set of all package names with their corresponding package
		// metadata and package version data.
		// Complication 1: PackageMetadata and Package resources are not CRs,
		// but rather aggregated API resources, which means a PackageInstall in
		// namespace "default" may refer to a Package (and metadata) in the same
		// "default" namespace, or from the global namespace.
		// Complication 2: PackageMetadata and Package resources for can be
		// present in different namespaces with the same name and different
		// versions etc.
		// With this in mind, we create a data structure that records the
		// metadata and versions per package refName per namespace.
		// TODO(minelson) move this struct initialization out.
		type pkgMetaAndVersionsData struct {
			meta     *datapackagingv1alpha1.PackageMetadata
			versions map[string]*datapackagingv1alpha1.Package
		}
		pkgData := make(map[string]map[string]*pkgMetaAndVersionsData)
		for _, pkgInstall := range pkgInstalls {
			pkgDataForNamespaces, ok := pkgData[pkgInstall.Spec.PackageRef.RefName]
			if !ok {
				pkgDataForNamespaces = map[string]*pkgMetaAndVersionsData{}
				pkgData[pkgInstall.Spec.PackageRef.RefName] = pkgDataForNamespaces
			}
			// As each package install could potentially be from a pkg in the same
			// namespace or a package in the global namespace, we track both.
			for _, ns := range []string{pkgInstall.Namespace, s.pluginConfig.globalPackagingNamespace} {
				pkgData, ok := pkgDataForNamespaces[ns]
				if !ok {
					pkgData = &pkgMetaAndVersionsData{
						versions: make(map[string]*datapackagingv1alpha1.Package),
					}
					pkgDataForNamespaces[ns] = pkgData
				}
				_, ok = pkgData.versions[pkgInstall.Status.Version]
				if !ok {
					pkgData.versions[pkgInstall.Status.Version] = nil
				}
			}
		}

		// First get all the package metadatas for the namespace (or across
		// namespaces) and filter to match the pkgInstalls. While filtering, we
		// populate the collected package data with the metadata.
		pkgMetadatas, err := s.getPkgMetadatas(ctx, request.Header(), cluster, namespace)
		if err != nil {
			return nil, connecterror.FromK8sError("get", "PackageMetadata", "", err)
		}
		for _, pm := range pkgMetadatas {
			if pkgDataForNamespaces, ok := pkgData[pm.Name]; ok {
				if pkgData, ok := pkgDataForNamespaces[pm.Namespace]; ok {
					pkgData.meta = pm
				}
			}
		}

		// Create a channel to receive all packages available in the namespace
		// (or across namespaces) Using a buffered channel so that we don't
		// block the network request if we can't process fast enough.
		getPkgsChannel := make(chan *datapackagingv1alpha1.Package, PACKAGES_CHANNEL_BUFFER_SIZE)
		var getPkgsError error
		go func() {
			getPkgsError = s.getPkgs(ctx, request.Header(), cluster, namespace, getPkgsChannel)
		}()

		// For each package, we check if we need it to populate our
		// package data.
		pkgsForVersionMap := []*datapackagingv1alpha1.Package{}
		for pkg := range getPkgsChannel {
			pkgDataForNamespaces, ok := pkgData[pkg.Spec.RefName]
			if !ok {
				continue
			}
			pkgData, ok := pkgDataForNamespaces[pkg.Namespace]
			if !ok {
				continue
			}
			pkgsForVersionMap = append(pkgsForVersionMap, pkg)
			_, ok = pkgData.versions[pkg.Spec.Version]
			if !ok {
				continue
			}
			pkgData.versions[pkg.Spec.Version] = pkg
		}

		// Verify no error during go routine.
		if getPkgsError != nil {
			return nil, connecterror.FromK8sError("get", "Package", "", err)
		}

		// Calculate the version map for all packages that we're interested
		// in for this paginated call.
		// TODO(minelson): Check if getPkgVersionsMap currently handles packages
		// in different namespaces (suspect not).
		pkgVersionsMap, err := getPkgVersionsMap(pkgsForVersionMap)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("Unable to calculate package versions map: %w", err))
		}

		// We now have all the data we need to return the result:
		for _, pkgi := range pkgInstalls {
			pkgName := pkgi.Spec.PackageRef.RefName
			pkgDataForNamespaces := pkgData[pkgName]
			// Check if there was package metadata for the specific namespace.
			pkgData := pkgDataForNamespaces[pkgi.Namespace]
			var ok bool
			if pkgData.meta == nil {
				pkgData, ok = pkgDataForNamespaces[s.pluginConfig.globalPackagingNamespace]
				// Ignore packages which do not have associated metadata
				// available. See https://github.com/vmware-tanzu/kubeapps/issues/4901
				if !ok || pkgData.meta == nil {
					log.Errorf("+kapp-controller GetInstalledPackageSummary: No corresponding package metadata found for package %q. Ignoring package.", pkgName)
					continue
				}
			}
			// generate the installedPackageSummary from the fetched information
			installedPackageSummary, err := s.buildInstalledPackageSummary(pkgi, pkgData.meta, pkgVersionsMap, cluster)
			if err != nil {
				return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("Unable to create the InstalledPackageSummary: %w", err))
			}

			// append the availablePackageSummary to the slice
			installedPkgSummaries = append(installedPkgSummaries, installedPackageSummary)
		}
	}

	// Only return a next page token if the request was for pagination and
	// the results are a full page.
	nextPageToken := ""
	if pageSize > 0 && len(installedPkgSummaries) == int(pageSize) {
		nextPageToken = fmt.Sprintf("%d", itemOffset+int(pageSize))
	}
	response := &corev1.GetInstalledPackageSummariesResponse{
		InstalledPackageSummaries: installedPkgSummaries,
		NextPageToken:             nextPageToken,
	}
	return connect.NewResponse(response), nil
}

// GetInstalledPackageDetail returns the package metadata managed by the 'kapp_controller' plugin
func (s *Server) GetInstalledPackageDetail(ctx context.Context, request *connect.Request[corev1.GetInstalledPackageDetailRequest]) (*connect.Response[corev1.GetInstalledPackageDetailResponse], error) {
	// Retrieve parameters from the request
	cluster := request.Msg.GetInstalledPackageRef().GetContext().GetCluster()
	namespace := request.Msg.GetInstalledPackageRef().GetContext().GetNamespace()
	installedPackageRefId := request.Msg.GetInstalledPackageRef().GetIdentifier()
	log.InfoS("+kapp-controller GetInstalledPackageDetail", "cluster", cluster, "namespace", namespace, "id", installedPackageRefId)

	if cluster == "" {
		cluster = s.globalPackagingCluster
	}

	typedClient, err := s.clientGetter.Typed(request.Header(), cluster)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("Unable to get the k8s client: '%w'", err))
	}

	// fetch the package install
	pkgInstall, err := s.getPkgInstall(ctx, request.Header(), cluster, namespace, installedPackageRefId)
	if err != nil {
		return nil, connecterror.FromK8sError("get", "PackageInstall", installedPackageRefId, err)
	}

	// fetch the resulting deployed app after the installation
	app, err := s.getApp(ctx, request.Header(), cluster, namespace, installedPackageRefId)
	if err != nil {
		return nil, connecterror.FromK8sError("get", "App", installedPackageRefId, err)
	}

	// retrieve the package metadata associated with this installed package
	pkgName := pkgInstall.Spec.PackageRef.RefName
	pkgMetadata, err := s.getPkgMetadata(ctx, request.Header(), cluster, namespace, pkgName)
	if err != nil {
		return nil, connecterror.FromK8sError("get", "PackageMetadata", pkgName, err)
	}

	// Use the field selector to return only Package CRs that match on the spec.refName.
	fieldSelector := fmt.Sprintf("spec.refName=%s", pkgMetadata.Name)
	pkgs, err := s.getPkgsWithFieldSelector(ctx, request.Header(), cluster, namespace, fieldSelector)
	if err != nil {
		return nil, connecterror.FromK8sError("get", "Package", "", err)
	}
	pkgVersionsMap, err := getPkgVersionsMap(pkgs)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("Unable to get the PkgVersionsMap: '%w'", err))
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
					log.Warningf("The referenced secret does not exist: %s", connecterror.FromK8sError("get", "Secret", secretRefName, err).Error())
				} else {
					return nil, connecterror.FromK8sError("get", "Secret", secretRefName, err)
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
	valuesApplied = strings.TrimSuffix(valuesApplied, "---")

	installedPackageDetail, err := s.buildInstalledPackageDetail(pkgInstall, pkgMetadata, pkgVersionsMap, app, valuesApplied, cluster)
	if err != nil {
		return nil, connecterror.FromK8sError("create", "InstalledPackageDetail", pkgInstall.Name, err)

	}

	response := &corev1.GetInstalledPackageDetailResponse{
		InstalledPackageDetail: installedPackageDetail,
	}
	return connect.NewResponse(response), nil
}

// CreateInstalledPackage creates an installed package managed by the 'kapp_controller' plugin
func (s *Server) CreateInstalledPackage(ctx context.Context, request *connect.Request[corev1.CreateInstalledPackageRequest]) (*connect.Response[corev1.CreateInstalledPackageResponse], error) {
	// Retrieve parameters from the request
	targetCluster := request.Msg.GetTargetContext().GetCluster()
	targetNamespace := request.Msg.GetTargetContext().GetNamespace()
	installedPackageName := request.Msg.GetName()

	log.InfoS("+kapp-controller CreateInstalledPackage %s", "cluster", targetCluster, "namespace", targetNamespace, "id", installedPackageName)

	// Validate the request
	if request.Msg.GetAvailablePackageRef().GetContext().GetNamespace() == "" || request.Msg.GetAvailablePackageRef().GetIdentifier() == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("Required context or identifier not provided"))
	}
	if request == nil || request.Msg.GetAvailablePackageRef() == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("No request AvailablePackageRef provided"))
	}
	if request.Msg.GetName() == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("No request Name provided"))
	}
	if request.Msg.GetTargetContext() == nil || request.Msg.GetTargetContext().GetNamespace() == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("No request TargetContext namespace provided"))
	}
	if request.Msg.GetReconciliationOptions() == nil || request.Msg.GetReconciliationOptions().GetServiceAccountName() == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("No request ReconciliationOptions serviceAccountName provided"))
	}

	// Retrieve additional parameters from the request
	identifier := request.Msg.GetAvailablePackageRef().GetIdentifier()
	packageCluster := request.Msg.GetAvailablePackageRef().GetContext().GetCluster()
	packageNamespace := request.Msg.GetAvailablePackageRef().GetContext().GetNamespace()
	reconciliationOptions := request.Msg.GetReconciliationOptions()
	pkgVersion := request.Msg.GetPkgVersionReference().GetVersion()
	values := request.Msg.GetValues()

	_, pkgName, err := pkgutils.SplitPackageIdentifier(identifier)
	if err != nil {
		return nil, err
	}

	if targetCluster == "" {
		targetCluster = s.globalPackagingCluster
	}

	if targetCluster != packageCluster {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("Installing packages in other clusters in not supported yet"))
	}

	typedClient, err := s.clientGetter.Typed(request.Header(), targetCluster)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("Unable to get the k8s client: '%w'", err))
	}

	// fetch the package metadata
	pkgMetadata, err := s.getPkgMetadata(ctx, request.Header(), packageCluster, packageNamespace, pkgName)
	if err != nil {
		return nil, connecterror.FromK8sError("get", "PackageMetadata", pkgName, err)
	}

	// build a new secret object with the values
	secret, err := s.buildSecret(installedPackageName, values, targetNamespace)
	if err != nil {
		return nil, connecterror.FromK8sError("create", "Secret", installedPackageName, err)
	}

	// build a new pkgInstall object
	newPkgInstall, err := s.buildPkgInstall(installedPackageName, targetCluster, targetNamespace, pkgMetadata.Name, pkgVersion, reconciliationOptions, secret)
	if err != nil {
		return nil, connect.NewError(connect.CodeOf(err), fmt.Errorf("Unable to create the PackageInstall '%s' due to '%w'", installedPackageName, err))
	}

	// create the Secret in the cluster
	// TODO(agamez): check when is the best moment to create this object.
	// See if we can delay the creation until the PackageInstall is successfully created.
	createdSecret, err := typedClient.CoreV1().Secrets(targetNamespace).Create(ctx, secret, metav1.CreateOptions{})
	if createdSecret == nil || err != nil {
		return nil, connecterror.FromK8sError("create", "Secret", secret.Name, err)
	}

	// create the PackageInstall in the cluster
	createdPkgInstall, err := s.createPkgInstall(ctx, request.Header(), targetCluster, targetNamespace, newPkgInstall)
	if err != nil {
		// clean-up the secret if something fails
		err := typedClient.CoreV1().Secrets(targetNamespace).Delete(ctx, secret.Name, metav1.DeleteOptions{})
		if err != nil {
			return nil, connecterror.FromK8sError("delete", "Secret", secret.Name, err)
		}
		return nil, connecterror.FromK8sError("create", "PackageInstall", newPkgInstall.Name, err)
	}

	resource, err := s.getAppResource(request.Header(), targetCluster, targetNamespace)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("Unable to get the App resource: '%w'", err))
	}
	// The InstalledPackage is considered as created once the associated kapp App gets created,
	// so we actively wait for the App CR to be present in the cluster before returning OK
	err = k8sutils.WaitForResource(ctx, resource, newPkgInstall.Name, time.Second*1, time.Second*time.Duration(s.pluginConfig.timeoutSeconds))
	if err != nil {
		// clean-up the secret if something fails
		err := typedClient.CoreV1().Secrets(targetNamespace).Delete(ctx, secret.Name, metav1.DeleteOptions{})
		if err != nil {
			return nil, connecterror.FromK8sError("delete", "Secret", secret.Name, err)
		}
		// clean-up the package install if something fails
		err = s.deletePkgInstall(ctx, request.Header(), targetCluster, targetNamespace, newPkgInstall.Name)
		if err != nil {
			return nil, connecterror.FromK8sError("delete", "PackageInstall", newPkgInstall.Name, err)
		}
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("Timeout exceeded (%v s) waiting for resource to be installed: '%w'", s.pluginConfig.timeoutSeconds, err))
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

	return connect.NewResponse(&corev1.CreateInstalledPackageResponse{
		InstalledPackageRef: installedRef,
	}), nil
}

// UpdateInstalledPackage Updates an installed package managed by the 'kapp_controller' plugin
func (s *Server) UpdateInstalledPackage(ctx context.Context, request *connect.Request[corev1.UpdateInstalledPackageRequest]) (*connect.Response[corev1.UpdateInstalledPackageResponse], error) {
	// Validate the request
	if request == nil || request.Msg.GetInstalledPackageRef() == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("No request AvailablePackageRef provided"))
	}
	// Retrieve parameters from the request
	packageCluster := request.Msg.GetInstalledPackageRef().GetContext().GetCluster()
	packageNamespace := request.Msg.GetInstalledPackageRef().GetContext().GetNamespace()
	installedPackageName := request.Msg.GetInstalledPackageRef().GetIdentifier()
	log.InfoS("+kapp-controller UpdateInstalledPackage", "cluster", packageCluster, "namespace", "packageNamespace", "id", installedPackageName)

	if request.Msg.GetInstalledPackageRef().GetContext().GetNamespace() == "" || request.Msg.GetInstalledPackageRef().GetIdentifier() == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("Required context or identifier not provided"))
	}
	if request.Msg.GetInstalledPackageRef().GetIdentifier() == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("No request Name provided"))
	}

	// Retrieve additional parameters from the request
	reconciliationOptions := request.Msg.GetReconciliationOptions()
	pkgVersion := request.Msg.GetPkgVersionReference().GetVersion()
	values := request.Msg.GetValues()

	if packageCluster == "" {
		packageCluster = s.globalPackagingCluster
	}

	typedClient, err := s.clientGetter.Typed(request.Header(), packageCluster)
	if err != nil {
		return nil, err
	}

	// fetch the package install
	pkgInstall, err := s.getPkgInstall(ctx, request.Header(), packageCluster, packageNamespace, installedPackageName)
	if err != nil {
		return nil, connecterror.FromK8sError("get", "PackageInstall", installedPackageName, err)
	}

	// Calculate the constraints and prerelease fields
	versionConstraints, err := pkgutils.VersionConstraintWithUpgradePolicy(pkgVersion, s.pluginConfig.defaultUpgradePolicy)
	if err != nil {
		return nil, err
	}
	prereleases := prereleasesVersionSelection(s.pluginConfig.defaultPrereleasesVersionSelection)

	versionSelection := &vendirversions.VersionSelectionSemver{
		Constraints: versionConstraints,
		Prereleases: prereleases,
	}

	// Ensure the selected version can be, actually installed to let the user know before installing
	eligibleVersion, err := versions.HighestConstrainedVersion([]string{pkgVersion}, vendirversions.VersionSelection{Semver: versionSelection})
	if eligibleVersion == "" || err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("The selected version %q is not eligible to be installed: %w", pkgVersion, err))
	}

	// Set the versionSelection
	pkgInstall.Spec.PackageRef.VersionSelection = versionSelection

	// Allow this PackageInstall to be downgraded
	// https://carvel.dev/kapp-controller/docs/v0.32.0/package-consumer-concepts/#downgrading
	if s.pluginConfig.defaultAllowDowngrades {
		if pkgInstall.ObjectMeta.Annotations == nil {
			pkgInstall.ObjectMeta.Annotations = map[string]string{}
		}
		pkgInstall.ObjectMeta.Annotations[kappctrlpackageinstall.DowngradableAnnKey] = ""
	}

	// Update the rest of the fields
	if reconciliationOptions != nil {
		if pkgInstall.Spec.SyncPeriod, err = pkgutils.ToDuration(reconciliationOptions.Interval); err != nil {
			return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("The interval is invalid: %w", err))
		}
		if reconciliationOptions.ServiceAccountName != "" {
			pkgInstall.Spec.ServiceAccountName = reconciliationOptions.ServiceAccountName
		}
		pkgInstall.Spec.Paused = reconciliationOptions.Suspend
	}

	// update the pkgInstall in the server
	updatedPkgInstall, err := s.updatePkgInstall(ctx, request.Header(), packageCluster, packageNamespace, pkgInstall)
	if err != nil {
		return nil, connecterror.FromK8sError("update", "PackageInstall", installedPackageName, err)
	}

	// Update the values.yaml values file if any is passed, otherwise, delete the values
	if values != "" {
		secret, err := s.buildSecret(installedPackageName, values, packageNamespace)
		if err != nil {
			return nil, connecterror.FromK8sError("update", "Secret", secret.Name, err)
		}
		updatedSecret, err := typedClient.CoreV1().Secrets(packageNamespace).Update(ctx, secret, metav1.UpdateOptions{})
		if updatedSecret == nil || err != nil {
			return nil, connecterror.FromK8sError("update", "Secret", secret.Name, err)
		}

		if updatedSecret != nil {
			// Similar logic as in https://github.com/vmware-tanzu/carvel-kapp-controller/blob/v0.32.0/cli/pkg/kctrl/cmd/package/installed/create_or_update.go#L505
			pkgInstall.Spec.Values = []packagingv1alpha1.PackageInstallValues{{
				SecretRef: &packagingv1alpha1.PackageInstallValuesSecretRef{
					// The secret name should have the format: <name>-<namespace> as per:
					// https://github.com/vmware-tanzu/carvel-kapp-controller/blob/v0.32.0/cli/pkg/kctrl/cmd/package/installed/created_resource_annotations.go#L19
					Name: updatedSecret.Name,
				},
			}}
		}

	} else {
		// Delete all the associated secrets
		// TODO(agamez): maybe it's too aggressive and we should be deleting only those secrets created by this plugin
		// See https://github.com/vmware-tanzu/kubeapps/pull/3790#discussion_r754797195
		for _, packageInstallValue := range pkgInstall.Spec.Values {
			secretId := packageInstallValue.SecretRef.Name
			err := typedClient.CoreV1().Secrets(packageNamespace).Delete(ctx, secretId, metav1.DeleteOptions{})
			if errors.IsNotFound(err) {
				log.Warningf("The referenced secret does not exist: %s", connecterror.FromK8sError("get", "Secret", secretId, err).Error())
			} else {
				return nil, connecterror.FromK8sError("delete", "Secret", secretId, err)
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
	return connect.NewResponse(&corev1.UpdateInstalledPackageResponse{
		InstalledPackageRef: updatedRef,
	}), nil
}

// DeleteInstalledPackage Deletes an installed package managed by the 'kapp_controller' plugin
func (s *Server) DeleteInstalledPackage(ctx context.Context, request *connect.Request[corev1.DeleteInstalledPackageRequest]) (*connect.Response[corev1.DeleteInstalledPackageResponse], error) {
	// Validate the request
	if request == nil || request.Msg.GetInstalledPackageRef() == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("No request InstalledPackageRef provided"))
	}
	// Retrieve parameters from the request
	namespace := request.Msg.GetInstalledPackageRef().GetContext().GetNamespace()
	cluster := request.Msg.GetInstalledPackageRef().GetContext().GetCluster()
	identifier := request.Msg.GetInstalledPackageRef().GetIdentifier()
	log.InfoS("+kapp-controller DeleteInstalledPackage", "namespace", namespace, "cluster", cluster, "id", identifier)

	if request.Msg.GetInstalledPackageRef().GetContext().GetNamespace() == "" || request.Msg.GetInstalledPackageRef().GetIdentifier() == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("Required context or identifier not provided"))
	}

	if cluster == "" {
		cluster = s.globalPackagingCluster
	}

	typedClient, err := s.clientGetter.Typed(request.Header(), cluster)
	if err != nil {
		return nil, err
	}

	pkgInstall, err := s.getPkgInstall(ctx, request.Header(), cluster, namespace, identifier)
	if err != nil {
		return nil, connecterror.FromK8sError("get", "PackageInstall", identifier, err)
	}

	// Delete the package install
	err = s.deletePkgInstall(ctx, request.Header(), cluster, namespace, identifier)
	if err != nil {
		return nil, connecterror.FromK8sError("delete", "PackageInstall", identifier, err)
	}

	// Delete all the associated secrets
	// TODO(agamez): maybe it's too aggressive and we should be deleting only those secrets created by this plugin
	// See https://github.com/vmware-tanzu/kubeapps/pull/3790#discussion_r754797195
	for _, packageInstallValue := range pkgInstall.Spec.Values {
		secretId := packageInstallValue.SecretRef.Name
		err := typedClient.CoreV1().Secrets(namespace).Delete(ctx, secretId, metav1.DeleteOptions{})
		if err != nil {
			if errors.IsNotFound(err) {
				log.Warningf("The referenced secret does not exist: %s", connecterror.FromK8sError("get", "Secret", secretId, err).Error())
			} else {
				return nil, connecterror.FromK8sError("delete", "Secret", secretId, err)
			}
		}
	}
	return connect.NewResponse(&corev1.DeleteInstalledPackageResponse{}), nil
}

// GetInstalledPackageResourceRefs returns the references for the k8s resources of an installed package managed by the 'kapp_controller' plugin
func (s *Server) GetInstalledPackageResourceRefs(ctx context.Context, request *connect.Request[corev1.GetInstalledPackageResourceRefsRequest]) (*connect.Response[corev1.GetInstalledPackageResourceRefsResponse], error) {
	// Retrieve parameters from the request
	cluster := request.Msg.GetInstalledPackageRef().GetContext().GetCluster()
	namespace := request.Msg.GetInstalledPackageRef().GetContext().GetNamespace()
	installedPackageRefId := request.Msg.GetInstalledPackageRef().GetIdentifier()
	log.InfoS("+kapp-controller GetInstalledPackageResourceRefs", "cluster", cluster, "namespace", namespace, "id", installedPackageRefId)

	if cluster == "" {
		cluster = s.globalPackagingCluster
	}

	// get the list of every k8s resource matching ResourceRef
	refs, err := s.inspectKappK8sResources(request.Header(), cluster, namespace, installedPackageRefId)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&corev1.GetInstalledPackageResourceRefsResponse{
		Context:      request.Msg.GetInstalledPackageRef().GetContext(),
		ResourceRefs: refs,
	}), nil
}

func (s *Server) GetAvailablePackageMetadatas(ctx context.Context, request *connect.Request[corev1.GetAvailablePackageMetadatasRequest]) (*connect.Response[corev1.GetAvailablePackageMetadatasResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, fmt.Errorf("Unimplemented"))
}
