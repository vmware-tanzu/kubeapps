// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package v1alpha1

import (
	context "context"
	v1alpha1 "github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// HelmPackagesServiceClient is the client API for HelmPackagesService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type HelmPackagesServiceClient interface {
	// GetAvailablePackageSummaries returns the available packages managed by the 'helm' plugin
	GetAvailablePackageSummaries(ctx context.Context, in *v1alpha1.GetAvailablePackageSummariesRequest, opts ...grpc.CallOption) (*v1alpha1.GetAvailablePackageSummariesResponse, error)
	// GetAvailablePackageDetail returns the package details managed by the 'helm' plugin
	GetAvailablePackageDetail(ctx context.Context, in *v1alpha1.GetAvailablePackageDetailRequest, opts ...grpc.CallOption) (*v1alpha1.GetAvailablePackageDetailResponse, error)
	// GetAvailablePackageVersions returns the package versions managed by the 'helm' plugin
	GetAvailablePackageVersions(ctx context.Context, in *v1alpha1.GetAvailablePackageVersionsRequest, opts ...grpc.CallOption) (*v1alpha1.GetAvailablePackageVersionsResponse, error)
	// GetInstalledPackageSummaries returns the installed packages managed by the 'helm' plugin
	GetInstalledPackageSummaries(ctx context.Context, in *v1alpha1.GetInstalledPackageSummariesRequest, opts ...grpc.CallOption) (*v1alpha1.GetInstalledPackageSummariesResponse, error)
	// GetInstalledPackageDetail returns the requested installed package managed by the 'helm' plugin
	GetInstalledPackageDetail(ctx context.Context, in *v1alpha1.GetInstalledPackageDetailRequest, opts ...grpc.CallOption) (*v1alpha1.GetInstalledPackageDetailResponse, error)
	// CreateInstalledPackage creates an installed package based on the request.
	CreateInstalledPackage(ctx context.Context, in *v1alpha1.CreateInstalledPackageRequest, opts ...grpc.CallOption) (*v1alpha1.CreateInstalledPackageResponse, error)
	// UpdateInstalledPackage updates an installed package based on the request.
	UpdateInstalledPackage(ctx context.Context, in *v1alpha1.UpdateInstalledPackageRequest, opts ...grpc.CallOption) (*v1alpha1.UpdateInstalledPackageResponse, error)
	// DeleteInstalledPackage deletes an installed package based on the request.
	DeleteInstalledPackage(ctx context.Context, in *v1alpha1.DeleteInstalledPackageRequest, opts ...grpc.CallOption) (*v1alpha1.DeleteInstalledPackageResponse, error)
	// RollbackInstalledPackage updates an installed package based on the request.
	RollbackInstalledPackage(ctx context.Context, in *RollbackInstalledPackageRequest, opts ...grpc.CallOption) (*RollbackInstalledPackageResponse, error)
	// GetInstalledPackageResourceRefs returns the references for the Kubernetes resources created by
	// an installed package.
	GetInstalledPackageResourceRefs(ctx context.Context, in *v1alpha1.GetInstalledPackageResourceRefsRequest, opts ...grpc.CallOption) (*v1alpha1.GetInstalledPackageResourceRefsResponse, error)
}

type helmPackagesServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewHelmPackagesServiceClient(cc grpc.ClientConnInterface) HelmPackagesServiceClient {
	return &helmPackagesServiceClient{cc}
}

