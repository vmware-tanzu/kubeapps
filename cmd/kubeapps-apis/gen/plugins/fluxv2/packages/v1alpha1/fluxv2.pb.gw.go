// Code generated by protoc-gen-grpc-gateway. DO NOT EDIT.
// source: kubeappsapis/plugins/fluxv2/packages/v1alpha1/fluxv2.proto

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
	v1alpha1_0 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
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
	filter_FluxV2PackagesService_GetAvailablePackageSummaries_0 = &utilities.DoubleArray{Encoding: map[string]int{}, Base: []int(nil), Check: []int(nil)}
)

func request_FluxV2PackagesService_GetAvailablePackageSummaries_0(ctx context.Context, marshaler runtime.Marshaler, client FluxV2PackagesServiceClient, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	var protoReq v1alpha1_0.GetAvailablePackageSummariesRequest
	var metadata runtime.ServerMetadata

	if err := req.ParseForm(); err != nil {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	}
	if err := runtime.PopulateQueryParameters(&protoReq, req.Form, filter_FluxV2PackagesService_GetAvailablePackageSummaries_0); err != nil {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	}

	msg, err := client.GetAvailablePackageSummaries(ctx, &protoReq, grpc.Header(&metadata.HeaderMD), grpc.Trailer(&metadata.TrailerMD))
	return msg, metadata, err

}

func local_request_FluxV2PackagesService_GetAvailablePackageSummaries_0(ctx context.Context, marshaler runtime.Marshaler, server FluxV2PackagesServiceServer, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	var protoReq v1alpha1_0.GetAvailablePackageSummariesRequest
	var metadata runtime.ServerMetadata

	if err := req.ParseForm(); err != nil {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	}
	if err := runtime.PopulateQueryParameters(&protoReq, req.Form, filter_FluxV2PackagesService_GetAvailablePackageSummaries_0); err != nil {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	}

	msg, err := server.GetAvailablePackageSummaries(ctx, &protoReq)
	return msg, metadata, err

}

var (
	filter_FluxV2PackagesService_GetAvailablePackageDetail_0 = &utilities.DoubleArray{Encoding: map[string]int{}, Base: []int(nil), Check: []int(nil)}
)

func request_FluxV2PackagesService_GetAvailablePackageDetail_0(ctx context.Context, marshaler runtime.Marshaler, client FluxV2PackagesServiceClient, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	var protoReq v1alpha1_0.GetAvailablePackageDetailRequest
	var metadata runtime.ServerMetadata

	if err := req.ParseForm(); err != nil {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	}
	if err := runtime.PopulateQueryParameters(&protoReq, req.Form, filter_FluxV2PackagesService_GetAvailablePackageDetail_0); err != nil {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	}

	msg, err := client.GetAvailablePackageDetail(ctx, &protoReq, grpc.Header(&metadata.HeaderMD), grpc.Trailer(&metadata.TrailerMD))
	return msg, metadata, err

}

func local_request_FluxV2PackagesService_GetAvailablePackageDetail_0(ctx context.Context, marshaler runtime.Marshaler, server FluxV2PackagesServiceServer, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	var protoReq v1alpha1_0.GetAvailablePackageDetailRequest
	var metadata runtime.ServerMetadata

	if err := req.ParseForm(); err != nil {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	}
	if err := runtime.PopulateQueryParameters(&protoReq, req.Form, filter_FluxV2PackagesService_GetAvailablePackageDetail_0); err != nil {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	}

	msg, err := server.GetAvailablePackageDetail(ctx, &protoReq)
	return msg, metadata, err

}

var (
	filter_FluxV2PackagesService_GetAvailablePackageVersions_0 = &utilities.DoubleArray{Encoding: map[string]int{}, Base: []int(nil), Check: []int(nil)}
)

func request_FluxV2PackagesService_GetAvailablePackageVersions_0(ctx context.Context, marshaler runtime.Marshaler, client FluxV2PackagesServiceClient, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	var protoReq v1alpha1_0.GetAvailablePackageVersionsRequest
	var metadata runtime.ServerMetadata

	if err := req.ParseForm(); err != nil {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	}
	if err := runtime.PopulateQueryParameters(&protoReq, req.Form, filter_FluxV2PackagesService_GetAvailablePackageVersions_0); err != nil {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	}

	msg, err := client.GetAvailablePackageVersions(ctx, &protoReq, grpc.Header(&metadata.HeaderMD), grpc.Trailer(&metadata.TrailerMD))
	return msg, metadata, err

}

