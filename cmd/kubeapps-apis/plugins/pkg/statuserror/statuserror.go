// Copyright 2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package statuserror

import (
	grpccodes "google.golang.org/grpc/codes"
	grpcstatus "google.golang.org/grpc/status"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
)

// FromK8sResourceError generates a grpc status error from a Kubernetes error
// when querying a resource.
func FromK8sError(verb, resource, identifier string, err error) error {
	if identifier == "" {
		identifier = "all"
	}
	if k8serrors.IsNotFound(err) {
		return grpcstatus.Errorf(grpccodes.NotFound, "unable to %s the %s '%s' due to '%v'", verb, resource, identifier, err)
	} else if k8serrors.IsForbidden(err) {
		return grpcstatus.Errorf(grpccodes.PermissionDenied, "Forbidden to %s the %s '%s' due to '%v'", verb, resource, identifier, err)
	} else if k8serrors.IsUnauthorized(err) {
		return grpcstatus.Errorf(grpccodes.Unauthenticated, "Authorization required to %s the %s '%s' due to '%v'", verb, resource, identifier, err)
	} else if k8serrors.IsAlreadyExists(err) {
		return grpcstatus.Errorf(grpccodes.AlreadyExists, "Cannot %s the %s '%s' due to '%v' as it already exists", verb, resource, identifier, err)
	}
	return grpcstatus.Errorf(grpccodes.Internal, "unable to %s the %s '%s' due to '%v'", verb, resource, identifier, err)
}
