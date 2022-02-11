# Kubeapps developer guides

### Environment setup

For setting up the environment for running Kubeapps, makefiles are provided.

Please refer to [How to set up the environment using the provided makefile targets](./using-makefiles.md)

### Kubeapps build guide

This guide explains how to build Kubeapps.

Please refer to [Kubeapps Build Guide](./build.md)
### Kubeapps dashboard

The `dashboard` is the main UI component of the Kubeapps project. Written in JavaScript, the dashboard uses the React JavaScript library for the frontend.

Please refer to the [Kubeapps Dashboard Developer Guide](dashboard.md) for the developer setup.

### Kubeapps Pinniped-Proxy

`pinniped-proxy` can be used by our Kubeapps frontend to ensure OIDC requests for the Kubernetes API server are forwarded through only after exchanging the OIDC id token for client certificates used by the Kubernetes API server, for situations where the Kubernetes API server is not configured for OIDC.

Please refer to [Kubeapps Pinniped-Proxy Developer Guide](./pinniped-proxy.md)


### Experimental Kubeapps APIs service

This is an experimental service that we are not yet including in our released Kubeapps chart, providing an extendable protobuf-based API service enabling Kubeapps (or other packaging UIs) to interact with various Kubernetes packaging formats in a consistent way.

Please refer to [Experimental Kubeapps APIs service](./kubeapps-apis.md)
### kubeops

The `kubeops` component is a micro-service that creates an API endpoint for accessing the Helm API and Kubernetes resources.

Please refer to the [Kubeapps Kubeops Developer Guide](kubeops.md) for the developer setup.

### assetsvc

The `assetsvc` component is a micro-service that creates an API endpoint for accessing the metadata for charts in Helm chart repositories that's populated in a Postgresql server.

Please refer to the [Kubeapps assetsvc Developer Guide](assetsvc.md) for the developer setup.

### asset-syncer

The `asset-syncer` component is a tool that scans a Helm chart repository and populates chart metadata in the database. This metadata is then served by the `assetsvc` component.

Please refer to the [Kubeapps asset-syncer Developer Guide](asset-syncer.md) for the developer setup.

### Kubeapps apprepository-controller

The `apprepository-controller` is a Kubernetes controller for managing Helm chart repositories added to Kubeapps.

Please refer to [Kubeapps apprepository-controller](./apprepository-controller.md)

### Kubeapps Releases Developer Guide

The purpose of this section is to guide maintainers through the process of releasing a new version of Kubeapps.

Please refer to [Kubeapps Releases Developer Guide](./release-process.md)


### Kubeapps issue triage

The purpose of this section is to guide maintainers to triage new issues and users to understand the triage process followed by Kubeapps team.

Please refer to [Kubeapps triage process](./issue-triage-process.md)