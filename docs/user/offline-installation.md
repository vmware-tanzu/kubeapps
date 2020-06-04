# Offline Installation of Kubeapps

Since the version 1.10.1 of Kubeapps (Chart version 3.7.0), it's possible to successfully install Kubeapps in an offline environment. To be able to able to install Kubeapps without Internet connection, it's necessary to:

 - Pre-download the Kubeapps chart.
 - Mirror Kubeapps images so they are accessible within the cluster.
 - [Optional] Have one or more offline App Repositories.

## 1. Download the Kubeapps chart

First, download the tarball containing the Kubeapps chart from the publicly available repository maintained by Bitnami. Note that Internet connection is necessary at this point:

```bash
helm pull --untar https://charts.bitnami.com/bitnami/kubeapps-3.7.0.tgz
helm dep update ./kubeapps
```

## 2. Mirror Kubeapps images

In order to be able to install Kubeapps, it's necessary to either have a copy of all the images that Kubeapps requires in each node of the cluster or push these images to an internal Docker registry that Kubernetes can access. You can obtain the list of images by checking the `values.yaml` of the chart. For example:

```yaml
  registry: docker.io
  repository: bitnami/nginx
  tag: 1.17.10-debian-10-r10
```

For simplicity, in this guide I am using a cluster with a single node using [Kubernetes in Docker (`kind`)](https://github.com/kubernetes-sigs/kind) so I just need to preload the images in this node.

```bash
docker pull bitnami/nginx:1.17.10-debian-10-r10
kind load docker-image bitnami/nginx:1.17.10-debian-10-r10
```

In case you are using a private Docker registry, you will need to re-tag the images and push them:

```bash
docker pull bitnami/nginx:1.17.10-debian-10-r10
docker tag bitnami/nginx:1.17.10-debian-10-r10 REPO_URL/bitnami/nginx:1.17.10-debian-10-r10
docker push REPO_URL/bitnami/nginx:1.17.10-debian-10-r10
```

You will need to follow a similar process for every image present in the values file.

## 3. [Optional] Prepare an offline App Repository

By default, Kubeapps install three App Repositories: `stable`, `incubator` and `bitnami`. Since, in order to sync those repositories, it's necessary to have Internet connection, you will need to mirror those or create your own repository (e.g. using Harbor) and configure it when installing Kubeapps.

For more information about how to create a private repository, follow this [guide](./private-app-repository.md).

## 4. Install Kubeapps

Now that you have everything pre-loaded in your cluster, it's possible to install Kubeapps using the chart directory from the first step:

**NOTE**: If during step 2), you were using a private docker registry, it's necessary to modify the global value used for the registry. This can be set by specifying `--set global.imageRegistry=REPO_URL`.

```bash
helm install kubeapps ./kubeapps [OPTIONS]
```
