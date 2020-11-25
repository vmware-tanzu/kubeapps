# How to set up the environment using the provided makefiles

The main file is [Makefile](https://github.com/kubeapps/kubeapps/blob/master/Makefile), which will compile and prepare the production assets for then generating a set of Docker images. It is the starting point when you want to build the Kubeapps different components.

For setting up the environment for running Kubeapps, we also provide (as it) makefiles for:
* Creating a multicluster environment with Kind ([cluster-kind.mk](https://github.com/kubeapps/kubeapps/blob/master/script/cluster-kind.mk))
* Deploying and configuring the components for getting Kubeapps running with OIDC login using Dex ([deploy-dev.mk](https://github.com/kubeapps/kubeapps/blob/master/script/deploy-dev.mk)).

## Makefile for generating images

### Commands

Find below a list of the most used commands:

``` bash
make # will make all the kubeapps images
make kubeapps/dashboard
make kubeapps/apprepository-controller
make kubeapps/kubeops
make kubeapps/assetsvc
make kubeapps/asset-syncer
```

> You can set the image tag manually: `IMAGE_TAG=myTag make`

## Makefile for setting up the environment

### Prerequisites

* Install `mkcert`; you gan get it from the [official repository](https://github.com/FiloSottile/mkcert/releases).
* Get the Kind network IP and replace it when necessary.
    * Retrieve the IP by executing  `kubectl --namespace=kube-system get pods -o wide | grep kube-apiserver-kubeapps-control-plane  | awk '{print $6}'`
    * Replace `172.18.0.2` with the previous IP the following files:
        * [kubeapps-local-dev-additional-apiserver-config.yaml](../user/manifests/kubeapps-local-dev-additional-apiserver-config.yaml)
        * [kubeapps-local-dev-additional-kind-cluster.yaml](../user/manifests/kubeapps-local-dev-additional-kind-cluster.yaml)
        * [kubeapps-local-dev-apiserver-config.yaml](../user/manifests/kubeapps-local-dev-apiserver-config.yaml)
        * [kubeapps-local-dev-auth-proxy-values.yaml](../user/manifests/kubeapps-local-dev-auth-proxy-values.yaml)
        * [kubeapps-local-dev-dex-values.yaml](../user/manifests/kubeapps-local-dev-dex-values.yaml)
### Commands

```bash
# Create two cluster with RBAC and Nginx Ingress controller
# and configure the kube apiserver with the oidc flags
make multi-cluster-kind

# Install dex (identity service using OIDC),
# install openldap and add default users,
# generate certs for tls,
# and deploy kubeapps with the proper configuration 
make deploy-dev
```

> Many commands are using `...|| true`, so you will have to manually make sure that every command is, indeed, running OK by checking the logs.
