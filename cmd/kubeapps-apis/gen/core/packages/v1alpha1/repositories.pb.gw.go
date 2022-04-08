// Code generated by protoc-gen-grpc-gateway. DO NOT EDIT.
// source: kubeappsapis/core/packages/v1alpha1/repositories.proto

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

func request_RepositoriesService_AddPackageRepository_0(ctx context.Context, marshaler runtime.Marshaler, client RepositoriesServiceClient, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	var protoReq AddPackageRepositoryRequest
	var metadata runtime.ServerMetadata

	newReader, berr := utilities.IOReaderFactory(req.Body)
	if berr != nil {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", berr)
	}
	if err := marshaler.NewDecoder(newReader()).Decode(&protoReq); err != nil && err != io.EOF {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	}

	msg, err := client.AddPackageRepository(ctx, &protoReq, grpc.Header(&metadata.HeaderMD), grpc.Trailer(&metadata.TrailerMD))
	return msg, metadata, err

}

func local_request_RepositoriesService_AddPackageRepository_0(ctx context.Context, marshaler runtime.Marshaler, server RepositoriesServiceServer, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	var protoReq AddPackageRepositoryRequest
	var metadata runtime.ServerMetadata

	newReader, berr := utilities.IOReaderFactory(req.Body)
	if berr != nil {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", berr)
	}
	if err := marshaler.NewDecoder(newReader()).Decode(&protoReq); err != nil && err != io.EOF {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	}

	msg, err := server.AddPackageRepository(ctx, &protoReq)
	return msg, metadata, err

}

var (
	filter_RepositoriesService_GetPackageRepositoryDetail_0 = &utilities.DoubleArray{Encoding: map[string]int{"package_repo_ref": 0, "plugin": 1, "name": 2, "version": 3, "context": 4, "cluster": 5, "namespace": 6, "identifier": 7}, Base: []int{1, 7, 1, 1, 2, 2, 2, 3, 6, 0, 0, 0, 5, 0, 7, 0}, Check: []int{0, 1, 2, 3, 2, 5, 2, 7, 2, 4, 6, 8, 9, 13, 2, 15}}
)

func request_RepositoriesService_GetPackageRepositoryDetail_0(ctx context.Context, marshaler runtime.Marshaler, client RepositoriesServiceClient, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	var protoReq GetPackageRepositoryDetailRequest
	var metadata runtime.ServerMetadata

	var (
		val string
		ok  bool
		err error
		_   = err
	)

	val, ok = pathParams["package_repo_ref.plugin.name"]
	if !ok {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "missing parameter %s", "package_repo_ref.plugin.name")
	}

	err = runtime.PopulateFieldFromPath(&protoReq, "package_repo_ref.plugin.name", val)
	if err != nil {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "type mismatch, parameter: %s, error: %v", "package_repo_ref.plugin.name", err)
	}

	val, ok = pathParams["package_repo_ref.plugin.version"]
	if !ok {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "missing parameter %s", "package_repo_ref.plugin.version")
	}

	err = runtime.PopulateFieldFromPath(&protoReq, "package_repo_ref.plugin.version", val)
	if err != nil {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "type mismatch, parameter: %s, error: %v", "package_repo_ref.plugin.version", err)
	}

	val, ok = pathParams["package_repo_ref.context.cluster"]
	if !ok {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "missing parameter %s", "package_repo_ref.context.cluster")
	}

	err = runtime.PopulateFieldFromPath(&protoReq, "package_repo_ref.context.cluster", val)
	if err != nil {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "type mismatch, parameter: %s, error: %v", "package_repo_ref.context.cluster", err)
	}

	val, ok = pathParams["package_repo_ref.context.namespace"]
	if !ok {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "missing parameter %s", "package_repo_ref.context.namespace")
	}

	err = runtime.PopulateFieldFromPath(&protoReq, "package_repo_ref.context.namespace", val)
	if err != nil {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "type mismatch, parameter: %s, error: %v", "package_repo_ref.context.namespace", err)
	}

	val, ok = pathParams["package_repo_ref.identifier"]
	if !ok {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "missing parameter %s", "package_repo_ref.identifier")
	}

	err = runtime.PopulateFieldFromPath(&protoReq, "package_repo_ref.identifier", val)
	if err != nil {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "type mismatch, parameter: %s, error: %v", "package_repo_ref.identifier", err)
	}

	if err := req.ParseForm(); err != nil {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	}
	if err := runtime.PopulateQueryParameters(&protoReq, req.Form, filter_RepositoriesService_GetPackageRepositoryDetail_0); err != nil {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	}

	msg, err := client.GetPackageRepositoryDetail(ctx, &protoReq, grpc.Header(&metadata.HeaderMD), grpc.Trailer(&metadata.TrailerMD))
	return msg, metadata, err

}

