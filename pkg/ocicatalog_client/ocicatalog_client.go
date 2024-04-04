// Copyright 2023-2024 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package ocicatalog_client

import (
	"fmt"

	ocicatalog "github.com/vmware-tanzu/kubeapps/cmd/oci-catalog/gen/catalog/v1alpha1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	CONTENT_TYPE_HELM = "helm"
)

func NewClient(ociCatalogAddr string) (ocicatalog.OCICatalogServiceClient, func(), error) {
	if ociCatalogAddr == "" {
		return nil, nil, fmt.Errorf("ociCatalogAddr must be specified")
	}
	conn, err := grpc.NewClient(ociCatalogAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, fmt.Errorf("unable to contact OCI Catalog at %q: %+v", ociCatalogAddr, err)
	}

	closer := func() { conn.Close() }

	return ocicatalog.NewOCICatalogServiceClient(conn), closer, nil
}
