// Copyright 2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package statuserror

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/apimachinery/pkg/api/errors"
)

// FromK8sResourceError generates a grpc status error from a Kubernetes error
// when querying a resource.
func FromK8sError(verb, resource, identifier string, err error) error {
	if identifier == "" {
		identifier = "all"
	}
	if errors.IsNotFound(err) {
		return status.Errorf(codes.NotFound, "unable to %s the %s '%s' due to '%v'", verb, resource, identifier, err)
	} else if errors.IsForbidden(err) {
		return status.Errorf(codes.PermissionDenied, "Forbidden to %s the %s '%s' due to '%v'", verb, resource, identifier, err)
	} else if errors.IsUnauthorized(err) {
		return status.Errorf(codes.Unauthenticated, "Authorization required to %s the %s '%s' due to '%v'", verb, resource, identifier, err)
	} else if errors.IsAlreadyExists(err) {
		return status.Errorf(codes.AlreadyExists, "Cannot %s the %s '%s' due to '%v' as it already exists", verb, resource, identifier, err)
	}
	return status.Errorf(codes.Internal, "unable to %s the %s '%s' due to '%v'", verb, resource, identifier, err)
}
