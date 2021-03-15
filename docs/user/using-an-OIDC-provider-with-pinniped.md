# Using an OIDC provider with Pinniped

The [Pinniped project](https://pinniped.dev/) exists to "Simplify user authentication for any Kubernetes cluster" and enables OIDC providers to be configured dynamically, rather than when a cluster is created. Kubeapps can be configured so that users must authenticate with the same OIDC provider and then authenticated requests from Kubeapps to the API server will be proxied via Pinniped, with the signed OIDC `id_token` being verified by Pinniped and exchanged for a client certificate accepted trusted by the API server.

## Installing Pinniped

Install Pinniped 0.6.0 into a `pinniped-concierge` namespace on your cluster with:

```bash
kubectl apply -f https://github.com/vmware-tanzu/pinniped/releases/download/v0.6.0/install-pinniped-concierge.yaml
```

**NOTE**: Due to a breaking change in [Pinniped 0.6.0](https://github.com/vmware-tanzu/pinniped/releases/tag/v0.6.0), the minimum version supported by Kubeapps is 0.6.0. Furthermore, [custom API suffixes](https://pinniped.dev/posts/multiple-pinnipeds) (introduced in Pinniped 0.5.0) are not yet fully supported. If your platform uses this feature, please [drop us an issue](https://github.com/kubeapps/kubeapps/issues/new).



## Configure Pinniped to trust your OIDC identity provider

Once Pinniped is running, you can add a `JWTAuthenticator` custom resource so that Pinniped knows to trust your OIDC identity provider. As an example, here is the `JWTAuthenticator` resource used in a local development environment where the Dex OIDC identity provider is running at `https://172.18.0.2:32000`:

```yaml
kind: JWTAuthenticator
apiVersion: authentication.concierge.pinniped.dev/v1alpha1
metadata:
  name: jwt-authenticator
spec:
  issuer: https://172.18.0.2:32000
  audience: default
  claims:
    groups: groups
    username: email
  tls:
    certificateAuthorityData: <removed-for-clarity>
```

> Note that in TMC, `authentication.concierge.pinniped.dev/v1alpha1` will become `authentication.concierge.pinniped.tmc.cloud.vmware.com/v1alpha1`

When the `pinniped-proxy` service of Kubeapps requests to exchange a JWT `id_token` for client certificates, Pinniped will verify the `id_token` is signed by the issuer identified here. Once verified, the claims for `username` and `groups` will be included on the generated client certificate so that the Kubernetes API server knows the username and groups associated with the request.

Note that the `spec.tls.certificateAuthorityData` field is required only if your TLS cert is signed by your own private certificate authority.

## Configuring Kubeapps to proxy requests via Pinniped

Ensure that the Kubeapps chart includes the pinniped service by enabling it in your values with:

```yaml
pinnipedProxy:
  enabled: true
```

Finally, because Kubeapps can be configured with multiple clusters, some of which may run with API servers configured with OIDC while others may be running Pinniped, your clusters configuration will need to identify that a specific cluster has pinniped enabled:

```yaml
clusters:
 - name: default
   pinnipedConfig:
    enable: true
```

The [Kubeapps auth-proxy configuration](./using-an-OIDC-provider.md#deploying-an-auth-proxy-to-access-kubeapps) remains the same as for the standard OIDC setup so that Kubeapps knows to deploy the auth-proxy service configured to redirect to your OIDC provider.

With those changes, Kubeapps is ready to send any request for a specific cluster via Pinniped so that the OIDC `id_token` can be exchanged for client certificates accepted by the Kubernetes API server.
