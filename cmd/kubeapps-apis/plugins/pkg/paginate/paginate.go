// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package paginate

import (
	"strconv"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// PageOffsetFromPageToken converts a page token to an integer offset
// representing the page of results.
// TODO(gfichtenholt): it'd be better if we ensure that the page_token
// contains an offset to the item, not the page so we can
// aggregate paginated results. Same as helm plug-in.
// Update this when helm plug-in does so
// TODO(minelson) This will be removed once the core pagination is updated.
func PageOffsetFromPageToken(pageToken string) (int, error) {
	if pageToken == "" {
		return 0, nil
	}
	offset, err := strconv.ParseInt(pageToken, 10, 0)
	if err != nil {
		return 0, status.Errorf(codes.InvalidArgument, "unable to interpret page token %q: %v",
			pageToken, err)
	}
	return int(offset), nil
}

// Plugins should be designed to use an offset to the next item, rather than the
// next page of items.
// Until we have a need for more structure, this can be an integer number and so
// is parsed in exactly the same way as a page offset.
func ItemOffsetFromPageToken(pageToken string) (int, error) {
	return PageOffsetFromPageToken(pageToken)
}
