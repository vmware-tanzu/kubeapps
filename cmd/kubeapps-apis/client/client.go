package main

import (
	"context"
	"log"

	packagesGRPCv1alpha1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	"github.com/kubeapps/kubeapps/cmd/kubeapps-apis/interceptors"
	"google.golang.org/grpc"
)

func main() {
	ctx := context.Background()

	// Create a connection and add our ClientPingCounter interceptor as a UnaryInterceptor to the connection
	conn, err := grpc.DialContext(ctx, "localhost:50051", grpc.WithInsecure(), grpc.WithBlock(), grpc.WithChainUnaryInterceptor(
		//interceptors.Identity{ID: "kubeapps-apis-client-junk"}.UnaryClient,
		interceptors.Identity{ID: "kubeapps-apis-client"}.UnaryClient,
	))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	// A new GRPC client to use
	client := packagesGRPCv1alpha1.NewPackagesServiceClient(conn)

	request := &packagesGRPCv1alpha1.GetAvailablePackageSummariesRequest{
		Context: &packagesGRPCv1alpha1.Context{
			Cluster:   "",
			Namespace: "kubeapps",
		},
	}

	availablePackageSummaries, err := client.GetAvailablePackageSummaries(ctx, request)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(availablePackageSummaries)
}
