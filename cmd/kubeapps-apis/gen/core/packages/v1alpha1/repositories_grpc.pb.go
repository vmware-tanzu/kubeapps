// Copyright 2021-2023 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             (unknown)
// source: kubeappsapis/core/packages/v1alpha1/repositories.proto

package v1alpha1

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

const (
	RepositoriesService_AddPackageRepository_FullMethodName            = "/kubeappsapis.core.packages.v1alpha1.RepositoriesService/AddPackageRepository"
	RepositoriesService_GetPackageRepositoryDetail_FullMethodName      = "/kubeappsapis.core.packages.v1alpha1.RepositoriesService/GetPackageRepositoryDetail"
	RepositoriesService_GetPackageRepositorySummaries_FullMethodName   = "/kubeappsapis.core.packages.v1alpha1.RepositoriesService/GetPackageRepositorySummaries"
	RepositoriesService_UpdatePackageRepository_FullMethodName         = "/kubeappsapis.core.packages.v1alpha1.RepositoriesService/UpdatePackageRepository"
	RepositoriesService_DeletePackageRepository_FullMethodName         = "/kubeappsapis.core.packages.v1alpha1.RepositoriesService/DeletePackageRepository"
	RepositoriesService_GetPackageRepositoryPermissions_FullMethodName = "/kubeappsapis.core.packages.v1alpha1.RepositoriesService/GetPackageRepositoryPermissions"
)

// RepositoriesServiceClient is the client API for RepositoriesService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type RepositoriesServiceClient interface {
	AddPackageRepository(ctx context.Context, in *AddPackageRepositoryRequest, opts ...grpc.CallOption) (*AddPackageRepositoryResponse, error)
	GetPackageRepositoryDetail(ctx context.Context, in *GetPackageRepositoryDetailRequest, opts ...grpc.CallOption) (*GetPackageRepositoryDetailResponse, error)
	GetPackageRepositorySummaries(ctx context.Context, in *GetPackageRepositorySummariesRequest, opts ...grpc.CallOption) (*GetPackageRepositorySummariesResponse, error)
	UpdatePackageRepository(ctx context.Context, in *UpdatePackageRepositoryRequest, opts ...grpc.CallOption) (*UpdatePackageRepositoryResponse, error)
	DeletePackageRepository(ctx context.Context, in *DeletePackageRepositoryRequest, opts ...grpc.CallOption) (*DeletePackageRepositoryResponse, error)
	GetPackageRepositoryPermissions(ctx context.Context, in *GetPackageRepositoryPermissionsRequest, opts ...grpc.CallOption) (*GetPackageRepositoryPermissionsResponse, error)
}

type repositoriesServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewRepositoriesServiceClient(cc grpc.ClientConnInterface) RepositoriesServiceClient {
	return &repositoriesServiceClient{cc}
}

