# Kubeapps

[![Build Status](https://travis-ci.org/kubeapps/kubeapps.svg?branch=master)](https://travis-ci.org/kubeapps/kubeapps)

<img src="./img/logo.png" width="100">

Kubeapps is a set of tools written by [Bitnami](https://bitnami.com) to super-charge your Kubernetes cluster with:
 * Your own application [dashboard](https://kubeapps.com/), allowing you to deploy Kubernetes-ready applications into your cluster with a single click.
 * [Kubeless](http://kubeless.io/), a Kubernetes-native Serverless Framework, compatible with [serverless.com](https://serverless.com).
 * [SealedSecrets](https://github.com/bitnami/sealed-secrets), a way to encrypt a Secret into a SealedSecret, which is safe to store...even for a public repository. 

These tools are easily deployed into your cluster with just one command: ```kubeapps up``` 

----

## Quickstart

Kubeapps assumes a working Kubernetes (v1.7+) with RBAC enabled and [Kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) installed and configured to talk to your Kubernetes cluster. Kubeapps binaries are available for both Linux and Darwin, and Kubeapps has been tested with both `minikube` and Google Kubernetes Engine (GKE).

The simplest way to try Kubeapps is to deploy it with the Kubeapps Installer on [minikube](https://github.com/kubernetes/minikube). For example, to install the latest binary on Linux, use these commands:

minikube start
sudo curl -s https://api.github.com/repos/kubeapps/kubeapps/releases/latest | grep linux | grep browser_download_url | cut -d '"' -f 4 | wget -i -
sudo mv kubeapps-linux-amd64 /usr/local/bin/kubeapps
chmod +x /usr/local/bin/kubeapps
kubeapps up
kubeapps dashboard

These commands will install Kubeapps for your cluster and launch a browser with the Kubeapps dashboard.

[image]

You can use the Kubeapps Dashboard to easily manage the deployments created by Helm in your cluster and to manage your Kubeless functions. Learn more about [using the Kubeapps Dashboard].

For a more detailed introduction to Kubeapps, read our [introductory walkthrough](docs/getting-started.md). You can also read [more detailed installation instructions] or [learn how to build Kubeapps from source].

----

## Installation

Kubeapps assumes a working Kubernetes (v1.7+) with RBAC enabled and [Kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) installed and configured to talk to your Kubernetes cluster.

Kubeapps has been tested with both `minikube` and Google Kubernetes Engine (GKE). 

> On GKE, you must either be an "Owner" or have the "Container Engine Admin" role in order to install Kubeapps.

### Install pre-built binary

* Download a binary version of the latest Kubeapps Installer for your platform from the [release page](https://github.com/kubeapps/kubeapps/releases). Currently, the Kubeapps Installer is distributed in binary form for two platforms: Linux amd64 and OS X amd64.
* Make the binary executable.

For example, to install 0.0.2 release on Linux, use this command:

```
sudo curl -L https://github.com/kubeapps/installer/releases/download/v0.0.2/kubeapps-linux-amd64 -o /usr/local/bin/kubeapps && sudo chmod +x /usr/local/bin/kubeapps
```

### Build binary from source

The Kubeapps Installer is a CLI tool written in Go that will deploy the Kubeapps components into your cluster.
You can build the latest Kubeapps Installer from source by following the steps below: 

* Visit [the Go website](https://golang.org), download the most recent [binary distribution of Go](https://golang.org/dl/) and install it following the [official instructions](https://golang.org/doc/install). 

  The remainder of this section assumes that Go is installed in `/usr/local/go`. Update the paths in subsequent commands if you used a different location.

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

----

## Basic Usage

[todo]

For a more detailed introduction to Kubeapps, read our [introductory walkthrough](docs/getting-started.md). You can also browse the [full documentation](docs/).

----

## Testing Kubeapps Installer with minikube

The simplest way to try Kubeapps is deploying it with Kubeapps Installer on [minikube](https://github.com/kubernetes/minikube). 

Kubeapps Installer is now working with RBAC-enabled Kubernetes (v1.7+). You can choose to start minikube vm with your preferred VM driver (virtualbox xhyve vmwarefusion) and a proper Kubernetes version. For example, the [latest minikube](https://github.com/kubernetes/minikube/releases/tag/v0.23.0) will start a Kubernetes v1.8.0.

```
minikube start
```

It's also required to have [Kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) installed on your machine. Verify Kubernetes version:

```
kubectl version
Client Version: version.Info{Major:"1", Minor:"7", GitVersion:"v1.7.0", GitCommit:"d3ada0119e776222f11ec7945e6d860061339aad", GitTreeState:"clean", BuildDate:"2017-06-29T23:15:59Z", GoVersion:"go1.8.3", Compiler:"gc", Platform:"darwin/amd64"}
Server Version: version.Info{Major:"1", Minor:"8", GitVersion:"v1.8.0", GitCommit:"0b9efaeb34a2fc51ff8e4d34ad9bc6375459c4a4", GitTreeState:"dirty", BuildDate:"2017-10-17T15:09:55Z", GoVersion:"go1.8.3", Compiler:"gc", Platform:"linux/amd64"}
```

Deploy Kubeapps with Kubeapps Installer

```
kubeapps up
```

Remove Kubeapps

```
kubeapps down
```