func local_request_RepositoriesService_GetPackageRepositoryDetail_0(ctx context.Context, marshaler runtime.Marshaler, server RepositoriesServiceServer, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	var protoReq GetPackageRepositoryDetailRequest
	var metadata runtime.ServerMetadata

	var (
		val string
		ok  bool
		err error
		_   = err
	)

	val, ok = pathParams["package_repo_ref.plugin.name"]
	if !ok {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "missing parameter %s", "package_repo_ref.plugin.name")
	}

	err = runtime.PopulateFieldFromPath(&protoReq, "package_repo_ref.plugin.name", val)
	if err != nil {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "type mismatch, parameter: %s, error: %v", "package_repo_ref.plugin.name", err)
	}

	val, ok = pathParams["package_repo_ref.plugin.version"]
	if !ok {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "missing parameter %s", "package_repo_ref.plugin.version")
	}

	err = runtime.PopulateFieldFromPath(&protoReq, "package_repo_ref.plugin.version", val)
	if err != nil {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "type mismatch, parameter: %s, error: %v", "package_repo_ref.plugin.version", err)
	}

	val, ok = pathParams["package_repo_ref.context.cluster"]
	if !ok {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "missing parameter %s", "package_repo_ref.context.cluster")
	}

	err = runtime.PopulateFieldFromPath(&protoReq, "package_repo_ref.context.cluster", val)
	if err != nil {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "type mismatch, parameter: %s, error: %v", "package_repo_ref.context.cluster", err)
	}

	val, ok = pathParams["package_repo_ref.context.namespace"]
	if !ok {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "missing parameter %s", "package_repo_ref.context.namespace")
	}

	err = runtime.PopulateFieldFromPath(&protoReq, "package_repo_ref.context.namespace", val)
	if err != nil {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "type mismatch, parameter: %s, error: %v", "package_repo_ref.context.namespace", err)
	}

	val, ok = pathParams["package_repo_ref.identifier"]
	if !ok {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "missing parameter %s", "package_repo_ref.identifier")
	}

	err = runtime.PopulateFieldFromPath(&protoReq, "package_repo_ref.identifier", val)
	if err != nil {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "type mismatch, parameter: %s, error: %v", "package_repo_ref.identifier", err)
	}

	if err := req.ParseForm(); err != nil {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	}
	if err := runtime.PopulateQueryParameters(&protoReq, req.Form, filter_RepositoriesService_GetPackageRepositoryDetail_0); err != nil {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	}

	msg, err := server.GetPackageRepositoryDetail(ctx, &protoReq)
	return msg, metadata, err

}

var (
	filter_RepositoriesService_GetPackageRepositorySummaries_0 = &utilities.DoubleArray{Encoding: map[string]int{}, Base: []int(nil), Check: []int(nil)}
)

func request_RepositoriesService_GetPackageRepositorySummaries_0(ctx context.Context, marshaler runtime.Marshaler, client RepositoriesServiceClient, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	var protoReq GetPackageRepositorySummariesRequest
	var metadata runtime.ServerMetadata

	if err := req.ParseForm(); err != nil {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	}
	if err := runtime.PopulateQueryParameters(&protoReq, req.Form, filter_RepositoriesService_GetPackageRepositorySummaries_0); err != nil {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	}

	msg, err := client.GetPackageRepositorySummaries(ctx, &protoReq, grpc.Header(&metadata.HeaderMD), grpc.Trailer(&metadata.TrailerMD))
	return msg, metadata, err

}

