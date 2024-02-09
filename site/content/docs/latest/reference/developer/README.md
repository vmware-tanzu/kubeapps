# The Kubeapps Components

## Kubeapps dashboard

The dashboard is the main UI component of the Kubeapps project. Written in JavaScript, the dashboard uses the React JavaScript library for the frontend.

Please refer to the [Kubeapps Dashboard Developer Guide](./dashboard.md) for the developer setup.

## Kubeapps APIs service

The Kubeapps APIs service is the main backend component of the Kubeapps project. Written in Go, the APIs service provides a pluggable gRPC service that is used to support different Kubernetes packaging formats.

See the [Kubeapps APIs Service Developer Guide](kubeapps-apis.md) for more information.

## asset-syncer

The `asset-syncer` component is a tool that scans a Helm chart repository and populates chart metadata in the database. This metadata is then served by the `kubeapps-apis` component.

Please refer to the [Kubeapps asset-syncer Developer Guide](asset-syncer.md) for the developer setup.

## pinniped-proxy

The `pinniped-proxy` service is an optional component that proxies incoming requests with an `Authorization: Bearer token` header, exchanging the token via the pinniped aggregate API for x509 short-lived client certificates, before forwarding the request onwards to the destination k8s API server.

Please refer to the [Kubeapps pinniped-proxy Developer Guide](pinniped-proxy.md) for the developer setup.

## oci-catalog

The `oci-catalog` service is an optional component that enables Kubeapps to display a catalog of apps for an OCI registry or a namespace of an OCI registry. The proposed implementation is for a stateless gRPC micro-service that can be run (though is not restricted to run) as a side-car of existing the asset-syncer job to provide lists of repositories for a (namespaced) registry, regardless of the registry provider.

Please refer to the [Kubeapps oci-catalog Developer Guide](oci-catalog.md) for the developer setup.
