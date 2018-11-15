# Kubeapps Tiller Proxy Developer Guide

The `tiller-proxy` component is a micro-service that creates a API endpoint for accessing the Helm Tiller server.

## Prerequisites

- [Git](https://git-scm.com/)
- [Make](https://www.gnu.org/software/make/)
- [Go programming language](https://golang.org/dl/)
- [Docker CE](https://www.docker.com/community-edition)
- [Kubernetes cluster (v1.8+)](https://kubernetes.io/docs/setup/pick-right-solution/). [Minikube](https://github.com/kubernetes/minikbue) is recommended.
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

The `tiller-proxy` sources are located under the `cmd/tiller-proxy/` and it uses packages from the `pkg` directory.

### Install Kubeapps in your cluster

Kubeapps is a Kubernetes-native application. To develop and test Kubeapps components we need a Kubernetes cluster with Kubeapps already installed. Follow the [Kubeapps installation guide](../../chart/kubeapps/README.md) to install Kubeapps in your cluster.

### Building the `tiller-proxy` binary

```bash
cd $KUBEAPPS_DIR/cmd/tiller-proxy
go build
```

This builds the `tiller-proxy` binary in the working directory.

### Running in development

If you are using Minikube it is important to start the cluster enabling RBAC (on by default in Minikube 0.26+) in order to check the authorization features:

```
minikube start
eval $(minikube docker-env)
```

Note: By default, Kubeapps will try to fetch the latest version of the image so in order to make this workflow work in Minikube you will need to update the imagePullPolicy first:

```
kubectl patch deployment kubeapps-internal-tiller-proxy -n kubeapps --type=json -p='[{"op": "replace", "path": "/spec/template/spec/containers/0/imagePullPolicy", "value": "IfNotPresent"}]'
```

The easiest way to create the `tiller-proxy` image is execute the Makefile task to do so:

```bash
IMAGE_TAG=dev make kubeapps/tiller-proxy
```

This will generate an image `kubeapps/tiller-proxy:dev` that you can use in the current deployment:

```
kubectl set image -n kubeapps deployment kubeapps-internal-tiller-proxy proxy=kubeapps/tiller-proxy:dev
```

For further redeploys you can change the version to deploy a different tag or rebuild the same image and restart the pod executing:

```
kubectl delete pod -n kubeapps -l app=kubeapps-internal-tiller-proxy
```

Note: If you using a cloud provider to develop the service you will need to retag the image and push it to a public registry.

### Running tests

You can run the tiller-proxy tests along with the tests of all the projecs

```bash
make test
```
