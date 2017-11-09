[![Build Status](https://travis-ci.org/kubeapps/installer.svg?branch=master)](https://travis-ci.org/kubeapps/installer)

# Kubeapps Installer

## Installation

Installation is made of two steps:

- Download latest Kubeapps Installer binary from the [release page](https://github.com/kubeapps/installer/releases). Currently Kubeapps Installer is distributed in two platforms: linux/amd64 and OSX/amd64
- Make the binary executable

## Build from source

You can build latest Kubeapps Installer from source. 

### Installing Go
- Visit https://golang.org/dl/
- Download the most recent Go version (here we used 1.9) and unpack the file
- Check the installation process on https://golang.org/doc/install
- Set the Go environment variables

```
$ export GOROOT=/GoDir/go
$ export GOPATH=/GoDir/go
$ export PATH=$GOPATH/bin:$PATH
```

### Create a working directory for the project
```
$ working_dir=$GOPATH/src/github.com/kubeapps/
$ mkdir -p $working_dir
```

### Clone kubeapps/installer repository
```
$ cd $working_dir
$ git clone https://github.com/kubeapps/installer
```

### Building local binary
```
$ cd installer
$ make binary
```

## Testing Kubeapps Installer with minikube

The simplest way to try Kubeapps is deploying it with Kubeapps Installer on [minikube](https://github.com/kubernetes/minikube). 

Kubeapps Installer is now working with RBAC-enabled Kubernetes (v1.7+). You can choose to start minikube vm with your preferred VM driver (virtualbox xhyve vmwarefusion) and a proper Kubernetes version. For example, the [latest minikube](https://github.com/kubernetes/minikube/releases/tag/v0.23.0) will start a Kubernetes v1.8.0.

```
$ minikube start
```

It's also required to have [Kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) installed on your machine. Verify Kubernetes version:

```
$ kubectl version
Client Version: version.Info{Major:"1", Minor:"7", GitVersion:"v1.7.0", GitCommit:"d3ada0119e776222f11ec7945e6d860061339aad", GitTreeState:"clean", BuildDate:"2017-06-29T23:15:59Z", GoVersion:"go1.8.3", Compiler:"gc", Platform:"darwin/amd64"}
Server Version: version.Info{Major:"1", Minor:"8", GitVersion:"v1.8.0", GitCommit:"0b9efaeb34a2fc51ff8e4d34ad9bc6375459c4a4", GitTreeState:"dirty", BuildDate:"2017-10-17T15:09:55Z", GoVersion:"go1.8.3", Compiler:"gc", Platform:"linux/amd64"}
```

Deploy Kubeapps with Kubeapps Installer

```
$ kubeapps up
```

Remove Kubeapps

```
$ kubeapps down
```