// Copyright 2021-2023 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package plugin_test

import (
	"context"
	"fmt"

	"github.com/bufbuild/connect-go"
	corev1 "github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	plugins "github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/gen/core/plugins/v1alpha1"
	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/plugins/pkg/paginate"
)

type TestPackagingPluginServer struct {
	corev1.UnimplementedPackagesServiceServer
	Plugin                    *plugins.Plugin
	AvailablePackageSummaries []*corev1.AvailablePackageSummary
	AvailablePackageDetail    *corev1.AvailablePackageDetail
	AvailablePackageMetadatas []*corev1.PackageMetadata
	InstalledPackageSummaries []*corev1.InstalledPackageSummary
	InstalledPackageDetail    *corev1.InstalledPackageDetail
	PackageAppVersions        []*corev1.PackageAppVersion
	ResourceRefs              []*corev1.ResourceRef
	Categories                []string
	NextPageToken             string
	ErrorCode                 connect.Code
}

func NewTestPackagingPlugin(plugin *plugins.Plugin) *TestPackagingPluginServer {
	return &TestPackagingPluginServer{
		Plugin: plugin,
	}
}

// GetAvailablePackages returns the packages based on the request.
func (s TestPackagingPluginServer) GetAvailablePackageSummaries(ctx context.Context, request *connect.Request[corev1.GetAvailablePackageSummariesRequest]) (*connect.Response[corev1.GetAvailablePackageSummariesResponse], error) {
	if s.ErrorCode != 0 {
		return nil, connect.NewError(s.ErrorCode, fmt.Errorf("Non-OK response"))
	}
	itemOffset, err := paginate.ItemOffsetFromPageToken(request.Msg.PaginationOptions.GetPageToken())
	if err != nil {
		return nil, err
	}
	summaries := s.AvailablePackageSummaries[itemOffset:]
	pageSize := int(request.Msg.PaginationOptions.GetPageSize())
	nextPageToken := ""
	if pageSize > 0 && pageSize < len(summaries) {
		summaries = summaries[:pageSize]
		nextPageToken = fmt.Sprintf("%d", itemOffset+pageSize)
	}
	return connect.NewResponse(&corev1.GetAvailablePackageSummariesResponse{
		AvailablePackageSummaries: summaries,
		Categories:                s.Categories,
		NextPageToken:             nextPageToken,
	}), nil
}

// GetAvailablePackageDetail returns the package details based on the request.
func (s TestPackagingPluginServer) GetAvailablePackageDetail(ctx context.Context, request *connect.Request[corev1.GetAvailablePackageDetailRequest]) (*connect.Response[corev1.GetAvailablePackageDetailResponse], error) {
	if s.ErrorCode != 0 {
		return nil, connect.NewError(s.ErrorCode, fmt.Errorf("Non-OK response"))
	}
	return connect.NewResponse(&corev1.GetAvailablePackageDetailResponse{
		AvailablePackageDetail: s.AvailablePackageDetail,
	}), nil
}

// GetAvailablePackageMetadatas returns the package metadatas based on the request.
func (s TestPackagingPluginServer) GetAvailablePackageMetadatas(ctx context.Context, request *connect.Request[corev1.GetAvailablePackageMetadatasRequest]) (*connect.Response[corev1.GetAvailablePackageMetadatasResponse], error) {
	if s.ErrorCode != 0 {
		return nil, connect.NewError(s.ErrorCode, fmt.Errorf("Non-OK response"))
	}
	return connect.NewResponse(&corev1.GetAvailablePackageMetadatasResponse{
		PackageMetadata: s.AvailablePackageMetadatas,
	}), nil
}

// GetInstalledPackageSummaries returns the installed package summaries based on the request.
func (s TestPackagingPluginServer) GetInstalledPackageSummaries(ctx context.Context, request *connect.Request[corev1.GetInstalledPackageSummariesRequest]) (*connect.Response[corev1.GetInstalledPackageSummariesResponse], error) {
	if s.ErrorCode != 0 {
		return nil, connect.NewError(s.ErrorCode, fmt.Errorf("Non-OK response"))
	}
	return connect.NewResponse(&corev1.GetInstalledPackageSummariesResponse{
		InstalledPackageSummaries: s.InstalledPackageSummaries,
		NextPageToken:             s.NextPageToken,
	}), nil
}

