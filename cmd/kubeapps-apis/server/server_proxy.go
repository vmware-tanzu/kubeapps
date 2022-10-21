// Copyright 2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"crypto/tls"
	"fmt"
	"github.com/mwitkow/grpc-proxy/proxy"
	httpclient "github.com/vmware-tanzu/kubeapps/pkg/http-client"
	"github.com/vmware-tanzu/kubeapps/pkg/kube"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/grpclog"
	grpcmetadata "google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	log "k8s.io/klog/v2"
	"net"
	"net/url"
	"regexp"
	"strconv"
)

const ProxyRequestHeader = "x-kubeapps-proxy"

// Only these characters are allowed on cluster names
var clusterNameRegex = regexp.MustCompile(`^[a-zA-Z\d-.]+$`)

func getProxyHandler(clustersConfig kube.ClustersConfig) proxy.StreamDirector {
	return func(ctx context.Context, fullMethodName string) (context.Context, *grpc.ClientConn, error) {
		// Find out to which cluster the request will be proxied
		targetCluster, err := getTargetClusterName(ctx)
		if err != nil {
			return nil, nil, status.Errorf(codes.InvalidArgument, "%v", err)
		}

		// Check cluster's configuration
		targetClusterConf, err := getClusterConfiguration(clustersConfig, targetCluster)
		if err != nil {
			return nil, nil, status.Errorf(codes.Internal, "unable to get cluster configuration when proxying. [%v]", err)
		}

		// Prepare outgoing context for the proxied request
		outgoingCtx, err := createOutgoingContext(ctx)
		if err != nil {
			return nil, nil, status.Errorf(codes.Internal, "%v", err)
		}

		grpcEndpoint, err := parseGrpcEndpoint(targetClusterConf.Ingress.Endpoint)
		if err != nil {
			return nil, nil, status.Errorf(codes.Internal, "unable to parse target cluster URL. [%v]", err)
		}

		// Create the connection to the target GRPC endpoint
		conn, err := createProxyConnection(outgoingCtx, grpcEndpoint, targetClusterConf.Ingress.CertificateAuthorityDataDecoded)
		if err != nil {
			return nil, nil, status.Errorf(codes.Internal, "error creating proxy connection [%v]", err)
		}

		log.Infof("++proxy GRPC req for method [%s] to [%s]", fullMethodName, grpcEndpoint)
		return outgoingCtx, conn, err
	}
}

func getTargetClusterName(ctx context.Context) (string, error) {
	md, ok := grpcmetadata.FromIncomingContext(ctx)
	if !ok {
		return "", fmt.Errorf("unable to obtain metadata from incoming context of proxied request")
	}
	if len(md[ProxyRequestHeader]) > 0 {
		clusterName := md[ProxyRequestHeader][0]
		if clusterName == "" || !isValidClusterName(clusterName) {
			return "", fmt.Errorf("provided cluster name [%q] is not valid", clusterName)
		}
		return clusterName, nil
	}
	return "", fmt.Errorf("no header [%s] is provided for proxying", ProxyRequestHeader)
}

func getClusterConfiguration(clustersConfig kube.ClustersConfig, clusterName string) (*kube.ClusterConfig, error) {
	targetClusterConf, ok := clustersConfig.Clusters[clusterName]
	if !ok {
		return nil, fmt.Errorf("cluster %q has no configuration", clusterName)
	}
	if targetClusterConf.Ingress.Endpoint == "" {
		return nil, fmt.Errorf("cluster %q has no ingress endpoint defined", clusterName)
	}
	return &targetClusterConf, nil
}

func createOutgoingContext(ctx context.Context) (context.Context, error) {
	md, ok := grpcmetadata.FromIncomingContext(ctx)
	if !ok {
		return nil, fmt.Errorf("unable to obtain metadata from incoming context of proxied request")
	}

	newMd := md.Copy()

	// remove the proxy header
	// metadata is always lowercase
	if len(newMd[ProxyRequestHeader]) > 0 {
		newMd.Delete(ProxyRequestHeader)
	}
	return grpcmetadata.NewOutgoingContext(ctx, newMd), nil
}

func createProxyConnection(ctx context.Context, endpoint, certificateAuthorityData string) (*grpc.ClientConn, error) {
	tlsConfig := &tls.Config{}

	// Add the certificate authority if specified
	if certificateAuthorityData != "" {
		caBytes := []byte(certificateAuthorityData)
		cp, err := httpclient.GetCertPool(caBytes)
		if err != nil {
			return nil, err
		}
		tlsConfig.RootCAs = cp
	}

	// Make sure we use DialContext so the dialing can be cancelled/timed out together with the context
	conn, err := grpc.DialContext(ctx, endpoint, grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))
	if err != nil {
		return nil, fmt.Errorf("unable to dial to proxied grpc service: [%v]", err)
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
	return conn, nil
}

func isValidClusterName(clusterName string) bool {
	return clusterNameRegex.MatchString(clusterName)
}

func parseGrpcEndpoint(endpoint string) (string, error) {
	u, err := url.Parse(endpoint)
	if err != nil {
		return "", fmt.Errorf("unable to obtain a valid GRPC endpoint from [%s]", endpoint)
	}

	if u.Host == "" {
		return parseGrpcEndpoint("https://" + endpoint)
	}

	port := u.Port()

	// If no port provided, it must be inferred
	if port == "" {
		if u.Scheme != "" {
			intPort, err := net.LookupPort("tcp", u.Scheme)
			if err != nil {
				return "", fmt.Errorf("unable to lookup a valid port for scheme [%s]. %v", u.Scheme, err)
			}
			port = strconv.Itoa(intPort)
		} else {
			return "", fmt.Errorf("either port or scheme should be provided for the cluster URL [%s]", endpoint)
		}
	}
	return fmt.Sprintf("%s:%s", u.Hostname(), port), nil
}
