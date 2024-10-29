# Offline Installation of Kubeapps

## Table of Contents

1. [Introduction](#introduction)
2. [Pre-requisites](#pre-requisites)
3. [Step 1. Download the Kubeapps chart](#step-1-download-the-kubeapps-chart)
4. [Step 2. Mirror Kubeapps images](#step-2-mirror-kubeapps-images)
5. [Step 3. [Optional] Prepare an offline Package Repository](#step-3-optional-prepare-an-offline-package-repository)
6. [Step 4. Install Kubeapps](#step-4-install-kubeapps)

## Introduction

[Kubeapps](https://kubeapps.dev/) provides a cloud-native solution to browse, deploy and manage the lifecycle of applications on a Kubernetes cluster. It is a one-time install that gives you many important benefits, including the ability to:

- browse and deploy packaged applications from public or private repositories
- customize deployments through an intuitive user interface
- upgrade, manage and delete the applications that are deployed in your Kubernetes cluster
- expose an API to manage your package repositories and your applications

This guide explains in detail how to **install Kubeapps in an air-gapped environment**.

## Pre-requisites

To be able to able to install Kubeapps without an Internet connection, it's necessary to:

- Kubernetes cluster (air-gapped).
- Pre-download the Kubeapps Helm chart.
- Mirror Kubeapps images so they are accessible within the cluster.
- [Optional] Have one or more offline Package Repositories.

> **Note**: Internet connection is necessary at this point to download charts and images for an offline installation

## Step 1. Download the Kubeapps chart

First, download the tarball containing the Kubeapps chart from the publicly available repository maintained by Bitnami.

```bash
helm pull --untar oci://registry-1.docker.io/bitnamicharts/kubeapps --version x.y.z
helm dep update ./kubeapps
```

> Notice `x.y.z` must be replaced by the latest version of the Kubeapps chart available at the [Bitnami Chart Repository](https://github.com/bitnami/charts/blob/main/bitnami/kubeapps/Chart.yaml#L3)

## Step 2. Mirror Kubeapps images

To be able to install Kubeapps, it's necessary to either have a copy of all the images that Kubeapps requires in each node of the cluster or push these images to an internal Docker registry that Kubernetes can access. You can obtain the list of images by checking the [`values.yaml` file](https://github.com/bitnami/charts/blob/main/bitnami/kubeapps/values.yaml) of the chart. For example:

```yaml
registry: docker.io
repository: bitnami/nginx
tag: 1.19.2-debian-10-r32
```

> The list of images to download includes (could change in time):
>
> - `bitnami/nginx`
> - `bitnami/kubeapps-dashboard`
> - `bitnami/kubeapps-apprepository-controller`
> - `bitnami/kubeapps-asset-syncer`
> - `bitnami/oauth2-proxy`
> - `bitnami/kubeapps-pinniped-proxy`
> - `bitnami/kubeapps-apis`
> - `bitnami/postgresql`

For simplicity, in this guide, use a single-node cluster created with [Kubernetes in Docker (`kind`)](https://github.com/kubernetes-sigs/kind), using a namespace called "kubeapps". In this environment, the images have to be pre-loaded: - first, pull the images (`docker pull`), - next load them into the cluster (`kind load docker-image`).

```bash
docker pull bitnami/nginx:1.19.2-debian-10-r32
kind load docker-image bitnami/nginx:1.19.2-debian-10-r32
```

> **Note**: tags must be updated according to the version of the images in the [`values.yaml` file](https://github.com/bitnami/charts/blob/main/bitnami/kubeapps/values.yaml)

In case you are using a private Docker registry, you need to re-tag the images and push them:

```bash
docker pull bitnami/nginx:1.19.2-debian-10-r32
docker tag bitnami/nginx:1.19.2-debian-10-r32 REPO_URL/bitnami/nginx:1.19.2-debian-10-r32
docker push REPO_URL/bitnami/nginx:1.19.2-debian-10-r32
```

Follow the same process for every image present in the values file.

## Step 3. [Optional] Prepare an offline Package Repository

By default, Kubeapps install the `bitnami` Package Repository. Since, to sync that repository, it's necessary to have an Internet connection, you will need to mirror it or create your repository (e.g. using Harbor) and configure it when installing Kubeapps.

For more information about how to create a private repository, follow this [how-to guide](./private-app-repository.md).

## Step 4. Install Kubeapps

Now that you have everything pre-loaded in your cluster, it's possible to install Kubeapps using the chart directory from the first step:

```bash
helm install kubeapps [OPTIONS] ./kubeapps
```

**NOTE**: If during step 2), you were using a private docker registry, it's necessary to modify the global value used for the registry. This can be set by specifying `--set global.imageRegistry=REPO_URL`.
If this registry, additionally, needs an ImagePullSecret, specify it with `--set global.imagePullSecrets[0]=SECRET_NAME`.