func local_request_FluxV2PackagesService_GetAvailablePackageVersions_0(ctx context.Context, marshaler runtime.Marshaler, server FluxV2PackagesServiceServer, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	var protoReq v1alpha1_0.GetAvailablePackageVersionsRequest
	var metadata runtime.ServerMetadata

	if err := req.ParseForm(); err != nil {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	}
	if err := runtime.PopulateQueryParameters(&protoReq, req.Form, filter_FluxV2PackagesService_GetAvailablePackageVersions_0); err != nil {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	}

	msg, err := server.GetAvailablePackageVersions(ctx, &protoReq)
	return msg, metadata, err

}

var (
	filter_FluxV2PackagesService_GetPackageRepositories_0 = &utilities.DoubleArray{Encoding: map[string]int{}, Base: []int(nil), Check: []int(nil)}
)

func request_FluxV2PackagesService_GetPackageRepositories_0(ctx context.Context, marshaler runtime.Marshaler, client FluxV2PackagesServiceClient, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	var protoReq GetPackageRepositoriesRequest
	var metadata runtime.ServerMetadata

	if err := req.ParseForm(); err != nil {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	}
	if err := runtime.PopulateQueryParameters(&protoReq, req.Form, filter_FluxV2PackagesService_GetPackageRepositories_0); err != nil {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	}

	msg, err := client.GetPackageRepositories(ctx, &protoReq, grpc.Header(&metadata.HeaderMD), grpc.Trailer(&metadata.TrailerMD))
	return msg, metadata, err

}

func local_request_FluxV2PackagesService_GetPackageRepositories_0(ctx context.Context, marshaler runtime.Marshaler, server FluxV2PackagesServiceServer, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	var protoReq GetPackageRepositoriesRequest
	var metadata runtime.ServerMetadata

	if err := req.ParseForm(); err != nil {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	}
	if err := runtime.PopulateQueryParameters(&protoReq, req.Form, filter_FluxV2PackagesService_GetPackageRepositories_0); err != nil {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	}

	msg, err := server.GetPackageRepositories(ctx, &protoReq)
	return msg, metadata, err

}

var (
	filter_FluxV2PackagesService_GetInstalledPackageSummaries_0 = &utilities.DoubleArray{Encoding: map[string]int{}, Base: []int(nil), Check: []int(nil)}
)

func request_FluxV2PackagesService_GetInstalledPackageSummaries_0(ctx context.Context, marshaler runtime.Marshaler, client FluxV2PackagesServiceClient, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	var protoReq v1alpha1_0.GetInstalledPackageSummariesRequest
	var metadata runtime.ServerMetadata

	if err := req.ParseForm(); err != nil {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	}
	if err := runtime.PopulateQueryParameters(&protoReq, req.Form, filter_FluxV2PackagesService_GetInstalledPackageSummaries_0); err != nil {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	}

	msg, err := client.GetInstalledPackageSummaries(ctx, &protoReq, grpc.Header(&metadata.HeaderMD), grpc.Trailer(&metadata.TrailerMD))
	return msg, metadata, err

}

func local_request_FluxV2PackagesService_GetInstalledPackageSummaries_0(ctx context.Context, marshaler runtime.Marshaler, server FluxV2PackagesServiceServer, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	var protoReq v1alpha1_0.GetInstalledPackageSummariesRequest
	var metadata runtime.ServerMetadata

	if err := req.ParseForm(); err != nil {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	}
	if err := runtime.PopulateQueryParameters(&protoReq, req.Form, filter_FluxV2PackagesService_GetInstalledPackageSummaries_0); err != nil {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	}

	msg, err := server.GetInstalledPackageSummaries(ctx, &protoReq)
	return msg, metadata, err

}

var (
	filter_FluxV2PackagesService_GetInstalledPackageDetail_0 = &utilities.DoubleArray{Encoding: map[string]int{}, Base: []int(nil), Check: []int(nil)}
)

func request_FluxV2PackagesService_GetInstalledPackageDetail_0(ctx context.Context, marshaler runtime.Marshaler, client FluxV2PackagesServiceClient, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	var protoReq v1alpha1_0.GetInstalledPackageDetailRequest
	var metadata runtime.ServerMetadata

	if err := req.ParseForm(); err != nil {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	}
	if err := runtime.PopulateQueryParameters(&protoReq, req.Form, filter_FluxV2PackagesService_GetInstalledPackageDetail_0); err != nil {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	}

	msg, err := client.GetInstalledPackageDetail(ctx, &protoReq, grpc.Header(&metadata.HeaderMD), grpc.Trailer(&metadata.TrailerMD))
	return msg, metadata, err

}

