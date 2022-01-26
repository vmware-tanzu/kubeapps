// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package paginate

import (
	"testing"

	pkgsGRPCv1alpha1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	grpccodes "google.golang.org/grpc/codes"
	grpcstatus "google.golang.org/grpc/status"
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
	if got, want := grpcstatus.Code(err), grpccodes.Unknown; got != want {
		t.Fatalf("got: %+v, want: %+v, err: %+v", got, want, err)
	}

	req1 := &pkgsGRPCv1alpha1.GetInstalledPackageSummariesRequest{
		Context: &pkgsGRPCv1alpha1.Context{Namespace: "namespace-1"},
	}
	offset, err = PageOffsetFromInstalledRequest(req1)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if offset != 0 {
		t.Fatalf("expected 1, got: %d", offset)
	}

	req2 := &pkgsGRPCv1alpha1.GetInstalledPackageSummariesRequest{
		Context: &pkgsGRPCv1alpha1.Context{Namespace: "namespace-1"},
		PaginationOptions: &pkgsGRPCv1alpha1.PaginationOptions{
			PageToken: "1",
		},
	}
	offset, err = PageOffsetFromInstalledRequest(req2)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if offset != 1 {
		t.Fatalf("expected 1, got: %d", offset)
	}

	req3 := &pkgsGRPCv1alpha1.GetAvailablePackageSummariesRequest{
		Context: &pkgsGRPCv1alpha1.Context{Namespace: "namespace-1"},
	}
	offset, err = PageOffsetFromAvailableRequest(req3)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if offset != 0 {
		t.Fatalf("expected 1, got: %d", offset)
	}

	req4 := &pkgsGRPCv1alpha1.GetAvailablePackageSummariesRequest{
		Context: &pkgsGRPCv1alpha1.Context{Namespace: "namespace-1"},
		PaginationOptions: &pkgsGRPCv1alpha1.PaginationOptions{
			PageToken: "1",
		},
	}
	offset, err = PageOffsetFromAvailableRequest(req4)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if offset != 1 {
		t.Fatalf("expected 1, got: %d", offset)
	}
}