func local_request_RepositoriesService_GetPackageRepositorySummaries_0(ctx context.Context, marshaler runtime.Marshaler, server RepositoriesServiceServer, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	var protoReq GetPackageRepositorySummariesRequest
	var metadata runtime.ServerMetadata

	if err := req.ParseForm(); err != nil {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	}
	if err := runtime.PopulateQueryParameters(&protoReq, req.Form, filter_RepositoriesService_GetPackageRepositorySummaries_0); err != nil {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	}

	msg, err := server.GetPackageRepositorySummaries(ctx, &protoReq)
	return msg, metadata, err

}

func request_RepositoriesService_UpdatePackageRepository_0(ctx context.Context, marshaler runtime.Marshaler, client RepositoriesServiceClient, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	var protoReq UpdatePackageRepositoryRequest
	var metadata runtime.ServerMetadata

	newReader, berr := utilities.IOReaderFactory(req.Body)
	if berr != nil {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", berr)
	}
	if err := marshaler.NewDecoder(newReader()).Decode(&protoReq); err != nil && err != io.EOF {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	}

	var (
		val string
		ok  bool
		err error
		_   = err
	)

	val, ok = pathParams["package_repo_ref.plugin.name"]
	if !ok {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "missing parameter %s", "package_repo_ref.plugin.name")
	}

	err = runtime.PopulateFieldFromPath(&protoReq, "package_repo_ref.plugin.name", val)
	if err != nil {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "type mismatch, parameter: %s, error: %v", "package_repo_ref.plugin.name", err)
	}

	val, ok = pathParams["package_repo_ref.plugin.version"]
	if !ok {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "missing parameter %s", "package_repo_ref.plugin.version")
	}

	err = runtime.PopulateFieldFromPath(&protoReq, "package_repo_ref.plugin.version", val)
	if err != nil {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "type mismatch, parameter: %s, error: %v", "package_repo_ref.plugin.version", err)
	}

	val, ok = pathParams["package_repo_ref.context.cluster"]
	if !ok {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "missing parameter %s", "package_repo_ref.context.cluster")
	}

	err = runtime.PopulateFieldFromPath(&protoReq, "package_repo_ref.context.cluster", val)
	if err != nil {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "type mismatch, parameter: %s, error: %v", "package_repo_ref.context.cluster", err)
	}

	val, ok = pathParams["package_repo_ref.context.namespace"]
	if !ok {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "missing parameter %s", "package_repo_ref.context.namespace")
	}

	err = runtime.PopulateFieldFromPath(&protoReq, "package_repo_ref.context.namespace", val)
	if err != nil {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "type mismatch, parameter: %s, error: %v", "package_repo_ref.context.namespace", err)
	}

	val, ok = pathParams["package_repo_ref.identifier"]
	if !ok {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "missing parameter %s", "package_repo_ref.identifier")
	}

	err = runtime.PopulateFieldFromPath(&protoReq, "package_repo_ref.identifier", val)
	if err != nil {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "type mismatch, parameter: %s, error: %v", "package_repo_ref.identifier", err)
	}

	msg, err := client.UpdatePackageRepository(ctx, &protoReq, grpc.Header(&metadata.HeaderMD), grpc.Trailer(&metadata.TrailerMD))
	return msg, metadata, err

}

