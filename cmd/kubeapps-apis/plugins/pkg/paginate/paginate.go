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
	"strconv"

	corev1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

func PageOffsetFromInstalledRequest(request *corev1.GetInstalledPackageSummariesRequest) (int, error) {
	offset, err := PageOffsetFromPageToken(request.GetPaginationOptions().GetPageToken())
	if err != nil {
		return 0, status.Errorf(codes.InvalidArgument, "unable to intepret page token %q: %v",
			request.GetPaginationOptions().GetPageToken(), err)
	} else {
		return offset, nil
	}
}

func PageOffsetFromAvailableRequest(request *corev1.GetAvailablePackageSummariesRequest) (int, error) {
	offset, err := PageOffsetFromPageToken(request.GetPaginationOptions().GetPageToken())
	if err != nil {
		return 0, status.Errorf(codes.InvalidArgument, "unable to intepret page token %q: %v",
			request.GetPaginationOptions().GetPageToken(), err)
	} else {
		return offset, nil
	}
}
