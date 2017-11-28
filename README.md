# <img src="./img/logo.png" width="40" align="left"> Kubeapps

[![Build Status](https://travis-ci.org/kubeapps/kubeapps.svg?branch=master)](https://travis-ci.org/kubeapps/kubeapps)

Kubeapps is a set of tools written by [Bitnami](https://bitnami.com) to super-charge your Kubernetes cluster with:
 * Your own application [dashboard](https://kubeapps.com/), allowing you to deploy Kubernetes-ready applications into your cluster with a single click.
 * [Kubeless](http://kubeless.io/), a Kubernetes-native Serverless Framework, compatible with [serverless.com](https://serverless.com).
 * [SealedSecrets](https://github.com/bitnami/sealed-secrets), a way to encrypt a Secret into a SealedSecret, which is safe to store...even for a public repository.

## Quickstart

Kubeapps assumes a working Kubernetes (v1.7+) with RBAC enabled and [`kubectl`](https://kubernetes.io/docs/tasks/tools/install-kubectl/) installed and configured to talk to your Kubernetes cluster. Kubeapps binaries are available for both Linux and OS X, and Kubeapps has been tested with both `minikube` and Google Kubernetes Engine (GKE).

> On GKE, you must either be an "Owner" or have the "Container Engine Admin" role in order to install Kubeapps.

The simplest way to try Kubeapps is to deploy it with the Kubeapps Installer on [minikube](https://github.com/kubernetes/minikube). For example, to install the latest binary on Linux, use these commands:

```
minikube start
sudo curl -s https://api.github.com/repos/kubeapps/kubeapps/releases/latest | grep linux | grep browser_download_url | cut -d '"' -f 4 | wget -i -
sudo mv kubeapps-linux-amd64 /usr/local/bin/kubeapps
chmod +x /usr/local/bin/kubeapps
kubeapps up
kubeapps dashboard
```

These commands will install Kubeapps for your cluster and launch a browser with the Kubeapps dashboard.

[image]

## Next Steps

Use the Kubeapps Dashboard to easily manage the deployments created by Helm in your cluster and to manage your Kubeless functions. Learn more about [using the Kubeapps Dashboard](docs/dashboard.md).

In case of difficulties installing Kubeapps, find [more detailed installation instructions](docs/install.md) or [learn how to build Kubeapps from source](docs/install.md).

For a more detailed and step-by-step introduction to Kubeapps, read our [introductory walkthrough](docs/get-started.md).

## Useful Resources

* [Walkthrough for new users](docs/get-started.md)
* [Detailed installation instructions](docs/install.md)
* [Kubeapps Dashboard documentation](docs/dashboard.md)
* [Kubeapps components](docs/components.md)
