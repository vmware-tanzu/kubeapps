// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	corev1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	"github.com/kubeapps/kubeapps/cmd/kubeapps-apis/plugins/pkg/k8sutils"
	"github.com/kubeapps/kubeapps/cmd/kubeapps-apis/plugins/pkg/paginate"
	"github.com/kubeapps/kubeapps/cmd/kubeapps-apis/plugins/pkg/pkgutils"
	"github.com/kubeapps/kubeapps/cmd/kubeapps-apis/plugins/pkg/statuserror"
	packagingv1alpha1 "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/packaging/v1alpha1"
	datapackagingv1alpha1 "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apiserver/apis/datapackaging/v1alpha1"
	kappctrlpackageinstall "github.com/vmware-tanzu/carvel-kapp-controller/pkg/packageinstall"
	"github.com/vmware-tanzu/carvel-vendir/pkg/vendir/versions"
	vendirversions "github.com/vmware-tanzu/carvel-vendir/pkg/vendir/versions/v1alpha1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	log "k8s.io/klog/v2"
)

const PACKAGES_CHANNEL_BUFFER_SIZE = 20

// GetAvailablePackageSummaries returns the available packages managed by the 'kapp_controller' plugin
func (s *Server) GetAvailablePackageSummaries(ctx context.Context, request *corev1.GetAvailablePackageSummariesRequest) (*corev1.GetAvailablePackageSummariesResponse, error) {
	// Retrieve parameters from the request
	namespace := request.GetContext().GetNamespace()
	cluster := request.GetContext().GetCluster()
	contextMsg := fmt.Sprintf("(cluster=%q, namespace=%q)", cluster, namespace)
	log.Infof("+kapp-controller GetAvailablePackageSummaries %s", contextMsg)

	// Retrieve additional parameters from the request
	pageSize := request.GetPaginationOptions().GetPageSize()
	pageOffset, err := paginate.PageOffsetFromAvailableRequest(request)
	if err != nil {
		return nil, err
	}
	// Assume the default cluster if none is specified
	if cluster == "" {
		cluster = s.globalPackagingCluster
	}
	// fetch all the package metadatas
	// TODO(minelson): We should be grabbing only the requested page
	// of results here.
	pkgMetadatas, err := s.getPkgMetadatas(ctx, cluster, namespace)
	if err != nil {
		return nil, statuserror.FromK8sError("get", "PackageMetadata", "", err)
	}
	// Until the above request uses the pagination, update the slice
	// to be the correct page of results.
	startAt := 0
	if pageSize > 0 {
		startAt = int(pageSize) * pageOffset
		if startAt > len(pkgMetadatas) {
			return nil, status.Errorf(codes.InvalidArgument, "invalid pagination arguments %v", request.GetPaginationOptions())
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
		getPkgsError = s.getPkgs(ctx, cluster, namespace, getPkgsChannel)
	}()

	// Skip through the packages until we get to the first item in our
	// paginated results.
	currentPkg := <-getPkgsChannel
	for currentPkg != nil && currentPkg.Spec.RefName != pkgMetadatas[0].Name {
		currentPkg = <-getPkgsChannel
	}

	availablePackageSummaries := make([]*corev1.AvailablePackageSummary, len(pkgMetadatas))
	categories := []string{}
	pkgsForMeta := []*datapackagingv1alpha1.Package{}
	for i, pkgMetadata := range pkgMetadatas {
		// currentPkg will be nil if the channel is closed and there's no
		// more items to consume.
		if currentPkg == nil {
			return nil, statuserror.FromK8sError("get", "Package", pkgMetadata.Name, fmt.Errorf("no package versions for the package %q", pkgMetadata.Name))
		}
		// The kapp-controller returns both packages and package metadata
		// in order.
		if currentPkg.Spec.RefName != pkgMetadata.Name {
			return nil, status.Errorf(codes.Internal, fmt.Sprintf("unexpected order for kapp-controller packages, expected %q, found %q", pkgMetadata.Name, currentPkg.Spec.RefName))
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
			return nil, status.Errorf(codes.Internal, fmt.Sprintf("unable to calculate package versions map for packages: %v, err: %v", pkgsForMeta, err))
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
		return nil, statuserror.FromK8sError("get", "Package", "", err)
	}

	// Only return a next page token if the request was for pagination and
	// the results are a full page.
	nextPageToken := ""
	if pageSize > 0 && len(availablePackageSummaries) == int(pageSize) {
		nextPageToken = fmt.Sprintf("%d", pageOffset+1)
	}
	response := &corev1.GetAvailablePackageSummariesResponse{
		AvailablePackageSummaries: availablePackageSummaries,
		Categories:                categories,
		NextPageToken:             nextPageToken,
	}
	return response, nil
}

// GetAvailablePackageVersions returns the package versions managed by the 'kapp_controller' plugin
func (s *Server) GetAvailablePackageVersions(ctx context.Context, request *corev1.GetAvailablePackageVersionsRequest) (*corev1.GetAvailablePackageVersionsResponse, error) {
	// Retrieve parameters from the request
	namespace := request.GetAvailablePackageRef().GetContext().GetNamespace()
	cluster := request.GetAvailablePackageRef().GetContext().GetCluster()
	identifier := request.GetAvailablePackageRef().GetIdentifier()
	contextMsg := fmt.Sprintf("(cluster=%q, namespace=%q, id=%q)", cluster, namespace, identifier)
	log.Infof("+kapp-controller GetAvailablePackageVersions %s", contextMsg)

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
		return nil, statuserror.FromK8sError("get", "Package", "", err)
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
	// Retrieve parameters from the request
	namespace := request.GetAvailablePackageRef().GetContext().GetNamespace()
	cluster := request.GetAvailablePackageRef().GetContext().GetCluster()
	identifier := request.GetAvailablePackageRef().GetIdentifier()
	contextMsg := fmt.Sprintf("(cluster=%q, namespace=%q, id=%q)", cluster, namespace, identifier)
	log.Infof("+kapp-controller GetAvailablePackageDetail %s", contextMsg)

	// Retrieve additional parameters from the request
	requestedPkgVersion := request.GetPkgVersion()

	// Validate the request
	if request.GetAvailablePackageRef().GetContext() == nil {
		return nil, status.Errorf(codes.InvalidArgument, "no request AvailablePackageRef.Context provided")
	}

	if cluster == "" {
		cluster = s.globalPackagingCluster
	}

	// fetch the package metadata
	pkgMetadata, err := s.getPkgMetadata(ctx, cluster, namespace, identifier)
	if err != nil {
		return nil, statuserror.FromK8sError("get", "PackageMetadata", identifier, err)
	}

	// Use the field selector to return only Package CRs that match on the spec.refName.
	fieldSelector := fmt.Sprintf("spec.refName=%s", identifier)
	pkgs, err := s.getPkgsWithFieldSelector(ctx, cluster, namespace, fieldSelector)
	if err != nil {
		return nil, statuserror.FromK8sError("get", "Package", identifier, err)
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
		return nil, statuserror.FromK8sError("create", "AvailablePackageDetail", pkgMetadata.Name, err)

	}

	return &corev1.GetAvailablePackageDetailResponse{
		AvailablePackageDetail: availablePackageDetail,
	}, nil
}

// GetInstalledPackageSummaries returns the installed packages managed by the 'kapp_controller' plugin
func (s *Server) GetInstalledPackageSummaries(ctx context.Context, request *corev1.GetInstalledPackageSummariesRequest) (*corev1.GetInstalledPackageSummariesResponse, error) {
	// Retrieve parameters from the request
	namespace := request.GetContext().GetNamespace()
	cluster := request.GetContext().GetCluster()
	contextMsg := fmt.Sprintf("(cluster=%q, namespace=%q)", cluster, namespace)
	log.Infof("+kapp-controller GetInstalledPackageSummaries %s", contextMsg)

	// Retrieve additional parameters from the request
	pageSize := request.GetPaginationOptions().GetPageSize()
	pageOffset, err := paginate.PageOffsetFromInstalledRequest(request)
	if err != nil {
		return nil, err
	}

	// Assume the default cluster if none is specified
	if cluster == "" {
		cluster = s.globalPackagingCluster
	}

	// retrieve the list of installed packages
	// TODO(agamez): we should be paginating this request rather than requesting everything every time
	pkgInstalls, err := s.getPkgInstalls(ctx, cluster, namespace)
	if err != nil {
		return nil, statuserror.FromK8sError("get", "PackageInstall", "", err)
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
					pkgMetadata, err := s.getPkgMetadata(ctx, cluster, pkgInstall.ObjectMeta.Namespace, pkgInstall.Spec.PackageRef.RefName)
					if err != nil {
						return statuserror.FromK8sError("get", "PackageMetadata", pkgInstall.Spec.PackageRef.RefName, err)
					}

					// Use the field selector to return only Package CRs that match on the spec.refName.
					fieldSelector := fmt.Sprintf("spec.refName=%s", pkgInstall.Spec.PackageRef.RefName)
					pkgs, err := s.getPkgsWithFieldSelector(ctx, cluster, pkgInstall.ObjectMeta.Namespace, fieldSelector)
					if err != nil {
						return statuserror.FromK8sError("get", "Package", "", err)
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
	// Retrieve parameters from the request
	cluster := request.GetInstalledPackageRef().GetContext().GetCluster()
	namespace := request.GetInstalledPackageRef().GetContext().GetNamespace()
	installedPackageRefId := request.GetInstalledPackageRef().GetIdentifier()
	contextMsg := fmt.Sprintf("(cluster=%q, namespace=%q, id=%q)", cluster, namespace, installedPackageRefId)
	log.Infof("+kapp-controller GetInstalledPackageDetail %s", contextMsg)

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
		return nil, statuserror.FromK8sError("get", "PackageInstall", installedPackageRefId, err)
	}

	// fetch the resulting deployed app after the installation
	app, err := s.getApp(ctx, cluster, namespace, installedPackageRefId)
	if err != nil {
		return nil, statuserror.FromK8sError("get", "App", installedPackageRefId, err)
	}

	// retrieve the package metadata associated with this installed package
	pkgName := pkgInstall.Spec.PackageRef.RefName
	pkgMetadata, err := s.getPkgMetadata(ctx, cluster, namespace, pkgName)
	if err != nil {
		return nil, statuserror.FromK8sError("get", "PackageMetadata", pkgName, err)
	}

	// Use the field selector to return only Package CRs that match on the spec.refName.
	fieldSelector := fmt.Sprintf("spec.refName=%s", pkgMetadata.Name)
	pkgs, err := s.getPkgsWithFieldSelector(ctx, cluster, namespace, fieldSelector)
	if err != nil {
		return nil, statuserror.FromK8sError("get", "Package", "", err)
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
					log.Warningf("The referenced secret does not exist: %s", statuserror.FromK8sError("get", "Secret", secretRefName, err).Error())
				} else {
					return nil, statuserror.FromK8sError("get", "Secret", secretRefName, err)
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
		return nil, statuserror.FromK8sError("create", "InstalledPackageDetail", pkgInstall.Name, err)

	}

	response := &corev1.GetInstalledPackageDetailResponse{
		InstalledPackageDetail: installedPackageDetail,
	}
	return response, nil
}

// CreateInstalledPackage creates an installed package managed by the 'kapp_controller' plugin
func (s *Server) CreateInstalledPackage(ctx context.Context, request *corev1.CreateInstalledPackageRequest) (*corev1.CreateInstalledPackageResponse, error) {
	// Retrieve parameters from the request
	targetCluster := request.GetTargetContext().GetCluster()
	targetNamespace := request.GetTargetContext().GetNamespace()
	installedPackageName := request.GetName()
	contextMsg := fmt.Sprintf("(cluster=%q, namespace=%q, id=%q)", targetCluster, targetNamespace, installedPackageName)
	log.Infof("+kapp-controller CreateInstalledPackage %s", contextMsg)

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

	// Retrieve additional parameters from the request
	packageRef := request.GetAvailablePackageRef()
	packageCluster := request.GetAvailablePackageRef().GetContext().GetCluster()
	packageNamespace := request.GetAvailablePackageRef().GetContext().GetNamespace()
	reconciliationOptions := request.GetReconciliationOptions()
	pkgVersion := request.GetPkgVersionReference().GetVersion()
	values := request.GetValues()

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
		return nil, statuserror.FromK8sError("get", "PackageMetadata", packageRef.Identifier, err)
	}

	// build a new secret object with the values
	secret, err := s.buildSecret(installedPackageName, values, targetNamespace)
	if err != nil {
		return nil, statuserror.FromK8sError("create", "Secret", installedPackageName, err)
	}

	// build a new pkgInstall object
	newPkgInstall, err := s.buildPkgInstall(installedPackageName, targetCluster, targetNamespace, pkgMetadata.Name, pkgVersion, reconciliationOptions, secret)
	if err != nil {
		return nil, status.Errorf(status.Code(err), "Unable to create the PackageInstall '%s' due to '%v'", installedPackageName, err)
	}

	// create the Secret in the cluster
	// TODO(agamez): check when is the best moment to create this object.
	// See if we can delay the creation until the PackageInstall is successfully created.
	createdSecret, err := typedClient.CoreV1().Secrets(targetNamespace).Create(ctx, secret, metav1.CreateOptions{})
	if createdSecret == nil || err != nil {
		return nil, statuserror.FromK8sError("create", "Secret", secret.Name, err)
	}

	// create the PackageInstall in the cluster
	createdPkgInstall, err := s.createPkgInstall(ctx, targetCluster, targetNamespace, newPkgInstall)
	if err != nil {
		// clean-up the secret if something fails
		err := typedClient.CoreV1().Secrets(targetNamespace).Delete(ctx, secret.Name, metav1.DeleteOptions{})
		if err != nil {
			return nil, statuserror.FromK8sError("delete", "Secret", secret.Name, err)
		}
		return nil, statuserror.FromK8sError("create", "PackageInstall", newPkgInstall.Name, err)
	}

	resource, err := s.getAppResource(ctx, targetCluster, targetNamespace)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "unable to get the App resource: '%v'", err)
	}
	// The InstalledPackage is considered as created once the associated kapp App gets created,
	// so we actively wait for the App CR to be present in the cluster before returning OK
	err = k8sutils.WaitForResource(ctx, resource, newPkgInstall.Name, time.Second*1, time.Second*time.Duration(s.pluginConfig.timeoutSeconds))
	if err != nil {
		// clean-up the secret if something fails
		err := typedClient.CoreV1().Secrets(targetNamespace).Delete(ctx, secret.Name, metav1.DeleteOptions{})
		if err != nil {
			return nil, statuserror.FromK8sError("delete", "Secret", secret.Name, err)
		}
		// clean-up the package install if something fails
		err = s.deletePkgInstall(ctx, targetCluster, targetNamespace, newPkgInstall.Name)
		if err != nil {
			return nil, statuserror.FromK8sError("delete", "PackageInstall", newPkgInstall.Name, err)
		}
		return nil, status.Errorf(codes.Internal, "timeout exceeded (%v s) waiting for resource to be installed: '%v'", s.pluginConfig.timeoutSeconds, err)
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
	// Retrieve parameters from the request
	packageCluster := request.GetInstalledPackageRef().GetContext().GetCluster()
	packageNamespace := request.GetInstalledPackageRef().GetContext().GetNamespace()
	installedPackageName := request.GetInstalledPackageRef().GetIdentifier()
	contextMsg := fmt.Sprintf("(cluster=%q, namespace=%q, id=%q)", packageCluster, packageNamespace, installedPackageName)
	log.Infof("+kapp-controller UpdateInstalledPackage %s", contextMsg)

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

	// Retrieve additional parameters from the request
	reconciliationOptions := request.GetReconciliationOptions()
	pkgVersion := request.GetPkgVersionReference().GetVersion()
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
		return nil, statuserror.FromK8sError("get", "PackageInstall", installedPackageName, err)
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
	elegibleVersion, err := versions.HighestConstrainedVersion([]string{pkgVersion}, vendirversions.VersionSelection{Semver: versionSelection})
	if elegibleVersion == "" || err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "The selected version %q is not elegible to be installed: %v", pkgVersion, err)
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
		return nil, statuserror.FromK8sError("update", "PackageInstall", installedPackageName, err)
	}

	// Update the values.yaml values file if any is passed, otherwise, delete the values
	if values != "" {
		secret, err := s.buildSecret(installedPackageName, values, packageNamespace)
		if err != nil {
			return nil, statuserror.FromK8sError("update", "Secret", secret.Name, err)
		}
		updatedSecret, err := typedClient.CoreV1().Secrets(packageNamespace).Update(ctx, secret, metav1.UpdateOptions{})
		if updatedSecret == nil || err != nil {
			return nil, statuserror.FromK8sError("update", "Secret", secret.Name, err)
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
		// See https://github.com/kubeapps/kubeapps/pull/3790#discussion_r754797195
		for _, packageInstallValue := range pkgInstall.Spec.Values {
			secretId := packageInstallValue.SecretRef.Name
			err := typedClient.CoreV1().Secrets(packageNamespace).Delete(ctx, secretId, metav1.DeleteOptions{})
			if errors.IsNotFound(err) {
				log.Warningf("The referenced secret does not exist: %s", statuserror.FromK8sError("get", "Secret", secretId, err).Error())
			} else {
				return nil, statuserror.FromK8sError("delete", "Secret", secretId, err)
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
	// Retrieve parameters from the request
	namespace := request.GetInstalledPackageRef().GetContext().GetNamespace()
	cluster := request.GetInstalledPackageRef().GetContext().GetCluster()
	identifier := request.GetInstalledPackageRef().GetIdentifier()
	contextMsg := fmt.Sprintf("(cluster=%q, namespace=%q, id=%q)", namespace, cluster, identifier)
	log.Infof("+kapp-controller DeleteInstalledPackage %s", contextMsg)

	// Validate the request
	if request == nil || request.GetInstalledPackageRef() == nil {
		return nil, status.Errorf(codes.InvalidArgument, "no request InstalledPackageRef provided")
	}
	if request.GetInstalledPackageRef().GetContext().GetNamespace() == "" || request.GetInstalledPackageRef().GetIdentifier() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "required context or identifier not provided")
	}

	if cluster == "" {
		cluster = s.globalPackagingCluster
	}

	typedClient, _, err := s.GetClients(ctx, cluster)
	if err != nil {
		return nil, err
	}

	pkgInstall, err := s.getPkgInstall(ctx, cluster, namespace, identifier)
	if err != nil {
		return nil, statuserror.FromK8sError("get", "PackageInstall", identifier, err)
	}

	// Delete the package install
	err = s.deletePkgInstall(ctx, cluster, namespace, identifier)
	if err != nil {
		return nil, statuserror.FromK8sError("delete", "PackageInstall", identifier, err)
	}

	// Delete all the associated secrets
	// TODO(agamez): maybe it's too aggressive and we should be deleting only those secrets created by this plugin
	// See https://github.com/kubeapps/kubeapps/pull/3790#discussion_r754797195
	for _, packageInstallValue := range pkgInstall.Spec.Values {
		secretId := packageInstallValue.SecretRef.Name
		err := typedClient.CoreV1().Secrets(namespace).Delete(ctx, secretId, metav1.DeleteOptions{})
		if err != nil {
			if errors.IsNotFound(err) {
				log.Warningf("The referenced secret does not exist: %s", statuserror.FromK8sError("get", "Secret", secretId, err).Error())
			} else {
				return nil, statuserror.FromK8sError("delete", "Secret", secretId, err)
			}
		}
	}
	return &corev1.DeleteInstalledPackageResponse{}, nil
}

// GetInstalledPackageResourceRefs returns the references for the k8s resources of an installed package managed by the 'kapp_controller' plugin
func (s *Server) GetInstalledPackageResourceRefs(ctx context.Context, request *corev1.GetInstalledPackageResourceRefsRequest) (*corev1.GetInstalledPackageResourceRefsResponse, error) {
	// Retrieve parameters from the request
	cluster := request.GetInstalledPackageRef().GetContext().GetCluster()
	namespace := request.GetInstalledPackageRef().GetContext().GetNamespace()
	installedPackageRefId := request.GetInstalledPackageRef().GetIdentifier()
	contextMsg := fmt.Sprintf("(cluster=%q, namespace=%q, id=%q)", namespace, cluster, installedPackageRefId)
	log.Infof("+kapp-controller GetInstalledPackageResourceRefs %s", contextMsg)

	if cluster == "" {
		cluster = s.globalPackagingCluster
	}

	// get the list of every k8s resource matching ResourceRef
	refs, err := s.inspectKappK8sResources(ctx, cluster, namespace, installedPackageRefId)
	if err != nil {
		return nil, err
	}

	return &corev1.GetInstalledPackageResourceRefsResponse{
		Context:      request.GetInstalledPackageRef().GetContext(),
		ResourceRefs: refs,
	}, nil
}
