# Kubeapps chart-repo Developer Guide

The `chart-repo` component is tool that scans a Helm chart repository and populates chart metadata in a MongoDB server. This metadata is then served by the `chartsvc` component. Its source is maintained in the [Monocular project repository](https://github.com/helm/monocular).

## Prerequisites

- [Git](https://git-scm.com/)
- [Make](https://www.gnu.org/software/make/)
- [Go programming language](https://golang.org/dl/)
- [Docker CE](https://www.docker.com/community-edition)
- [Kubernetes cluster (v1.8+)](https://kubernetes.io/docs/setup/pick-right-solution/). [Minikube](https://github.com/kubernetes/minikbue) is recommended.
- [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/)
- [Telepresence](https://telepresence.io)

## Environment

```bash
export GOPATH=~/gopath/
export PATH=$GOPATH/bin:$PATH
export MONOCULAR_DIR=$GOPATH
```

## Download the Monocular source code

```bash
git clone https://github.com/helm/monocular $MONOCULAR_DIR
```

The `chart-repo` sources are located under the `cmd/chart-repo/` directory.

### Install Kubeapps in your cluster

Kubeapps is a Kubernetes-native application. To develop and test Kubeapps components we need a Kubernetes cluster with Kubeapps already installed. Follow the [Kubeapps installation guide](../../chart/kubeapps/README.md) to install Kubeapps in your cluster.

### Building the `chart-repo` image

```bash
cd $MONOCULAR_DIR
go mod tidy
make -C cmd/chart-repo docker-build
```

This builds the `chart-repo` Docker image. Please refer to [Monocular Developers Guide](https://github.com/helm/monocular/blob/master/docs/development.md) for more details.

### Running in development

```bash
export MONGO_PASSWORD=$(kubectl get secret --namespace kubeapps kubeapps-mongodb -o jsonpath="{.data.mongodb-root-password}" | base64 --decode)
telepresence --namespace kubeapps --docker-run -e MONGO_PASSWORD=$MONGO_PASSWORD --rm -ti quay.io/helmpack/chart-repo /chart-repo sync --mongo-user=root --mongo-url=kubeapps-mongodb stable https://kubernetes-charts.storage.googleapis.com
```

Note that the chart-repo should be rebuilt for new changes to take effect.

### Running tests

You can run the chart-repo tests along with the tests for the Monocular project:

```bash
go test -v ./...
```
