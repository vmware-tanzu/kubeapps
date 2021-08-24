// Code generated by protoc-gen-grpc-gateway. DO NOT EDIT.
// source: kubeappsapis/plugins/kapp_controller/packages/v1alpha1/kapp_controller.proto

/*
Package v1alpha1 is a reverse proxy.

It translates gRPC into RESTful JSON APIs.
*/
package v1alpha1

import (
	"context"
	"io"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/grpc-ecosystem/grpc-gateway/v2/utilities"
	"github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

// Suppress "imported and not used" errors
var _ codes.Code
var _ io.Reader
var _ status.Status
var _ = runtime.String
var _ = utilities.NewDoubleArray
var _ = metadata.Join

var (
	filter_KappControllerPackagesService_GetAvailablePackageSummaries_0 = &utilities.DoubleArray{Encoding: map[string]int{}, Base: []int(nil), Check: []int(nil)}
)

func request_KappControllerPackagesService_GetAvailablePackageSummaries_0(ctx context.Context, marshaler runtime.Marshaler, client KappControllerPackagesServiceClient, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	var protoReq v1alpha1.GetAvailablePackageSummariesRequest
	var metadata runtime.ServerMetadata

	if err := req.ParseForm(); err != nil {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	}
	if err := runtime.PopulateQueryParameters(&protoReq, req.Form, filter_KappControllerPackagesService_GetAvailablePackageSummaries_0); err != nil {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	}

	msg, err := client.GetAvailablePackageSummaries(ctx, &protoReq, grpc.Header(&metadata.HeaderMD), grpc.Trailer(&metadata.TrailerMD))
	return msg, metadata, err

}

func local_request_KappControllerPackagesService_GetAvailablePackageSummaries_0(ctx context.Context, marshaler runtime.Marshaler, server KappControllerPackagesServiceServer, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	var protoReq v1alpha1.GetAvailablePackageSummariesRequest
	var metadata runtime.ServerMetadata

	if err := req.ParseForm(); err != nil {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	}
	if err := runtime.PopulateQueryParameters(&protoReq, req.Form, filter_KappControllerPackagesService_GetAvailablePackageSummaries_0); err != nil {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	}

	msg, err := server.GetAvailablePackageSummaries(ctx, &protoReq)
	return msg, metadata, err

}

var (
	filter_KappControllerPackagesService_GetAvailablePackageDetail_0 = &utilities.DoubleArray{Encoding: map[string]int{}, Base: []int(nil), Check: []int(nil)}
)

func request_KappControllerPackagesService_GetAvailablePackageDetail_0(ctx context.Context, marshaler runtime.Marshaler, client KappControllerPackagesServiceClient, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	var protoReq v1alpha1.GetAvailablePackageDetailRequest
	var metadata runtime.ServerMetadata

	if err := req.ParseForm(); err != nil {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	}
	if err := runtime.PopulateQueryParameters(&protoReq, req.Form, filter_KappControllerPackagesService_GetAvailablePackageDetail_0); err != nil {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	}

	msg, err := client.GetAvailablePackageDetail(ctx, &protoReq, grpc.Header(&metadata.HeaderMD), grpc.Trailer(&metadata.TrailerMD))
	return msg, metadata, err

}

func local_request_KappControllerPackagesService_GetAvailablePackageDetail_0(ctx context.Context, marshaler runtime.Marshaler, server KappControllerPackagesServiceServer, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	var protoReq v1alpha1.GetAvailablePackageDetailRequest
	var metadata runtime.ServerMetadata

	if err := req.ParseForm(); err != nil {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	}
	if err := runtime.PopulateQueryParameters(&protoReq, req.Form, filter_KappControllerPackagesService_GetAvailablePackageDetail_0); err != nil {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	}

	msg, err := server.GetAvailablePackageDetail(ctx, &protoReq)
	return msg, metadata, err

}

var (
	filter_KappControllerPackagesService_GetPackageRepositories_0 = &utilities.DoubleArray{Encoding: map[string]int{}, Base: []int(nil), Check: []int(nil)}
)

func request_KappControllerPackagesService_GetPackageRepositories_0(ctx context.Context, marshaler runtime.Marshaler, client KappControllerPackagesServiceClient, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	var protoReq GetPackageRepositoriesRequest
	var metadata runtime.ServerMetadata

	if err := req.ParseForm(); err != nil {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	}
	if err := runtime.PopulateQueryParameters(&protoReq, req.Form, filter_KappControllerPackagesService_GetPackageRepositories_0); err != nil {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	}

	msg, err := client.GetPackageRepositories(ctx, &protoReq, grpc.Header(&metadata.HeaderMD), grpc.Trailer(&metadata.TrailerMD))
	return msg, metadata, err

}

func local_request_KappControllerPackagesService_GetPackageRepositories_0(ctx context.Context, marshaler runtime.Marshaler, server KappControllerPackagesServiceServer, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	var protoReq GetPackageRepositoriesRequest
	var metadata runtime.ServerMetadata

	if err := req.ParseForm(); err != nil {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	}
	if err := runtime.PopulateQueryParameters(&protoReq, req.Form, filter_KappControllerPackagesService_GetPackageRepositories_0); err != nil {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	}

	msg, err := server.GetPackageRepositories(ctx, &protoReq)
	return msg, metadata, err

}

var (
	filter_KappControllerPackagesService_GetAvailablePackageVersions_0 = &utilities.DoubleArray{Encoding: map[string]int{}, Base: []int(nil), Check: []int(nil)}
)

func request_KappControllerPackagesService_GetAvailablePackageVersions_0(ctx context.Context, marshaler runtime.Marshaler, client KappControllerPackagesServiceClient, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	var protoReq v1alpha1.GetAvailablePackageVersionsRequest
	var metadata runtime.ServerMetadata

	if err := req.ParseForm(); err != nil {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	}
	if err := runtime.PopulateQueryParameters(&protoReq, req.Form, filter_KappControllerPackagesService_GetAvailablePackageVersions_0); err != nil {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	}

	msg, err := client.GetAvailablePackageVersions(ctx, &protoReq, grpc.Header(&metadata.HeaderMD), grpc.Trailer(&metadata.TrailerMD))
	return msg, metadata, err

}

func local_request_KappControllerPackagesService_GetAvailablePackageVersions_0(ctx context.Context, marshaler runtime.Marshaler, server KappControllerPackagesServiceServer, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	var protoReq v1alpha1.GetAvailablePackageVersionsRequest
	var metadata runtime.ServerMetadata

	if err := req.ParseForm(); err != nil {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	}
	if err := runtime.PopulateQueryParameters(&protoReq, req.Form, filter_KappControllerPackagesService_GetAvailablePackageVersions_0); err != nil {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	}

	msg, err := server.GetAvailablePackageVersions(ctx, &protoReq)
	return msg, metadata, err

}

var (
	filter_KappControllerPackagesService_GetInstalledPackageSummaries_0 = &utilities.DoubleArray{Encoding: map[string]int{}, Base: []int(nil), Check: []int(nil)}
)

func request_KappControllerPackagesService_GetInstalledPackageSummaries_0(ctx context.Context, marshaler runtime.Marshaler, client KappControllerPackagesServiceClient, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	var protoReq v1alpha1.GetInstalledPackageSummariesRequest
	var metadata runtime.ServerMetadata

	if err := req.ParseForm(); err != nil {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	}
	if err := runtime.PopulateQueryParameters(&protoReq, req.Form, filter_KappControllerPackagesService_GetInstalledPackageSummaries_0); err != nil {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	}

	msg, err := client.GetInstalledPackageSummaries(ctx, &protoReq, grpc.Header(&metadata.HeaderMD), grpc.Trailer(&metadata.TrailerMD))
	return msg, metadata, err

}

func local_request_KappControllerPackagesService_GetInstalledPackageSummaries_0(ctx context.Context, marshaler runtime.Marshaler, server KappControllerPackagesServiceServer, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	var protoReq v1alpha1.GetInstalledPackageSummariesRequest
	var metadata runtime.ServerMetadata

	if err := req.ParseForm(); err != nil {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	}
	if err := runtime.PopulateQueryParameters(&protoReq, req.Form, filter_KappControllerPackagesService_GetInstalledPackageSummaries_0); err != nil {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	}

	msg, err := server.GetInstalledPackageSummaries(ctx, &protoReq)
	return msg, metadata, err

}

var (
	filter_KappControllerPackagesService_GetInstalledPackageDetail_0 = &utilities.DoubleArray{Encoding: map[string]int{}, Base: []int(nil), Check: []int(nil)}
)

func request_KappControllerPackagesService_GetInstalledPackageDetail_0(ctx context.Context, marshaler runtime.Marshaler, client KappControllerPackagesServiceClient, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	var protoReq v1alpha1.GetInstalledPackageDetailRequest
	var metadata runtime.ServerMetadata

	if err := req.ParseForm(); err != nil {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	}
	if err := runtime.PopulateQueryParameters(&protoReq, req.Form, filter_KappControllerPackagesService_GetInstalledPackageDetail_0); err != nil {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	}

	msg, err := client.GetInstalledPackageDetail(ctx, &protoReq, grpc.Header(&metadata.HeaderMD), grpc.Trailer(&metadata.TrailerMD))
	return msg, metadata, err

}

func local_request_KappControllerPackagesService_GetInstalledPackageDetail_0(ctx context.Context, marshaler runtime.Marshaler, server KappControllerPackagesServiceServer, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	var protoReq v1alpha1.GetInstalledPackageDetailRequest
	var metadata runtime.ServerMetadata

	if err := req.ParseForm(); err != nil {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	}
	if err := runtime.PopulateQueryParameters(&protoReq, req.Form, filter_KappControllerPackagesService_GetInstalledPackageDetail_0); err != nil {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	}

	msg, err := server.GetInstalledPackageDetail(ctx, &protoReq)
	return msg, metadata, err

}

var (
	filter_KappControllerPackagesService_CreateInstalledPackage_0 = &utilities.DoubleArray{Encoding: map[string]int{}, Base: []int(nil), Check: []int(nil)}
)

func request_KappControllerPackagesService_CreateInstalledPackage_0(ctx context.Context, marshaler runtime.Marshaler, client KappControllerPackagesServiceClient, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	var protoReq v1alpha1.CreateInstalledPackageRequest
	var metadata runtime.ServerMetadata

	if err := req.ParseForm(); err != nil {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	}
	if err := runtime.PopulateQueryParameters(&protoReq, req.Form, filter_KappControllerPackagesService_CreateInstalledPackage_0); err != nil {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	}

	msg, err := client.CreateInstalledPackage(ctx, &protoReq, grpc.Header(&metadata.HeaderMD), grpc.Trailer(&metadata.TrailerMD))
	return msg, metadata, err

}

func local_request_KappControllerPackagesService_CreateInstalledPackage_0(ctx context.Context, marshaler runtime.Marshaler, server KappControllerPackagesServiceServer, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	var protoReq v1alpha1.CreateInstalledPackageRequest
	var metadata runtime.ServerMetadata

	if err := req.ParseForm(); err != nil {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	}
	if err := runtime.PopulateQueryParameters(&protoReq, req.Form, filter_KappControllerPackagesService_CreateInstalledPackage_0); err != nil {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	}

	msg, err := server.CreateInstalledPackage(ctx, &protoReq)
	return msg, metadata, err

}

// RegisterKappControllerPackagesServiceHandlerServer registers the http handlers for service KappControllerPackagesService to "mux".
// UnaryRPC     :call KappControllerPackagesServiceServer directly.
// StreamingRPC :currently unsupported pending https://github.com/grpc/grpc-go/issues/906.
// Note that using this registration option will cause many gRPC library features to stop working. Consider using RegisterKappControllerPackagesServiceHandlerFromEndpoint instead.
func RegisterKappControllerPackagesServiceHandlerServer(ctx context.Context, mux *runtime.ServeMux, server KappControllerPackagesServiceServer) error {

	mux.Handle("GET", pattern_KappControllerPackagesService_GetAvailablePackageSummaries_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		var stream runtime.ServerTransportStream
		ctx = grpc.NewContextWithServerTransportStream(ctx, &stream)
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateIncomingContext(ctx, mux, req, "/kubeappsapis.plugins.kapp_controller.packages.v1alpha1.KappControllerPackagesService/GetAvailablePackageSummaries", runtime.WithHTTPPathPattern("/plugins/kapp_controller/packages/v1alpha1/availablepackagesummaries"))
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}
		resp, md, err := local_request_KappControllerPackagesService_GetAvailablePackageSummaries_0(rctx, inboundMarshaler, server, req, pathParams)
		md.HeaderMD, md.TrailerMD = metadata.Join(md.HeaderMD, stream.Header()), metadata.Join(md.TrailerMD, stream.Trailer())
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}

		forward_KappControllerPackagesService_GetAvailablePackageSummaries_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})

	mux.Handle("GET", pattern_KappControllerPackagesService_GetAvailablePackageDetail_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		var stream runtime.ServerTransportStream
		ctx = grpc.NewContextWithServerTransportStream(ctx, &stream)
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateIncomingContext(ctx, mux, req, "/kubeappsapis.plugins.kapp_controller.packages.v1alpha1.KappControllerPackagesService/GetAvailablePackageDetail", runtime.WithHTTPPathPattern("/plugins/kapp_controller/packages/v1alpha1/availablepackagedetails"))
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}
		resp, md, err := local_request_KappControllerPackagesService_GetAvailablePackageDetail_0(rctx, inboundMarshaler, server, req, pathParams)
		md.HeaderMD, md.TrailerMD = metadata.Join(md.HeaderMD, stream.Header()), metadata.Join(md.TrailerMD, stream.Trailer())
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}

		forward_KappControllerPackagesService_GetAvailablePackageDetail_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})

	mux.Handle("GET", pattern_KappControllerPackagesService_GetPackageRepositories_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		var stream runtime.ServerTransportStream
		ctx = grpc.NewContextWithServerTransportStream(ctx, &stream)
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateIncomingContext(ctx, mux, req, "/kubeappsapis.plugins.kapp_controller.packages.v1alpha1.KappControllerPackagesService/GetPackageRepositories", runtime.WithHTTPPathPattern("/plugins/kapp_controller/packages/v1alpha1/packagerepositories"))
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}
		resp, md, err := local_request_KappControllerPackagesService_GetPackageRepositories_0(rctx, inboundMarshaler, server, req, pathParams)
		md.HeaderMD, md.TrailerMD = metadata.Join(md.HeaderMD, stream.Header()), metadata.Join(md.TrailerMD, stream.Trailer())
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}

		forward_KappControllerPackagesService_GetPackageRepositories_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})

	mux.Handle("GET", pattern_KappControllerPackagesService_GetAvailablePackageVersions_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		var stream runtime.ServerTransportStream
		ctx = grpc.NewContextWithServerTransportStream(ctx, &stream)
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateIncomingContext(ctx, mux, req, "/kubeappsapis.plugins.kapp_controller.packages.v1alpha1.KappControllerPackagesService/GetAvailablePackageVersions", runtime.WithHTTPPathPattern("/plugins/kapp_controller/packages/v1alpha1/availablepackageversions"))
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}
		resp, md, err := local_request_KappControllerPackagesService_GetAvailablePackageVersions_0(rctx, inboundMarshaler, server, req, pathParams)
		md.HeaderMD, md.TrailerMD = metadata.Join(md.HeaderMD, stream.Header()), metadata.Join(md.TrailerMD, stream.Trailer())
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}

		forward_KappControllerPackagesService_GetAvailablePackageVersions_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})

	mux.Handle("GET", pattern_KappControllerPackagesService_GetInstalledPackageSummaries_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		var stream runtime.ServerTransportStream
		ctx = grpc.NewContextWithServerTransportStream(ctx, &stream)
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateIncomingContext(ctx, mux, req, "/kubeappsapis.plugins.kapp_controller.packages.v1alpha1.KappControllerPackagesService/GetInstalledPackageSummaries", runtime.WithHTTPPathPattern("/plugins/kapp_controller/packages/v1alpha1/installedpackagesummaries"))
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}
		resp, md, err := local_request_KappControllerPackagesService_GetInstalledPackageSummaries_0(rctx, inboundMarshaler, server, req, pathParams)
		md.HeaderMD, md.TrailerMD = metadata.Join(md.HeaderMD, stream.Header()), metadata.Join(md.TrailerMD, stream.Trailer())
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}

		forward_KappControllerPackagesService_GetInstalledPackageSummaries_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})

	mux.Handle("GET", pattern_KappControllerPackagesService_GetInstalledPackageDetail_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		var stream runtime.ServerTransportStream
		ctx = grpc.NewContextWithServerTransportStream(ctx, &stream)
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateIncomingContext(ctx, mux, req, "/kubeappsapis.plugins.kapp_controller.packages.v1alpha1.KappControllerPackagesService/GetInstalledPackageDetail", runtime.WithHTTPPathPattern("/plugins/kapp_controller/packages/v1alpha1/installedpackagedetail"))
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}
		resp, md, err := local_request_KappControllerPackagesService_GetInstalledPackageDetail_0(rctx, inboundMarshaler, server, req, pathParams)
		md.HeaderMD, md.TrailerMD = metadata.Join(md.HeaderMD, stream.Header()), metadata.Join(md.TrailerMD, stream.Trailer())
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}

		forward_KappControllerPackagesService_GetInstalledPackageDetail_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})

	mux.Handle("POST", pattern_KappControllerPackagesService_CreateInstalledPackage_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		var stream runtime.ServerTransportStream
		ctx = grpc.NewContextWithServerTransportStream(ctx, &stream)
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateIncomingContext(ctx, mux, req, "/kubeappsapis.plugins.kapp_controller.packages.v1alpha1.KappControllerPackagesService/CreateInstalledPackage", runtime.WithHTTPPathPattern("/plugins/kapp_controller/packages/v1alpha1/installedpackages"))
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}
		resp, md, err := local_request_KappControllerPackagesService_CreateInstalledPackage_0(rctx, inboundMarshaler, server, req, pathParams)
		md.HeaderMD, md.TrailerMD = metadata.Join(md.HeaderMD, stream.Header()), metadata.Join(md.TrailerMD, stream.Trailer())
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}

		forward_KappControllerPackagesService_CreateInstalledPackage_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})

	return nil
}