// GetInstalledPackageDetail returns the package versions based on the request.
func (s TestPackagingPluginServer) GetInstalledPackageDetail(ctx context.Context, request *connect.Request[corev1.GetInstalledPackageDetailRequest]) (*connect.Response[corev1.GetInstalledPackageDetailResponse], error) {
	if s.ErrorCode != 0 {
		return nil, connect.NewError(s.ErrorCode, fmt.Errorf("Non-OK response"))
	}
	return connect.NewResponse(&corev1.GetInstalledPackageDetailResponse{
		InstalledPackageDetail: s.InstalledPackageDetail,
	}), nil
}

// GetAvailablePackageVersions returns the package versions based on the request.
func (s TestPackagingPluginServer) GetAvailablePackageVersions(ctx context.Context, request *connect.Request[corev1.GetAvailablePackageVersionsRequest]) (*connect.Response[corev1.GetAvailablePackageVersionsResponse], error) {
	if s.ErrorCode != 0 {
		return nil, connect.NewError(s.ErrorCode, fmt.Errorf("Non-OK response"))
	}
	return connect.NewResponse(&corev1.GetAvailablePackageVersionsResponse{
		PackageAppVersions: s.PackageAppVersions,
	}), nil
}

// GetInstalledPackageResourceRefs returns the resource references based on the request.
func (s TestPackagingPluginServer) GetInstalledPackageResourceRefs(ctx context.Context, request *connect.Request[corev1.GetInstalledPackageResourceRefsRequest]) (*connect.Response[corev1.GetInstalledPackageResourceRefsResponse], error) {
	if s.ErrorCode != 0 {
		return nil, connect.NewError(s.ErrorCode, fmt.Errorf("Non-OK response"))
	}
	return connect.NewResponse(&corev1.GetInstalledPackageResourceRefsResponse{
		Context:      request.Msg.GetInstalledPackageRef().GetContext(),
		ResourceRefs: s.ResourceRefs,
	}), nil
}

func (s TestPackagingPluginServer) CreateInstalledPackage(ctx context.Context, request *connect.Request[corev1.CreateInstalledPackageRequest]) (*connect.Response[corev1.CreateInstalledPackageResponse], error) {
	if s.ErrorCode != 0 {
		return nil, connect.NewError(s.ErrorCode, fmt.Errorf("Non-OK response"))
	}
	return connect.NewResponse(&corev1.CreateInstalledPackageResponse{
		InstalledPackageRef: &corev1.InstalledPackageReference{
			Context:    request.Msg.GetTargetContext(),
			Identifier: request.Msg.GetName(),
			Plugin:     s.Plugin,
		},
	}), nil
}

func (s TestPackagingPluginServer) UpdateInstalledPackage(ctx context.Context, request *connect.Request[corev1.UpdateInstalledPackageRequest]) (*connect.Response[corev1.UpdateInstalledPackageResponse], error) {
	if s.ErrorCode != 0 {
		return nil, connect.NewError(s.ErrorCode, fmt.Errorf("Non-OK response"))
	}
	return connect.NewResponse(&corev1.UpdateInstalledPackageResponse{
		InstalledPackageRef: &corev1.InstalledPackageReference{
			Context:    request.Msg.GetInstalledPackageRef().GetContext(),
			Identifier: request.Msg.GetInstalledPackageRef().GetIdentifier(),
			Plugin:     s.Plugin,
		},
	}), nil
}

func (s TestPackagingPluginServer) DeleteInstalledPackage(ctx context.Context, request *connect.Request[corev1.DeleteInstalledPackageRequest]) (*connect.Response[corev1.DeleteInstalledPackageResponse], error) {
	if s.ErrorCode != 0 {
		return nil, connect.NewError(s.ErrorCode, fmt.Errorf("Non-OK response"))
	}
	return connect.NewResponse(&corev1.DeleteInstalledPackageResponse{}), nil
}

type TestRepositoriesPluginServer struct {
	corev1.UnimplementedRepositoriesServiceServer
	Plugin                     *plugins.Plugin
	PackageRepositoryDetail    *corev1.PackageRepositoryDetail
	PackageRepositorySummaries []*corev1.PackageRepositorySummary
	ErrorCode                  connect.Code
}

