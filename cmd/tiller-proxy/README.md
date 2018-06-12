# Tiller Proxy

This proxy is a service for Kubeapps that connects the Dashboard with Tiller. The goal of this Proxy is to provide a secure proxy for authenticated users to deploy, upgrade and delete charts in different namespaces.

Part of the logic of this tool has been extracted from [helm-CRD](https://github.com/bitnami-labs/helm-crd). That tool has been deprecated in Kubeapps to avoid having to synchronize the state of a release in two different places (Tiller and the CRD object).

The client should provide the header `Authorization: Bearer TOKEN` being TOKEN the Kubernetes API Token in order to perform any action.

The client should have permissions to:

 - "List" access to all the resources and all the groups for listing releases in a namespace.
 - "Read" access to all the release resources in a release when doing a HTTP GET over a specific release.
 - "Create" and "update" permissions to all the release resources when doing an HTTP POST or PUT to install or update a release.

# Configuration

It is possible to configure this proxy with the following flags:

```
      --debug                           enable verbose output
      --home string                     location of your Helm config. Overrides $HELM_HOME (default "/Users/andresmartinez/.helm")
      --host string                     address of Tiller. Overrides $HELM_HOST
      --kube-context string             name of the kubeconfig context to use
      --tiller-connection-timeout int   the duration (in seconds) Helm will wait to establish a connection to tiller (default 300)
      --tiller-namespace string         namespace of Tiller (default "kube-system")
```

# Routes

This proxy provides 6 different routes:

 - `GET` `/v1/releases`: List all the releases of the Tiller
 - `GET` `/namespaces/{namespace}/releases`: List all the releases within a namespace
 - `POST` `/namespaces/{namespace}/releases`: Create a new release
 - `GET` `/namespaces/{namespace}/releases/{release}`: Get release info
 - `PUT` `/namespaces/{namespace}/releases/{release}`: Update release info
 - `DELETE` `/namespaces/{namespace}/releases/{release}`: Delete a release
