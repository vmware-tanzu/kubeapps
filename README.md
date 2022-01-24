# <img src="./docs/img/logo.png" width="40" align="left"> Kubeapps

[![CircleCI](https://circleci.com/gh/kubeapps/kubeapps/tree/main.svg?style=svg)](https://circleci.com/gh/kubeapps/kubeapps/tree/main)

## Overview

Kubeapps is an in-cluster web-based application that enables users with a one-time installation to deploy, manage, and upgrade applications on a Kubernetes cluster.

With Kubeapps you can:

- Customize deployments through an intuitive, form-based user interface
- Inspect, upgrade and delete applications installed in the cluster
- Browse and deploy [Helm](https://github.com/helm/helm) charts from public or private chart repositories (including [VMware Marketplace™](https://marketplace.cloud.vmware.com) and [Bitnami Application Catalog](https://bitnami.com/application-catalog))
- Browse and deploy [Kubernetes Operators](https://operatorhub.io/)
- Secure authentication to Kubeapps using a [standalone OAuth2/OIDC provider](./docs/user/using-an-OIDC-provider.md) or [using Pinniped](./docs/user/using-an-OIDC-provider-with-pinniped.md)
- Secure authorization based on Kubernetes [Role-Based Access Control](./docs/user/access-control.md)

**_Note:_** Kubeapps 2.0 and onwards supports Helm 3 only. While only the Helm 3 API is supported, in most cases, charts made for Helm 2 will still work.

## Getting started with Kubeapps

Installing Kubeapps is as simple as:

```bash
helm repo add bitnami https://charts.bitnami.com/bitnami
kubectl create namespace kubeapps
helm install kubeapps --namespace kubeapps bitnami/kubeapps
```

See the [Getting Started Guide](./docs/user/getting-started.md) for detailed instructions on how to install and use Kubeapps.

> Kubeapps is deployed using the official [Bitnami Kubeapps chart](https://github.com/bitnami/charts/tree/master/bitnami/kubeapps) from the separate Bitnami charts repository. Although the Kubeapps repository also defines a chart, this is intended for development purposes only.

## Documentation

Please refer to:

- [Getting started guide](./docs/user/getting-started.md)
- [Detailed installation instructions](./chart/kubeapps/README.md)
- [Kubeapps user guide](./docs/user/dashboard.md) to easily manage your applications running in your cluster.
- [Kubeapps FAQs](./chart/kubeapps/README.md#faq).

See how to deploy and configure [Kubeapps on VMware Tanzu™ Kubernetes Grid™](./docs/step-by-step/kubeapps-on-tkg/README.md)

## Troubleshooting

If you encounter issues, review the [troubleshooting docs](./chart/kubeapps/README.md#troubleshooting), review our [project board](https://github.com/kubeapps/kubeapps/projects/11), file an [issue](https://github.com/kubeapps/kubeapps/issues), or talk to us on the [#Kubeapps channel](https://kubernetes.slack.com/messages/kubeapps) on the Kubernetes Slack server.

- Click [here](https://slack.k8s.io) to sign up to the Kubernetes Slack org.

- Review our FAQs section on the [Kubeapps chart README](./chart/kubeapps/README.md#faq).

## Contributing

If you are ready to jump in and test, add code, or help with documentation, follow the instructions on our [start contributing](./CONTRIBUTING.md) documentation for guidance on how to setup Kubeapps for development.

## Changelog

Take a look at the list of [releases](https://github.com/kubeapps/kubeapps/releases) to stay tuned for the latest features and changes.
