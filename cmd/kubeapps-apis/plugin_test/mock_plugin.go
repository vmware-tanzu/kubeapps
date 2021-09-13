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
	availablePackageSummaries []*corev1.AvailablePackageSummary
	availablePackageDetail    *corev1.AvailablePackageDetail
	installedPackageSummaries []*corev1.InstalledPackageSummary
	installedPackageDetail    *corev1.InstalledPackageDetail
	packageAppVersions        []*corev1.PackageAppVersion
	categories                []string
	nextPageToken             string
	status                    codes.Code
}

func (s *TestPackagingPluginServer) SetAvailablePackageSummaries(availablePackageSummaries []*corev1.AvailablePackageSummary) {
	s.availablePackageSummaries = availablePackageSummaries
}
func (s *TestPackagingPluginServer) SetAvailablePackageDetail(availablePackageDetail *corev1.AvailablePackageDetail) {
	s.availablePackageDetail = availablePackageDetail
}
func (s *TestPackagingPluginServer) SetInstalledPackageSummary(installedPackageSummaries []*corev1.InstalledPackageSummary) {
	s.installedPackageSummaries = installedPackageSummaries
}
func (s *TestPackagingPluginServer) SetInstalledPackageDetail(installedPackageDetail *corev1.InstalledPackageDetail) {
	s.installedPackageDetail = installedPackageDetail
}
func (s *TestPackagingPluginServer) SetPackageAppVersion(packageAppVersions []*corev1.PackageAppVersion) {
	s.packageAppVersions = packageAppVersions
}
func (s *TestPackagingPluginServer) SetCategories(categories []string) {
	s.categories = categories
}
func (s *TestPackagingPluginServer) SetNextPageToken(nextPageToken string) {
	s.nextPageToken = nextPageToken
}
func (s *TestPackagingPluginServer) SetStatus(status codes.Code) {
	s.status = status
}

func NewTestPackagingPlugin(plugin *plugins.Plugin) *TestPackagingPluginServer {
	return &TestPackagingPluginServer{
		Plugin: plugin,
	}
}

// GetAvailablePackages returns the packages based on the request.
func (s TestPackagingPluginServer) GetAvailablePackageSummaries(ctx context.Context, request *packages.GetAvailablePackageSummariesRequest) (*packages.GetAvailablePackageSummariesResponse, error) {
	if s.status != codes.OK {
		return nil, status.Errorf(s.status, "Non-OK response")
	}
	return &packages.GetAvailablePackageSummariesResponse{
		AvailablePackageSummaries: s.availablePackageSummaries,
		Categories:                s.categories,
		NextPageToken:             s.nextPageToken,
	}, nil
}

// GetAvailablePackageDetail returns the package details based on the request.
func (s TestPackagingPluginServer) GetAvailablePackageDetail(ctx context.Context, request *packages.GetAvailablePackageDetailRequest) (*packages.GetAvailablePackageDetailResponse, error) {
	if s.status != codes.OK {
		return nil, status.Errorf(s.status, "Non-OK response")
	}
	return &packages.GetAvailablePackageDetailResponse{
		AvailablePackageDetail: s.availablePackageDetail,
	}, nil
}

// GetInstalledPackageSummaries returns the installed package summaries based on the request.
func (s TestPackagingPluginServer) GetInstalledPackageSummaries(ctx context.Context, request *packages.GetInstalledPackageSummariesRequest) (*packages.GetInstalledPackageSummariesResponse, error) {
	if s.status != codes.OK {
		return nil, status.Errorf(s.status, "Non-OK response")
	}
	return &packages.GetInstalledPackageSummariesResponse{
		InstalledPackageSummaries: s.installedPackageSummaries,
		NextPageToken:             s.nextPageToken,
	}, nil
}

// GetInstalledPackageDetail returns the package versions based on the request.
func (s TestPackagingPluginServer) GetInstalledPackageDetail(ctx context.Context, request *packages.GetInstalledPackageDetailRequest) (*packages.GetInstalledPackageDetailResponse, error) {
	if s.status != codes.OK {
		return nil, status.Errorf(s.status, "Non-OK response")
	}
	return &packages.GetInstalledPackageDetailResponse{
		InstalledPackageDetail: s.installedPackageDetail,
	}, nil
}

// GetAvailablePackageVersions returns the package versions based on the request.
func (s TestPackagingPluginServer) GetAvailablePackageVersions(ctx context.Context, request *packages.GetAvailablePackageVersionsRequest) (*packages.GetAvailablePackageVersionsResponse, error) {
	if s.status != codes.OK {
		return nil, status.Errorf(s.status, "Non-OK response")
	}
	return &packages.GetAvailablePackageVersionsResponse{
		PackageAppVersions: s.packageAppVersions,
	}, nil
}
