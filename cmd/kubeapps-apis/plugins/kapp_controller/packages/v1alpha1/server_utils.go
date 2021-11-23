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
	"fmt"
	"sort"
	"strconv"

	"github.com/Masterminds/semver/v3"
	corev1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	kappctrlv1alpha1 "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/kappctrl/v1alpha1"
	datapackagingv1alpha1 "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apiserver/apis/datapackaging/v1alpha1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/apimachinery/pkg/api/errors"
)

type pkgSemver struct {
	pkg     *datapackagingv1alpha1.Package
	version *semver.Version
}

// pkgVersionsMap recturns a map of packages keyed by the packagemetadataName.
//
// A Package CR in carvel is really a particular version of a package, so we need
// to sort them by the package metadata name, since this is what they share in common.
// The packages are then sorted by version.
func getPkgVersionsMap(packages []*datapackagingv1alpha1.Package) (map[string][]pkgSemver, error) {
	pkgVersionsMap := map[string][]pkgSemver{}
	for _, pkg := range packages {
		semverVersion, err := semver.NewVersion(pkg.Spec.Version)
		if err != nil {
			return nil, fmt.Errorf("required field spec.version was not semver compatible on kapp-controller Package: %v\n%v", err, pkg)
		}
		pkgVersionsMap[pkg.Spec.RefName] = append(pkgVersionsMap[pkg.Spec.RefName], pkgSemver{pkg, semverVersion})
	}

	for _, pkgVersions := range pkgVersionsMap {
		sort.Slice(pkgVersions, func(i, j int) bool {
			return pkgVersions[i].version.GreaterThan(pkgVersions[j].version)
		})
	}

	return pkgVersionsMap, nil
}

// statusReasonForKappStatus returns the reason for a given status
func statusReasonForKappStatus(status kappctrlv1alpha1.AppConditionType) corev1.InstalledPackageStatus_StatusReason {
	switch status {
	case kappctrlv1alpha1.ReconcileSucceeded:
		return corev1.InstalledPackageStatus_STATUS_REASON_INSTALLED
	case "ValuesSchemaCheckFailed", kappctrlv1alpha1.ReconcileFailed:
		return corev1.InstalledPackageStatus_STATUS_REASON_FAILED
	case kappctrlv1alpha1.Reconciling:
		return corev1.InstalledPackageStatus_STATUS_REASON_PENDING
	}
	// Fall back to unknown/unspecified.
	return corev1.InstalledPackageStatus_STATUS_REASON_UNSPECIFIED
}

// userReasonForKappStatus returns the reason for a given status
func userReasonForKappStatus(status kappctrlv1alpha1.AppConditionType) string {
	switch status {
	case kappctrlv1alpha1.ReconcileSucceeded:
		return "Reconcile succeeded"
	case "ValuesSchemaCheckFailed", kappctrlv1alpha1.ReconcileFailed:
		return "Reconcile failed"
	case kappctrlv1alpha1.Reconciling:
		return "Reconciling"
	}
	// Fall back to unknown/unspecified.
	return "Unknown"
}

// errorByStatus generates a meaningful error message
func errorByStatus(verb, resource, identifier string, err error) error {
	if identifier == "" {
		identifier = "all"
	}
	if errors.IsNotFound(err) {
		return status.Errorf(codes.NotFound, "unable to %s the %s '%s' due to '%v'", verb, resource, identifier, err)
	} else if errors.IsForbidden(err) || errors.IsUnauthorized(err) {
		return status.Errorf(codes.Unauthenticated, "Unauthorized to %s the %s '%s' due to '%v'", verb, resource, identifier, err)
	}
	return status.Errorf(codes.Internal, "unable to %s the %s '%s' due to '%v'", verb, resource, identifier, err)
}

// pageOffsetFromPageToken converts a page token to an integer offset representing the page of results.
//
// TODO(mnelson): When aggregating results from different plugins, we'll
// need to update the actual query in GetPaginatedChartListWithFilters to
// use a row offset rather than a page offset (as not all rows may be consumed
// for a specific plugin when combining).
func pageOffsetFromPageToken(pageToken string) (int, error) {
	if pageToken == "" {
		return 0, nil
	}
	offset, err := strconv.ParseUint(pageToken, 10, 0)
	if err != nil {
		return 0, err
	}

	return int(offset), nil
}
