# The Kubeapps Overview

This document describes the Kubeapps architecture at a high level.

## The purpose of Kubeapps

Kubeapps is a Kubernetes dashboard that provides simple browse and click deployment of applications and Kubeless functions. The Kubeapps dashboard can do the following:

* List charts from Helm chart repositories
* Install and uninstall Helm charts from repositories
* Manage installed chart releases
* Create, edit and test Kubeless functions
* Browse, provision and manage external cloud services

## Components

### Kubeapps installer

The Kubeapps installer is a command-line tool for installing, upgrading and uninstalling the Kubeapps in-cluster components. The tool is written in the Go programming language and the Kubernetes manifests are written in the Jsonnet data templating language.

### Kubeapps dashboard

At the heart of Kubeapps is a in-cluster Kubernetes dashboard that provides you a simple browse and click experience for installing applications packaged as Helm charts. It also provides you a simple and easy to use web interface to develop, test and deploy Kubeless functions.

Additionally, the dashboard integrates with the [Kubernetes service catalog](https://github.com/kubernetes-incubator/service-catalog) and enables you to browse and provision cloud services via the [Open Service Broker API](https://github.com/openservicebrokerapi/servicebroker).

The dashboard is written in the Javascript programming language and is developed using the React Javascript library.

### Helm CRD

[Helm CRD](https://github.com/bitnami-labs/helm-crd) is an experimental Kubernetes controller for Helm releases. It enables cluster operators to secure the Helm tiller server and is used to install, upgrade and uninstall Helm charts.

The controller is developed independently of Kubeapps at [bitnami-labs/helm-crd](https://github.com/bitnami-labs/helm-crd) and is written in the Go programming language.

### Kubeless

[Kubeless](http://kubeless.io/) is a Kubernetes-native serverless framework that lets you deploy small bits of code (functions) without having to worry about the underlying infrastructure. It leverages Kubernetes resources to provide auto-scaling, API routing, monitoring, troubleshooting and more.

Kubeless is also written in the Go programming language and is also developed independently of Kubeapps at [kubeless/kubeless](https://github.com/kubeless/kubeless).
