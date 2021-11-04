/*
Copyright © 2021 VMware
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
package plugin_test

import (
	"context"

	corev1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	packages "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	plugins "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/plugins/v1alpha1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type TestPackagingPluginServer struct {
	packages.UnimplementedPackagesServiceServer
	Plugin                    *plugins.Plugin
	AvailablePackageSummaries []*corev1.AvailablePackageSummary
	AvailablePackageDetail    *corev1.AvailablePackageDetail
	InstalledPackageSummaries []*corev1.InstalledPackageSummary
	InstalledPackageDetail    *corev1.InstalledPackageDetail
	PackageAppVersions        []*corev1.PackageAppVersion
	ResourceRefs              []*corev1.ResourceRef
	Categories                []string
	NextPageToken             string
	Status                    codes.Code
}

func NewTestPackagingPlugin(plugin *plugins.Plugin) *TestPackagingPluginServer {
	return &TestPackagingPluginServer{
		Plugin: plugin,
	}
}

// GetAvailablePackages returns the packages based on the request.
func (s TestPackagingPluginServer) GetAvailablePackageSummaries(ctx context.Context, request *packages.GetAvailablePackageSummariesRequest) (*packages.GetAvailablePackageSummariesResponse, error) {
	if s.Status != codes.OK {
		return nil, status.Errorf(s.Status, "Non-OK response")
	}
	return &packages.GetAvailablePackageSummariesResponse{
		AvailablePackageSummaries: s.AvailablePackageSummaries,
		Categories:                s.Categories,
		NextPageToken:             s.NextPageToken,
	}, nil
}

// GetAvailablePackageDetail returns the package details based on the request.
func (s TestPackagingPluginServer) GetAvailablePackageDetail(ctx context.Context, request *packages.GetAvailablePackageDetailRequest) (*packages.GetAvailablePackageDetailResponse, error) {
	if s.Status != codes.OK {
		return nil, status.Errorf(s.Status, "Non-OK response")
	}
	return &packages.GetAvailablePackageDetailResponse{
		AvailablePackageDetail: s.AvailablePackageDetail,
	}, nil
}

// GetInstalledPackageSummaries returns the installed package summaries based on the request.
func (s TestPackagingPluginServer) GetInstalledPackageSummaries(ctx context.Context, request *packages.GetInstalledPackageSummariesRequest) (*packages.GetInstalledPackageSummariesResponse, error) {
	if s.Status != codes.OK {
		return nil, status.Errorf(s.Status, "Non-OK response")
	}
	return &packages.GetInstalledPackageSummariesResponse{
		InstalledPackageSummaries: s.InstalledPackageSummaries,
		NextPageToken:             s.NextPageToken,
	}, nil
}

// GetInstalledPackageDetail returns the package versions based on the request.
func (s TestPackagingPluginServer) GetInstalledPackageDetail(ctx context.Context, request *packages.GetInstalledPackageDetailRequest) (*packages.GetInstalledPackageDetailResponse, error) {
	if s.Status != codes.OK {
		return nil, status.Errorf(s.Status, "Non-OK response")
	}
	return &packages.GetInstalledPackageDetailResponse{
		InstalledPackageDetail: s.InstalledPackageDetail,
	}, nil
}

// GetAvailablePackageVersions returns the package versions based on the request.
func (s TestPackagingPluginServer) GetAvailablePackageVersions(ctx context.Context, request *packages.GetAvailablePackageVersionsRequest) (*packages.GetAvailablePackageVersionsResponse, error) {
	if s.Status != codes.OK {
		return nil, status.Errorf(s.Status, "Non-OK response")
	}
	return &packages.GetAvailablePackageVersionsResponse{
		PackageAppVersions: s.PackageAppVersions,
	}, nil
}

// GetInstalledPackageResourceRefs returns the resource references based on the request.
func (s TestPackagingPluginServer) GetInstalledPackageResourceRefs(ctx context.Context, request *packages.GetInstalledPackageResourceRefsRequest) (*packages.GetInstalledPackageResourceRefsResponse, error) {
	if s.Status != codes.OK {
		return nil, status.Errorf(s.Status, "Non-OK response")
	}
	return &packages.GetInstalledPackageResourceRefsResponse{
		Context:      request.GetInstalledPackageRef().GetContext(),
		ResourceRefs: s.ResourceRefs,
	}, nil
}

func (s TestPackagingPluginServer) CreateInstalledPackage(ctx context.Context, request *packages.CreateInstalledPackageRequest) (*packages.CreateInstalledPackageResponse, error) {
	if s.Status != codes.OK {
		return nil, status.Errorf(s.Status, "Non-OK response")
	}
	return &packages.CreateInstalledPackageResponse{
		InstalledPackageRef: &packages.InstalledPackageReference{
			Context:    request.GetTargetContext(),
			Identifier: request.GetName(),
			Plugin:     s.Plugin,
		},
	}, nil
}

func (s TestPackagingPluginServer) UpdateInstalledPackage(ctx context.Context, request *packages.UpdateInstalledPackageRequest) (*packages.UpdateInstalledPackageResponse, error) {
	if s.Status != codes.OK {
		return nil, status.Errorf(s.Status, "Non-OK response")
	}
	return &packages.UpdateInstalledPackageResponse{
		InstalledPackageRef: &packages.InstalledPackageReference{
			Context:    request.GetInstalledPackageRef().GetContext(),
			Identifier: request.GetInstalledPackageRef().GetIdentifier(),
			Plugin:     s.Plugin,
		},
	}, nil
}

func (s TestPackagingPluginServer) DeleteInstalledPackage(ctx context.Context, request *packages.DeleteInstalledPackageRequest) (*packages.DeleteInstalledPackageResponse, error) {
	if s.Status != codes.OK {
		return nil, status.Errorf(s.Status, "Non-OK response")
	}
	return &packages.DeleteInstalledPackageResponse{}, nil
}
