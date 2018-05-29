# <img src="./img/logo.png" width="40" align="left"> Kubeapps

[![Build Status](https://travis-ci.org/kubeapps/kubeapps.svg?branch=master)](https://travis-ci.org/kubeapps/kubeapps)

Kubeapps is a set of tools written by [Bitnami](https://bitnami.com) to super-charge your Kubernetes cluster with:

* Your own application [dashboard](https://kubeapps.com/), allowing you to deploy Kubernetes-ready applications into your cluster with a single click.
* [Kubeless](http://kubeless.io/), a Kubernetes-native Serverless Framework, compatible with [serverless.com](https://serverless.com).
* [SealedSecrets](https://github.com/bitnami/sealed-secrets), a way to encrypt a Secret into a SealedSecret, which is safe to store...even for a public repository.

## Quickstart

Kubeapps assumes a working Kubernetes cluster (v1.8+) and [`kubectl`](https://kubernetes.io/docs/tasks/tools/install-kubectl/) installed and configured to talk to your Kubernetes cluster. Kubeapps binaries are available for Linux, OS X and Windows, and Kubeapps has been tested with `minikube`, Google Kubernetes Engine (GKE) and Azure Container Service (AKS). Kubeapps works on RBAC-enabled clusters and this configuration is encouraged for a more secure install.

> On GKE, you must either be an "Owner" or have the "Container Engine Admin" role in order to install Kubeapps.

The simplest way to try Kubeapps is to deploy it with the Kubeapps Installer on [minikube](https://github.com/kubernetes/minikube). Assuming you are using Linux or OS X, run the following commands to download and install the Kubeapps Installer binary:

```bash
curl -s https://api.github.com/repos/kubeapps/kubeapps/releases/latest | grep -i $(uname -s) | grep browser_download_url | cut -d '"' -f 4 | wget -i -
sudo mv kubeapps-$(uname -s| tr '[:upper:]' '[:lower:]')-amd64 /usr/local/bin/kubeapps
sudo chmod +x /usr/local/bin/kubeapps
kubeapps up
kubeapps dashboard
```

These commands will deploy Kubeapps in your cluster and launch a browser with the Kubeapps dashboard.

![Dashboard login page](img/dashboard-login.png)

Access to the dashboard requires a Kubernetes API token to authenticate with the Kubernetes API server. Read the [Access Control](docs/access-control.md) documentation for more information on configuring users for Kubeapps.

The following commands create a ServiceAccount and ClusterRoleBinding named `kubeapps-operator` which will enable the dashboard to authenticate and manage resources on the Kubernetes cluster:

```bash
kubectl create serviceaccount kubeapps-operator
kubectl create clusterrolebinding kubeapps-operator --clusterrole=cluster-admin --serviceaccount=default:kubeapps-operator
```

Use the following command to reveal the authorization token that should be used to authenticate the Kubeapps dashboard with the Kubernetes API:

```bash
kubectl get secret $(kubectl get serviceaccount kubeapps-operator -o jsonpath='{.secrets[].name}') -o jsonpath='{.data.token}' | base64 --decode
```

**NOTE**: It's not recommended to create cluster-admin users for Kubeapps. Please refer to the [Access Control](docs/access-control.md) documentation to configure more fine-grained access.

![Dashboard main page](img/dashboard-home.png)

To remove Kubeapps from your cluster, simply run:

```bash
kubeapps down
```

To delete the `kubeapps-operator` ServiceAccount and ClusterRoleBinding,

```bash
kubectl delete clusterrolebinding kubeapps-operator
kubectl delete serviceaccount kubeapps-operator
```

## Installation

Get the latest release of Kubeapps Installer on the [Github releases](https://github.com/kubeapps/kubeapps/releases) page.

Alternatively, when you have configured a proper Go environment (refer to the first two steps of [Build from Source](#build-from-source) section), the latest Kubeapps Installer can be get-able from source:

```bash
go get github.com/kubeapps/kubeapps
```

## Build from Source

The Kubeapps Installer is a CLI tool written in Go that will deploy the Kubeapps components into your cluster.
You can build the latest Kubeapps Installer from source by following the steps below:

* Visit [the Go website](https://golang.org), download the most recent [binary distribution of Go](https://golang.org/dl/) and install it following the [official instructions](https://golang.org/doc/install).

  > The remainder of this section assumes that Go is installed in `/usr/local/go`. Update the paths in subsequent commands if you used a different location.

* Set the Go environment variables:

  ```bash
  export GOPATH=$HOME/gopath
  export PATH=$GOPATH/bin:$PATH
  ```

* Install kubeapps build dependencies:

  ```bash
  go get github.com/ksonnet/kubecfg
  ```

* Create a working directory for the project:

  ```bash
  working_dir=$GOPATH/src/github.com/kubeapps/
  mkdir -p $working_dir
  ```

* Clone the Kubeapps source repository:

  ```bash
  cd $working_dir
  git clone --recurse-submodules https://github.com/kubeapps/kubeapps
  ```

* Build the Kubeapps binary and move it to a location in your path:

  ```bash
  cd kubeapps
  make kubeapps
  ```

## Running tests

Run Go tests using `make test`. See [dashboard documentation](dashboard/README.md) for information on running dashboard tests.

```bash
make test
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
