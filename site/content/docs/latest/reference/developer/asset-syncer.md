# Kubeapps asset-syncer Developer Guide

The `asset-syncer` component is a tool that scans a Helm chart repository and populates chart metadata in the database. This metadata is then served by the Helm plugin of the `kubeapps-apis` component.

## Prerequisites

- [Git](https://git-scm.com/)
- [Make](https://www.gnu.org/software/make/)
- [Go programming language](https://go.dev/dl/)
- [Docker](https://docs.docker.com/engine/install/)
- [Kubernetes cluster](https://kubernetes.io/docs/setup/).
  - [Kind](https://kind.sigs.k8s.io/docs/user/quick-start/) is recommended.
- [kubectl](https://kubernetes.io/docs/tasks/tools/#kubectl)
- [Telepresence](https://telepresence.io)
  - _Telepresence is not a hard requirement, but is recommended for a better developer experience_

## Download the Kubeapps source code

```bash
git clone https://github.com/vmware-tanzu/kubeapps $KUBEAPPS_DIR
```

The `asset-syncer` sources are located under the `cmd/asset-syncer/` directory.

### Install Kubeapps in your cluster

Kubeapps is a Kubernetes-native application. To develop and test Kubeapps components we need a Kubernetes cluster with Kubeapps already installed. Follow the [Kubeapps installation guide](https://github.com/vmware-tanzu/kubeapps/blob/main/chart/kubeapps/README.md) to install Kubeapps in your cluster.

### Building the `asset-syncer` image

```bash
cd $KUBEAPPS_DIR
make kubeapps/asset-syncer
```

This builds the `asset-syncer` Docker image.

### Running in development

```bash
export DB_PASSWORD=$(kubectl get secret --namespace kubeapps kubeapps-postgresql -o go-template='{{index .data "postgres-password" | base64decode}}')
telepresence --namespace kubeapps --docker-run -e DB_PASSWORD=$DB_PASSWORD --rm -ti kubeapps/asset-syncer /asset-syncer sync --database-user=postgres --database-url=kubeapps-postgresql:5432 --database-name=assets stable https://kubernetes-charts.storage.googleapis.com
```

Note that the asset-syncer should be rebuilt for new changes to take effect.

### Running tests

You can run the asset-syncer tests along with the tests for the Kubeapps project:

```bash
go test -v ./...
```
