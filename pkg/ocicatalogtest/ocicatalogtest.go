// Copyright 2023 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package ocicatalogtest

import (
	"net"
	"testing"

	ocicatalog "github.com/vmware-tanzu/kubeapps/cmd/oci-catalog/gen/catalog/v1alpha1"
	"google.golang.org/grpc"
)

type OCICatalogDouble struct {
	Repositories []ocicatalog.Repository
	Tags         []ocicatalog.Tag
	ocicatalog.UnimplementedOCICatalogServiceServer
}

// Dummy implementation that just sends the canned repositories.
func (c OCICatalogDouble) ListRepositoriesForRegistry(r *ocicatalog.ListRepositoriesForRegistryRequest, stream ocicatalog.OCICatalogService_ListRepositoriesForRegistryServer) error {
	for _, repo := range c.Repositories {
		stream.Send(&repo)
	}
	return nil
}

func (c OCICatalogDouble) ListTagsForRepository(r *ocicatalog.ListTagsForRepositoryRequest, stream ocicatalog.OCICatalogService_ListTagsForRepositoryServer) error {
	for _, tag := range c.Tags {
		stream.Send(&tag)
	}
	return nil
}

// SetupTestDouble starts the gRPC service, returns the test double and a
// function to cleanup
func SetupTestDouble(t *testing.T) (string, *OCICatalogDouble, func()) {
	lis, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatalf("%+v", err)
	}

	catalogDouble := &OCICatalogDouble{}
	grpcServer := grpc.NewServer()
	ocicatalog.RegisterOCICatalogServiceServer(grpcServer, catalogDouble)

	go func() {
		grpcServer.Serve(lis)
	}()

	return lis.Addr().String(), catalogDouble, func() {
		grpcServer.Stop()
		lis.Close()
	}
}