// RegisterKappControllerPackagesServiceHandlerFromEndpoint is same as RegisterKappControllerPackagesServiceHandler but
// automatically dials to "endpoint" and closes the connection when "ctx" gets done.
func RegisterKappControllerPackagesServiceHandlerFromEndpoint(ctx context.Context, mux *runtime.ServeMux, endpoint string, opts []grpc.DialOption) (err error) {
	conn, err := grpc.Dial(endpoint, opts...)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			if cerr := conn.Close(); cerr != nil {
				grpclog.Infof("Failed to close conn to %s: %v", endpoint, cerr)
			}
			return
		}
		go func() {
			<-ctx.Done()
			if cerr := conn.Close(); cerr != nil {
				grpclog.Infof("Failed to close conn to %s: %v", endpoint, cerr)
			}
		}()
	}()

	return RegisterKappControllerPackagesServiceHandler(ctx, mux, conn)
}

// RegisterKappControllerPackagesServiceHandler registers the http handlers for service KappControllerPackagesService to "mux".
// The handlers forward requests to the grpc endpoint over "conn".
func RegisterKappControllerPackagesServiceHandler(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error {
	return RegisterKappControllerPackagesServiceHandlerClient(ctx, mux, NewKappControllerPackagesServiceClient(conn))
}

// RegisterKappControllerPackagesServiceHandlerClient registers the http handlers for service KappControllerPackagesService
// to "mux". The handlers forward requests to the grpc endpoint over the given implementation of "KappControllerPackagesServiceClient".
// Note: the gRPC framework executes interceptors within the gRPC handler. If the passed in "KappControllerPackagesServiceClient"
// doesn't go through the normal gRPC flow (creating a gRPC client etc.) then it will be up to the passed in
// "KappControllerPackagesServiceClient" to call the correct interceptors.
func RegisterKappControllerPackagesServiceHandlerClient(ctx context.Context, mux *runtime.ServeMux, client KappControllerPackagesServiceClient) error {

	mux.Handle("GET", pattern_KappControllerPackagesService_GetAvailablePackageSummaries_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateContext(ctx, mux, req, "/kubeappsapis.plugins.kapp_controller.packages.v1alpha1.KappControllerPackagesService/GetAvailablePackageSummaries", runtime.WithHTTPPathPattern("/plugins/kapp_controller/packages/v1alpha1/availablepackagesummaries"))
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}
		resp, md, err := request_KappControllerPackagesService_GetAvailablePackageSummaries_0(rctx, inboundMarshaler, client, req, pathParams)
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}

		forward_KappControllerPackagesService_GetAvailablePackageSummaries_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})

	mux.Handle("GET", pattern_KappControllerPackagesService_GetAvailablePackageDetail_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateContext(ctx, mux, req, "/kubeappsapis.plugins.kapp_controller.packages.v1alpha1.KappControllerPackagesService/GetAvailablePackageDetail", runtime.WithHTTPPathPattern("/plugins/kapp_controller/packages/v1alpha1/availablepackagedetails"))
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}
		resp, md, err := request_KappControllerPackagesService_GetAvailablePackageDetail_0(rctx, inboundMarshaler, client, req, pathParams)
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}

		forward_KappControllerPackagesService_GetAvailablePackageDetail_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})

	mux.Handle("GET", pattern_KappControllerPackagesService_GetPackageRepositories_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateContext(ctx, mux, req, "/kubeappsapis.plugins.kapp_controller.packages.v1alpha1.KappControllerPackagesService/GetPackageRepositories", runtime.WithHTTPPathPattern("/plugins/kapp_controller/packages/v1alpha1/packagerepositories"))
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}
		resp, md, err := request_KappControllerPackagesService_GetPackageRepositories_0(rctx, inboundMarshaler, client, req, pathParams)
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}

		forward_KappControllerPackagesService_GetPackageRepositories_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})

	mux.Handle("GET", pattern_KappControllerPackagesService_GetAvailablePackageVersions_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateContext(ctx, mux, req, "/kubeappsapis.plugins.kapp_controller.packages.v1alpha1.KappControllerPackagesService/GetAvailablePackageVersions", runtime.WithHTTPPathPattern("/plugins/kapp_controller/packages/v1alpha1/availablepackageversions"))
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}
		resp, md, err := request_KappControllerPackagesService_GetAvailablePackageVersions_0(rctx, inboundMarshaler, client, req, pathParams)
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}

		forward_KappControllerPackagesService_GetAvailablePackageVersions_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})

	mux.Handle("GET", pattern_KappControllerPackagesService_GetInstalledPackageSummaries_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateContext(ctx, mux, req, "/kubeappsapis.plugins.kapp_controller.packages.v1alpha1.KappControllerPackagesService/GetInstalledPackageSummaries", runtime.WithHTTPPathPattern("/plugins/kapp_controller/packages/v1alpha1/installedpackagesummaries"))
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}
		resp, md, err := request_KappControllerPackagesService_GetInstalledPackageSummaries_0(rctx, inboundMarshaler, client, req, pathParams)
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}

		forward_KappControllerPackagesService_GetInstalledPackageSummaries_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})

	mux.Handle("GET", pattern_KappControllerPackagesService_GetInstalledPackageDetail_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateContext(ctx, mux, req, "/kubeappsapis.plugins.kapp_controller.packages.v1alpha1.KappControllerPackagesService/GetInstalledPackageDetail", runtime.WithHTTPPathPattern("/plugins/kapp_controller/packages/v1alpha1/installedpackagedetail"))
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}
		resp, md, err := request_KappControllerPackagesService_GetInstalledPackageDetail_0(rctx, inboundMarshaler, client, req, pathParams)
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}

		forward_KappControllerPackagesService_GetInstalledPackageDetail_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})

	mux.Handle("POST", pattern_KappControllerPackagesService_CreateInstalledPackage_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateContext(ctx, mux, req, "/kubeappsapis.plugins.kapp_controller.packages.v1alpha1.KappControllerPackagesService/CreateInstalledPackage", runtime.WithHTTPPathPattern("/plugins/kapp_controller/packages/v1alpha1/installedpackages"))
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}
		resp, md, err := request_KappControllerPackagesService_CreateInstalledPackage_0(rctx, inboundMarshaler, client, req, pathParams)
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}

		forward_KappControllerPackagesService_CreateInstalledPackage_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})

	return nil
}

