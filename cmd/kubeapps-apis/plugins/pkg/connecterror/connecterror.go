// Copyright 2023 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package connecterror

import (
	"fmt"

	"github.com/bufbuild/connect-go"
	"k8s.io/apimachinery/pkg/api/errors"
)

// FromK8sResourceError generates a grpc connect error from a Kubernetes error
// when querying a resource.
func FromK8sError(verb, resource, identifier string, err error) error {
	if identifier == "" {
		identifier = "all"
	}
	if errors.IsNotFound(err) {
		return connect.NewError(connect.CodeNotFound, fmt.Errorf("Unable to %s the %s '%s' due to '%w'", verb, resource, identifier, err))
	} else if errors.IsForbidden(err) {
		return connect.NewError(connect.CodePermissionDenied, fmt.Errorf("Forbidden to %s the %s '%s' due to '%w'", verb, resource, identifier, err))
	} else if errors.IsUnauthorized(err) {
		return connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("Authorization required to %s the %s '%s' due to '%w'", verb, resource, identifier, err))
	} else if errors.IsAlreadyExists(err) {
		return connect.NewError(connect.CodeAlreadyExists, fmt.Errorf("Cannot %s the %s '%s' due to '%w' as it already exists", verb, resource, identifier, err))
	}
	return connect.NewError(connect.CodeInternal, fmt.Errorf("Unable to %s the %s '%s' due to '%w'", verb, resource, identifier, err))
}
