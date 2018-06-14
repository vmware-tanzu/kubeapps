# Kubeapps Installer Developer Guide

The Kubeapps installer is a command-line tool for installing, upgrading and uninstalling the Kubeapps in-cluster components. The tool is written using the Go programming language and the Kubernetes manifests are written in the Jsonnet data templating language.

## Prerequisites

- [Git](https://git-scm.com/)
- [Make](https://www.gnu.org/software/make/)
- [Go programming language](https://golang.org/dl/)
- [kubecfg](https://github.com/ksonnet/kubecfg)
- [Kubernetes cluster (v1.8+)](https://kubernetes.io/docs/setup/pick-right-solution/)
- [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/)

## Environment

```bash
export GOPATH=~/gopath/
export PATH=$GOPATH/bin:$PATH
export KUBEAPPS_DIR=$GOPATH/src/github.com/kubeapps/kubeapps
```

## Download the kubeapps source code

```bash
git clone --recurse-submodules https://github.com/kubeapps/kubeapps $KUBEAPPS_DIR
```

The installer sources are located under the `cmd/kubeapps/` directory while the jsonnet manifests for the Kubernetes resources are located under the `manifests/` directory of the repository.

```bash
cd $KUBEAPPS_DIR
```

## Building kubeapps installer

The `kubeapps` installer should be compiled from the root of the Kubeapps repository.

```bash
make VERSION=myver kubeapps
```

This builds the `kubeapps` binary in the working directory.

### Running in development

The `kubeapps` binary is a tool to install Kubeapps components to a Kubernetes cluster. Testing the `kubeapps` binary requires a Kubernetes cluster with `kubectl` configured to talk to the cluster.

Simply execute the `kubeapps` binary:

```bash
./kubeapps up
```

Add the `--dry-run` argument to the command to quickly verify changes in jsonnet manifests without making any change to your cluster.

```bash
./kubeapps up --dry-run
```

### Running tests

To execture the kubeapps installer tests execute the following command from the Kubeapps repository root.

```bash
make test-kubeapps
```
