# Offline Installation of Kubeapps

Since the version 1.10.1 of Kubeapps (Chart version 3.7.0), it's possible to successfully install Kubeapps in an offline environment. To be able to able to install Kubeapps without Internet connection, it's necessary to:

 - Pre-download the Kubeapps chart.
 - Mirror Kubeapps images so they are accessible within the cluster.
 - [Optional] Have one or more offline App Repositories.

## 1. Download the Kubeapps chart

First, download the tarball containing the Kubeapps chart from the publicly available repository maintained by Bitnami. Note that Internet connection is necessary at this point:

```bash
wget https://charts.bitnami.com/bitnami/kubeapps-3.7.0.tgz
```

Note that Kubeapps has a dependency with another chart used for the database. This is either MongoDB or PostgreSQL. You also need to download these charts in advance. To get the exact version needed, you can check the `requirements.lock` file from the chart:

```bash
wget -P charts/ https://charts.bitnami.com/bitnami/mongodb-7.10.10.tgz
wget -P charts/ https://charts.bitnami.com/bitnami/postgresql-8.9.1.tgz
```

Note that the chart should be kept in a `charts/` subfolder.

## 2. Mirror Kubeapps images

In order to be able to install Kubeapps, it's necessary for the Kubernetes cluster to contain a copy of the images that Kubeapps requires. You can obtain the list of images by checking the `values.yaml` of the chart. For example:

```yaml
  registry: docker.io
  repository: bitnami/nginx
  tag: 1.17.10-debian-10-r10
```

For this guide I am using [Kubernetes in Docker (`kind`)](https://github.com/kubernetes-sigs/kind) so depending on your cluster provider, this process will change. In this case, it's enough to pull the image and load it with `kind`:

```bash
docker pull bitnami/nginx:1.17.10-debian-10-r10
kind load docker-image bitnami/nginx:1.17.10-debian-10-r10
```

You will need to follow a similar process for every image present in the values file.

## 3. [Optional] Prepare an offline App Repository

By default, Kubeapps install three App Repositories: `stable`, `incubator` and `bitnami`. Since, in order to sync those repositories, it's necessary to have Internet connection, you will need to mirror those or create your own repository (e.g. using Harbor) and configure it when installing Kubeapps.

For more information about how to create a private repository, follow this [guide](./private-app-repository.md).

## 4. Install Kubeapps

Now that you have everything pre-loaded in your cluster, it's possible to install Kubeapps using the tarball from the first step:

```bash
helm install kubeapps ./kubeapps-3.7.0.tgz [OPTIONS]
```
