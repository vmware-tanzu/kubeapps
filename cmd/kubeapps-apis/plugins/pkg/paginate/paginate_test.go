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
package paginate

import (
	"testing"

	corev1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
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
	if got, want := status.Code(err), codes.Unknown; got != want {
		t.Fatalf("got: %+v, want: %+v, err: %+v", got, want, err)
	}

	req1 := &corev1.GetInstalledPackageSummariesRequest{
		Context: &corev1.Context{Namespace: "namespace-1"},
	}
	offset, err = PageOffsetFromInstalledRequest(req1)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if offset != 0 {
		t.Fatalf("expected 1, got: %d", offset)
	}

	req2 := &corev1.GetInstalledPackageSummariesRequest{
		Context: &corev1.Context{Namespace: "namespace-1"},
		PaginationOptions: &corev1.PaginationOptions{
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

	req3 := &corev1.GetAvailablePackageSummariesRequest{
		Context: &corev1.Context{Namespace: "namespace-1"},
	}
	offset, err = PageOffsetFromAvailableRequest(req3)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if offset != 0 {
		t.Fatalf("expected 1, got: %d", offset)
	}

	req4 := &corev1.GetAvailablePackageSummariesRequest{
		Context: &corev1.Context{Namespace: "namespace-1"},
		PaginationOptions: &corev1.PaginationOptions{
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
