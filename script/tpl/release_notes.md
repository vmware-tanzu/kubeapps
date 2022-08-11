Kubeapps vX.Y.Z (chart version A.B.C) is a <major|minor|patch> release that... <!-- ADD SUMMARY HERE -->

## Installation

To install this release, ensure you add the [Bitnami charts repository](https://github.com/bitnami/charts) to your local Helm cache:

```bash
helm repo add bitnami https://charts.bitnami.com/bitnami
helm repo update
```

Install the Kubeapps Helm chart:

```bash
kubectl create namespace kubeapps
helm install kubeapps --namespace kubeapps bitnami/kubeapps
```

To get started with Kubeapps, check out this [walkthrough](https://github.com/vmware-tanzu/kubeapps/blob/main/site/content/docs/latest/tutorials/getting-started.md).

<!-- CLICK ON THE "Auto-generate release notes" BUTTON -->
