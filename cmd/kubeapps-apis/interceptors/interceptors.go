package interceptors

import (
	"context"

	log "k8s.io/klog/v2"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type Identity struct {
	ID string
}

func (i Identity) UnaryClient(
	ctx context.Context,
	method string,
	req interface{},
	reply interface{},
	cc *grpc.ClientConn,
	invoker grpc.UnaryInvoker,
	opts ...grpc.CallOption,
) error {
	md := metadata.Pairs()
	md.Set("client-id", i.ID)

	ctx = metadata.NewOutgoingContext(ctx, md)

	err := invoker(ctx, method, req, reply, cc, opts...)

	return err
}

func VerifyUnaryServer(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (resp interface{}, err error) {

	// clients can use kubeapps-apis-client in grpc.DialContext
	// that can be verified here

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok || len(md["client-id"]) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "missing metadata")
	}

	if md["client-id"][0] != "kubeapps-apis-client" {
		return nil, status.Error(codes.PermissionDenied, "unexpected client")
	}
	log.Info("called VerifyUnaryServer Interceptor")
	res, err := handler(ctx, req)
	return res, err
}

// LogRequest is a gRPC UnaryServerInterceptor that will log the API call to stdOut
func LogRequest(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (response interface{}, err error) {
	log.Infof("Interceptor Request for : %s\n", info.FullMethod)
	res, err := handler(ctx, req)
	return res, err
}
