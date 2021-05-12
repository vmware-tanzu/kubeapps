# Kubeapps on Tanzu Kubernetes Grid

In this guide, we will walk you through performing a production-ready configuration and using Kubeapps on Tanzu Kubernetes Grid (TKG) clusters.

[VMware Tanzu™ Kubernetes Grid™](https://tanzu.vmware.com/kubernetes-grid) is an enterprise-ready Kubernetes runtime that streamlines operations across a multi-cloud infrastructure. It can run both on-premise in vSphere and in the public cloud and includes signed and supported versions of open source applications to provide all the key services required for a production Kubernetes environment.

One of the key aspects of Day 2 operations is managing your workloads. In this scenario, to simplify application deployment and management on Kubernetes, [Kubeapps](https://kubeapps.com/) provides a web-based dashboard to deploy, manage, and upgrade applications on a Kubernetes cluster.

Kubeapps includes a built-in catalog of Helm charts and operators that allows users to deploy trusted and continuously maintained content on their clusters. In addition, you can configure Kubeapps to use your private application repository as its source. This feature gives you the option of extending your catalog with your charts located in a [private Helm repository](https://github.com/kubeapps/kubeapps/blob/master/docs/user/private-app-repository.md) such as ChartMuseum or Harbor, and even to use your customized Helm chart catalog directly from the [VMware Tanzu™ Application Catalog™](https://tanzu.vmware.com/application-catalog).

This tutorial will show you how to: i) successfully install Kubeapps with production-ready values, leveraging the existing identity management capabilities in TKG; ii) configure two application repositories: the public [VMware Marketplace™](https://marketplace.cloud.vmware.com/) and your private [VMware Tanzu™ Application Catalog™ (TAC)](https://tanzu.vmware.com/application-catalog) instance.

## Assumptions and prerequisites

As this guide will cover the installation of Kubeapps in a [VMware Tanzu™ Kubernetes Grid™](https://tanzu.vmware.com/kubernetes-grid) cluster and the configuration of an application repository for [VMware Tanzu™ Application Catalog™ (TAC)](https://tanzu.vmware.com/application-catalog), it assumes that you have access to a TKG cluster as well as a TAC pre-built or custom catalog. If you do not have access, please reach out to your [VMware sales representative](https://www.vmware.com/company/contact_sales.html).

This tutorial, therefore, assumes that:

- You have a Tanzu™ Kubernetes Grid v1.3.1 or later cluster properly configured with identy management. Check out the [VMware Tanzu™ Kubernetes Grid 1.3 Documentation](https://docs.vmware.com/en/VMware-Tanzu-Kubernetes-Grid/1.3/vmware-tanzu-kubernetes-grid-13/GUID-index.html) and the [Enabling Identity Management in Tanzu Kubernetes Grid](https://docs.vmware.com/en/VMware-Tanzu-Kubernetes-Grid/1.3/vmware-tanzu-kubernetes-grid-13/GUID-mgmt-clusters-enabling-id-mgmt.html) guide.
- You have access to the [VMWare Cloud Services Portal (CSP)](https://console.cloud.vmware.com/). If not, talk to your [VMware sales representative](https://www.vmware.com/company/contact_sales.html) to request access.
- You have access to, at least, the Tanzu™ Application Catalog™ Demo environment. If not, reach out to your [VMware sales representative](https://www.vmware.com/company/contact_sales.html).
- You have the _kubectl_ CLI and the Helm v3.x package manager installed. Learn how to [install kubectl and Helm v3.x](https://docs.bitnami.com/kubernetes/get-started-kubernetes/#step-3-install-kubectl-command-line).

## Intended Audience

This information is intended for administrators who want to install Kubeapps on a Tanzu Kubernetes Grid cluster and use it to deploy and manage applications from the VMware Marketplace™ and the VMware Tanzu™ Application Catalog™. Besides, this content is also intended for application administrators and developers who want to use Kubeaps to deploy and manage modern apps in a Kubernetes architecture.

The information is written for users who have a basic understanding of Kubernetes and are familiar with container deployment concepts. In-depth knowledge of Kubernetes is not required.

# Steps

<!-- To be extracted to a separate file -->

## Step 1: Prepare the identity management

<!-- Work in progress -->

### Install or use the existing Pinniped in TKG?

- [Enabling Identity Management in Tanzu Kubernetes Grid](https://docs.vmware.com/en/VMware-Tanzu-Kubernetes-Grid/1.3/vmware-tanzu-kubernetes-grid-13/GUID-mgmt-clusters-enabling-id-mgmt.html).
- [Tanzu Kubernetes Grid 1.3 with Identity Management](https://liveandletlearn.net/post/kubeapps-on-tanzu-kubernetes-grid-13/)

### Configure an OIDC provider

- [VMware Cloud Services as OIDC provider](https://github.com/kubeapps/kubeapps/blob/master/docs/user/using-an-OIDC-provider.md#vmware-cloud-services)

<!-- To be extracted to a separate file -->

## Step 2: Configure and install Kubeapps

<!-- Work in progress -->

- [Configure Pinniped to trust the ODIC provider](https://github.com/kubeapps/kubeapps/blob/master/docs/user/using-an-OIDC-provider-with-pinniped.md#configure-pinniped-to-trust-your-oidc-identity-provider)
- [Configuring Kubeapps to proxy requests via Pinniped](https://github.com/kubeapps/kubeapps/blob/master/docs/user/using-an-OIDC-provider-with-pinniped.md#configuring-kubeapps-to-proxy-requests-via-pinniped)
- Use the VMware custom look&feel <!-- Undocumented -->

<!-- To be extracted to a separate file -->

## Step 3: Add the VMware Marketplace™ as an application repository

<!-- Work in progress -->

- [Adding an public application repository](https://github.com/kubeapps/kubeapps/blob/master/docs/user/dashboard.md)

<!-- To be extracted to a separate file -->

## Step 4: Add the VMware Tanzu™ Application Catalog™ as an application repository

<!-- Work in progress -->

- [Consume Tanzu Application Catalog Helm Charts using Kubeapps](https://docs.vmware.com/en/VMware-Tanzu-Application-Catalog/services/tac-docs/GUID-using-tac-consume-tac-kubeapps.html)
- [Adding an private application repository](https://github.com/kubeapps/kubeapps/blob/master/docs/user/private-app-repository.md)
