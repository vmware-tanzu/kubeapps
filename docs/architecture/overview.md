# The Kubeapps Overview

This document describes the Kubeapps architecture at a high level.

## Components

### Kubeapps dashboard

At the heart of Kubeapps is a in-cluster Kubernetes dashboard that provides you a simple browse and click experience for installing and manage Kubernetes applications packaged as Helm charts.

Additionally, the dashboard integrates with the [Kubernetes service catalog](https://github.com/kubernetes-incubator/service-catalog) and enables you to browse and provision cloud services via the [Open Service Broker API](https://github.com/openservicebrokerapi/servicebroker).

The dashboard is written in the Javascript programming language and is developed using the React Javascript library.

### Tiller proxy

In order to secure the access to Tiller and allow the dashboard to contact the Helm Tiller server we deploy a proxy that handles the communication with Tiller. The goal of this proxy is to validate that the user doing the request has sufficent permissions to create or delete all the resources that are part of the specific chart being deployed or deleted.

This proxy is written in Go. Check more details about the implementation in this [document](/cmd/tiller-proxy/README.md).

### Apprepository CRD and Controller

Chart repositories in Kubeapps are managed with a `CustomResourceDefinition` called `apprepositories.kubeapps.com`. Each repository added to Kubeapps is an object of type `AppRepository` and the `apprepository-controller` will watch for changes on those type of objects to update the list of available charts to deploy.

### `chart-repo`

The `chart-repo` component is tool that scans a Helm chart repository and populates chart metadata in a MongoDB database. This metadata is then served by the chartsvc component. It is maintained as part of the [Helm Monocular project](https://github.com/helm/monocular/tree/master/cmd/chart-repo).

### `chartsvc`

The `chartsvc` component is a micro-service that creates an API endpoint for accessing the metadata for charts in Helm chart repositories that's populated in a MongoDB database. It is maintained as part of the [Helm Monocular project](https://github.com/helm/monocular/tree/master/cmd/chartsvc).