var (
	pattern_KappControllerPackagesService_GetAvailablePackageSummaries_0 = runtime.MustPattern(runtime.NewPattern(1, []int{2, 0, 2, 1, 2, 2, 2, 3, 2, 4}, []string{"plugins", "kapp_controller", "packages", "v1alpha1", "availablepackagesummaries"}, ""))

	pattern_KappControllerPackagesService_GetAvailablePackageDetail_0 = runtime.MustPattern(runtime.NewPattern(1, []int{2, 0, 2, 1, 2, 2, 2, 3, 2, 4}, []string{"plugins", "kapp_controller", "packages", "v1alpha1", "availablepackagedetails"}, ""))

	pattern_KappControllerPackagesService_GetPackageRepositories_0 = runtime.MustPattern(runtime.NewPattern(1, []int{2, 0, 2, 1, 2, 2, 2, 3, 2, 4}, []string{"plugins", "kapp_controller", "packages", "v1alpha1", "packagerepositories"}, ""))

	pattern_KappControllerPackagesService_GetAvailablePackageVersions_0 = runtime.MustPattern(runtime.NewPattern(1, []int{2, 0, 2, 1, 2, 2, 2, 3, 2, 4}, []string{"plugins", "kapp_controller", "packages", "v1alpha1", "availablepackageversions"}, ""))

	pattern_KappControllerPackagesService_GetInstalledPackageSummaries_0 = runtime.MustPattern(runtime.NewPattern(1, []int{2, 0, 2, 1, 2, 2, 2, 3, 2, 4}, []string{"plugins", "kapp_controller", "packages", "v1alpha1", "installedpackagesummaries"}, ""))

	pattern_KappControllerPackagesService_GetInstalledPackageDetail_0 = runtime.MustPattern(runtime.NewPattern(1, []int{2, 0, 2, 1, 2, 2, 2, 3, 2, 4}, []string{"plugins", "kapp_controller", "packages", "v1alpha1", "installedpackagedetail"}, ""))

	pattern_KappControllerPackagesService_CreateInstalledPackage_0 = runtime.MustPattern(runtime.NewPattern(1, []int{2, 0, 2, 1, 2, 2, 2, 3, 2, 4}, []string{"plugins", "kapp_controller", "packages", "v1alpha1", "installedpackages"}, ""))
)

var (
	forward_KappControllerPackagesService_GetAvailablePackageSummaries_0 = runtime.ForwardResponseMessage

	forward_KappControllerPackagesService_GetAvailablePackageDetail_0 = runtime.ForwardResponseMessage

	forward_KappControllerPackagesService_GetPackageRepositories_0 = runtime.ForwardResponseMessage

	forward_KappControllerPackagesService_GetAvailablePackageVersions_0 = runtime.ForwardResponseMessage

	forward_KappControllerPackagesService_GetInstalledPackageSummaries_0 = runtime.ForwardResponseMessage

	forward_KappControllerPackagesService_GetInstalledPackageDetail_0 = runtime.ForwardResponseMessage

	forward_KappControllerPackagesService_CreateInstalledPackage_0 = runtime.ForwardResponseMessage
)
