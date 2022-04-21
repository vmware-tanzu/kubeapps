# How-to guides

How-to guides can be thought of as recipes, directions that guide the reader through the steps to achieve a specific end. They’ll help you achieve an end result but may require you to understand and adapt the steps to fit your specific requirements. Here you’ll find short answers to “How do I….?” types of questions.

| How-to-guides                                                                          | Get stuff done                                                                                                                            |
| -------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------- |
| [Access Control](./access-control.md)                                                  | Establish how users will authenticate with the Kubernetes clusters on which Kubeapps operates.                                            |
| [Basic Form Support](./basic-form-support.md)                                          | Configure your Helm chart in order to present a simple intuitive form during installation.                                                               |
| [Custom App View Support](./custom-app-view-support.md)                                | Inject custom app views for specific deployments.                                                                                         |
| [Custom Form Component Support](./custom-form-component-support.md)                    | Extend basic form with custom UI component or thrid party APIs for component values and validation.                                       |
| [Dashboard](./dashboard.md)                                                            | Manage and deploy applications in your cluster by using Kubeapps dashboard.                                                               |
| [Multi-cluster Support](./deploying-to-multiple-clusters.md)                           | Configure Kubeapps to target other clusters when deploying a package, in addition to the cluster on which Kubeapps is itself deployed.    |
| [Offline Installation](./offline-installation.md)                                      | Install Kubeapps in an offline environment (without Internet connection)                                                                  |
| [Private App Repository](./private-app-repository.md)                                  | Configure Kubeapps to use a private app repository.                                                                                       |
| [Syncing App Repositories](./syncing-apprepository-webhook.md)                         | Change default configuration for scheduling the syncing process of the App Repositories (globally or specific for a given App Repository) |
| [Using an OIDC provider with Pinniped](./OIDC/using-an-OIDC-provider-with-pinniped.md) | Install and configure Pinniped in your cluster to trust your identity provider and configure Kubeapps to proxy requests via Pinniped.     |

Alternatively, our [Tutorials section](../tutorials/README.md) contain step-by-step tutorials to help outline what Kubeapps is capable of while helping you achieve specific aims, such as installing Kubeapps or using an OIDC provider.

Take a look at our [Reference section](../reference/README.md) when you need to know design decisions, what functions the API supports, detailed developer guides, etc.

Finally, for a better understanding of Kubeapps architecture, our [Background section](../background/README.md) enable you to expand your knowledge.
