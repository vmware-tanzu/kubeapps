// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package plugin_test

import (
	"context"

	corev1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	plugins "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/plugins/v1alpha1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type TestPackagingPluginServer struct {
	corev1.UnimplementedPackagesServiceServer
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
func (s TestPackagingPluginServer) GetAvailablePackageSummaries(ctx context.Context, request *corev1.GetAvailablePackageSummariesRequest) (*corev1.GetAvailablePackageSummariesResponse, error) {
	if s.Status != codes.OK {
		return nil, status.Errorf(s.Status, "Non-OK response")
	}
	return &corev1.GetAvailablePackageSummariesResponse{
		AvailablePackageSummaries: s.AvailablePackageSummaries,
		Categories:                s.Categories,
		NextPageToken:             s.NextPageToken,
	}, nil
}

// GetAvailablePackageDetail returns the package details based on the request.
func (s TestPackagingPluginServer) GetAvailablePackageDetail(ctx context.Context, request *corev1.GetAvailablePackageDetailRequest) (*corev1.GetAvailablePackageDetailResponse, error) {
	if s.Status != codes.OK {
		return nil, status.Errorf(s.Status, "Non-OK response")
	}
	return &corev1.GetAvailablePackageDetailResponse{
		AvailablePackageDetail: s.AvailablePackageDetail,
	}, nil
}

// GetInstalledPackageSummaries returns the installed package summaries based on the request.
func (s TestPackagingPluginServer) GetInstalledPackageSummaries(ctx context.Context, request *corev1.GetInstalledPackageSummariesRequest) (*corev1.GetInstalledPackageSummariesResponse, error) {
	if s.Status != codes.OK {
		return nil, status.Errorf(s.Status, "Non-OK response")
	}
	return &corev1.GetInstalledPackageSummariesResponse{
		InstalledPackageSummaries: s.InstalledPackageSummaries,
		NextPageToken:             s.NextPageToken,
	}, nil
}

// GetInstalledPackageDetail returns the package versions based on the request.
func (s TestPackagingPluginServer) GetInstalledPackageDetail(ctx context.Context, request *corev1.GetInstalledPackageDetailRequest) (*corev1.GetInstalledPackageDetailResponse, error) {
	if s.Status != codes.OK {
		return nil, status.Errorf(s.Status, "Non-OK response")
	}
	return &corev1.GetInstalledPackageDetailResponse{
		InstalledPackageDetail: s.InstalledPackageDetail,
	}, nil
}

// GetAvailablePackageVersions returns the package versions based on the request.
func (s TestPackagingPluginServer) GetAvailablePackageVersions(ctx context.Context, request *corev1.GetAvailablePackageVersionsRequest) (*corev1.GetAvailablePackageVersionsResponse, error) {
	if s.Status != codes.OK {
		return nil, status.Errorf(s.Status, "Non-OK response")
	}
	return &corev1.GetAvailablePackageVersionsResponse{
		PackageAppVersions: s.PackageAppVersions,
	}, nil
}

// GetInstalledPackageResourceRefs returns the resource references based on the request.
func (s TestPackagingPluginServer) GetInstalledPackageResourceRefs(ctx context.Context, request *corev1.GetInstalledPackageResourceRefsRequest) (*corev1.GetInstalledPackageResourceRefsResponse, error) {
	if s.Status != codes.OK {
		return nil, status.Errorf(s.Status, "Non-OK response")
	}
	return &corev1.GetInstalledPackageResourceRefsResponse{
		Context:      request.GetInstalledPackageRef().GetContext(),
		ResourceRefs: s.ResourceRefs,
	}, nil
}

func (s TestPackagingPluginServer) CreateInstalledPackage(ctx context.Context, request *corev1.CreateInstalledPackageRequest) (*corev1.CreateInstalledPackageResponse, error) {
	if s.Status != codes.OK {
		return nil, status.Errorf(s.Status, "Non-OK response")
	}
	return &corev1.CreateInstalledPackageResponse{
		InstalledPackageRef: &corev1.InstalledPackageReference{
			Context:    request.GetTargetContext(),
			Identifier: request.GetName(),
			Plugin:     s.Plugin,
		},
	}, nil
}

func (s TestPackagingPluginServer) UpdateInstalledPackage(ctx context.Context, request *corev1.UpdateInstalledPackageRequest) (*corev1.UpdateInstalledPackageResponse, error) {
	if s.Status != codes.OK {
		return nil, status.Errorf(s.Status, "Non-OK response")
	}
	return &corev1.UpdateInstalledPackageResponse{
		InstalledPackageRef: &corev1.InstalledPackageReference{
			Context:    request.GetInstalledPackageRef().GetContext(),
			Identifier: request.GetInstalledPackageRef().GetIdentifier(),
			Plugin:     s.Plugin,
		},
	}, nil
}

func (s TestPackagingPluginServer) DeleteInstalledPackage(ctx context.Context, request *corev1.DeleteInstalledPackageRequest) (*corev1.DeleteInstalledPackageResponse, error) {
	if s.Status != codes.OK {
		return nil, status.Errorf(s.Status, "Non-OK response")
	}
	return &corev1.DeleteInstalledPackageResponse{}, nil
}

type TestRepositoriesPluginServer struct {
	corev1.UnimplementedRepositoriesServiceServer
	Plugin *plugins.Plugin
	Status codes.Code
}

func NewTestRepositoriesPlugin(plugin *plugins.Plugin) *TestRepositoriesPluginServer {
	return &TestRepositoriesPluginServer{
		Plugin: plugin,
	}
}

func (s TestRepositoriesPluginServer) AddPackageRepository(ctx context.Context, request *corev1.AddPackageRepositoryRequest) (*corev1.AddPackageRepositoryResponse, error) {
	if s.Status != codes.OK {
		return nil, status.Errorf(s.Status, "Non-OK response")
	}
	return &corev1.AddPackageRepositoryResponse{}, nil
}