func local_request_RepositoriesService_UpdatePackageRepository_0(ctx context.Context, marshaler runtime.Marshaler, server RepositoriesServiceServer, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	var protoReq UpdatePackageRepositoryRequest
	var metadata runtime.ServerMetadata

	newReader, berr := utilities.IOReaderFactory(req.Body)
	if berr != nil {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", berr)
	}
	if err := marshaler.NewDecoder(newReader()).Decode(&protoReq); err != nil && err != io.EOF {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	}

	var (
		val string
		ok  bool
		err error
		_   = err
	)

	val, ok = pathParams["package_repo_ref.plugin.name"]
	if !ok {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "missing parameter %s", "package_repo_ref.plugin.name")
	}

	err = runtime.PopulateFieldFromPath(&protoReq, "package_repo_ref.plugin.name", val)
	if err != nil {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "type mismatch, parameter: %s, error: %v", "package_repo_ref.plugin.name", err)
	}

	val, ok = pathParams["package_repo_ref.plugin.version"]
	if !ok {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "missing parameter %s", "package_repo_ref.plugin.version")
	}

	err = runtime.PopulateFieldFromPath(&protoReq, "package_repo_ref.plugin.version", val)
	if err != nil {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "type mismatch, parameter: %s, error: %v", "package_repo_ref.plugin.version", err)
	}

	val, ok = pathParams["package_repo_ref.context.cluster"]
	if !ok {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "missing parameter %s", "package_repo_ref.context.cluster")
	}

	err = runtime.PopulateFieldFromPath(&protoReq, "package_repo_ref.context.cluster", val)
	if err != nil {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "type mismatch, parameter: %s, error: %v", "package_repo_ref.context.cluster", err)
	}

	val, ok = pathParams["package_repo_ref.context.namespace"]
	if !ok {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "missing parameter %s", "package_repo_ref.context.namespace")
	}

	err = runtime.PopulateFieldFromPath(&protoReq, "package_repo_ref.context.namespace", val)
	if err != nil {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "type mismatch, parameter: %s, error: %v", "package_repo_ref.context.namespace", err)
	}

	val, ok = pathParams["package_repo_ref.identifier"]
	if !ok {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "missing parameter %s", "package_repo_ref.identifier")
	}

	err = runtime.PopulateFieldFromPath(&protoReq, "package_repo_ref.identifier", val)
	if err != nil {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "type mismatch, parameter: %s, error: %v", "package_repo_ref.identifier", err)
	}

	msg, err := server.UpdatePackageRepository(ctx, &protoReq)
	return msg, metadata, err

}

// RegisterRepositoriesServiceHandlerServer registers the http handlers for service RepositoriesService to "mux".
// UnaryRPC     :call RepositoriesServiceServer directly.
// StreamingRPC :currently unsupported pending https://github.com/grpc/grpc-go/issues/906.
// Note that using this registration option will cause many gRPC library features to stop working. Consider using RegisterRepositoriesServiceHandlerFromEndpoint instead.
func RegisterRepositoriesServiceHandlerServer(ctx context.Context, mux *runtime.ServeMux, server RepositoriesServiceServer) error {

	mux.Handle("POST", pattern_RepositoriesService_AddPackageRepository_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		var stream runtime.ServerTransportStream
		ctx = grpc.NewContextWithServerTransportStream(ctx, &stream)
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateIncomingContext(ctx, mux, req, "/kubeappsapis.core.packages.v1alpha1.RepositoriesService/AddPackageRepository", runtime.WithHTTPPathPattern("/core/packages/v1alpha1/repositories"))
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}
		resp, md, err := local_request_RepositoriesService_AddPackageRepository_0(rctx, inboundMarshaler, server, req, pathParams)
		md.HeaderMD, md.TrailerMD = metadata.Join(md.HeaderMD, stream.Header()), metadata.Join(md.TrailerMD, stream.Trailer())
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}

		forward_RepositoriesService_AddPackageRepository_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})

	mux.Handle("GET", pattern_RepositoriesService_GetPackageRepositoryDetail_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		var stream runtime.ServerTransportStream
		ctx = grpc.NewContextWithServerTransportStream(ctx, &stream)
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateIncomingContext(ctx, mux, req, "/kubeappsapis.core.packages.v1alpha1.RepositoriesService/GetPackageRepositoryDetail", runtime.WithHTTPPathPattern("/core/packages/v1alpha1/repositories/plugin/{package_repo_ref.plugin.name}/{package_repo_ref.plugin.version}/c/{package_repo_ref.context.cluster}/ns/{package_repo_ref.context.namespace}/{package_repo_ref.identifier=**}"))
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}
		resp, md, err := local_request_RepositoriesService_GetPackageRepositoryDetail_0(rctx, inboundMarshaler, server, req, pathParams)
		md.HeaderMD, md.TrailerMD = metadata.Join(md.HeaderMD, stream.Header()), metadata.Join(md.TrailerMD, stream.Trailer())
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}

		forward_RepositoriesService_GetPackageRepositoryDetail_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})

	mux.Handle("GET", pattern_RepositoriesService_GetPackageRepositorySummaries_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		var stream runtime.ServerTransportStream
		ctx = grpc.NewContextWithServerTransportStream(ctx, &stream)
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateIncomingContext(ctx, mux, req, "/kubeappsapis.core.packages.v1alpha1.RepositoriesService/GetPackageRepositorySummaries", runtime.WithHTTPPathPattern("/core/packages/v1alpha1/repositories"))
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}
		resp, md, err := local_request_RepositoriesService_GetPackageRepositorySummaries_0(rctx, inboundMarshaler, server, req, pathParams)
		md.HeaderMD, md.TrailerMD = metadata.Join(md.HeaderMD, stream.Header()), metadata.Join(md.TrailerMD, stream.Trailer())
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}

		forward_RepositoriesService_GetPackageRepositorySummaries_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})

	mux.Handle("PUT", pattern_RepositoriesService_UpdatePackageRepository_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		var stream runtime.ServerTransportStream
		ctx = grpc.NewContextWithServerTransportStream(ctx, &stream)
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateIncomingContext(ctx, mux, req, "/kubeappsapis.core.packages.v1alpha1.RepositoriesService/UpdatePackageRepository", runtime.WithHTTPPathPattern("/core/packages/v1alpha1/repositories/plugin/{package_repo_ref.plugin.name}/{package_repo_ref.plugin.version}/c/{package_repo_ref.context.cluster}/ns/{package_repo_ref.context.namespace}/{package_repo_ref.identifier=**}"))
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}
		resp, md, err := local_request_RepositoriesService_UpdatePackageRepository_0(rctx, inboundMarshaler, server, req, pathParams)
		md.HeaderMD, md.TrailerMD = metadata.Join(md.HeaderMD, stream.Header()), metadata.Join(md.TrailerMD, stream.Trailer())
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}

		forward_RepositoriesService_UpdatePackageRepository_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})

	return nil
}

