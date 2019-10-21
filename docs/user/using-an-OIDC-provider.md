# Using an OAuth2/OIDC Provider with Kubeapps

OpenID Connect (OIDC) is a simple identity layer on top of the OAuth 2.0 protocol which allows clients to verify the identity of an user based on the authentication performed by an authorization server, as well as to obtain basic profile information about the user.

It is possible to configure your Kubernetes cluster to use an OIDC provider in order to manage accounts, groups and roles with a single application. Additionally, some managed Kubernetes environments enable authenticating via plain OAuth2 (GKE).
This guide will explain how you can use an existing OAuth2 provider, including OIDC, to authenticate users within Kubeapps.

## Pre-requisites

For this guide we assume that you have a Kubernetes cluster that is properly configured to use an Identity Provider (IdP) to handle the authentication to your cluster. You can find more information about how Kubernetes uses OIDC tokens [here](https://kubernetes.io/docs/reference/access-authn-authz/authentication/#openid-connect-tokens). This means that the Kubernetes API server should be configured to use that OIDC provider.

There are several Identity Providers (IdP) that can be used in a Kubernetes cluster. The steps of this guide have been validated using the following providers:

- [Keycloak](https://www.keycloak.org/): Open Source Identity and Access Management.
- [Dex](https://github.com/dexidp/dex): Open Source OIDC and OAuth 2.0 Provider with Pluggable Connectors.
- [Azure Active Directory](https://docs.microsoft.com/en-us/azure/active-directory/fundamentals/active-directory-whatis): Identity Provider that can be used for AKS.
- [Google OpenID Connect](https://developers.google.com/identity/protocols/OpenIDConnect): OAuth 2.0 for Google accounts.

For Kubeapps to use an Identity Provider it's necessary to configure at least the following parameters:

- Client ID: Client ID of the IdP.
- Client Secret: (If configured) Secret used to validate the Client ID.
- Provider name (which can be oidc, in which case the OIDC Issuer URL is also required).
- Cookie secret: a 16, 24 or 32 byte base64 encoded seed string used to encrypt sensitive data (eg. `echo "not-good-secret" | base64`).

**Note**: Depending on the configuration of the Identity Provider more parameters may be needed.

Kubeapps uses [OAuth2 Proxy](https://github.com/pusher/oauth2_proxy) to handle the OAuth2/OpenIDConnect authentication. The following sections explain how you can find the parameters above for some of the identity providers tested. If you have configured your cluster to use an Identity Provider you will already know some of these parameters. More detailed information can be found on the [OAuth2 Proxy Auth configuration page](https://pusher.github.io/oauth2_proxy/auth-configuration).

### Keycloak

In the case of Keycloak, you can find the parameters in the Keycloak admin console:

- Client-ID: Keycloak client ID.
- Client-secret: Secret associated to the client above.
- OIDC Issuer URL: `https://<keycloak.domain>/auth/realms/<realm>`.

### Dex

For Dex, you can find the parameters that you need to set in the configuration (a ConfigMap if Dex is deployed within the cluster) that the server reads the configuration from. Note that since Dex is only a connector you need to configure it with some third-party credentials that may be a client-id and client-secret as well. Don't confuse those credentials with the ones of the application that you can find under the `staticClients` section.

- Client-ID: Static client ID.
- Client-secret: Static client secret.
- OIDC Issuer URL: Dex URL (i.e. https://dex.example.com:32000).

### Azure Active Directory

For setting up an Azure Kubernetes cluster (aks) with Azure Active Directory you can follow [this guide](https://docs.microsoft.com/en-us/azure/aks/aad-integration). At the end of the tutorial you should have an Active Directory Application (Server). That's the Application from which we will get the needed parameters.

- Client-ID: Azure Active Directory server Application ID.
- Client-secret: A "Password" Key of the server Application.
- OIDC Issuer URL: `https://sts.windows.net/<Tenant-ID>/`. The Tenant-ID can be found at `Home > Default Directory - Properties > Directory ID`.

### Google OIDC

In the case of Google we can use an OAuth 2.0 client ID. You can find more information [here](https://developers.google.com/identity/protocols/OpenIDConnect). In particular we will use a "Web Application".

- Client-ID: `<abc>.apps.googleusercontent.com`.
- Client-Secret: Secret for the Web application.
- OIDC Issuer URL: https://accounts.google.com.

## Deploying a proxy to access Kubeapps

The main difference in the authentication is that instead of accessing the Kubeapps service, we will be accessing an oauth2 proxy service that is in charge of authenticating users with the identity provider and injecting the required credentials in the requests to Kubeapps. There are a number of available solutions for this use-case, like [keycloak-gatekeeper](https://github.com/keycloak/keycloak-gatekeeper) and [oauth2_proxy](https://github.com/pusher/oauth2_proxy). For this guide we will use `oauth2_proxy` since it supports both OIDC and plain OAuth2 for many providers.

Once the proxy is accessible, you will be redirected to the identity provider to authenticate. After successfully authenticating, you will be redirected to Kubeapps and be authenticated with your user's OIDC token.

The next sections explain how you can deploy this proxy either using the Kubeapps chart or manually.

### Using the chart

Kubeapps chart allows you to automatically deploy the proxy for you as a sidecar container if you specify the necessary flags. In a nutshell you need to enable the feature and set the client ID, secret and the IdP URL. The following examples use Google as the Identity Provider, modify the flags below to adapt them:

**Example 1: Using the OIDC provider**

This example uses `oauth2-proxy`'s generic OIDC provider with Google, but is applicable to any OIDC provider such as Keycloak, Dex, Okta or Azure Active Directory etc. Note that the issuer url is passed as an additional flag here, together with an option to enable the cookie being set over an insecure connection for local development only:

```
helm install bitnami/kubeapps \
  --namespace kubeapps --name kubeapps \
  --set authProxy.enabled=true \
  --set authProxy.provider=oidc \
  --set authProxy.clientID=my-client-id.apps.googleusercontent.com \
  --set authProxy.clientSecret=my-client-secret \
  --set authProxy.cookieSecret=$(echo "not-good-secret" | base64) \
  --set authProxy.additionalFlags="{-cookie-secure=false,-oidc-issuer-url=https://accounts.google.com}" \
```

**Example 2: Using a custom oauth2-proxy provider**

Some of the specific providers that come with `oauth2-proxy` are using OpenIDConnect to obtain the required IDToken and can be used instead of the generic oidc provider. Currently this includes only the GitLab, Google and LoginGov providers (see [OAuth2_Proxy's provider configuration](https://pusher.github.io/oauth2_proxy/auth-configuration) for the full list of OAuth2 providers). The user authentication flow is the same as above, with some small UI differences, such as the default login button is customized to the provider (rather than "Login with OpenID Connect"), or improved presentation when accepting the requested scopes (as is the case with Google, but only visible if you request extra scopes).

Here we no longer need to provide the issuer -url as an additional flag:
```
helm install bitnami/kubeapps \
  --namespace kubeapps --name kubeapps \
  --set authProxy.enabled=true \
  --set authProxy.provider=google \
  --set authProxy.clientID=my-client-id.apps.googleusercontent.com \
  --set authProxy.clientSecret=my-client-secret \
  --set authProxy.cookieSecret=$(echo "not-good-secret" | base64) \
  --set authProxy.additionalFlags="{-cookie-secure=false}"
```

**Example 3: Authentication for Kubeapps on a GKE cluster**

Google Kubernetes Engine does not allow an OIDC IDToken to be used to authenticate requests to the managed API server, instead requiring the standard OAuth2 access token.
For this reason, when deploying Kubeapps on GKE we need to ensure that

* The scopes required by the user to interact with cloud platform are included, and
* The Kubeapps frontend uses the OAuth2 `access_key` as the bearer token when communicating with the managed Kubernetes API

Note that using the custom `google` provider here enables google to prompt the user for consent for the specific permissions requested in the scopes below, in a user-friendly way. You can also use the `oidc` provider but in this case the user is not prompted for the extra consent:

```
helm install bitnami/kubeapps \
  --namespace kubeapps --name kubeapps \
  --set authProxy.enabled=true \
  --set authProxy.provider=google \
  --set authProxy.clientID=my-client-id.apps.googleusercontent.com \
  --set authProxy.clientSecret=my-client-secret \
  --set authProxy.cookieSecret=$(echo "not-good-secret" | base64) \
  --set authProxy.additionalFlags="{-cookie-secure=false,-scope=https://www.googleapis.com/auth/userinfo.email https://www.googleapis.com/auth/cloud-platform}" \
  --set frontend.proxypassAccessTokenAsBearer=true
```

### Manual deployment

In case you want to manually deploy the proxy, first you will create a Kubernetes deployment and service for the proxy. For the snippet below, you need to set the environment variables `AUTH_PROXY_CLIENT_ID`, `AUTH_PROXY_CLIENT_SECRET`, `AUTH_PROXY_DISCOVERY_URL` with the information from the IdP and `KUBEAPPS_NAMESPACE`.

```
export AUTH_PROXY_CLIENT_ID=<ID>
export AUTH_PROXY_CLIENT_SECRET=<SECRET>
export AUTH_PROXY_DISCOVERY_URL=<URL>
export AUTH_PROXY_COOKIE_SECRET=$(echo "not-good-secret" | base64)
kubectl create -n $KUBEAPPS_NAMESPACE -f - -o yaml << EOF
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    name: kubeapps-auth-proxy
  name: kubeapps-auth-proxy
spec:
  replicas: 1
  selector:
    matchLabels:
      name: kubeapps-auth-proxy
  strategy:
    rollingUpdate:
      maxSurge: 1
      maxUnavailable: 1
    type: RollingUpdate
  template:
    metadata:
      labels:
        name: kubeapps-auth-proxy
    spec:
      containers:
      - args:
        - -provider=oidc
        - -client-id=$AUTH_PROXY_CLIENT_ID
        - -client-secret=$AUTH_PROXY_CLIENT_SECRET
        - -oidc-issuer-url=$AUTH_PROXY_DISCOVERY_URL
        - -cookie-secret=$AUTH_PROXY_COOKIE_SECRET
        - -upstream=http://localhost:8080/
        - -http-address=0.0.0.0:3000
        - -email-domain="*"
        - -pass-basic-auth=false
        - -pass-access-token=true
        - -pass-authorization-header=true
        image: bitnami/oauth2-proxy
        imagePullPolicy: IfNotPresent
        name: kubeapps-auth-proxy
---
apiVersion: v1
kind: Service
metadata:
  labels:
    name: kubeapps-auth-proxy
  name: kubeapps-auth-proxy
spec:
  ports:
  - name: http
    port: 3000
    protocol: TCP
    targetPort: 3000
  selector:
    name: kubeapps-auth-proxy
  sessionAffinity: None
  type: ClusterIP
EOF
```

The above is a sample deployment, depending on the configuration of the Identity Provider those flags may vary. For this example we use:

- `-client-id`, `-client-secret` and `-oidc-issuer-url`: Client ID, Secret and IdP URL as stated in the section above.
- `-upstream`: Internal URL for the `kubeapps` service.
- `-http-address=0.0.0.0:3000`: Listen in all the interfaces.

**NOTE**: If the identity provider is deployed with a self-signed certificate (which may be the case for Keycloak or Dex) you will need to disable the TLS and cookie verification. For doing so you can add the flags `-ssl-insecure-skip-verify` and `--cookie-secure=false` to the deployment above. You can find more options for `oauth2-proxy` [here](https://pusher.github.io/oauth2_proxy/docs/configuration).

#### Exposing the proxy

Once the proxy is in place and it's able to connect to the IdP we will need to expose it to access it as the main endpoint for Kubeapps (instead of the `kubeapps` service). We can do that with an Ingress object. Note that for doing so an [Ingress Controller](https://kubernetes.io/docs/concepts/services-networking/ingress/#ingress-controllers) is needed. There are also other methods to expose the `kubeapps-auth-proxy` service, for example using `LoadBalancer` as type in a cloud environment. In case an Ingress is used, remember to modify the host `kubeapps.local` for the value that you want to use as a hostname for Kubeapps:

```
kubectl create -n $KUBEAPPS_NAMESPACE -f - -o yaml << EOF
apiVersion: extensions/v1beta1
kind: Ingress
metadata:
  annotations:
    nginx.ingress.kubernetes.io/connection-proxy-header: keep-alive
    nginx.ingress.kubernetes.io/proxy-read-timeout: "600"
  name: kubeapps
spec:
  rules:
  - host: kubeapps.local
    http:
      paths:
      - backend:
          serviceName: kubeapps-auth-proxy
          servicePort: 3000
        path: /
EOF
```
