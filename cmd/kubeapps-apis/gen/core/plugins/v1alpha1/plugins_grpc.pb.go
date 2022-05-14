// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

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

// PluginsServiceClient is the client API for PluginsService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type PluginsServiceClient interface {
	// GetConfiguredPlugins returns a map of short and longnames for the configured plugins.
	GetConfiguredPlugins(ctx context.Context, in *GetConfiguredPluginsRequest, opts ...grpc.CallOption) (*GetConfiguredPluginsResponse, error)
}

type pluginsServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewPluginsServiceClient(cc grpc.ClientConnInterface) PluginsServiceClient {
	return &pluginsServiceClient{cc}
}

func (c *pluginsServiceClient) GetConfiguredPlugins(ctx context.Context, in *GetConfiguredPluginsRequest, opts ...grpc.CallOption) (*GetConfiguredPluginsResponse, error) {
	out := new(GetConfiguredPluginsResponse)
	err := c.cc.Invoke(ctx, "/kubeappsapis.core.plugins.v1alpha1.PluginsService/GetConfiguredPlugins", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// PluginsServiceServer is the server API for PluginsService service.
// All implementations should embed UnimplementedPluginsServiceServer
// for forward compatibility
type PluginsServiceServer interface {
	// GetConfiguredPlugins returns a map of short and longnames for the configured plugins.
	GetConfiguredPlugins(context.Context, *GetConfiguredPluginsRequest) (*GetConfiguredPluginsResponse, error)
}

// UnimplementedPluginsServiceServer should be embedded to have forward compatible implementations.
type UnimplementedPluginsServiceServer struct {
}

func (UnimplementedPluginsServiceServer) GetConfiguredPlugins(context.Context, *GetConfiguredPluginsRequest) (*GetConfiguredPluginsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetConfiguredPlugins not implemented")
}

// UnsafePluginsServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to PluginsServiceServer will
// result in compilation errors.
type UnsafePluginsServiceServer interface {
	mustEmbedUnimplementedPluginsServiceServer()
}

func RegisterPluginsServiceServer(s grpc.ServiceRegistrar, srv PluginsServiceServer) {
	s.RegisterService(&PluginsService_ServiceDesc, srv)
}

func _PluginsService_GetConfiguredPlugins_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetConfiguredPluginsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PluginsServiceServer).GetConfiguredPlugins(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/kubeappsapis.core.plugins.v1alpha1.PluginsService/GetConfiguredPlugins",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PluginsServiceServer).GetConfiguredPlugins(ctx, req.(*GetConfiguredPluginsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// PluginsService_ServiceDesc is the grpc.ServiceDesc for PluginsService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var PluginsService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "kubeappsapis.core.plugins.v1alpha1.PluginsService",
	HandlerType: (*PluginsServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetConfiguredPlugins",
			Handler:    _PluginsService_GetConfiguredPlugins_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "kubeappsapis/core/plugins/v1alpha1/plugins.proto",
}
