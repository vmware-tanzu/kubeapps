# <img src="./img/logo.png" width="40" align="left"> Kubeapps

[![Build Status](https://travis-ci.org/kubeapps/kubeapps.svg?branch=master)](https://travis-ci.org/kubeapps/kubeapps)

Kubeapps is a set of tools written by [Bitnami](https://bitnami.com) to super-charge your Kubernetes cluster with:
 * Your own application [dashboard](https://kubeapps.com/), allowing you to deploy Kubernetes-ready applications into your cluster with a single click.
 * [Kubeless](http://kubeless.io/), a Kubernetes-native Serverless Framework, compatible with [serverless.com](https://serverless.com).
 * [SealedSecrets](https://github.com/bitnami/sealed-secrets), a way to encrypt a Secret into a SealedSecret, which is safe to store...even for a public repository.

## Quickstart

Kubeapps assumes a working Kubernetes (v1.7+) with RBAC enabled and [`kubectl`](https://kubernetes.io/docs/tasks/tools/install-kubectl/) installed and configured to talk to your Kubernetes cluster. Kubeapps binaries are available for both Linux and OS X, and Kubeapps has been tested with both `minikube` and Google Kubernetes Engine (GKE).

> On GKE, you must either be an "Owner" or have the "Container Engine Admin" role in order to install Kubeapps.

The simplest way to try Kubeapps is to deploy it with the Kubeapps Installer on [minikube](https://github.com/kubernetes/minikube). Assuming you are deploying a binary installer on Linux, here are the commands to run: 

```
curl -s https://api.github.com/repos/kubeapps/kubeapps/releases/latest | grep linux | grep browser_download_url | cut -d '"' -f 4 | wget -i -
sudo mv kubeapps-linux-amd64 /usr/local/bin/kubeapps
sudo chmod +x /usr/local/bin/kubeapps
kubeapps up
kubeapps dashboard
```

These commands will deploy Kubeapps in your cluster and launch a browser with the Kubeapps dashboard.

![Dashboard main page](img/dashboard-home.png)

To remove Kubeapps from your cluster, simply run:

```
kubeapps down
```

## Build from Source

The Kubeapps Installer is a CLI tool written in Go that will deploy the Kubeapps components into your cluster.
You can build the latest Kubeapps Installer from source by following the steps below:

* Visit [the Go website](https://golang.org), download the most recent [binary distribution of Go](https://golang.org/dl/) and install it following the [official instructions](https://golang.org/doc/install).

  > The remainder of this section assumes that Go is installed in `/usr/local/go`. Update the paths in subsequent commands if you used a different location.

* Set the Go environment variables:

  ```
  export GOROOT=/usr/local/go
  export GOPATH=/usr/local/go
  export PATH=$GOPATH/bin:$PATH
  ```

* Create a working directory for the project:

  ```
  working_dir=$GOPATH/src/github.com/kubeapps/
  mkdir -p $working_dir
  ```

* Clone the Kubeapps source repository:

  ```
  cd $working_dir
  git clone https://github.com/kubeapps/kubeapps
  ```

* Build the Kubeapps binary and move it to a location in your path:

  ```
  cd kubeapps
  make binary
  cp kubeapps /usr/local
  ```

## Next Steps

[Use the Kubeapps Dashboard](docs/dashboard.md) to easily manage the deployments created by Helm in your cluster and to manage your Kubeless functions, or [look under the hood to see what's included in Kubeapps](docs/components.md).

In case of difficulties installing Kubeapps, find [more detailed installation instructions](docs/install.md).

For a more detailed and step-by-step introduction to Kubeapps, read our [introductory walkthrough](docs/getting-started.md).

## Useful Resources

* [Walkthrough for first-time users](docs/getting-started.md)
* [Detailed installation instructions](docs/install.md)
* [Kubeapps Dashboard documentation](docs/dashboard.md)
* [Kubeapps components](docs/components.md)
