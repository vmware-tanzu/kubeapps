# Managing Carvel Packages with Kubeapps

> **NOTE**: This guide is about a feature that is under active development and, therefore, not yet ready for production use. The information herein stated is subject to change without notice until we reach a stable release.

## Table of Contents

1. [Introduction](#introduction)
2. [Installing kapp-controller in your cluster](#installing-kapp-controller-in-your-cluster)
   1. [Quick overview of the kapp-controller CRs](#quick-overview-of-the-kapp-controller-crs)
3. [xxxxx](#xxxxxx)

## Introduction

Historically, Kubeapps was initially developed to solely manage [Helm Charts](https://helm.sh) on your Kubernetes clusters. However, it has evolved to support multiple formats, such as [Carvel Packages](https://carvel.dev/kapp-controller/docs/latest/packaging/#package) and [Helm releases via Fluxv2](https://fluxcd.io/docs/guides/helmreleases/).

> Find more information about the architectural evolution at [this video](https://www.youtube.com/watch?v=rS2AhcIPQEs) and [this technical documentation](../developer/kubeapps-apis.md).

[Carvel](https://carvel.dev/) is often defined as a set of reliable, single-purpose, composable tools that aid in your application building, configuration, and deployment to Kubernetes.
Particularly, two of these tools have paramount importance for Kubeapps: [kapp](https://carvel.dev/kapp/) and [kapp-controller](https://carvel.dev/kapp-controller/).

On the one hand, [kapp](https://carvel.dev/kapp/) is a CLI for deploying and viewing groups of Kubernetes resources as [Applications](https://carvel.dev/kapp/docs/latest/apps/). On the other hand, [kapp-controller](https://carvel.dev/kapp-controller/) is a controller for managing the lifecycle of those applications, which also helps package software into distributable [Packages](https://carvel.dev/kapp-controller/docs/latest/packaging/#package) and enables users to discover, configure, and install these packages on a Kubernetes cluster.

This guide walks you through the process of using Kubeapps for configuring and deploying [Packages](https://carvel.dev/kapp-controller/docs/latest/packaging/#package) and managing [Applications](https://carvel.dev/kapp/docs/latest/apps/).

## Installing kapp-controller in your cluster

> **NOTE**: This section can safely be skipped if you already have kapp-controller installed in your cluster.

In order to manage Carvel Packages, first of all, you will need to install kapp-controller in your cluster. That is, applying a set of Kubernetes resources and CRDs.

According to the [Carvel kapp-controller official documentation](https://carvel.dev/kapp-controller/docs/latest/install/), you can install everything it kapp-controller requires just by running the following command:

```bash
kubectl apply -f https://github.com/vmware-tanzu/carvel-kapp-controller/releases/latest/download/release.yml
```

After running this command, you should have everything you need to manage Carvel Packages in your cluster. Next section will give you an overview of the relevant Custom Resources included in kapp-controller.

### Quick overview of the kapp-controller CRs

At the time of writing this guide, kapp-controller will install the following Custom Resources Definitions:

- [PackageRepository](https://carvel.dev/kapp-controller/docs/latest/packaging/#package-repository): is a collection of packages and their metadata. Similar to a maven repository or a rpm repository, adding a package repository to a cluster gives users of that cluster the ability to install any of the packages from that repository.

  - [Package](https://carvel.dev/kapp-controller/docs/latest/packaging/#package): is a combination of configuration metadata and OCI images that informs the package manager what software it holds and how to install itself onto a Kubernetes cluster.

  - [PackageMetadata](https://carvel.dev/kapp-controller/docs/latest/packaging/#package-metadata): are attributes of a single package that do not change frequently and that are shared across multiple versions of a single package. It contains information similar to a project’s README.md.

- [PackageInstall](https://carvel.dev/kapp-controller/docs/latest/packaging/#package-install) is an actual installation of a package and its underlying resources on a Kubernetes cluster.

- [App](https://carvel.dev/kapp-controller/docs/latest/app-overview/) is a set of Kubernetes resources. These resources could span any number of namespaces or could be cluster-wide.

The following image depicts the relationship between the different kapp-controller CRs:

![kapp-controller CRs](../img/kapp-crs.png)

## xxxxxx

### Confiugring Kubeapps to support Carvel Packages

As any other packaging format, the kapp-controller support is broght into Kubeapps by means of a plugin. 

This `kapp-controller` plugin is currently being built by default in the Kubeapps release and it is just a matter of enabling it when installing Kubeapps.

> **NOTE**: Please refer to the [getting started documentation](./getting-started.md) for more information on how to install Kubeapps and pass custom configuration values.

In the [values.yaml](../../chart/kubeapps/values.yaml) file, under `kubeappsapis.enabledPlugins` add 
 `kapp-controller` to the list of enabled plugins. For example:

```yaml
kubeappsapis:
    - resources 
    - helm 
    - kapp-controller # add this one
```

### Installing a Package Repository

> **NOTE**: Currently, Kubeapps does not offer any graphical way to manage [Carvel Packages Repositories](https://carvel.dev/kapp-controller/docs/latest/packaging/#package-repository). Therefore, you will need to install the package repository manually.

Since we are actively working on [refactor the Application Repository management in Kubeapps](https://github.com/kubeapps/kubeapps/projects/11?card_filter_query=milestone%3A%22app+repository+refactor%22), [Carvel Packages Repositories](https://carvel.dev/kapp-controller/docs/latest/packaging/#package-repository) cannot be currently manged by Kubeapps.
This section covers how to manage repositories manually.

First, you need to find a Carvel Package Repository already published. If not, you can always [create your own manually](https://carvel.dev/kapp-controller/docs/latest/packaging-tutorial/#creating-a-package-repository).
In this section, we will use the `Tanzu Community Edition` repository.

> In [this Carvel website](https://carvel.dev/kapp-controller/docs/latest/oss-packages/) you will find a list of Carvel Packages and Package Repositories that are available to open source users.

Next, you need to create a Package Repository CR. This is done by running the following command:


```bash
cat > repo.yaml << EOF
---
apiVersion: packaging.carvel.dev/v1alpha1
kind: PackageRepository
metadata:
  name: tce-repo
spec:
  fetch:
    imgpkgBundle:
      image: projects.registry.vmware.com/tce/main:0.9.1
EOF
```

Then, you need to apply the `PackageRepository` CR to your cluster using `kubectl` (or, alternatively, the `kapp` CLI), by running the following command:

```bash
kubectl apply -f repo.yaml
```

Under the hood, kapp-controller will create `Package` and `PackageMetadata` CRs for each of the packages in the repository.

> Run `kubectl get packagerepository`, `kubectl get packages` and `kubectl get packagemetadatas` to get the created CRs.

### Installing a Package

<!-- 

intro: same experience as usual

disclaimer: big repos with high reponse times, link issue

screenshot

how to install

need of a service account, link to carvel docs

how the values are stored as secrets

discuss why the default values are commented out

-->

### Viewing the Installed Applications

<!-- 

intro: same experience as usual

apps will auto-reconcile, explain what it is

screenshot overview

app details page

installed k8s resources

screenshot app details 

upgrade and delete applications

rollback is not a kapp-ctrl feature

view apps via kubectl

-->

### Conclusions

<!-- 

this guide covered how to enable kapp-ctrl support in kubeapps, installing a repo, installing a package, and viewing the installed apps.

we are actively working on this plugin, so if you encounter any problems, please file an issue xxx

 -->
