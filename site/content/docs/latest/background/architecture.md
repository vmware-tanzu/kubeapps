# The Kubeapps Overview

This document describes the Kubeapps architecture at a high level.

## Components

### Kubeapps dashboard

At the heart of Kubeapps is an in-cluster Kubernetes dashboard that provides you a simple browse and click experience for installing and managing Kubernetes applications packaged as Helm charts.

The dashboard is written in the JavaScript programming language and is developed using the React JavaScript library.

Please refer to the [Kubeapps Dashboard Developer Guide](../reference/developer/dashboard.md) for the developer setup.

### Kubeapps-APIs

The Kubeapps APIs service provides a pluggable, gRPC-based API service enabling the Kubeapps UI (or other clients) to interact with different Kubernetes packaging formats in a consistent, extensible way.

You can read more details about the architecture, implementation and getting started in the [Kubeapps APIs developer documentation](../reference/developer/kubeapps-apis.md).

### Apprepository CRD and Controller

Chart repositories in Kubeapps are managed with a `CustomResourceDefinition` called `apprepositories.kubeapps.com`. Each repository added to Kubeapps is an object of type `AppRepository` and the `apprepository-controller` will watch for changes on those types of objects to update the list of available charts to deploy.

Please refer to the [Kubeapps Apprepository Controller Developer Guide](../reference/developer/apprepository-controller.md) for the developer setup.

### asset-syncer

The `asset-syncer` component is a tool that scans a Helm chart repository and populates chart metadata in a database. This metadata is then served by the Helm plugin of the `kubeapps-apis` component. Check more details about the implementation in this [document](../reference/developer/asset-syncer.md).

### pinniped-proxy

The `pinniped-proxy` service is an optional component that proxies incoming requests with an `Authorization: Bearer token` header, exchanging the token via the pinniped aggregate API for x509 short-lived client certificates, before forwarding the request onwards to the destination k8s API server.Check more details about the implementation in this [document](../reference/developer/pinniped-proxy.md).

Please refer to the [Kubeapps Pinniped Proxy Developer Guide](../reference/developer/pinniped-proxy.md) for the developer setup.

### oci-catalog

The `oci-catalog` service is an optional component that enables Kubeapps to display a catalog of apps for an OCI registry or a namespace of an OCI registry. The proposed implementation is for a stateless gRPC micro-service that can be run (though is not restricted to run) as a side-car of existing the asset-syncer job to provide lists of repositories for a (namespaced) registry, regardless of the registry provider.

Please refer to the [Kubeapps OCI Catalog Developer Guide](../reference/developer/oci-catalog.md) for the developer setup.