func local_request_FluxV2PackagesService_GetInstalledPackageDetail_0(ctx context.Context, marshaler runtime.Marshaler, server FluxV2PackagesServiceServer, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	var protoReq v1alpha1_0.GetInstalledPackageDetailRequest
	var metadata runtime.ServerMetadata

	if err := req.ParseForm(); err != nil {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	}
	if err := runtime.PopulateQueryParameters(&protoReq, req.Form, filter_FluxV2PackagesService_GetInstalledPackageDetail_0); err != nil {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	}

	msg, err := server.GetInstalledPackageDetail(ctx, &protoReq)
	return msg, metadata, err

}

func request_FluxV2PackagesService_CreateInstalledPackage_0(ctx context.Context, marshaler runtime.Marshaler, client FluxV2PackagesServiceClient, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	var protoReq v1alpha1_0.CreateInstalledPackageRequest
	var metadata runtime.ServerMetadata

	newReader, berr := utilities.IOReaderFactory(req.Body)
	if berr != nil {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", berr)
	}
	if err := marshaler.NewDecoder(newReader()).Decode(&protoReq); err != nil && err != io.EOF {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	}

	msg, err := client.CreateInstalledPackage(ctx, &protoReq, grpc.Header(&metadata.HeaderMD), grpc.Trailer(&metadata.TrailerMD))
	return msg, metadata, err

}

func local_request_FluxV2PackagesService_CreateInstalledPackage_0(ctx context.Context, marshaler runtime.Marshaler, server FluxV2PackagesServiceServer, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	var protoReq v1alpha1_0.CreateInstalledPackageRequest
	var metadata runtime.ServerMetadata

	newReader, berr := utilities.IOReaderFactory(req.Body)
	if berr != nil {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", berr)
	}
	if err := marshaler.NewDecoder(newReader()).Decode(&protoReq); err != nil && err != io.EOF {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	}

	msg, err := server.CreateInstalledPackage(ctx, &protoReq)
	return msg, metadata, err

}

func request_FluxV2PackagesService_UpdateInstalledPackage_0(ctx context.Context, marshaler runtime.Marshaler, client FluxV2PackagesServiceClient, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	var protoReq v1alpha1_0.UpdateInstalledPackageRequest
	var metadata runtime.ServerMetadata

	newReader, berr := utilities.IOReaderFactory(req.Body)
	if berr != nil {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", berr)
	}
	if err := marshaler.NewDecoder(newReader()).Decode(&protoReq); err != nil && err != io.EOF {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	}

	msg, err := client.UpdateInstalledPackage(ctx, &protoReq, grpc.Header(&metadata.HeaderMD), grpc.Trailer(&metadata.TrailerMD))
	return msg, metadata, err

}

func local_request_FluxV2PackagesService_UpdateInstalledPackage_0(ctx context.Context, marshaler runtime.Marshaler, server FluxV2PackagesServiceServer, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	var protoReq v1alpha1_0.UpdateInstalledPackageRequest
	var metadata runtime.ServerMetadata

	newReader, berr := utilities.IOReaderFactory(req.Body)
	if berr != nil {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", berr)
	}
	if err := marshaler.NewDecoder(newReader()).Decode(&protoReq); err != nil && err != io.EOF {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	}

	msg, err := server.UpdateInstalledPackage(ctx, &protoReq)
	return msg, metadata, err

}

