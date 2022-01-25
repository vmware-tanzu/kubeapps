// Copyright 2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package statuserror

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestErrorByStatus(t *testing.T) {
	tests := []struct {
		name        string
		verb        string
		resource    string
		identifier  string
		err         error
		expectedErr error
	}{
		{
			"error msg for all resources ",
			"get",
			"my-resource",
			"",
			status.Errorf(codes.InvalidArgument, "boom!"),
			status.Errorf(codes.Internal, "unable to get the my-resource 'all' due to 'rpc error: code = InvalidArgument desc = boom!'"),
		},
		{
			"error msg for a single resources ",
			"get",
			"my-resource",
			"my-id",
			status.Errorf(codes.InvalidArgument, "boom!"),
			status.Errorf(codes.Internal, "unable to get the my-resource 'my-id' due to 'rpc error: code = InvalidArgument desc = boom!'"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := FromK8sError(tt.verb, tt.resource, tt.identifier, tt.err)
			if got, want := err.Error(), tt.expectedErr.Error(); !cmp.Equal(want, got) {
				t.Errorf("in %s: mismatch (-want +got):\n%s", tt.name, cmp.Diff(want, got))
			}
		})
	}
}
