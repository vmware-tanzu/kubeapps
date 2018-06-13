# The Kubeapps Build Guide

This guide explains how to build Kubeapps.

## Prerequisites

- [Git](https://git-scm.com/)
- [Make](https://www.gnu.org/software/make/)
- [Go programming language](https://golang.org/)
- [kubecfg](https://github.com/ksonnet/kubecfg)
- [Docker CE](https://www.docker.com/community-edition)

## Environment setup

```bash
export GOPATH=~/gopath/
export PATH=$GOPATH/bin:$PATH
export KUBEAPPS_DIR=$GOPATH/src/github.com/kubeapps/kubeapps
```
## Download kubeapps source code

```bash
git clone --recurse-submodules https://github.com/kubeapps/kubeapps $KUBEAPPS_DIR
cd $KUBEAPPS_DIR
```

## Build kubeapps

```bash
make VERSION=myver
```

The above command builds the `kubeapps` binary in the current directory. The `VERSION` argument sets build version to `myver ` and when not specified the build date and time is used to generate the version string.

Kubeapps consists of a number of in-cluster components. To build all these components in one go:

```bash
make VERSION=myver all
```

Or if you wish to build specific component(s):

```bash
# to build the kubeapps binary
make VERSION=myver kubeapps

# to build the kubeapps/dashboard docker image
make VERSION=myver kubeapps/dashboard

# to build the kubeapps/chartsvc docker image
make VERSION=myver kubeapps/chartsvc

# to build the kubeapps/chart-repo docker image
make VERSION=myver kubeapps/chart-repo

# to build the kubeapps/apprepository-controller docker image
make VERSION=myver kubeapps/apprepository-controller
```

## Running tests

```bash
make test
```

The above command runs the tests for the `kubeapps` binary. To run the tests on all Kubeapps components:

```bash
make test-all
```

Or if you wish to test specific component(s):

```bash
# to test the kubeapps binary
make test test-kubeapps

# to test kubeapps/dashboard
make test test-dashboard

# to test the cmd/chartsvc package
make test test-chartsvc

# to test the cmd/chart-repo package
make test test-chart-repo

# to test the cmd/apprepository-controller package
make test test-apprepository-controller
```
