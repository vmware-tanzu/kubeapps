# Deploy and Configure Kubeapps on VMware Tanzu™ Community Edition

## Introduction

[VMware Tanzu™ Community Edition (TCE)](https://tanzucommunityedition.io) is a full-featured, easy-to-manage Kubernetes platform for learners and users. It is a freely available, community supported, open source distribution of VMware Tanzu that can be installed and configured in minutes on your local workstation or your favorite cloud.

[Kubeapps](https://kubeapps.com/) provides a cloud native dashboard to deploy, manage, and upgrade applications on a Kubernetes cluster. It is a one-time install that gives you a number of important benefits, including the ability to:

- browse and deploy packaged applications from public or private chart repositories
- customize deployments through an intuitive, form-based user interface
- upgrade, manage and delete the applications that are deployed in your Kubernetes cluster
- expose an API to manage your package repositories and your applications

Kubeapps can be configured with public catalogs, such as the [VMware Marketplace™](https://marketplace.cloud.vmware.com/) catalog, the [Bitnami Application Catalog](https://bitnami.com/stacks/helm) or with private Helm repositories such as ChartMuseum or Harbor. It also integrates with [VMware Tanzu™ Application Catalog™ (TAC) for Tanzu™ Advanced](https://tanzu.vmware.com/application-catalog), which provides an enterprise-ready Helm chart catalog.

This guide walks you through the process of configuring, deploying and using Kubeapps on a VMware Tanzu™ Community Edition cluster of your choice. It covers the following tasks:

- Selection and deployment of the appropriate TCE cluster type
- Configuring an identity management provider in the cluster
- Configuring access control in Kubeapps
- Deploying and configuring Kubeapps in the TCE cluster
- Management of applications through Kubeapps using the TCE catalog

## Intended Audience

This guide is intended for the following user roles:

- System administrators who want to install Kubeapps on a VMware Tanzu™ Community Edition cluster and use it to deploy and manage applications from any package repository.
- Application administrators and developers who want to use Kubeapps to deploy and manage modern applications in a Kubernetes architecture.
- Any user willing to play around with Kubeapps and TCE

In-depth knowledge of Kubernetes is not required.

## Tutorial structure

The tutorial is organized in the following sections:

1. [TCE cluster decision and creation](./01-TCE-cluster.md)
2. [Deploying Kubeapps](./03-deploying-kubeapps.md)
3. [Configuring web access for Kubeapps](./04-ingress-traffic.md)
4. [Managing applications in Kubeapps](./05-Managing-applications.md)

## Begin

> Begin the tutorial by [installing a TCE cluster](./01-TCE-cluster.md).
