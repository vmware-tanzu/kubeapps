# Developers Guide

This guide explains how to set up your environment for developing on Helm and Tiller.

## Prerequisites
* Docker 1.10 or later
* Docker Compose 1.6 or later
* A Kubernetes cluster with Helm/Tiller installed (optional)
* kubectl 1.2 or later (optional)
* Git

## Running Monocular

We use [Docker Compose](https://docs.docker.com/compose/) to orchestrate the UI frontend and API backend containers for local development. The simplest way to get started is:

```
$ docker-compose up
```

After a few minutes, you will be able to visit http://localhost:4200/ in your browser.

### Running a local cluster

Monocular allows you to install and view installed charts in your Kubernetes cluster, to test and develop these features we highly recommend using the Kubernetes Minikube developer-oriented distribution. Once this is installed, you can use helm init to install Tiller into the cluster.

```
$ minikube start
$ helm init
$ kubectl config set-credentials minikube --client-certificate=$HOME/.minikube/apiserver.crt --client-key=$HOME/.minikube/apiserver.key --embed-certs
$ kubectl config set-cluster minikube --certificate-authority=$HOME/.minikube/ca.crt --embed-certs
```

Since your kubeconfig is mounted into the API container, we embed the certificates to authenticate with the cluster.

## Architecture

The UI is an Angular 2 application located in `src/ui/`. This path is mounted into the UI container. The server watches for file changes and automatically rebuilds the application.

* [UI documentation](../src/ui/README.md)

The API is a Go HTTP server located in `src/api/`.

* [API documentation](../src/api/README.md)
