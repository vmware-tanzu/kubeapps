/*
Copyright © 2022 VMware
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
