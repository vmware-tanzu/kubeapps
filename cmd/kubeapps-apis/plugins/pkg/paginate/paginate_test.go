// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package paginate

import (
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestPageOffsetFromPageToken(t *testing.T) {
	offset, err := PageOffsetFromPageToken("1021")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if offset != 1021 {
		t.Fatalf("expected 1021, got: %d", offset)
	}

	_, err = PageOffsetFromPageToken("not a number")
	if got, want := status.Code(err), codes.InvalidArgument; got != want {
		t.Fatalf("got: %+v, want: %+v, err: %+v", got, want, err)
	}
}
