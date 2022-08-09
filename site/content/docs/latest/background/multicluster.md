# Multicluster mode on Kubeapps

Kubeapps is a tool for managing applications in Kubernetes clusters. Initially it was possible to work only in a single cluster, but with the implementation of the [proposal for multicluster support in Kubeapps](https://github.com/vmware-tanzu/kubeapps/blob/main/site/content/docs/latest/reference/proposals/multi-cluster-support.md), this is possible also in additional clusters.

As a global overview, multicluster mode is based on a _default cluster_ (the one acting as orchestrator, with Kubeapps installed on it) and _additional clusters_ (target clusters without Kubeapps, where apps will be installed). Actions are performed by users in the default cluster but results are applied in a chosen additional cluster.

## Requirements for multiclustering

The main requirement for working in multicluster mode is that users can be authenticated with multiple clusters using the same credentials. This is usually achieved by sharing the same OIDC client, or having the OIDC provider configured so that tokens in one cluster are allowed to include the client-ids used by the other clusters. OIDC setup can be done directly in K8s API server (with `--oidc*` flags) or [using an OIDC provider with Pinniped](https://github.com/vmware-tanzu/kubeapps/blob/main/site/content/docs/latest/howto/OIDC/using-an-OIDC-provider-with-pinniped.md).

For more information on requirements and how to get Kubeapps working in multicluster mode, read the [how-to Configuring Kubeapps for multiple clusters](https://github.com/vmware-tanzu/kubeapps/blob/main/site/content/docs/latest/howto/deploying-to-multiple-clusters.md).

## Features

Kubeapps offers many features, but not all are available in the multicluster mode. In the following table there is the breakdown of features for single and additional clusters.

| Feature group | Kubeapps runtime feature                                          | Default cluster | Additional clusters | Comments                                                                                           |
| ------------- | ----------------------------------------------------------------- | --------------- | ------------------- | -------------------------------------------------------------------------------------------------- |
| Global        | List namespaces                                                   | Yes             | Yes                 |                                                                                                    |
|               | Create namespace                                                  | Yes             | Yes                 |                                                                                                    |
|               | Change context                                                    | Yes             | Yes                 |                                                                                                    |
| Repositories  | List package repositories                                         | Yes             | No                  |                                                                                                    |
|               | Add/Update/Delete package repository - Global                     | Yes             | No                  | Source of truth for package repositories is the default cluster                                    |
|               | Add/Update/Delete package repository - Namespaced                 | Yes             | No                  | Source of truth for package repositories is the default cluster                                    |
| Clusters      | Add/Update/Delete cluster                                         | No              | No                  | Clusters can only be defined at deployment time through input values                               |
| Packages      | HELM - List available packages from global public repository      | Yes             | Yes                 |                                                                                                    |
|               | HELM - List available packages from namespaced public repository  | Yes             | No                  |                                                                                                    |
|               | HELM - List available packages from global private repository     | Yes             | Yes                 |                                                                                                    |
|               | HELM - List available packages from namespaced private repository | Yes             | No                  |                                                                                                    |
|               | CARVEL - List available packages from global repository           | Yes             | No\*                | \* It could be done if kubeconfig provided to Kapp                                                 |
|               | CARVEL - List available packages from namespaced repository       | Yes             | No\*                | \* It could be done if kubeconfig provided to Kapp                                                 |
|               | FLUX - List available packages                                    | Yes             | No\*                | \* Throws error message `not supported yet: request.Context.Cluster: \[%v\]`                     |
|               | HELM - List installed packages in namespace                       | Yes             | Yes                 |                                                                                                    |
|               | CARVEL - List installed packages in namespace                     | Yes             | No\*                | \* It could be done if kubeconfig provided to Kapp                                                 |
|               | FLUX - List installed packages in namespace                       | Yes             | No\*                | \* Throws error message `not supported yet: request.Context.Cluster: \[%v\]`                     |
|               | HELM - Get installed package details                              | Yes             | Yes                 |                                                                                                    |
|               | CARVEL - Get installed package details                            | Yes             | No\*                | \* It could be done if kubeconfig provided to Kapp                                                 |
|               | FLUX - Get installed package details                              | Yes             | No                  |                                                                                                    |
|               | HELM - Package management (install, delete, etc.) without imagePullSecrets                   | Yes             | Yes                 | In additional clusters, only possible from Global repositories                                     |
|               | HELM - Package management (install, delete, etc.) with imagePullSecrets                      | Yes             | No                  |                                                                                                    |
|               | CARVEL - Package management (install, delete, etc.)                                         | Yes             | No\*                | \* Throws error message `installing packages in other clusters in not supported yet`             |
|               | FLUX - Package management (install, delete, etc.)                                         | Yes             | No\*                | \* Throws error message `not supported yet: request.AvailablePackageRef.Context.Cluster: \[%v\]` |

## Limitations

As it can be seen in the table of features, Kubeapps can work in multicluster mode only by using Helm plugin together with global repositories.
At the moment, dynamically managing additional clusters, Carvel and Flux operations, or namespaced repositories are not supported.