// RegisterFluxV2PackagesServiceHandlerServer registers the http handlers for service FluxV2PackagesService to "mux".
// UnaryRPC     :call FluxV2PackagesServiceServer directly.
// StreamingRPC :currently unsupported pending https://github.com/grpc/grpc-go/issues/906.
// Note that using this registration option will cause many gRPC library features to stop working. Consider using RegisterFluxV2PackagesServiceHandlerFromEndpoint instead.
func RegisterFluxV2PackagesServiceHandlerServer(ctx context.Context, mux *runtime.ServeMux, server FluxV2PackagesServiceServer) error {

	mux.Handle("GET", pattern_FluxV2PackagesService_GetAvailablePackageSummaries_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		var stream runtime.ServerTransportStream
		ctx = grpc.NewContextWithServerTransportStream(ctx, &stream)
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateIncomingContext(ctx, mux, req, "/kubeappsapis.plugins.fluxv2.packages.v1alpha1.FluxV2PackagesService/GetAvailablePackageSummaries", runtime.WithHTTPPathPattern("/plugins/fluxv2/packages/v1alpha1/availablepackagesummaries"))
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}
		resp, md, err := local_request_FluxV2PackagesService_GetAvailablePackageSummaries_0(rctx, inboundMarshaler, server, req, pathParams)
		md.HeaderMD, md.TrailerMD = metadata.Join(md.HeaderMD, stream.Header()), metadata.Join(md.TrailerMD, stream.Trailer())
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}

		forward_FluxV2PackagesService_GetAvailablePackageSummaries_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})

	mux.Handle("GET", pattern_FluxV2PackagesService_GetAvailablePackageDetail_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		var stream runtime.ServerTransportStream
		ctx = grpc.NewContextWithServerTransportStream(ctx, &stream)
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateIncomingContext(ctx, mux, req, "/kubeappsapis.plugins.fluxv2.packages.v1alpha1.FluxV2PackagesService/GetAvailablePackageDetail", runtime.WithHTTPPathPattern("/plugins/fluxv2/packages/v1alpha1/availablepackagedetails"))
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}
		resp, md, err := local_request_FluxV2PackagesService_GetAvailablePackageDetail_0(rctx, inboundMarshaler, server, req, pathParams)
		md.HeaderMD, md.TrailerMD = metadata.Join(md.HeaderMD, stream.Header()), metadata.Join(md.TrailerMD, stream.Trailer())
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}

		forward_FluxV2PackagesService_GetAvailablePackageDetail_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})

	mux.Handle("GET", pattern_FluxV2PackagesService_GetAvailablePackageVersions_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		var stream runtime.ServerTransportStream
		ctx = grpc.NewContextWithServerTransportStream(ctx, &stream)
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateIncomingContext(ctx, mux, req, "/kubeappsapis.plugins.fluxv2.packages.v1alpha1.FluxV2PackagesService/GetAvailablePackageVersions", runtime.WithHTTPPathPattern("/plugins/fluxv2/packages/v1alpha1/availablepackageversions"))
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}
		resp, md, err := local_request_FluxV2PackagesService_GetAvailablePackageVersions_0(rctx, inboundMarshaler, server, req, pathParams)
		md.HeaderMD, md.TrailerMD = metadata.Join(md.HeaderMD, stream.Header()), metadata.Join(md.TrailerMD, stream.Trailer())
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}

		forward_FluxV2PackagesService_GetAvailablePackageVersions_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})

	mux.Handle("GET", pattern_FluxV2PackagesService_GetPackageRepositories_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		var stream runtime.ServerTransportStream
		ctx = grpc.NewContextWithServerTransportStream(ctx, &stream)
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateIncomingContext(ctx, mux, req, "/kubeappsapis.plugins.fluxv2.packages.v1alpha1.FluxV2PackagesService/GetPackageRepositories", runtime.WithHTTPPathPattern("/plugins/fluxv2/packages/v1alpha1/packagerepositories"))
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}
		resp, md, err := local_request_FluxV2PackagesService_GetPackageRepositories_0(rctx, inboundMarshaler, server, req, pathParams)
		md.HeaderMD, md.TrailerMD = metadata.Join(md.HeaderMD, stream.Header()), metadata.Join(md.TrailerMD, stream.Trailer())
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}

		forward_FluxV2PackagesService_GetPackageRepositories_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})

	mux.Handle("GET", pattern_FluxV2PackagesService_GetInstalledPackageSummaries_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		var stream runtime.ServerTransportStream
		ctx = grpc.NewContextWithServerTransportStream(ctx, &stream)
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateIncomingContext(ctx, mux, req, "/kubeappsapis.plugins.fluxv2.packages.v1alpha1.FluxV2PackagesService/GetInstalledPackageSummaries", runtime.WithHTTPPathPattern("/plugins/fluxv2/packages/v1alpha1/installedpackagesummaries"))
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}
		resp, md, err := local_request_FluxV2PackagesService_GetInstalledPackageSummaries_0(rctx, inboundMarshaler, server, req, pathParams)
		md.HeaderMD, md.TrailerMD = metadata.Join(md.HeaderMD, stream.Header()), metadata.Join(md.TrailerMD, stream.Trailer())
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}

		forward_FluxV2PackagesService_GetInstalledPackageSummaries_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})

	mux.Handle("GET", pattern_FluxV2PackagesService_GetInstalledPackageDetail_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		var stream runtime.ServerTransportStream
		ctx = grpc.NewContextWithServerTransportStream(ctx, &stream)
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateIncomingContext(ctx, mux, req, "/kubeappsapis.plugins.fluxv2.packages.v1alpha1.FluxV2PackagesService/GetInstalledPackageDetail", runtime.WithHTTPPathPattern("/plugins/fluxv2/packages/v1alpha1/installedpackagedetail"))
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}
		resp, md, err := local_request_FluxV2PackagesService_GetInstalledPackageDetail_0(rctx, inboundMarshaler, server, req, pathParams)
		md.HeaderMD, md.TrailerMD = metadata.Join(md.HeaderMD, stream.Header()), metadata.Join(md.TrailerMD, stream.Trailer())
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}

		forward_FluxV2PackagesService_GetInstalledPackageDetail_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})

	mux.Handle("POST", pattern_FluxV2PackagesService_CreateInstalledPackage_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		var stream runtime.ServerTransportStream
		ctx = grpc.NewContextWithServerTransportStream(ctx, &stream)
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateIncomingContext(ctx, mux, req, "/kubeappsapis.plugins.fluxv2.packages.v1alpha1.FluxV2PackagesService/CreateInstalledPackage", runtime.WithHTTPPathPattern("/plugins/fluxv2/packages/v1alpha1/installedpackages"))
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}
		resp, md, err := local_request_FluxV2PackagesService_CreateInstalledPackage_0(rctx, inboundMarshaler, server, req, pathParams)
		md.HeaderMD, md.TrailerMD = metadata.Join(md.HeaderMD, stream.Header()), metadata.Join(md.TrailerMD, stream.Trailer())
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}

		forward_FluxV2PackagesService_CreateInstalledPackage_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})

	mux.Handle("PUT", pattern_FluxV2PackagesService_UpdateInstalledPackage_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		var stream runtime.ServerTransportStream
		ctx = grpc.NewContextWithServerTransportStream(ctx, &stream)
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateIncomingContext(ctx, mux, req, "/kubeappsapis.plugins.fluxv2.packages.v1alpha1.FluxV2PackagesService/UpdateInstalledPackage", runtime.WithHTTPPathPattern("/plugins/fluxv2/packages/v1alpha1/installedpackages"))
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}
		resp, md, err := local_request_FluxV2PackagesService_UpdateInstalledPackage_0(rctx, inboundMarshaler, server, req, pathParams)
		md.HeaderMD, md.TrailerMD = metadata.Join(md.HeaderMD, stream.Header()), metadata.Join(md.TrailerMD, stream.Trailer())
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}

		forward_FluxV2PackagesService_UpdateInstalledPackage_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})

	return nil
}