func (c *helmPackagesServiceClient) GetAvailablePackageSummaries(ctx context.Context, in *v1alpha1.GetAvailablePackageSummariesRequest, opts ...grpc.CallOption) (*v1alpha1.GetAvailablePackageSummariesResponse, error) {
	out := new(v1alpha1.GetAvailablePackageSummariesResponse)
	err := c.cc.Invoke(ctx, "/kubeappsapis.plugins.helm.packages.v1alpha1.HelmPackagesService/GetAvailablePackageSummaries", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *helmPackagesServiceClient) GetAvailablePackageDetail(ctx context.Context, in *v1alpha1.GetAvailablePackageDetailRequest, opts ...grpc.CallOption) (*v1alpha1.GetAvailablePackageDetailResponse, error) {
	out := new(v1alpha1.GetAvailablePackageDetailResponse)
	err := c.cc.Invoke(ctx, "/kubeappsapis.plugins.helm.packages.v1alpha1.HelmPackagesService/GetAvailablePackageDetail", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *helmPackagesServiceClient) GetAvailablePackageVersions(ctx context.Context, in *v1alpha1.GetAvailablePackageVersionsRequest, opts ...grpc.CallOption) (*v1alpha1.GetAvailablePackageVersionsResponse, error) {
	out := new(v1alpha1.GetAvailablePackageVersionsResponse)
	err := c.cc.Invoke(ctx, "/kubeappsapis.plugins.helm.packages.v1alpha1.HelmPackagesService/GetAvailablePackageVersions", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *helmPackagesServiceClient) GetInstalledPackageSummaries(ctx context.Context, in *v1alpha1.GetInstalledPackageSummariesRequest, opts ...grpc.CallOption) (*v1alpha1.GetInstalledPackageSummariesResponse, error) {
	out := new(v1alpha1.GetInstalledPackageSummariesResponse)
	err := c.cc.Invoke(ctx, "/kubeappsapis.plugins.helm.packages.v1alpha1.HelmPackagesService/GetInstalledPackageSummaries", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *helmPackagesServiceClient) GetInstalledPackageDetail(ctx context.Context, in *v1alpha1.GetInstalledPackageDetailRequest, opts ...grpc.CallOption) (*v1alpha1.GetInstalledPackageDetailResponse, error) {
	out := new(v1alpha1.GetInstalledPackageDetailResponse)
	err := c.cc.Invoke(ctx, "/kubeappsapis.plugins.helm.packages.v1alpha1.HelmPackagesService/GetInstalledPackageDetail", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *helmPackagesServiceClient) CreateInstalledPackage(ctx context.Context, in *v1alpha1.CreateInstalledPackageRequest, opts ...grpc.CallOption) (*v1alpha1.CreateInstalledPackageResponse, error) {
	out := new(v1alpha1.CreateInstalledPackageResponse)
	err := c.cc.Invoke(ctx, "/kubeappsapis.plugins.helm.packages.v1alpha1.HelmPackagesService/CreateInstalledPackage", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *helmPackagesServiceClient) UpdateInstalledPackage(ctx context.Context, in *v1alpha1.UpdateInstalledPackageRequest, opts ...grpc.CallOption) (*v1alpha1.UpdateInstalledPackageResponse, error) {
	out := new(v1alpha1.UpdateInstalledPackageResponse)
	err := c.cc.Invoke(ctx, "/kubeappsapis.plugins.helm.packages.v1alpha1.HelmPackagesService/UpdateInstalledPackage", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *helmPackagesServiceClient) DeleteInstalledPackage(ctx context.Context, in *v1alpha1.DeleteInstalledPackageRequest, opts ...grpc.CallOption) (*v1alpha1.DeleteInstalledPackageResponse, error) {
	out := new(v1alpha1.DeleteInstalledPackageResponse)
	err := c.cc.Invoke(ctx, "/kubeappsapis.plugins.helm.packages.v1alpha1.HelmPackagesService/DeleteInstalledPackage", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *helmPackagesServiceClient) RollbackInstalledPackage(ctx context.Context, in *RollbackInstalledPackageRequest, opts ...grpc.CallOption) (*RollbackInstalledPackageResponse, error) {
	out := new(RollbackInstalledPackageResponse)
	err := c.cc.Invoke(ctx, "/kubeappsapis.plugins.helm.packages.v1alpha1.HelmPackagesService/RollbackInstalledPackage", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *helmPackagesServiceClient) GetInstalledPackageResourceRefs(ctx context.Context, in *v1alpha1.GetInstalledPackageResourceRefsRequest, opts ...grpc.CallOption) (*v1alpha1.GetInstalledPackageResourceRefsResponse, error) {
	out := new(v1alpha1.GetInstalledPackageResourceRefsResponse)
	err := c.cc.Invoke(ctx, "/kubeappsapis.plugins.helm.packages.v1alpha1.HelmPackagesService/GetInstalledPackageResourceRefs", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// HelmPackagesServiceServer is the server API for HelmPackagesService service.
// All implementations should embed UnimplementedHelmPackagesServiceServer
// for forward compatibility
type HelmPackagesServiceServer interface {
	// GetAvailablePackageSummaries returns the available packages managed by the 'helm' plugin
	GetAvailablePackageSummaries(context.Context, *v1alpha1.GetAvailablePackageSummariesRequest) (*v1alpha1.GetAvailablePackageSummariesResponse, error)
	// GetAvailablePackageDetail returns the package details managed by the 'helm' plugin
	GetAvailablePackageDetail(context.Context, *v1alpha1.GetAvailablePackageDetailRequest) (*v1alpha1.GetAvailablePackageDetailResponse, error)
	// GetAvailablePackageVersions returns the package versions managed by the 'helm' plugin
	GetAvailablePackageVersions(context.Context, *v1alpha1.GetAvailablePackageVersionsRequest) (*v1alpha1.GetAvailablePackageVersionsResponse, error)
	// GetInstalledPackageSummaries returns the installed packages managed by the 'helm' plugin
	GetInstalledPackageSummaries(context.Context, *v1alpha1.GetInstalledPackageSummariesRequest) (*v1alpha1.GetInstalledPackageSummariesResponse, error)
	// GetInstalledPackageDetail returns the requested installed package managed by the 'helm' plugin
	GetInstalledPackageDetail(context.Context, *v1alpha1.GetInstalledPackageDetailRequest) (*v1alpha1.GetInstalledPackageDetailResponse, error)
	// CreateInstalledPackage creates an installed package based on the request.
	CreateInstalledPackage(context.Context, *v1alpha1.CreateInstalledPackageRequest) (*v1alpha1.CreateInstalledPackageResponse, error)
	// UpdateInstalledPackage updates an installed package based on the request.
	UpdateInstalledPackage(context.Context, *v1alpha1.UpdateInstalledPackageRequest) (*v1alpha1.UpdateInstalledPackageResponse, error)
	// DeleteInstalledPackage deletes an installed package based on the request.
	DeleteInstalledPackage(context.Context, *v1alpha1.DeleteInstalledPackageRequest) (*v1alpha1.DeleteInstalledPackageResponse, error)
	// RollbackInstalledPackage updates an installed package based on the request.
	RollbackInstalledPackage(context.Context, *RollbackInstalledPackageRequest) (*RollbackInstalledPackageResponse, error)
	// GetInstalledPackageResourceRefs returns the references for the Kubernetes resources created by
	// an installed package.
	GetInstalledPackageResourceRefs(context.Context, *v1alpha1.GetInstalledPackageResourceRefsRequest) (*v1alpha1.GetInstalledPackageResourceRefsResponse, error)
}

// UnimplementedHelmPackagesServiceServer should be embedded to have forward compatible implementations.
type UnimplementedHelmPackagesServiceServer struct {
}

func (UnimplementedHelmPackagesServiceServer) GetAvailablePackageSummaries(context.Context, *v1alpha1.GetAvailablePackageSummariesRequest) (*v1alpha1.GetAvailablePackageSummariesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetAvailablePackageSummaries not implemented")
}
func (UnimplementedHelmPackagesServiceServer) GetAvailablePackageDetail(context.Context, *v1alpha1.GetAvailablePackageDetailRequest) (*v1alpha1.GetAvailablePackageDetailResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetAvailablePackageDetail not implemented")
}
func (UnimplementedHelmPackagesServiceServer) GetAvailablePackageVersions(context.Context, *v1alpha1.GetAvailablePackageVersionsRequest) (*v1alpha1.GetAvailablePackageVersionsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetAvailablePackageVersions not implemented")
}
func (UnimplementedHelmPackagesServiceServer) GetInstalledPackageSummaries(context.Context, *v1alpha1.GetInstalledPackageSummariesRequest) (*v1alpha1.GetInstalledPackageSummariesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetInstalledPackageSummaries not implemented")
}
func (UnimplementedHelmPackagesServiceServer) GetInstalledPackageDetail(context.Context, *v1alpha1.GetInstalledPackageDetailRequest) (*v1alpha1.GetInstalledPackageDetailResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetInstalledPackageDetail not implemented")
}
func (UnimplementedHelmPackagesServiceServer) CreateInstalledPackage(context.Context, *v1alpha1.CreateInstalledPackageRequest) (*v1alpha1.CreateInstalledPackageResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateInstalledPackage not implemented")
}
func (UnimplementedHelmPackagesServiceServer) UpdateInstalledPackage(context.Context, *v1alpha1.UpdateInstalledPackageRequest) (*v1alpha1.UpdateInstalledPackageResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateInstalledPackage not implemented")
}
func (UnimplementedHelmPackagesServiceServer) DeleteInstalledPackage(context.Context, *v1alpha1.DeleteInstalledPackageRequest) (*v1alpha1.DeleteInstalledPackageResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteInstalledPackage not implemented")
}
func (UnimplementedHelmPackagesServiceServer) RollbackInstalledPackage(context.Context, *RollbackInstalledPackageRequest) (*RollbackInstalledPackageResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RollbackInstalledPackage not implemented")
}
func (UnimplementedHelmPackagesServiceServer) GetInstalledPackageResourceRefs(context.Context, *v1alpha1.GetInstalledPackageResourceRefsRequest) (*v1alpha1.GetInstalledPackageResourceRefsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetInstalledPackageResourceRefs not implemented")
}

// UnsafeHelmPackagesServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to HelmPackagesServiceServer will
// result in compilation errors.
type UnsafeHelmPackagesServiceServer interface {
	mustEmbedUnimplementedHelmPackagesServiceServer()
}

func RegisterHelmPackagesServiceServer(s grpc.ServiceRegistrar, srv HelmPackagesServiceServer) {
	s.RegisterService(&HelmPackagesService_ServiceDesc, srv)
}

func _HelmPackagesService_GetAvailablePackageSummaries_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(v1alpha1.GetAvailablePackageSummariesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(HelmPackagesServiceServer).GetAvailablePackageSummaries(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/kubeappsapis.plugins.helm.packages.v1alpha1.HelmPackagesService/GetAvailablePackageSummaries",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(HelmPackagesServiceServer).GetAvailablePackageSummaries(ctx, req.(*v1alpha1.GetAvailablePackageSummariesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _HelmPackagesService_GetAvailablePackageDetail_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(v1alpha1.GetAvailablePackageDetailRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(HelmPackagesServiceServer).GetAvailablePackageDetail(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/kubeappsapis.plugins.helm.packages.v1alpha1.HelmPackagesService/GetAvailablePackageDetail",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(HelmPackagesServiceServer).GetAvailablePackageDetail(ctx, req.(*v1alpha1.GetAvailablePackageDetailRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _HelmPackagesService_GetAvailablePackageVersions_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(v1alpha1.GetAvailablePackageVersionsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(HelmPackagesServiceServer).GetAvailablePackageVersions(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/kubeappsapis.plugins.helm.packages.v1alpha1.HelmPackagesService/GetAvailablePackageVersions",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(HelmPackagesServiceServer).GetAvailablePackageVersions(ctx, req.(*v1alpha1.GetAvailablePackageVersionsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _HelmPackagesService_GetInstalledPackageSummaries_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(v1alpha1.GetInstalledPackageSummariesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(HelmPackagesServiceServer).GetInstalledPackageSummaries(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/kubeappsapis.plugins.helm.packages.v1alpha1.HelmPackagesService/GetInstalledPackageSummaries",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(HelmPackagesServiceServer).GetInstalledPackageSummaries(ctx, req.(*v1alpha1.GetInstalledPackageSummariesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _HelmPackagesService_GetInstalledPackageDetail_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(v1alpha1.GetInstalledPackageDetailRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(HelmPackagesServiceServer).GetInstalledPackageDetail(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/kubeappsapis.plugins.helm.packages.v1alpha1.HelmPackagesService/GetInstalledPackageDetail",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(HelmPackagesServiceServer).GetInstalledPackageDetail(ctx, req.(*v1alpha1.GetInstalledPackageDetailRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _HelmPackagesService_CreateInstalledPackage_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(v1alpha1.CreateInstalledPackageRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(HelmPackagesServiceServer).CreateInstalledPackage(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/kubeappsapis.plugins.helm.packages.v1alpha1.HelmPackagesService/CreateInstalledPackage",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(HelmPackagesServiceServer).CreateInstalledPackage(ctx, req.(*v1alpha1.CreateInstalledPackageRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _HelmPackagesService_UpdateInstalledPackage_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(v1alpha1.UpdateInstalledPackageRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(HelmPackagesServiceServer).UpdateInstalledPackage(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/kubeappsapis.plugins.helm.packages.v1alpha1.HelmPackagesService/UpdateInstalledPackage",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(HelmPackagesServiceServer).UpdateInstalledPackage(ctx, req.(*v1alpha1.UpdateInstalledPackageRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _HelmPackagesService_DeleteInstalledPackage_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(v1alpha1.DeleteInstalledPackageRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(HelmPackagesServiceServer).DeleteInstalledPackage(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/kubeappsapis.plugins.helm.packages.v1alpha1.HelmPackagesService/DeleteInstalledPackage",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(HelmPackagesServiceServer).DeleteInstalledPackage(ctx, req.(*v1alpha1.DeleteInstalledPackageRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _HelmPackagesService_RollbackInstalledPackage_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RollbackInstalledPackageRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(HelmPackagesServiceServer).RollbackInstalledPackage(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/kubeappsapis.plugins.helm.packages.v1alpha1.HelmPackagesService/RollbackInstalledPackage",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(HelmPackagesServiceServer).RollbackInstalledPackage(ctx, req.(*RollbackInstalledPackageRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _HelmPackagesService_GetInstalledPackageResourceRefs_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(v1alpha1.GetInstalledPackageResourceRefsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(HelmPackagesServiceServer).GetInstalledPackageResourceRefs(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/kubeappsapis.plugins.helm.packages.v1alpha1.HelmPackagesService/GetInstalledPackageResourceRefs",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(HelmPackagesServiceServer).GetInstalledPackageResourceRefs(ctx, req.(*v1alpha1.GetInstalledPackageResourceRefsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// HelmPackagesService_ServiceDesc is the grpc.ServiceDesc for HelmPackagesService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var HelmPackagesService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "kubeappsapis.plugins.helm.packages.v1alpha1.HelmPackagesService",
	HandlerType: (*HelmPackagesServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetAvailablePackageSummaries",
			Handler:    _HelmPackagesService_GetAvailablePackageSummaries_Handler,
		},
		{
			MethodName: "GetAvailablePackageDetail",
			Handler:    _HelmPackagesService_GetAvailablePackageDetail_Handler,
		},
		{
			MethodName: "GetAvailablePackageVersions",
			Handler:    _HelmPackagesService_GetAvailablePackageVersions_Handler,
		},
		{
			MethodName: "GetInstalledPackageSummaries",
			Handler:    _HelmPackagesService_GetInstalledPackageSummaries_Handler,
		},
		{
			MethodName: "GetInstalledPackageDetail",
			Handler:    _HelmPackagesService_GetInstalledPackageDetail_Handler,
		},
		{
			MethodName: "CreateInstalledPackage",
			Handler:    _HelmPackagesService_CreateInstalledPackage_Handler,
		},
		{
			MethodName: "UpdateInstalledPackage",
			Handler:    _HelmPackagesService_UpdateInstalledPackage_Handler,
		},
		{
			MethodName: "DeleteInstalledPackage",
			Handler:    _HelmPackagesService_DeleteInstalledPackage_Handler,
		},
		{
			MethodName: "RollbackInstalledPackage",
			Handler:    _HelmPackagesService_RollbackInstalledPackage_Handler,
		},
		{
			MethodName: "GetInstalledPackageResourceRefs",
			Handler:    _HelmPackagesService_GetInstalledPackageResourceRefs_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "kubeappsapis/plugins/helm/packages/v1alpha1/helm.proto",
}
