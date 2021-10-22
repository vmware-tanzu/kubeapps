# Setup Kubeapps testing environment

This guide explains how to setup your environment to test Kubeapps integration with other services.

## Background

Kubeapps can be integrated with other services to extend its capabilities. Find more information about these integrations in the link below:

- [Using Private App Repositories with Kubeapps](../user/private-app-repository.md).

This guide aims to provide the instructions to easily setup the environment to test these integrations.

## Prerequisites

- [Kubernetes cluster (v1.12+)](https://kubernetes.io/docs/setup/).
- [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/).
- [Helm client](https://helm.sh/docs/intro/install/).

## Environment setup

We are providing scripts to automatically setup both Kubeapps and the services to integrate on a K8s cluster. Find them under the [script](../../script) directory.

Currently supported integrations:

- Kubeapps integration with Harbor.

### Kubeapps integration with Harbor

You can setup environment to test Kubeapps integration with Harbor using the scripts below:

- [setup-kubeapps](./scripts/setup-kubeapps.sh).
- [setup-harbor](./scripts/setup-harbor.sh).

These scripts will create the necessary namespaces, install the charts, wait for them to be available, and perform any extra action that might be needed. Find detailed information about how to use these scripts running the commands below:

```bash
./setup-kubeapps.sh --help
./setup-harbor.sh --help
```

You can also use the [setup-kubeapps-harbor](./scripts/setup-kubeapps-harbor.sh) script which is a wrapper that uses both the scripts mentioned above with some default values:

- Install Harbor under the `harbor` namespace.
- Install Kubeapps under the `kubeapps` namespace.
- Adds Harbor as an extra initial repository to Kubeapps, based on its service hostname.

#### Cleaning up the environment

You can use the scripts [delete-kubeapps](./scripts/delete-kubeapps.sh) and [delete-harbor](./scripts/delete-harbor.sh) to uninstall Kubeapps and Harbor releases from the cluster, respectively. These scripts will also remove the associated namespaces and resources.

> Note: you can use the [delete-kubeapps-harbor](./scripts/delete-kubeapps-harbor.sh) script to clean up the environment if you used the [setup-kubeapps-harbor](./scripts/setup-kubeapps-harbor.sh) script to setup the environment.