func NewTestRepositoriesPlugin(plugin *plugins.Plugin) *TestRepositoriesPluginServer {
	return &TestRepositoriesPluginServer{
		Plugin: plugin,
	}
}

func (s TestRepositoriesPluginServer) AddPackageRepository(ctx context.Context, request *connect.Request[corev1.AddPackageRepositoryRequest]) (*connect.Response[corev1.AddPackageRepositoryResponse], error) {
	if s.ErrorCode != 0 {
		return nil, connect.NewError(s.ErrorCode, fmt.Errorf("Non-OK response"))
	}
	return connect.NewResponse(&corev1.AddPackageRepositoryResponse{
		PackageRepoRef: &corev1.PackageRepositoryReference{
			Context:    request.Msg.GetContext(),
			Identifier: request.Msg.GetName(),
			Plugin:     s.Plugin,
		},
	}), nil
}

func (s TestRepositoriesPluginServer) GetPackageRepositoryDetail(ctx context.Context, request *connect.Request[corev1.GetPackageRepositoryDetailRequest]) (*connect.Response[corev1.GetPackageRepositoryDetailResponse], error) {
	if s.ErrorCode != 0 {
		return nil, connect.NewError(s.ErrorCode, fmt.Errorf("Non-OK response"))
	}
	return connect.NewResponse(&corev1.GetPackageRepositoryDetailResponse{
		Detail: s.PackageRepositoryDetail,
	}), nil
}

// GetPackageRepositorySummaries returns the package repository summaries based on the request.
func (s TestRepositoriesPluginServer) GetPackageRepositorySummaries(ctx context.Context, request *connect.Request[corev1.GetPackageRepositorySummariesRequest]) (*connect.Response[corev1.GetPackageRepositorySummariesResponse], error) {
	if s.ErrorCode != 0 {
		return nil, connect.NewError(s.ErrorCode, fmt.Errorf("Non-OK response"))
	}
	return connect.NewResponse(&corev1.GetPackageRepositorySummariesResponse{
		PackageRepositorySummaries: s.PackageRepositorySummaries,
	}), nil
}

func (s TestRepositoriesPluginServer) UpdatePackageRepository(ctx context.Context, request *connect.Request[corev1.UpdatePackageRepositoryRequest]) (*connect.Response[corev1.UpdatePackageRepositoryResponse], error) {
	if s.ErrorCode != 0 {
		return nil, connect.NewError(s.ErrorCode, fmt.Errorf("Non-OK response"))
	}
	return connect.NewResponse(&corev1.UpdatePackageRepositoryResponse{
		PackageRepoRef: &corev1.PackageRepositoryReference{
			Context:    request.Msg.GetPackageRepoRef().GetContext(),
			Identifier: request.Msg.GetPackageRepoRef().GetIdentifier(),
			Plugin:     s.Plugin,
		},
	}), nil
}

func (s TestRepositoriesPluginServer) DeletePackageRepository(ctx context.Context, request *connect.Request[corev1.DeletePackageRepositoryRequest]) (*connect.Response[corev1.DeletePackageRepositoryResponse], error) {
	if s.ErrorCode != 0 {
		return nil, connect.NewError(s.ErrorCode, fmt.Errorf("Non-OK response"))
	}
	return connect.NewResponse(&corev1.DeletePackageRepositoryResponse{}), nil
}

func (s TestRepositoriesPluginServer) GetPackageRepositoryPermissions(ctx context.Context, request *connect.Request[corev1.GetPackageRepositoryPermissionsRequest]) (*connect.Response[corev1.GetPackageRepositoryPermissionsResponse], error) {
	if s.ErrorCode != 0 {
		return nil, connect.NewError(s.ErrorCode, fmt.Errorf("Non-OK response"))
	}
	return connect.NewResponse(&corev1.GetPackageRepositoryPermissionsResponse{
		Permissions: []*corev1.PackageRepositoriesPermissions{
			{
				Plugin: s.Plugin,
				Namespace: map[string]bool{
					"ns-verb": true,
				},
				Global: map[string]bool{
					"global-verb": true,
				},
			},
		},
	}), nil
}
