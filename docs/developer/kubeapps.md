# Kubeapps Installer Developer Guide

The Kubeapps installer is a command-line tool for installing, upgrading and uninstalling the Kubeapps in-cluster components. The tool is written in the Go programming language and the Kubernetes manifests are written in the Jsonnet data templating language.

## Prerequisites

- [Git](https://git-scm.com/)
- [Make](https://www.gnu.org/software/make/)
- [Kubernetes cluster](https://kubernetes.io/docs/setup/pick-right-solution/)
- [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/)

## Environment setup

```bash
export GOPATH=~/gopath/
export PATH=$GOPATH/bin:$PATH
export KUBEAPPS_DIR=$GOPATH/src/github.com/kubeapps/kubeapps
```
## Download kubeapps source code

The kubeapps installer code is located under the `cmd/kubeapps/` directory of the Kubeapps repository.

```bash
git clone --recurse-submodules https://github.com/kubeapps/kubeapps $KUBEAPPS_DIR
cd $KUBEAPPS_DIR/cmd/kubeapps
```

## Building kubeapps installer

The `kubeapps` installer should be compiled from the root of the Kubeapps repository.

```bash
cd $KUBEAPPS_DIR
make VERSION=myver kubeapps
```

The above command builds the `kubeapps` binary in the working directory.

### Running in development

Kubeapps is a tool for installing and uninstalling Kubeapps components to a Kubernetes cluster. Testing the `kubeapps` installer requires a Kubernetes cluster with `kubectl` configure correctly.

To test the installation procedure,

```bash
./kubeapps up
```

Adding the `--dry-run` argument to the above command lets you to quickly verify changes in the jsonnet manifests without making any changes to your development cluster.

```bash
./kubeapps up --dry-run
```

### Running tests

To start the tests on the `kubeapps` installer execute the following command from the Kubeapps repository root.

```bash
make test-kubeapps
```
