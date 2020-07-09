# Kubeapps Login

By default, Kubeapps allow you to login using a Kubernetes service account token. While this is enabled by default to make it easier to get started, the **recommended way of authenticate in Kubeapps is using an OAuth2/OIDC provider**. Depending on your cluster provider, the instructions may differ. Check this [guide](./using-an-OIDC-provider.md) for more information.

When using the service account authentication, the login form is shown for the user to introduce a Kubernetes API token:

![Dashboard Login](../img/dashboard-login.png)

The goal of the login form is to identify the user and associate it with a Kubernetes service account. This identity information will be used by Kubeapps to authenticate the user against the Kubernetes API. You can find more information about access control in Kubeapps in this [document](./access-control.md).
