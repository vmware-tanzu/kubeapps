# Step 1: VMware Tanzu™ Community Edition cluster preparation

VMware Tanzu™ Community Edition consists of a variety of components that enable the bootstrapping and management of Kubernetes clusters.

The main components are:

- **Tanzu CLI**

    A CLI installed on your local machine and then used to operate with clusters on your chosen target platform.

- **Managed clusters**

    This is the primary deployment model for clusters in the Tanzu ecosystem and is recommended for production scenarios.

- **Unmanaged clusters**

    Offer a single node, local workstation cluster suitable for a development/test environment.

In this step of the tutorial, we will prepare the installation of the TCE cluster.

## Step 1.1: Install Tanzu CLI

Tanzu CLI can be installed on the three major operating systems (Linux, MacOS and Windows).

[Follow the instructions here to install Tanzu CLI selecting your operating system](https://tanzucommunityedition.io/docs/v0.12/cli-installation/).

In order to check that Tanzu CLI is properly installed, run this in the command line:

```bash
tanzu version
```

And the output should be similar to this:

```bash
version: v0.11.4
buildDate: 2022-05-17
sha: a9b8f3a
```

## Step 1.2: Choose the type of your cluster

Tanzu Community Edition has different flavours of clusters, that can adapt to different needs.

You must decide, which of the following suits your desired outcome:

- Do you want a single node, local workstation cluster suitable for a development/test environment? If so, deploy an [unmanaged cluster](./02-TCE-unmanaged-cluster.md).

or

- Do you want a full-featured, scalable Kubernetes implementation suitable for a development or production environment? If so, deploy a [managed cluster](./02-TCE-managed-cluster.md).

If you want to know more about planning your deployment check out [the official TCE documentation](https://tanzucommunityedition.io/docs/v0.12/planning/).
