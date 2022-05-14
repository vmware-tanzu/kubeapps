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

// KappControllerPackagesServiceClient is the client API for KappControllerPackagesService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type KappControllerPackagesServiceClient interface {
	// GetAvailablePackageSummaries returns the available packages managed by the 'kapp_controller' plugin
	GetAvailablePackageSummaries(ctx context.Context, in *v1alpha1.GetAvailablePackageSummariesRequest, opts ...grpc.CallOption) (*v1alpha1.GetAvailablePackageSummariesResponse, error)
	// GetAvailablePackageDetail returns the package details managed by the 'kapp_controller' plugin
	GetAvailablePackageDetail(ctx context.Context, in *v1alpha1.GetAvailablePackageDetailRequest, opts ...grpc.CallOption) (*v1alpha1.GetAvailablePackageDetailResponse, error)
	// GetPackageRepositories returns the repositories managed by the 'kapp_controller' plugin
	GetPackageRepositories(ctx context.Context, in *GetPackageRepositoriesRequest, opts ...grpc.CallOption) (*GetPackageRepositoriesResponse, error)
	// GetAvailablePackageVersions returns the package versions managed by the 'kapp_controller' plugin
	GetAvailablePackageVersions(ctx context.Context, in *v1alpha1.GetAvailablePackageVersionsRequest, opts ...grpc.CallOption) (*v1alpha1.GetAvailablePackageVersionsResponse, error)
	// GetInstalledPackageSummaries returns the installed packages managed by the 'kapp_controller' plugin
	GetInstalledPackageSummaries(ctx context.Context, in *v1alpha1.GetInstalledPackageSummariesRequest, opts ...grpc.CallOption) (*v1alpha1.GetInstalledPackageSummariesResponse, error)
	// GetInstalledPackageDetail returns the requested installed package managed by the 'kapp_controller' plugin
	GetInstalledPackageDetail(ctx context.Context, in *v1alpha1.GetInstalledPackageDetailRequest, opts ...grpc.CallOption) (*v1alpha1.GetInstalledPackageDetailResponse, error)
	// CreateInstalledPackage creates an installed package based on the request.
	CreateInstalledPackage(ctx context.Context, in *v1alpha1.CreateInstalledPackageRequest, opts ...grpc.CallOption) (*v1alpha1.CreateInstalledPackageResponse, error)
	// UpdateInstalledPackage updates an installed package based on the request.
	UpdateInstalledPackage(ctx context.Context, in *v1alpha1.UpdateInstalledPackageRequest, opts ...grpc.CallOption) (*v1alpha1.UpdateInstalledPackageResponse, error)
	// DeleteInstalledPackage deletes an installed package based on the request.
	DeleteInstalledPackage(ctx context.Context, in *v1alpha1.DeleteInstalledPackageRequest, opts ...grpc.CallOption) (*v1alpha1.DeleteInstalledPackageResponse, error)
	// GetInstalledPackageResourceRefs returns the references for the Kubernetes resources created by
	// an installed package.
	GetInstalledPackageResourceRefs(ctx context.Context, in *v1alpha1.GetInstalledPackageResourceRefsRequest, opts ...grpc.CallOption) (*v1alpha1.GetInstalledPackageResourceRefsResponse, error)
}

type kappControllerPackagesServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewKappControllerPackagesServiceClient(cc grpc.ClientConnInterface) KappControllerPackagesServiceClient {
	return &kappControllerPackagesServiceClient{cc}
}

