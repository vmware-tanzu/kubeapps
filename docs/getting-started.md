# Getting Started

Kubeapps is a set of tools written by [Bitnami](https://bitnami.com) to super-charge your Kubernetes cluster with:
 * Your own applications [dashboard](https://kubeapps.com/), allowing you to deploy Kubernetes-ready applications into your cluster with a single click.
 * [Kubeless](http://kubeless.io/) - a Kubernetes-native Serverless Framework, compatible with [serverless.com](https://serverless.com).
 * [SealedSecrets](https://github.com/bitnami/sealed-secrets) - a SealedSecret can be decrypted only by the controller running in the cluster and nobody else (not even the original author).

These tools are easily deployed into your cluster with just one command: ```kubeapps up``` 

## Installation of the Kubeapps Installer

Kubeapps assumes a working Kubernetes (v1.7+) with RBAC enabled and [Kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) installed in your machine and configured to talk to your Kubernetes cluster. Kubeapps has been tested in `minikube` and Google Kubernetes Engine.

- Download latest Kubeapps Installer binary from the [release page](https://github.com/kubeapps/kubeapps/releases). Currently Kubeapps Installer is distributed in two platforms: linux/amd64 and OSX/amd64
- Make the binary executable

## Deploy the Kubeapps components into your cluster

```
kubeapps up
```

This command will deploy the Kubeapps components into your cluster. The deployments may take few minutes until they are ready.

## What is it that I am actually deploying?

You can check the [latest version of the Kubernetes manifest](https://github.com/kubeapps/kubeapps/blob/master/static/kubeapps-objs.yaml) that Kubeapps will deploy for you.

Check the [components documentation page](components.md) for a brief description of the list of components Kubeapps is deploying into your cluster.

## Kubeapps Dashboard

Kubeapps provides an in-cluster toolset for simplified deployment of over 100 Kubernetes ready applications as Helm charts and Kubeless functions.

To open the Kubeapps Dashboard:

```
kubeapps dashboard
```
Check the [documentation specific to the Dashboard](dashboard.md)

## Clean the Kubeapps components

You can easily clean all the Kubeapps deployment with a single command:

```
kubeapps down
```