<!-- ADD SUMMARY HERE -->

## Installation

To install this release, ensure you add the [Bitnami charts repository](https://github.com/bitnami/charts) to your local Helm cache:

```
helm repo add bitnami https://charts.bitnami.com/bitnami
helm repo update
```

Install the Kubeapps Helm chart:

For Helm 2:

```
helm install --name kubeapps --namespace kubeapps bitnami/kubeapps
```

For Helm 3:

```
kubectl create namespace kubeapps
helm install kubeapps --namespace kubeapps bitnami/kubeapps --set useHelm3=true
```

To get started with Kubeapps, checkout this [walkthrough](https://github.com/kubeapps/kubeapps/blob/master/docs/user/getting-started.md).

## Changelog

