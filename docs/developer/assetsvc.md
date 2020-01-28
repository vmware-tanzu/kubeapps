# Kubeapps assetsvc Developer Guide

The `assetsvc` component is a micro-service that creates an API endpoint for accessing the metadata for charts in Helm chart repositories that's populated in a MongoDB server.

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

## Download the Kubeapps source code

```bash
git clone https://github.com/kubeapps/kubeapps $KUBEAPPS_DIR
```

The `assetsvc` sources are located under the `cmd/assetsvc/` directory.

### Install Kubeapps in your cluster

Kubeapps is a Kubernetes-native application. To develop and test Kubeapps components we need a Kubernetes cluster with Kubeapps already installed. Follow the [Kubeapps installation guide](../../chart/kubeapps/README.md) to install Kubeapps in your cluster.

### Building the `assetsvc` image

```bash
cd $KUBEAPPS_DIR
make kubeapps/assetsvc
```

This builds the `assetsvc` Docker image.

### Running in development

#### Option 1: Using Telepresence (recommended)

When using MongoDB:

```bash
telepresence --swap-deployment kubeapps-internal-assetsvc --namespace kubeapps --expose 8080:8080 --docker-run --rm -ti kubeapps/assetsvc /assetsvc --database-user=root --database-url=kubeapps-mongodb --database-type=mongodb --database-name=charts
```

When using PostgreSQL:

```bash
telepresence --swap-deployment kubeapps-internal-assetsvc --namespace kubeapps --expose 8080:8080 --docker-run --rm -ti kubeapps/assetsvc /assetsvc --database-user=postgres --database-url=kubeapps-postgresql:5432 --database-type=postgresql --database-name=assets
```

Note that the assetsvc should be rebuilt for new changes to take effect.

#### Option 2: Replacing the image in the assetsvc Deployment

Note: By default, Kubeapps will try to fetch the latest version of the image so in order to make this workflow work in Minikube you will need to update the imagePullPolicy first:

```bash
kubectl patch deployment kubeapps-internal-assetsvc -n kubeapps --type=json -p='[{"op": "replace", "path": "/spec/template/spec/containers/0/imagePullPolicy", "value": "IfNotPresent"}]'
```

```bash
kubectl set image -n kubeapps deployment kubeapps-internal-assetsvc assetsvc=kubeapps/assetsvc:latest
```

For further redeploys you can change the version to deploy a different tag or rebuild the same image and restart the pod executing:

```bash
kubectl delete pod -n kubeapps -l app=kubeapps-internal-assetsvc
```

Note: If you using a cloud provider to develop the service you will need to retag the image and push it to a public registry.

### Running tests

You can run the assetsvc tests along with the tests for the Kubeapps project:

```bash
go test -v ./...
```
