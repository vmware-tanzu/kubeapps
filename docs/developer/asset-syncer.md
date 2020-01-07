# Kubeapps asset-syncer Developer Guide

The `chart-repo` component is a tool that scans a Helm chart repository and populates chart metadata in the database. This metadata is then served by the `asse-svc` component.

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
export GOPATH=~/gopath
export PATH=$GOPATH/bin:$PATH
export KUBEAPPS_DIR=$GOPATH/src/github.com/kubeapps/kubeapps
```

## Download the Monocular source code

```bash
git clone https://github.com/kubeapps/kubeapps $KUBEAPPS_DIR
```

The `asset-syncer` sources are located under the `cmd/asset-syncer/` directory.

### Install Kubeapps in your cluster

Kubeapps is a Kubernetes-native application. To develop and test Kubeapps components we need a Kubernetes cluster with Kubeapps already installed. Follow the [Kubeapps installation guide](../../chart/kubeapps/README.md) to install Kubeapps in your cluster.

### Building the `chart-repo` image

```bash
cd $KUBEAPPS_DIR
make kubeapps/asset-syncer
```

This builds the `asset-syncer` Docker image.

### Running in development

When using MongoDB:

```bash
export DB_PASSWORD=$(kubectl get secret --namespace kubeapps kubeapps-mongodb -o go-template='{{index .data "mongodb-root-password" | base64decode}}')
telepresence --namespace kubeapps --docker-run -e DB_PASSWORD=$DB_PASSWORD --rm -ti kubeapps/asset-syncer /asset-syncer sync --database-user=root --database-url=kubeapps-mongodb --database-type=mongodb --database-name=charts stable https://kubernetes-charts.storage.googleapis.com
```

When using PostgreSQL:

```bash
export DB_PASSWORD=$(kubectl get secret --namespace kubeapps kubeapps-db -o go-template='{{index .data "postgresql-password" | base64decode}}')
telepresence --namespace kubeapps --docker-run -e DB_PASSWORD=$DB_PASSWORD --rm -ti kubeapps/asset-syncer /asset-syncer sync --database-user=postgres --database-url=kubeapps-postgresql:5432 --database-type=postgresql --database-name=assets stable https://kubernetes-charts.storage.googleapis.com
```

Note that the chart-repo should be rebuilt for new changes to take effect.

### Running tests

You can run the chart-repo tests along with the tests for the Monocular project:

```bash
go test -v ./...
```
