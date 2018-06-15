# Kubeapps ChartSVC Developer Guide

The `chartsvc` component is a micro-service that creates a API endpoint for accessing the metadata for charts in Helm chart repositories that's populated in a MongoDB server.

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

The `chartsvc` sources are located under the `cmd/chartsvc/` directory of the repository.

```bash
cd $KUBEAPPS_DIR/cmd/chartsvc
```

### Install Kubeapps in your cluster

Kubeapps is a Kubernetes-native application. To develop and test Kubeapps components we need a Kubernetes cluster with Kubeapps already installed. Follow the [Kubeapps installation guide](../user/install.md) to install Kubeapps in your cluster.

### Building the `chartsvc` binary

```bash
go build
```

This builds the `chartsvc` binary in the working directory.

### Running in development

[Telepresence](https://www.telepresence.io/) is a local development tool for Kubernetes microservices. As `chartsvc` is a service running in the Kubernetes cluster we use telepresence to proxy requests to the `chartsvc` running in your cluster to your local development host.

Create a `telepresence` shell to swap the `chartsvc` deployment in the `kubeapps` namespace, forwarding local port `9000` to port `8080` of the `chartsvc` pod.

```bash
telepresence --namespace kubeapps --method inject-tcp --swap-deployment chartsvc --expose 9000:8080 --run-shell
```

> **NOTE**: If you encounter issues getting this setup working correctly, please try switching the telepresence proxying method in the above command to `vpn-tcp`. Refer to [the telepresence docs](https://www.telepresence.io/reference/methods) to learn more about the available proxying methods and their limitations.

Next, launch the `chartsvc` locally within the telepresence shell:

```bash
export PORT=9000
./chartsvc --mongo-url=mongodb.kubeapps --mongo-user=root
```

From this point any API requests made to the `chartsvc` will be served by the service running locally on your development host.

### Running tests

To start the tests on the `chartsvc` execute the following command:

```bash
cd $KUBEAPPS_DIR/cmd/chartsvc
go test
```

## Building the kubeapps/chartsvc Docker image

To build the `kubeapps/chartsvc` docker image with the docker image tag `myver`:

```bash
cd $KUBEAPPS_DIR
make VERSION=myver kubeapps/chartsvc
```