// RegisterRepositoriesServiceHandlerFromEndpoint is same as RegisterRepositoriesServiceHandler but
// automatically dials to "endpoint" and closes the connection when "ctx" gets done.
func RegisterRepositoriesServiceHandlerFromEndpoint(ctx context.Context, mux *runtime.ServeMux, endpoint string, opts []grpc.DialOption) (err error) {
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

	return RegisterRepositoriesServiceHandler(ctx, mux, conn)
}

// RegisterRepositoriesServiceHandler registers the http handlers for service RepositoriesService to "mux".
// The handlers forward requests to the grpc endpoint over "conn".
func RegisterRepositoriesServiceHandler(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error {
	return RegisterRepositoriesServiceHandlerClient(ctx, mux, NewRepositoriesServiceClient(conn))
}

// RegisterRepositoriesServiceHandlerClient registers the http handlers for service RepositoriesService
// to "mux". The handlers forward requests to the grpc endpoint over the given implementation of "RepositoriesServiceClient".
// Note: the gRPC framework executes interceptors within the gRPC handler. If the passed in "RepositoriesServiceClient"
// doesn't go through the normal gRPC flow (creating a gRPC client etc.) then it will be up to the passed in
// "RepositoriesServiceClient" to call the correct interceptors.
func RegisterRepositoriesServiceHandlerClient(ctx context.Context, mux *runtime.ServeMux, client RepositoriesServiceClient) error {

	mux.Handle("POST", pattern_RepositoriesService_AddPackageRepository_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateContext(ctx, mux, req, "/kubeappsapis.core.packages.v1alpha1.RepositoriesService/AddPackageRepository", runtime.WithHTTPPathPattern("/core/packages/v1alpha1/repositories"))
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}
		resp, md, err := request_RepositoriesService_AddPackageRepository_0(rctx, inboundMarshaler, client, req, pathParams)
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}

		forward_RepositoriesService_AddPackageRepository_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})

	mux.Handle("GET", pattern_RepositoriesService_GetPackageRepositoryDetail_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateContext(ctx, mux, req, "/kubeappsapis.core.packages.v1alpha1.RepositoriesService/GetPackageRepositoryDetail", runtime.WithHTTPPathPattern("/core/packages/v1alpha1/repositories/plugin/{package_repo_ref.plugin.name}/{package_repo_ref.plugin.version}/c/{package_repo_ref.context.cluster}/ns/{package_repo_ref.context.namespace}/{package_repo_ref.identifier=**}"))
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}
		resp, md, err := request_RepositoriesService_GetPackageRepositoryDetail_0(rctx, inboundMarshaler, client, req, pathParams)
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}

		forward_RepositoriesService_GetPackageRepositoryDetail_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})

	mux.Handle("GET", pattern_RepositoriesService_GetPackageRepositorySummaries_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateContext(ctx, mux, req, "/kubeappsapis.core.packages.v1alpha1.RepositoriesService/GetPackageRepositorySummaries", runtime.WithHTTPPathPattern("/core/packages/v1alpha1/repositories"))
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}
		resp, md, err := request_RepositoriesService_GetPackageRepositorySummaries_0(rctx, inboundMarshaler, client, req, pathParams)
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}

		forward_RepositoriesService_GetPackageRepositorySummaries_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})

	mux.Handle("PUT", pattern_RepositoriesService_UpdatePackageRepository_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateContext(ctx, mux, req, "/kubeappsapis.core.packages.v1alpha1.RepositoriesService/UpdatePackageRepository", runtime.WithHTTPPathPattern("/core/packages/v1alpha1/repositories/plugin/{package_repo_ref.plugin.name}/{package_repo_ref.plugin.version}/c/{package_repo_ref.context.cluster}/ns/{package_repo_ref.context.namespace}/{package_repo_ref.identifier=**}"))
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}
		resp, md, err := request_RepositoriesService_UpdatePackageRepository_0(rctx, inboundMarshaler, client, req, pathParams)
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}

		forward_RepositoriesService_UpdatePackageRepository_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})

	return nil
}