func (c *repositoriesServiceClient) AddPackageRepository(ctx context.Context, in *AddPackageRepositoryRequest, opts ...grpc.CallOption) (*AddPackageRepositoryResponse, error) {
	out := new(AddPackageRepositoryResponse)
	err := c.cc.Invoke(ctx, RepositoriesService_AddPackageRepository_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *repositoriesServiceClient) GetPackageRepositoryDetail(ctx context.Context, in *GetPackageRepositoryDetailRequest, opts ...grpc.CallOption) (*GetPackageRepositoryDetailResponse, error) {
	out := new(GetPackageRepositoryDetailResponse)
	err := c.cc.Invoke(ctx, RepositoriesService_GetPackageRepositoryDetail_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *repositoriesServiceClient) GetPackageRepositorySummaries(ctx context.Context, in *GetPackageRepositorySummariesRequest, opts ...grpc.CallOption) (*GetPackageRepositorySummariesResponse, error) {
	out := new(GetPackageRepositorySummariesResponse)
	err := c.cc.Invoke(ctx, RepositoriesService_GetPackageRepositorySummaries_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *repositoriesServiceClient) UpdatePackageRepository(ctx context.Context, in *UpdatePackageRepositoryRequest, opts ...grpc.CallOption) (*UpdatePackageRepositoryResponse, error) {
	out := new(UpdatePackageRepositoryResponse)
	err := c.cc.Invoke(ctx, RepositoriesService_UpdatePackageRepository_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *repositoriesServiceClient) DeletePackageRepository(ctx context.Context, in *DeletePackageRepositoryRequest, opts ...grpc.CallOption) (*DeletePackageRepositoryResponse, error) {
	out := new(DeletePackageRepositoryResponse)
	err := c.cc.Invoke(ctx, RepositoriesService_DeletePackageRepository_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *repositoriesServiceClient) GetPackageRepositoryPermissions(ctx context.Context, in *GetPackageRepositoryPermissionsRequest, opts ...grpc.CallOption) (*GetPackageRepositoryPermissionsResponse, error) {
	out := new(GetPackageRepositoryPermissionsResponse)
	err := c.cc.Invoke(ctx, RepositoriesService_GetPackageRepositoryPermissions_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// RepositoriesServiceServer is the server API for RepositoriesService service.
// All implementations should embed UnimplementedRepositoriesServiceServer
// for forward compatibility
type RepositoriesServiceServer interface {
	AddPackageRepository(context.Context, *AddPackageRepositoryRequest) (*AddPackageRepositoryResponse, error)
	GetPackageRepositoryDetail(context.Context, *GetPackageRepositoryDetailRequest) (*GetPackageRepositoryDetailResponse, error)
	GetPackageRepositorySummaries(context.Context, *GetPackageRepositorySummariesRequest) (*GetPackageRepositorySummariesResponse, error)
	UpdatePackageRepository(context.Context, *UpdatePackageRepositoryRequest) (*UpdatePackageRepositoryResponse, error)
	DeletePackageRepository(context.Context, *DeletePackageRepositoryRequest) (*DeletePackageRepositoryResponse, error)
	GetPackageRepositoryPermissions(context.Context, *GetPackageRepositoryPermissionsRequest) (*GetPackageRepositoryPermissionsResponse, error)
}

// UnimplementedRepositoriesServiceServer should be embedded to have forward compatible implementations.
type UnimplementedRepositoriesServiceServer struct {
}

func (UnimplementedRepositoriesServiceServer) AddPackageRepository(context.Context, *AddPackageRepositoryRequest) (*AddPackageRepositoryResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AddPackageRepository not implemented")
}
func (UnimplementedRepositoriesServiceServer) GetPackageRepositoryDetail(context.Context, *GetPackageRepositoryDetailRequest) (*GetPackageRepositoryDetailResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetPackageRepositoryDetail not implemented")
}
func (UnimplementedRepositoriesServiceServer) GetPackageRepositorySummaries(context.Context, *GetPackageRepositorySummariesRequest) (*GetPackageRepositorySummariesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetPackageRepositorySummaries not implemented")
}
func (UnimplementedRepositoriesServiceServer) UpdatePackageRepository(context.Context, *UpdatePackageRepositoryRequest) (*UpdatePackageRepositoryResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdatePackageRepository not implemented")
}
func (UnimplementedRepositoriesServiceServer) DeletePackageRepository(context.Context, *DeletePackageRepositoryRequest) (*DeletePackageRepositoryResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeletePackageRepository not implemented")
}
func (UnimplementedRepositoriesServiceServer) GetPackageRepositoryPermissions(context.Context, *GetPackageRepositoryPermissionsRequest) (*GetPackageRepositoryPermissionsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetPackageRepositoryPermissions not implemented")
}

// UnsafeRepositoriesServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to RepositoriesServiceServer will
// result in compilation errors.
type UnsafeRepositoriesServiceServer interface {
	mustEmbedUnimplementedRepositoriesServiceServer()
}

func RegisterRepositoriesServiceServer(s grpc.ServiceRegistrar, srv RepositoriesServiceServer) {
	s.RegisterService(&RepositoriesService_ServiceDesc, srv)
}

func _RepositoriesService_AddPackageRepository_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AddPackageRepositoryRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RepositoriesServiceServer).AddPackageRepository(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: RepositoriesService_AddPackageRepository_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RepositoriesServiceServer).AddPackageRepository(ctx, req.(*AddPackageRepositoryRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _RepositoriesService_GetPackageRepositoryDetail_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetPackageRepositoryDetailRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RepositoriesServiceServer).GetPackageRepositoryDetail(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: RepositoriesService_GetPackageRepositoryDetail_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RepositoriesServiceServer).GetPackageRepositoryDetail(ctx, req.(*GetPackageRepositoryDetailRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _RepositoriesService_GetPackageRepositorySummaries_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetPackageRepositorySummariesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RepositoriesServiceServer).GetPackageRepositorySummaries(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: RepositoriesService_GetPackageRepositorySummaries_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RepositoriesServiceServer).GetPackageRepositorySummaries(ctx, req.(*GetPackageRepositorySummariesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _RepositoriesService_UpdatePackageRepository_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdatePackageRepositoryRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RepositoriesServiceServer).UpdatePackageRepository(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: RepositoriesService_UpdatePackageRepository_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RepositoriesServiceServer).UpdatePackageRepository(ctx, req.(*UpdatePackageRepositoryRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _RepositoriesService_DeletePackageRepository_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeletePackageRepositoryRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RepositoriesServiceServer).DeletePackageRepository(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: RepositoriesService_DeletePackageRepository_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RepositoriesServiceServer).DeletePackageRepository(ctx, req.(*DeletePackageRepositoryRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _RepositoriesService_GetPackageRepositoryPermissions_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetPackageRepositoryPermissionsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RepositoriesServiceServer).GetPackageRepositoryPermissions(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: RepositoriesService_GetPackageRepositoryPermissions_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RepositoriesServiceServer).GetPackageRepositoryPermissions(ctx, req.(*GetPackageRepositoryPermissionsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// RepositoriesService_ServiceDesc is the grpc.ServiceDesc for RepositoriesService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var RepositoriesService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "kubeappsapis.core.packages.v1alpha1.RepositoriesService",
	HandlerType: (*RepositoriesServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "AddPackageRepository",
			Handler:    _RepositoriesService_AddPackageRepository_Handler,
		},
		{
			MethodName: "GetPackageRepositoryDetail",
			Handler:    _RepositoriesService_GetPackageRepositoryDetail_Handler,
		},
		{
			MethodName: "GetPackageRepositorySummaries",
			Handler:    _RepositoriesService_GetPackageRepositorySummaries_Handler,
		},
		{
			MethodName: "UpdatePackageRepository",
			Handler:    _RepositoriesService_UpdatePackageRepository_Handler,
		},
		{
			MethodName: "DeletePackageRepository",
			Handler:    _RepositoriesService_DeletePackageRepository_Handler,
		},
		{
			MethodName: "GetPackageRepositoryPermissions",
			Handler:    _RepositoriesService_GetPackageRepositoryPermissions_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "kubeappsapis/core/packages/v1alpha1/repositories.proto",
}