func (c *kappControllerPackagesServiceClient) GetAvailablePackageSummaries(ctx context.Context, in *v1alpha1.GetAvailablePackageSummariesRequest, opts ...grpc.CallOption) (*v1alpha1.GetAvailablePackageSummariesResponse, error) {
	out := new(v1alpha1.GetAvailablePackageSummariesResponse)
	err := c.cc.Invoke(ctx, "/kubeappsapis.plugins.kapp_controller.packages.v1alpha1.KappControllerPackagesService/GetAvailablePackageSummaries", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *kappControllerPackagesServiceClient) GetAvailablePackageDetail(ctx context.Context, in *v1alpha1.GetAvailablePackageDetailRequest, opts ...grpc.CallOption) (*v1alpha1.GetAvailablePackageDetailResponse, error) {
	out := new(v1alpha1.GetAvailablePackageDetailResponse)
	err := c.cc.Invoke(ctx, "/kubeappsapis.plugins.kapp_controller.packages.v1alpha1.KappControllerPackagesService/GetAvailablePackageDetail", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *kappControllerPackagesServiceClient) GetPackageRepositories(ctx context.Context, in *GetPackageRepositoriesRequest, opts ...grpc.CallOption) (*GetPackageRepositoriesResponse, error) {
	out := new(GetPackageRepositoriesResponse)
	err := c.cc.Invoke(ctx, "/kubeappsapis.plugins.kapp_controller.packages.v1alpha1.KappControllerPackagesService/GetPackageRepositories", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *kappControllerPackagesServiceClient) GetAvailablePackageVersions(ctx context.Context, in *v1alpha1.GetAvailablePackageVersionsRequest, opts ...grpc.CallOption) (*v1alpha1.GetAvailablePackageVersionsResponse, error) {
	out := new(v1alpha1.GetAvailablePackageVersionsResponse)
	err := c.cc.Invoke(ctx, "/kubeappsapis.plugins.kapp_controller.packages.v1alpha1.KappControllerPackagesService/GetAvailablePackageVersions", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *kappControllerPackagesServiceClient) GetInstalledPackageSummaries(ctx context.Context, in *v1alpha1.GetInstalledPackageSummariesRequest, opts ...grpc.CallOption) (*v1alpha1.GetInstalledPackageSummariesResponse, error) {
	out := new(v1alpha1.GetInstalledPackageSummariesResponse)
	err := c.cc.Invoke(ctx, "/kubeappsapis.plugins.kapp_controller.packages.v1alpha1.KappControllerPackagesService/GetInstalledPackageSummaries", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *kappControllerPackagesServiceClient) GetInstalledPackageDetail(ctx context.Context, in *v1alpha1.GetInstalledPackageDetailRequest, opts ...grpc.CallOption) (*v1alpha1.GetInstalledPackageDetailResponse, error) {
	out := new(v1alpha1.GetInstalledPackageDetailResponse)
	err := c.cc.Invoke(ctx, "/kubeappsapis.plugins.kapp_controller.packages.v1alpha1.KappControllerPackagesService/GetInstalledPackageDetail", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *kappControllerPackagesServiceClient) CreateInstalledPackage(ctx context.Context, in *v1alpha1.CreateInstalledPackageRequest, opts ...grpc.CallOption) (*v1alpha1.CreateInstalledPackageResponse, error) {
	out := new(v1alpha1.CreateInstalledPackageResponse)
	err := c.cc.Invoke(ctx, "/kubeappsapis.plugins.kapp_controller.packages.v1alpha1.KappControllerPackagesService/CreateInstalledPackage", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *kappControllerPackagesServiceClient) UpdateInstalledPackage(ctx context.Context, in *v1alpha1.UpdateInstalledPackageRequest, opts ...grpc.CallOption) (*v1alpha1.UpdateInstalledPackageResponse, error) {
	out := new(v1alpha1.UpdateInstalledPackageResponse)
	err := c.cc.Invoke(ctx, "/kubeappsapis.plugins.kapp_controller.packages.v1alpha1.KappControllerPackagesService/UpdateInstalledPackage", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *kappControllerPackagesServiceClient) DeleteInstalledPackage(ctx context.Context, in *v1alpha1.DeleteInstalledPackageRequest, opts ...grpc.CallOption) (*v1alpha1.DeleteInstalledPackageResponse, error) {
	out := new(v1alpha1.DeleteInstalledPackageResponse)
	err := c.cc.Invoke(ctx, "/kubeappsapis.plugins.kapp_controller.packages.v1alpha1.KappControllerPackagesService/DeleteInstalledPackage", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *kappControllerPackagesServiceClient) GetInstalledPackageResourceRefs(ctx context.Context, in *v1alpha1.GetInstalledPackageResourceRefsRequest, opts ...grpc.CallOption) (*v1alpha1.GetInstalledPackageResourceRefsResponse, error) {
	out := new(v1alpha1.GetInstalledPackageResourceRefsResponse)
	err := c.cc.Invoke(ctx, "/kubeappsapis.plugins.kapp_controller.packages.v1alpha1.KappControllerPackagesService/GetInstalledPackageResourceRefs", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// KappControllerPackagesServiceServer is the server API for KappControllerPackagesService service.
// All implementations should embed UnimplementedKappControllerPackagesServiceServer
// for forward compatibility
type KappControllerPackagesServiceServer interface {
	// GetAvailablePackageSummaries returns the available packages managed by the 'kapp_controller' plugin
	GetAvailablePackageSummaries(context.Context, *v1alpha1.GetAvailablePackageSummariesRequest) (*v1alpha1.GetAvailablePackageSummariesResponse, error)
	// GetAvailablePackageDetail returns the package details managed by the 'kapp_controller' plugin
	GetAvailablePackageDetail(context.Context, *v1alpha1.GetAvailablePackageDetailRequest) (*v1alpha1.GetAvailablePackageDetailResponse, error)
	// GetPackageRepositories returns the repositories managed by the 'kapp_controller' plugin
	GetPackageRepositories(context.Context, *GetPackageRepositoriesRequest) (*GetPackageRepositoriesResponse, error)
	// GetAvailablePackageVersions returns the package versions managed by the 'kapp_controller' plugin
	GetAvailablePackageVersions(context.Context, *v1alpha1.GetAvailablePackageVersionsRequest) (*v1alpha1.GetAvailablePackageVersionsResponse, error)
	// GetInstalledPackageSummaries returns the installed packages managed by the 'kapp_controller' plugin
	GetInstalledPackageSummaries(context.Context, *v1alpha1.GetInstalledPackageSummariesRequest) (*v1alpha1.GetInstalledPackageSummariesResponse, error)
	// GetInstalledPackageDetail returns the requested installed package managed by the 'kapp_controller' plugin
	GetInstalledPackageDetail(context.Context, *v1alpha1.GetInstalledPackageDetailRequest) (*v1alpha1.GetInstalledPackageDetailResponse, error)
	// CreateInstalledPackage creates an installed package based on the request.
	CreateInstalledPackage(context.Context, *v1alpha1.CreateInstalledPackageRequest) (*v1alpha1.CreateInstalledPackageResponse, error)
	// UpdateInstalledPackage updates an installed package based on the request.
	UpdateInstalledPackage(context.Context, *v1alpha1.UpdateInstalledPackageRequest) (*v1alpha1.UpdateInstalledPackageResponse, error)
	// DeleteInstalledPackage deletes an installed package based on the request.
	DeleteInstalledPackage(context.Context, *v1alpha1.DeleteInstalledPackageRequest) (*v1alpha1.DeleteInstalledPackageResponse, error)
	// GetInstalledPackageResourceRefs returns the references for the Kubernetes resources created by
	// an installed package.
	GetInstalledPackageResourceRefs(context.Context, *v1alpha1.GetInstalledPackageResourceRefsRequest) (*v1alpha1.GetInstalledPackageResourceRefsResponse, error)
}

// UnimplementedKappControllerPackagesServiceServer should be embedded to have forward compatible implementations.
type UnimplementedKappControllerPackagesServiceServer struct {
}

func (UnimplementedKappControllerPackagesServiceServer) GetAvailablePackageSummaries(context.Context, *v1alpha1.GetAvailablePackageSummariesRequest) (*v1alpha1.GetAvailablePackageSummariesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetAvailablePackageSummaries not implemented")
}
func (UnimplementedKappControllerPackagesServiceServer) GetAvailablePackageDetail(context.Context, *v1alpha1.GetAvailablePackageDetailRequest) (*v1alpha1.GetAvailablePackageDetailResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetAvailablePackageDetail not implemented")
}
func (UnimplementedKappControllerPackagesServiceServer) GetPackageRepositories(context.Context, *GetPackageRepositoriesRequest) (*GetPackageRepositoriesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetPackageRepositories not implemented")
}
func (UnimplementedKappControllerPackagesServiceServer) GetAvailablePackageVersions(context.Context, *v1alpha1.GetAvailablePackageVersionsRequest) (*v1alpha1.GetAvailablePackageVersionsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetAvailablePackageVersions not implemented")
}
func (UnimplementedKappControllerPackagesServiceServer) GetInstalledPackageSummaries(context.Context, *v1alpha1.GetInstalledPackageSummariesRequest) (*v1alpha1.GetInstalledPackageSummariesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetInstalledPackageSummaries not implemented")
}
func (UnimplementedKappControllerPackagesServiceServer) GetInstalledPackageDetail(context.Context, *v1alpha1.GetInstalledPackageDetailRequest) (*v1alpha1.GetInstalledPackageDetailResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetInstalledPackageDetail not implemented")
}
func (UnimplementedKappControllerPackagesServiceServer) CreateInstalledPackage(context.Context, *v1alpha1.CreateInstalledPackageRequest) (*v1alpha1.CreateInstalledPackageResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateInstalledPackage not implemented")
}
func (UnimplementedKappControllerPackagesServiceServer) UpdateInstalledPackage(context.Context, *v1alpha1.UpdateInstalledPackageRequest) (*v1alpha1.UpdateInstalledPackageResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateInstalledPackage not implemented")
}
func (UnimplementedKappControllerPackagesServiceServer) DeleteInstalledPackage(context.Context, *v1alpha1.DeleteInstalledPackageRequest) (*v1alpha1.DeleteInstalledPackageResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteInstalledPackage not implemented")
}
func (UnimplementedKappControllerPackagesServiceServer) GetInstalledPackageResourceRefs(context.Context, *v1alpha1.GetInstalledPackageResourceRefsRequest) (*v1alpha1.GetInstalledPackageResourceRefsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetInstalledPackageResourceRefs not implemented")
}

// UnsafeKappControllerPackagesServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to KappControllerPackagesServiceServer will
// result in compilation errors.
type UnsafeKappControllerPackagesServiceServer interface {
	mustEmbedUnimplementedKappControllerPackagesServiceServer()
}

func RegisterKappControllerPackagesServiceServer(s grpc.ServiceRegistrar, srv KappControllerPackagesServiceServer) {
	s.RegisterService(&KappControllerPackagesService_ServiceDesc, srv)
}

func _KappControllerPackagesService_GetAvailablePackageSummaries_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(v1alpha1.GetAvailablePackageSummariesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(KappControllerPackagesServiceServer).GetAvailablePackageSummaries(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/kubeappsapis.plugins.kapp_controller.packages.v1alpha1.KappControllerPackagesService/GetAvailablePackageSummaries",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(KappControllerPackagesServiceServer).GetAvailablePackageSummaries(ctx, req.(*v1alpha1.GetAvailablePackageSummariesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _KappControllerPackagesService_GetAvailablePackageDetail_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(v1alpha1.GetAvailablePackageDetailRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(KappControllerPackagesServiceServer).GetAvailablePackageDetail(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/kubeappsapis.plugins.kapp_controller.packages.v1alpha1.KappControllerPackagesService/GetAvailablePackageDetail",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(KappControllerPackagesServiceServer).GetAvailablePackageDetail(ctx, req.(*v1alpha1.GetAvailablePackageDetailRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _KappControllerPackagesService_GetPackageRepositories_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetPackageRepositoriesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(KappControllerPackagesServiceServer).GetPackageRepositories(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/kubeappsapis.plugins.kapp_controller.packages.v1alpha1.KappControllerPackagesService/GetPackageRepositories",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(KappControllerPackagesServiceServer).GetPackageRepositories(ctx, req.(*GetPackageRepositoriesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _KappControllerPackagesService_GetAvailablePackageVersions_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(v1alpha1.GetAvailablePackageVersionsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(KappControllerPackagesServiceServer).GetAvailablePackageVersions(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/kubeappsapis.plugins.kapp_controller.packages.v1alpha1.KappControllerPackagesService/GetAvailablePackageVersions",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(KappControllerPackagesServiceServer).GetAvailablePackageVersions(ctx, req.(*v1alpha1.GetAvailablePackageVersionsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _KappControllerPackagesService_GetInstalledPackageSummaries_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(v1alpha1.GetInstalledPackageSummariesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(KappControllerPackagesServiceServer).GetInstalledPackageSummaries(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/kubeappsapis.plugins.kapp_controller.packages.v1alpha1.KappControllerPackagesService/GetInstalledPackageSummaries",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(KappControllerPackagesServiceServer).GetInstalledPackageSummaries(ctx, req.(*v1alpha1.GetInstalledPackageSummariesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _KappControllerPackagesService_GetInstalledPackageDetail_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(v1alpha1.GetInstalledPackageDetailRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(KappControllerPackagesServiceServer).GetInstalledPackageDetail(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/kubeappsapis.plugins.kapp_controller.packages.v1alpha1.KappControllerPackagesService/GetInstalledPackageDetail",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(KappControllerPackagesServiceServer).GetInstalledPackageDetail(ctx, req.(*v1alpha1.GetInstalledPackageDetailRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _KappControllerPackagesService_CreateInstalledPackage_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(v1alpha1.CreateInstalledPackageRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(KappControllerPackagesServiceServer).CreateInstalledPackage(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/kubeappsapis.plugins.kapp_controller.packages.v1alpha1.KappControllerPackagesService/CreateInstalledPackage",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(KappControllerPackagesServiceServer).CreateInstalledPackage(ctx, req.(*v1alpha1.CreateInstalledPackageRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _KappControllerPackagesService_UpdateInstalledPackage_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(v1alpha1.UpdateInstalledPackageRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(KappControllerPackagesServiceServer).UpdateInstalledPackage(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/kubeappsapis.plugins.kapp_controller.packages.v1alpha1.KappControllerPackagesService/UpdateInstalledPackage",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(KappControllerPackagesServiceServer).UpdateInstalledPackage(ctx, req.(*v1alpha1.UpdateInstalledPackageRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _KappControllerPackagesService_DeleteInstalledPackage_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(v1alpha1.DeleteInstalledPackageRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(KappControllerPackagesServiceServer).DeleteInstalledPackage(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/kubeappsapis.plugins.kapp_controller.packages.v1alpha1.KappControllerPackagesService/DeleteInstalledPackage",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(KappControllerPackagesServiceServer).DeleteInstalledPackage(ctx, req.(*v1alpha1.DeleteInstalledPackageRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _KappControllerPackagesService_GetInstalledPackageResourceRefs_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(v1alpha1.GetInstalledPackageResourceRefsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(KappControllerPackagesServiceServer).GetInstalledPackageResourceRefs(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/kubeappsapis.plugins.kapp_controller.packages.v1alpha1.KappControllerPackagesService/GetInstalledPackageResourceRefs",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(KappControllerPackagesServiceServer).GetInstalledPackageResourceRefs(ctx, req.(*v1alpha1.GetInstalledPackageResourceRefsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// KappControllerPackagesService_ServiceDesc is the grpc.ServiceDesc for KappControllerPackagesService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var KappControllerPackagesService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "kubeappsapis.plugins.kapp_controller.packages.v1alpha1.KappControllerPackagesService",
	HandlerType: (*KappControllerPackagesServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetAvailablePackageSummaries",
			Handler:    _KappControllerPackagesService_GetAvailablePackageSummaries_Handler,
		},
		{
			MethodName: "GetAvailablePackageDetail",
			Handler:    _KappControllerPackagesService_GetAvailablePackageDetail_Handler,
		},
		{
			MethodName: "GetPackageRepositories",
			Handler:    _KappControllerPackagesService_GetPackageRepositories_Handler,
		},
		{
			MethodName: "GetAvailablePackageVersions",
			Handler:    _KappControllerPackagesService_GetAvailablePackageVersions_Handler,
		},
		{
			MethodName: "GetInstalledPackageSummaries",
			Handler:    _KappControllerPackagesService_GetInstalledPackageSummaries_Handler,
		},
		{
			MethodName: "GetInstalledPackageDetail",
			Handler:    _KappControllerPackagesService_GetInstalledPackageDetail_Handler,
		},
		{
			MethodName: "CreateInstalledPackage",
			Handler:    _KappControllerPackagesService_CreateInstalledPackage_Handler,
		},
		{
			MethodName: "UpdateInstalledPackage",
			Handler:    _KappControllerPackagesService_UpdateInstalledPackage_Handler,
		},
		{
			MethodName: "DeleteInstalledPackage",
			Handler:    _KappControllerPackagesService_DeleteInstalledPackage_Handler,
		},
		{
			MethodName: "GetInstalledPackageResourceRefs",
			Handler:    _KappControllerPackagesService_GetInstalledPackageResourceRefs_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "kubeappsapis/plugins/kapp_controller/packages/v1alpha1/kapp_controller.proto",
}