// RegisterFluxV2PackagesServiceHandlerFromEndpoint is same as RegisterFluxV2PackagesServiceHandler but
// automatically dials to "endpoint" and closes the connection when "ctx" gets done.
func RegisterFluxV2PackagesServiceHandlerFromEndpoint(ctx context.Context, mux *runtime.ServeMux, endpoint string, opts []grpc.DialOption) (err error) {
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

	return RegisterFluxV2PackagesServiceHandler(ctx, mux, conn)
}

// RegisterFluxV2PackagesServiceHandler registers the http handlers for service FluxV2PackagesService to "mux".
// The handlers forward requests to the grpc endpoint over "conn".
func RegisterFluxV2PackagesServiceHandler(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error {
	return RegisterFluxV2PackagesServiceHandlerClient(ctx, mux, NewFluxV2PackagesServiceClient(conn))
}

// RegisterFluxV2PackagesServiceHandlerClient registers the http handlers for service FluxV2PackagesService
// to "mux". The handlers forward requests to the grpc endpoint over the given implementation of "FluxV2PackagesServiceClient".
// Note: the gRPC framework executes interceptors within the gRPC handler. If the passed in "FluxV2PackagesServiceClient"
// doesn't go through the normal gRPC flow (creating a gRPC client etc.) then it will be up to the passed in
// "FluxV2PackagesServiceClient" to call the correct interceptors.
func RegisterFluxV2PackagesServiceHandlerClient(ctx context.Context, mux *runtime.ServeMux, client FluxV2PackagesServiceClient) error {

	mux.Handle("GET", pattern_FluxV2PackagesService_GetAvailablePackageSummaries_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateContext(ctx, mux, req, "/kubeappsapis.plugins.fluxv2.packages.v1alpha1.FluxV2PackagesService/GetAvailablePackageSummaries", runtime.WithHTTPPathPattern("/plugins/fluxv2/packages/v1alpha1/availablepackagesummaries"))
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}
		resp, md, err := request_FluxV2PackagesService_GetAvailablePackageSummaries_0(rctx, inboundMarshaler, client, req, pathParams)
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}

		forward_FluxV2PackagesService_GetAvailablePackageSummaries_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})

	mux.Handle("GET", pattern_FluxV2PackagesService_GetAvailablePackageDetail_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateContext(ctx, mux, req, "/kubeappsapis.plugins.fluxv2.packages.v1alpha1.FluxV2PackagesService/GetAvailablePackageDetail", runtime.WithHTTPPathPattern("/plugins/fluxv2/packages/v1alpha1/availablepackagedetails"))
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}
		resp, md, err := request_FluxV2PackagesService_GetAvailablePackageDetail_0(rctx, inboundMarshaler, client, req, pathParams)
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}

		forward_FluxV2PackagesService_GetAvailablePackageDetail_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})

	mux.Handle("GET", pattern_FluxV2PackagesService_GetAvailablePackageVersions_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateContext(ctx, mux, req, "/kubeappsapis.plugins.fluxv2.packages.v1alpha1.FluxV2PackagesService/GetAvailablePackageVersions", runtime.WithHTTPPathPattern("/plugins/fluxv2/packages/v1alpha1/availablepackageversions"))
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}
		resp, md, err := request_FluxV2PackagesService_GetAvailablePackageVersions_0(rctx, inboundMarshaler, client, req, pathParams)
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}

		forward_FluxV2PackagesService_GetAvailablePackageVersions_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})

	mux.Handle("GET", pattern_FluxV2PackagesService_GetPackageRepositories_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateContext(ctx, mux, req, "/kubeappsapis.plugins.fluxv2.packages.v1alpha1.FluxV2PackagesService/GetPackageRepositories", runtime.WithHTTPPathPattern("/plugins/fluxv2/packages/v1alpha1/packagerepositories"))
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}
		resp, md, err := request_FluxV2PackagesService_GetPackageRepositories_0(rctx, inboundMarshaler, client, req, pathParams)
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}

		forward_FluxV2PackagesService_GetPackageRepositories_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})

	mux.Handle("GET", pattern_FluxV2PackagesService_GetInstalledPackageSummaries_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateContext(ctx, mux, req, "/kubeappsapis.plugins.fluxv2.packages.v1alpha1.FluxV2PackagesService/GetInstalledPackageSummaries", runtime.WithHTTPPathPattern("/plugins/fluxv2/packages/v1alpha1/installedpackagesummaries"))
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}
		resp, md, err := request_FluxV2PackagesService_GetInstalledPackageSummaries_0(rctx, inboundMarshaler, client, req, pathParams)
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}

		forward_FluxV2PackagesService_GetInstalledPackageSummaries_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})

	mux.Handle("GET", pattern_FluxV2PackagesService_GetInstalledPackageDetail_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateContext(ctx, mux, req, "/kubeappsapis.plugins.fluxv2.packages.v1alpha1.FluxV2PackagesService/GetInstalledPackageDetail", runtime.WithHTTPPathPattern("/plugins/fluxv2/packages/v1alpha1/installedpackagedetail"))
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}
		resp, md, err := request_FluxV2PackagesService_GetInstalledPackageDetail_0(rctx, inboundMarshaler, client, req, pathParams)
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}

		forward_FluxV2PackagesService_GetInstalledPackageDetail_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})

	mux.Handle("POST", pattern_FluxV2PackagesService_CreateInstalledPackage_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateContext(ctx, mux, req, "/kubeappsapis.plugins.fluxv2.packages.v1alpha1.FluxV2PackagesService/CreateInstalledPackage", runtime.WithHTTPPathPattern("/plugins/fluxv2/packages/v1alpha1/installedpackages"))
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}
		resp, md, err := request_FluxV2PackagesService_CreateInstalledPackage_0(rctx, inboundMarshaler, client, req, pathParams)
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}

		forward_FluxV2PackagesService_CreateInstalledPackage_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})

	mux.Handle("PUT", pattern_FluxV2PackagesService_UpdateInstalledPackage_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateContext(ctx, mux, req, "/kubeappsapis.plugins.fluxv2.packages.v1alpha1.FluxV2PackagesService/UpdateInstalledPackage", runtime.WithHTTPPathPattern("/plugins/fluxv2/packages/v1alpha1/installedpackages"))
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}
		resp, md, err := request_FluxV2PackagesService_UpdateInstalledPackage_0(rctx, inboundMarshaler, client, req, pathParams)
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}

		forward_FluxV2PackagesService_UpdateInstalledPackage_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})

	return nil
}

