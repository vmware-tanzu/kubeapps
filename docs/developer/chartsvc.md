# Kubeapps chartsvc Developer Guide

The `chartsvc` component is a micro-service that creates an API endpoint for accessing the metadata for charts in Helm chart repositories that's populated in a MongoDB server. Its source is maintained in the [Monocular project repository](https://github.com/helm/monocular).

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

The `chartsvc` sources are located under the `cmd/chartsvc/` directory.

### Install Kubeapps in your cluster

Kubeapps is a Kubernetes-native application. To develop and test Kubeapps components we need a Kubernetes cluster with Kubeapps already installed. Follow the [Kubeapps installation guide](../../chart/kubeapps/README.md) to install Kubeapps in your cluster.

### Building the `chartsvc` image

```bash
cd $MONOCULAR_DIR
dep ensure
make -C cmd/chartsvc docker-build
```

This builds the `chartsvc` Docker image.

### Running in development

#### Option 1: Using Telepresence (recommended)

```bash
telepresence --swap-deployment kubeapps-internal-chartsvc --namespace kubeapps --expose 8080:8080 --docker-run --rm -ti quay.io/helmpack/chartsvc /chartsvc --mongo-user=root --mongo-url=kubeapps-mongodb
```

Note that the chartsvc should be rebuilt for new changes to take effect.

#### Option 2: Replacing the image in the chartsvc Deployment

Note: By default, Kubeapps will try to fetch the latest version of the image so in order to make this workflow work in Minikube you will need to update the imagePullPolicy first:

```
kubectl patch deployment kubeapps-internal-chartsvc -n kubeapps --type=json -p='[{"op": "replace", "path": "/spec/template/spec/containers/0/imagePullPolicy", "value": "IfNotPresent"}]'
```

```
kubectl set image -n kubeapps deployment kubeapps-internal-chartsvc chartsvc=quay.io/helmpack/chartsvc:latest
```

For further redeploys you can change the version to deploy a different tag or rebuild the same image and restart the pod executing:

```
kubectl delete pod -n kubeapps -l app=kubeapps-internal-chartsvc
```

Note: If you using a cloud provider to develop the service you will need to retag the image and push it to a public registry.

### Running tests

You can run the chartsvc tests along with the tests for the Monocular project:

```bash
go test -v ./...
```