var (
	pattern_RepositoriesService_AddPackageRepository_0 = runtime.MustPattern(runtime.NewPattern(1, []int{2, 0, 2, 1, 2, 2, 2, 3}, []string{"core", "packages", "v1alpha1", "repositories"}, ""))

	pattern_RepositoriesService_GetPackageRepositoryDetail_0 = runtime.MustPattern(runtime.NewPattern(1, []int{2, 0, 2, 1, 2, 2, 2, 3, 2, 4, 1, 0, 4, 1, 5, 5, 1, 0, 4, 1, 5, 6, 2, 7, 1, 0, 4, 1, 5, 8, 2, 9, 1, 0, 4, 1, 5, 10, 3, 0, 4, 1, 5, 11}, []string{"core", "packages", "v1alpha1", "repositories", "plugin", "package_repo_ref.plugin.name", "package_repo_ref.plugin.version", "c", "package_repo_ref.context.cluster", "ns", "package_repo_ref.context.namespace", "package_repo_ref.identifier"}, ""))

	pattern_RepositoriesService_GetPackageRepositorySummaries_0 = runtime.MustPattern(runtime.NewPattern(1, []int{2, 0, 2, 1, 2, 2, 2, 3}, []string{"core", "packages", "v1alpha1", "repositories"}, ""))

	pattern_RepositoriesService_UpdatePackageRepository_0 = runtime.MustPattern(runtime.NewPattern(1, []int{2, 0, 2, 1, 2, 2, 2, 3, 2, 4, 1, 0, 4, 1, 5, 5, 1, 0, 4, 1, 5, 6, 2, 7, 1, 0, 4, 1, 5, 8, 2, 9, 1, 0, 4, 1, 5, 10, 3, 0, 4, 1, 5, 11}, []string{"core", "packages", "v1alpha1", "repositories", "plugin", "package_repo_ref.plugin.name", "package_repo_ref.plugin.version", "c", "package_repo_ref.context.cluster", "ns", "package_repo_ref.context.namespace", "package_repo_ref.identifier"}, ""))
)

var (
	forward_RepositoriesService_AddPackageRepository_0 = runtime.ForwardResponseMessage

	forward_RepositoriesService_GetPackageRepositoryDetail_0 = runtime.ForwardResponseMessage

	forward_RepositoriesService_GetPackageRepositorySummaries_0 = runtime.ForwardResponseMessage

	forward_RepositoriesService_UpdatePackageRepository_0 = runtime.ForwardResponseMessage
)
