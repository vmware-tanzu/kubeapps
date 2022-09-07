# Deploy and Configure Kubeapps on VMware Tanzu™ Community Edition

## Introduction

[VMware Tanzu™ Community Edition (TCE)](https://tanzucommunityedition.io) is a full-featured, easy-to-manage Kubernetes platform for learners and users. It is a freely available, community supported, open source distribution of VMware Tanzu that can be installed and configured in minutes on your local workstation or your favorite cloud.

[Kubeapps](https://kubeapps.com/) provides a cloud native dashboard to deploy, manage, and upgrade applications on a Kubernetes cluster. It is a one-time install that gives you a number of important benefits, including the ability to:

- browse and deploy packaged applications from public or private repositories
- customize deployments through an intuitive user interface
- upgrade, manage and delete the applications that are deployed in your Kubernetes cluster
- expose an API to manage your package repositories and your applications

Kubeapps can be configured with public catalogs, such as the [VMware Marketplace™](https://marketplace.cloud.vmware.com/) catalog, the [Bitnami Application Catalog](https://bitnami.com/application-catalog) or with private Helm repositories such as ChartMuseum or Harbor. It also integrates with [VMware Application Catalog](https://tanzu.vmware.com/application-catalog), which provides an enterprise-ready Helm chart catalog.

This guide walks you through the process of configuring, deploying and using Kubeapps on a VMware Tanzu™ Community Edition cluster of your choice. It covers the following tasks:

- Selection and deployment of the appropriate TCE cluster type
- Configuring an identity management provider in the cluster
- Configuring access control in Kubeapps
- Deploying and configuring Kubeapps in the TCE cluster
- Management of applications through Kubeapps using the TCE catalog

## Intended Audience

This guide is intended for the following user roles:

- **System administrators** who want to install Kubeapps on a VMware Tanzu™ Community Edition cluster and use it to deploy and manage applications from any package repository.
- **Application administrators** and **developers** who want to use Kubeapps to deploy and manage modern applications in a Kubernetes environment.
- Any user willing to play around with Kubeapps and TCE

In-depth knowledge of Kubernetes is not required.

## Tutorial index

The tutorial is organized in the following sections:

1. [TCE cluster deployment preparation](./01-TCE-cluster-preparation.md)
2. [Deploying a managed cluster](./02-TCE-managed-cluster.md) or [Deploy an unmanaged cluster](./02-TCE-unmanaged-cluster.md)
3. [Preparing the Kubeapps deployment](./03-preparing-kubeapps-deployment.md)
4. [Deploying Kubeapps](./04-deploying-kubeapps.md)
5. [Further documentation for managing applications in Kubeapps](./05-managing-applications.md)

## Begin

> Begin the tutorial by [preparing your TCE cluster deployment](./01-TCE-cluster-preparation.md).