var (
	pattern_FluxV2PackagesService_GetAvailablePackageSummaries_0 = runtime.MustPattern(runtime.NewPattern(1, []int{2, 0, 2, 1, 2, 2, 2, 3, 2, 4}, []string{"plugins", "fluxv2", "packages", "v1alpha1", "availablepackagesummaries"}, ""))

	pattern_FluxV2PackagesService_GetAvailablePackageDetail_0 = runtime.MustPattern(runtime.NewPattern(1, []int{2, 0, 2, 1, 2, 2, 2, 3, 2, 4}, []string{"plugins", "fluxv2", "packages", "v1alpha1", "availablepackagedetails"}, ""))

	pattern_FluxV2PackagesService_GetAvailablePackageVersions_0 = runtime.MustPattern(runtime.NewPattern(1, []int{2, 0, 2, 1, 2, 2, 2, 3, 2, 4}, []string{"plugins", "fluxv2", "packages", "v1alpha1", "availablepackageversions"}, ""))

	pattern_FluxV2PackagesService_GetPackageRepositories_0 = runtime.MustPattern(runtime.NewPattern(1, []int{2, 0, 2, 1, 2, 2, 2, 3, 2, 4}, []string{"plugins", "fluxv2", "packages", "v1alpha1", "packagerepositories"}, ""))

	pattern_FluxV2PackagesService_GetInstalledPackageSummaries_0 = runtime.MustPattern(runtime.NewPattern(1, []int{2, 0, 2, 1, 2, 2, 2, 3, 2, 4}, []string{"plugins", "fluxv2", "packages", "v1alpha1", "installedpackagesummaries"}, ""))

	pattern_FluxV2PackagesService_GetInstalledPackageDetail_0 = runtime.MustPattern(runtime.NewPattern(1, []int{2, 0, 2, 1, 2, 2, 2, 3, 2, 4}, []string{"plugins", "fluxv2", "packages", "v1alpha1", "installedpackagedetail"}, ""))

	pattern_FluxV2PackagesService_CreateInstalledPackage_0 = runtime.MustPattern(runtime.NewPattern(1, []int{2, 0, 2, 1, 2, 2, 2, 3, 2, 4}, []string{"plugins", "fluxv2", "packages", "v1alpha1", "installedpackages"}, ""))

	pattern_FluxV2PackagesService_UpdateInstalledPackage_0 = runtime.MustPattern(runtime.NewPattern(1, []int{2, 0, 2, 1, 2, 2, 2, 3, 2, 4}, []string{"plugins", "fluxv2", "packages", "v1alpha1", "installedpackages"}, ""))
)

var (
	forward_FluxV2PackagesService_GetAvailablePackageSummaries_0 = runtime.ForwardResponseMessage

	forward_FluxV2PackagesService_GetAvailablePackageDetail_0 = runtime.ForwardResponseMessage

	forward_FluxV2PackagesService_GetAvailablePackageVersions_0 = runtime.ForwardResponseMessage

	forward_FluxV2PackagesService_GetPackageRepositories_0 = runtime.ForwardResponseMessage

	forward_FluxV2PackagesService_GetInstalledPackageSummaries_0 = runtime.ForwardResponseMessage

	forward_FluxV2PackagesService_GetInstalledPackageDetail_0 = runtime.ForwardResponseMessage

	forward_FluxV2PackagesService_CreateInstalledPackage_0 = runtime.ForwardResponseMessage

	forward_FluxV2PackagesService_UpdateInstalledPackage_0 = runtime.ForwardResponseMessage
)
