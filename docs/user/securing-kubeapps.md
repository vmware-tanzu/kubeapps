# Securing Kubeapps installation

In this guide we will explain how to secure the installation of Kubeapps in a multi-tenant cluster. Following these steps are only necessary if different people with different permissions have access to the same cluster. Generic instructions to secure Helm can be found [here](https://github.com/kubernetes/helm/blob/master/docs/securing_installation.md).

The main goal is to secure the access to [Tiller](https://github.com/kubernetes/helm/blob/master/docs/securing_installation.md) (Helm server-side component). Tiller has access to create or delete any resource in the cluster so we should be careful on how we expose the functionality it provides.

In order to take advantage of Kubeapps security features you will need to configure two things: a **TLS certificate** to control the access to Tiller and [**RBAC roles**](https://kubernetes.io/docs/reference/access-authn-authz/rbac/) to authorize requests.

## Install Tiller Securely

You can follow the Helm documentation for deploying Tiller in a secure way. In particular we are interested in:

- Using a TLS certificate to control the access to the Tiller deployment: https://docs.helm.sh/using_helm/#using-ssl-between-helm-and-tiller
- Storing release info as secrets: https://docs.helm.sh/using_helm/#tiller-s-release-information

From these guides you can find out how to create the TLS certificate and the necessary flags to install Tiller securely:

```
helm init --tiller-tls --tiller-tls-verify \
  --override 'spec.template.spec.containers[0].command'='{/tiller,--storage=secret}' \
  --tiller-tls-cert ./tiller.cert.pem \
  --tiller-tls-key ./tiller.key.pem \
  --tls-ca-cert ca.cert.pem
```

## Deploy Kubeapps with a TLS certificate

This is the command to install Kubeapps with our certificate:

```
helm repo add bitnami https://charts.bitnami.com/bitnami
helm install \
  --tls --tls-ca-cert ca.cert.pem --tls-cert helm.cert.pem --tls-key helm.key.pem \
  --set tillerProxy.tls.ca="$(cat ca.cert.pem)" \
  --set tillerProxy.tls.key="$(cat helm.key.pem)" \
  --set tillerProxy.tls.cert="$(cat helm.cert.pem)" \
  --namespace kubeapps \
  bitnami/kubeapps
```

> Note: To use the `tls-verify` flag (and validate Tiller hostname), the certificate should have configured the host of Tiller within the cluster: `tiller-deploy.kube-system` by default.

## Enable RBAC

In order to be able to authorize requests from users it is necessary to enable [RBAC](https://kubernetes.io/docs/reference/access-authn-authz/rbac/) in the Kubernetes cluster. Some providers have it enabled by default but in some cases you need to set it up explicitly. Check out your provider documentation to know how to enable it. To verify if your cluster has RBAC available you can check if the API group exists:

```
$ kubectl api-versions | grep rbac.authorization
rbac.authorization.k8s.io/v1
```

Once your cluster has RBAC enabled read [this document](/docs/user/access-control.md) to know how to login in Kubeapps using a token that identifies a user account and how you can create users with different permissions.

In a nutshell, Kubeapps authorization validates:

- When getting a release details, it checks that the user have "read" access to all the components of the release.
- When creating, upgrading or deleting a release it checks that the user is allowed to create, update or delete all the components contained in the release chart.

For example, if the user account `foo` wants to deploy a chart `bar` that is composed of a `Deployment` and a `Service` it should have enough permissions to create each one of those. In other case it will receive an error message with the missing permissions required to deploy the chart.
