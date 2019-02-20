# Kubeapps Login Form

By default, when first visiting Kubeapps, a login form is shown for the user to introduce a personal token for a specific service account:

![Dashboard Login](../img/dashboard-login.png)

You can find more information about access control in Kubeapps in this [document](./access-control.md).

However, it's possible to disable this form or delegate the authentication to an OIDC provider so Kubeapps users don't need to introduce a token in the login form.

## Bypassing authentication

Kubeapps expects an `Authorization` header that will be used to validate operations agains the Kubernetes API. This is usually set with the login form but if Kubeapps is exposed with an Ingress object, it's possible to hardcode a valid token in the Ingress configuration and automatically include it in all the requests.

**NOTE**: This is not suitable for production since anyone with access to Kubeapps would be granted with the permissions associated with the hardcoded token.

This is an example of the values that you can configure in the Kubeapps chart in order to set a valid token:

```yaml
ingress:
  enabled: true
  hosts:
    - name: kubeapps.local
      path: /
      tls: false
  annotations:
    nginx.ingress.kubernetes.io/proxy-read-timeout: "600"
    nginx.ingress.kubernetes.io/configuration-snippet: |
      add_header Authorization "Bearer TOKEN";
```

You just need to substitute TOKEN with the actual value of the token. The above assumes an Nginx [Ingress Controller](https://kubernetes.io/docs/concepts/services-networking/ingress/#ingress-controllers), in case a different controller is being used that annotation will need to be adapted.

# Using an OIDC provider

In case you want to use OAuth 2.0 to authenticate Kubeapps users follow this [guide](./using-an-OIDC-provider.md).
