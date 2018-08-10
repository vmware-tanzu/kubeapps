# Kubeapps chart-repo Developer Guide

The `chart-repo` component is tool that scans a Helm chart repository and populates chart metadata in a MongoDB server. This metadata is then served by the `chartsvc` component.

## Prerequisites

- [Git](https://git-scm.com/)
- [Make](https://www.gnu.org/software/make/)
- [Go programming language](https://golang.org/dl/)
- [Docker CE](https://www.docker.com/community-edition)
- [Kubernetes cluster (v1.8+)](https://kubernetes.io/docs/setup/pick-right-solution/)
- [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/)
- [Telepresence](https://telepresence.io)

*Telepresence is not a hard requirement, but is recommended for a better developer experience*

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

The `chart-repo` sources are located under the `cmd/chart-repo/` directory of the repository.

```bash
cd $KUBEAPPS_DIR/cmd/chart-repo
```

### Install Kubeapps in your cluster

Kubeapps is a Kubernetes-native application. To develop and test Kubeapps components we need a Kubernetes cluster with Kubeapps already installed. Follow the [Kubeapps installation guide](../../chart/kubeapps/README.md) to install Kubeapps in your cluster.

### Building `chart-repo` binary

```bash
go build
```

This builds the `chart-repo` binary in the working directory.

### Running in development

[Telepresence](https://www.telepresence.io/) is a local development tool for Kubernetes microservices. As `chart-repo` is a tool that is executed in the Kubernetes cluster we use telepresence so that `chart-repo` can access the services running on your local development server.

Create a `telepresence` shell,

```bash
telepresence --namespace kubeapps --method inject-tcp --run-shell
```

> **NOTE**: If you encounter issues getting this setup working correctly, please try switching the telepresence proxying method in the above command to `vpn-tcp`. Refer to [the telepresence docs](https://www.telepresence.io/reference/methods) to learn more about the available proxying methods and their limitations.

To test and debug the `chart-repo` tool launch the command locally within the telepresence shell:

```bash
export MONGO_PASSWORD=$(kubectl get secret --namespace kubeapps mongodb -o jsonpath="{.data.mongodb-root-password}" | base64 --decode)
./chart-repo sync --mongo-url=mongodb.kubeapps --mongo-user=root stable https://kubernetes-charts.storage.googleapis.com
```

### Running tests

To start the tests on the `chart-repo` execute the following command:

```bash
go test
```

## Building the `kubeapps/chart-repo` Docker image

To build the `kubeapps/chart-repo` docker image with the docker image tag `myver`:

```bash
cd $KUBEAPPS_DIR
make VERSION=myver kubeapps/chart-repo
```
