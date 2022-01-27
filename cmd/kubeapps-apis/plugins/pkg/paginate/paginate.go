// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package paginate

import (
	"strconv"

	pkgsGRPCv1alpha1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	grpccodes "google.golang.org/grpc/codes"
	grpcstatus "google.golang.org/grpc/status"
)

// PageOffsetFromPageToken converts a page token to an integer offset
// representing the page of results.
// TODO(gfichtenholt): it'd be better if we ensure that the page_token
// contains an offset to the item, not the page so we can
// aggregate paginated results. Same as helm plug-in.
// Update this when helm plug-in does so
func PageOffsetFromPageToken(pageToken string) (int, error) {
	if pageToken == "" {
		return 0, nil
	}
	offset, err := strconv.ParseUint(pageToken, 10, 0)
	if err != nil {
		return 0, err
	}
	return int(offset), nil
}

func PageOffsetFromInstalledRequest(request *pkgsGRPCv1alpha1.GetInstalledPackageSummariesRequest) (int, error) {
	offset, err := PageOffsetFromPageToken(request.GetPaginationOptions().GetPageToken())
	if err != nil {
		return 0, grpcstatus.Errorf(grpccodes.InvalidArgument, "unable to intepret page token %q: %v",
			request.GetPaginationOptions().GetPageToken(), err)
	} else {
		return offset, nil
	}
}

func PageOffsetFromAvailableRequest(request *pkgsGRPCv1alpha1.GetAvailablePackageSummariesRequest) (int, error) {
	offset, err := PageOffsetFromPageToken(request.GetPaginationOptions().GetPageToken())
	if err != nil {
		return 0, grpcstatus.Errorf(grpccodes.InvalidArgument, "unable to intepret page token %q: %v",
			request.GetPaginationOptions().GetPageToken(), err)
	} else {
		return offset, nil
	}
}
