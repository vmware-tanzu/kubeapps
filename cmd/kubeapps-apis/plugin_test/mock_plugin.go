// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package plugin_test

import (
	"context"

	pkgsGRPCv1alpha1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	pluginsGRPCv1alpha1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/plugins/v1alpha1"
	grpccodes "google.golang.org/grpc/codes"
	grpcstatus "google.golang.org/grpc/status"
)

type TestPackagingPluginServer struct {
	pkgsGRPCv1alpha1.UnimplementedPackagesServiceServer
	Plugin                    *pluginsGRPCv1alpha1.Plugin
	AvailablePackageSummaries []*pkgsGRPCv1alpha1.AvailablePackageSummary
	AvailablePackageDetail    *pkgsGRPCv1alpha1.AvailablePackageDetail
	InstalledPackageSummaries []*pkgsGRPCv1alpha1.InstalledPackageSummary
	InstalledPackageDetail    *pkgsGRPCv1alpha1.InstalledPackageDetail
	PackageAppVersions        []*pkgsGRPCv1alpha1.PackageAppVersion
	ResourceRefs              []*pkgsGRPCv1alpha1.ResourceRef
	Categories                []string
	NextPageToken             string
	Status                    grpccodes.Code
}

func NewTestPackagingPlugin(plugin *pluginsGRPCv1alpha1.Plugin) *TestPackagingPluginServer {
	return &TestPackagingPluginServer{
		Plugin: plugin,
	}
}

// GetAvailablePackages returns the packages based on the request.
func (s TestPackagingPluginServer) GetAvailablePackageSummaries(ctx context.Context, request *pkgsGRPCv1alpha1.GetAvailablePackageSummariesRequest) (*pkgsGRPCv1alpha1.GetAvailablePackageSummariesResponse, error) {
	if s.Status != grpccodes.OK {
		return nil, grpcstatus.Errorf(s.Status, "Non-OK response")
	}
	return &pkgsGRPCv1alpha1.GetAvailablePackageSummariesResponse{
		AvailablePackageSummaries: s.AvailablePackageSummaries,
		Categories:                s.Categories,
		NextPageToken:             s.NextPageToken,
	}, nil
}

// GetAvailablePackageDetail returns the package details based on the request.
func (s TestPackagingPluginServer) GetAvailablePackageDetail(ctx context.Context, request *pkgsGRPCv1alpha1.GetAvailablePackageDetailRequest) (*pkgsGRPCv1alpha1.GetAvailablePackageDetailResponse, error) {
	if s.Status != grpccodes.OK {
		return nil, grpcstatus.Errorf(s.Status, "Non-OK response")
	}
	return &pkgsGRPCv1alpha1.GetAvailablePackageDetailResponse{
		AvailablePackageDetail: s.AvailablePackageDetail,
	}, nil
}

// GetInstalledPackageSummaries returns the installed package summaries based on the request.
func (s TestPackagingPluginServer) GetInstalledPackageSummaries(ctx context.Context, request *pkgsGRPCv1alpha1.GetInstalledPackageSummariesRequest) (*pkgsGRPCv1alpha1.GetInstalledPackageSummariesResponse, error) {
	if s.Status != grpccodes.OK {
		return nil, grpcstatus.Errorf(s.Status, "Non-OK response")
	}
	return &pkgsGRPCv1alpha1.GetInstalledPackageSummariesResponse{
		InstalledPackageSummaries: s.InstalledPackageSummaries,
		NextPageToken:             s.NextPageToken,
	}, nil
}

// GetInstalledPackageDetail returns the package versions based on the request.
func (s TestPackagingPluginServer) GetInstalledPackageDetail(ctx context.Context, request *pkgsGRPCv1alpha1.GetInstalledPackageDetailRequest) (*pkgsGRPCv1alpha1.GetInstalledPackageDetailResponse, error) {
	if s.Status != grpccodes.OK {
		return nil, grpcstatus.Errorf(s.Status, "Non-OK response")
	}
	return &pkgsGRPCv1alpha1.GetInstalledPackageDetailResponse{
		InstalledPackageDetail: s.InstalledPackageDetail,
	}, nil
}

// GetAvailablePackageVersions returns the package versions based on the request.
func (s TestPackagingPluginServer) GetAvailablePackageVersions(ctx context.Context, request *pkgsGRPCv1alpha1.GetAvailablePackageVersionsRequest) (*pkgsGRPCv1alpha1.GetAvailablePackageVersionsResponse, error) {
	if s.Status != grpccodes.OK {
		return nil, grpcstatus.Errorf(s.Status, "Non-OK response")
	}
	return &pkgsGRPCv1alpha1.GetAvailablePackageVersionsResponse{
		PackageAppVersions: s.PackageAppVersions,
	}, nil
}

// GetInstalledPackageResourceRefs returns the resource references based on the request.
func (s TestPackagingPluginServer) GetInstalledPackageResourceRefs(ctx context.Context, request *pkgsGRPCv1alpha1.GetInstalledPackageResourceRefsRequest) (*pkgsGRPCv1alpha1.GetInstalledPackageResourceRefsResponse, error) {
	if s.Status != grpccodes.OK {
		return nil, grpcstatus.Errorf(s.Status, "Non-OK response")
	}
	return &pkgsGRPCv1alpha1.GetInstalledPackageResourceRefsResponse{
		Context:      request.GetInstalledPackageRef().GetContext(),
		ResourceRefs: s.ResourceRefs,
	}, nil
}

func (s TestPackagingPluginServer) CreateInstalledPackage(ctx context.Context, request *pkgsGRPCv1alpha1.CreateInstalledPackageRequest) (*pkgsGRPCv1alpha1.CreateInstalledPackageResponse, error) {
	if s.Status != grpccodes.OK {
		return nil, grpcstatus.Errorf(s.Status, "Non-OK response")
	}
	return &pkgsGRPCv1alpha1.CreateInstalledPackageResponse{
		InstalledPackageRef: &pkgsGRPCv1alpha1.InstalledPackageReference{
			Context:    request.GetTargetContext(),
			Identifier: request.GetName(),
			Plugin:     s.Plugin,
		},
	}, nil
}

func (s TestPackagingPluginServer) UpdateInstalledPackage(ctx context.Context, request *pkgsGRPCv1alpha1.UpdateInstalledPackageRequest) (*pkgsGRPCv1alpha1.UpdateInstalledPackageResponse, error) {
	if s.Status != grpccodes.OK {
		return nil, grpcstatus.Errorf(s.Status, "Non-OK response")
	}
	return &pkgsGRPCv1alpha1.UpdateInstalledPackageResponse{
		InstalledPackageRef: &pkgsGRPCv1alpha1.InstalledPackageReference{
			Context:    request.GetInstalledPackageRef().GetContext(),
			Identifier: request.GetInstalledPackageRef().GetIdentifier(),
			Plugin:     s.Plugin,
		},
	}, nil
}

func (s TestPackagingPluginServer) DeleteInstalledPackage(ctx context.Context, request *pkgsGRPCv1alpha1.DeleteInstalledPackageRequest) (*pkgsGRPCv1alpha1.DeleteInstalledPackageResponse, error) {
	if s.Status != grpccodes.OK {
		return nil, grpcstatus.Errorf(s.Status, "Non-OK response")
	}
	return &pkgsGRPCv1alpha1.DeleteInstalledPackageResponse{}, nil
}
