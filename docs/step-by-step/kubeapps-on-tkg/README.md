# Kubeapps on VMware Tanzu™ Kubernetes Grid

In this guide, we will walk you through performing a ready-to-use configuration and using Kubeapps on a VMware Tanzu™ Kubernetes cluster (TKC).

These clusters can be spun up by [VMware Tanzu™ Kubernetes Grid™ (TKG)](https://tanzu.vmware.com/kubernetes-grid), an enterprise-ready Kubernetes runtime that streamlines operations across a multi-cloud infrastructure. It can run both on-premise in vSphere and in the public cloud and includes signed and supported versions of open source applications to provide all the key services required for a production Kubernetes environment.

One of the key aspects of Day 2 operations is managing your workloads. In this scenario, to simplify application deployment and management on Kubernetes, [Kubeapps](https://kubeapps.com/) provides a web-based dashboard to deploy, manage, and upgrade applications on a Kubernetes cluster.

Kubeapps can be configured with public catalogs, such as the [VMware Marketplace™](https://marketplace.cloud.vmware.com/) or the [Bitnami Application Catalog](https://bitnami.com/stacks/helm). But also, you can configure Kubeapps to use your private application repository as its source.
This feature gives you the option of extending your catalog with your charts located in a private Helm repository such as ChartMuseum or Harbor, and even to use your enterprise-ready customized Helm chart catalog directly from the [VMware Tanzu™ Application Catalog™ (TAC) for Tanzu™ Advanced](https://tanzu.vmware.com/application-catalog).

This tutorial will specifically show you how to: i) successfully install Kubeapps with ready-to-use values, leveraging the existing identity management capabilities in TKG; ii) configure two application repositories: the public [VMware Marketplace™](https://marketplace.cloud.vmware.com/) and your private [VMware Tanzu™ Application Catalog™ for Tanzu™ Advanced](https://tanzu.vmware.com/application-catalog) instance.

## Assumptions and Prerequisites

This guide will assume henceforth that:

- You have a VMware Tanzu™ Kubernetes Grid v1.3.1 or later cluster. Check out the [VMware Tanzu™ Kubernetes Grid 1.3 Documentation](https://docs.vmware.com/en/VMware-Tanzu-Kubernetes-Grid/1.3/vmware-tanzu-kubernetes-grid-13/GUID-index.html) for more information.
- You have access to the [VMWare Cloud Services Portal (CSP)](https://console.cloud.vmware.com/). If not, talk to your [VMware sales representative](https://www.vmware.com/company/contact_sales.html) to request access.
- You have access to, at least, the Tanzu™ Application Catalog™ for Tanzu™ Advanced Demo environment. If not, reach out to your [VMware sales representative](https://www.vmware.com/company/contact_sales.html).
- You have the _kubectl_ CLI and the Helm v3.x package manager installed. Learn how to [install kubectl and Helm v3.x](https://docs.bitnami.com/kubernetes/get-started-kubernetes/#step-3-install-kubectl-command-line).

## Intended Audience

This information is intended for administrators who want to install Kubeapps on a Tanzu Kubernetes Grid cluster and use it to deploy and manage applications from the VMware Marketplace™ and the VMware Tanzu™ Application Catalog™ for Tanzu™ Advanced. Besides, this content is also intended for application administrators and developers who want to use Kubeaps to deploy and manage modern apps in a Kubernetes architecture. Nonetheless, in-depth knowledge of Kubernetes is not required.

## Steps

1. [Step 1 - Configure an Identity Management Provider in your Cluster](./step-1.md)
2. [Step 2 - Configure and Install Kubeapps](./step-2.md)
3. [Step 3 - Add Application Repositories](./step-3.md)
5. [Step 4 - Deploy an Application](./step-4.md)

# What to Do Next

Check out the [Useful Links](#useful-links) section to learn more about TKG, TAC and other advanced features in Kubeapps.

# Useful Links

- [Enabling Identity Management in Tanzu Kubernetes Grid](https://docs.vmware.com/en/VMware-Tanzu-Kubernetes-Grid/1.3/vmware-tanzu-kubernetes-grid-13/GUID-mgmt-clusters-enabling-id-mgmt.html)
- [Installing Kubeapps in airgapped environments](https://github.com/kubeapps/kubeapps/blob/master/docs/user/offline-installation.md)
- [Syncing app repositories using webhooks](https://github.com/kubeapps/kubeapps/blob/master/docs/user/syncing-apprepository-webhook.md)
- [Using Kubeapps to deploy in multiple clusters](https://github.com/kubeapps/kubeapps/blob/master/docs/user/deploying-to-multiple-clusters.md)
- [Using Operators in Kubeapps](https://github.com/kubeapps/kubeapps/blob/master/docs/user/operators.md)
