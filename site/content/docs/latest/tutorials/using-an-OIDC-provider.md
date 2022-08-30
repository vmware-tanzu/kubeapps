# Using an OAuth2/OIDC provider in Kubeapps

## Table of Contents

1. [Introduction](#introduction)
2. [Pre-requisites](#pre-requisites)
3. [Identity providers (IdP)](#identity-providers-idp)
4. [OIDC provider configuration](#oidc-provider-configuration)
   1. [OIDC parameters](#configure-parameters)
   2. [Redirection](#configure-redirection)
5. [Troubleshooting](#troubleshoothing)

---

## Introduction

_OpenID Connect (OIDC)_ is a simple identity layer on top of the OAuth 2.0 protocol that allows clients to verify the identity of a user based on the authentication performed by an authorization server, as well as to obtain basic profile information about the user.

In Kubernetes, one of the authentication strategies for incoming requests to the API server is using [OpenID Connect tokens](https://kubernetes.io/docs/reference/access-authn-authz/authentication/#openid-connect-tokens). The cluster can be configured to trust an external OIDC provider so that authenticated requests can be matched with defined RBAC. Additionally, some managed Kubernetes environments enable authenticating via plain OAuth2.

However, for Kubernetes to be able to use OIDC-based authentication it is required to [enable certain flags in the API server](https://kubernetes.io/docs/reference/access-authn-authz/authentication/#configuring-the-api-server), which is not always possible on certain environments (such as managed Kubernetes distributions by a cloud provider). To work around this limitation, we recommend having a look at the [Pinniped project](https://pinniped.dev/).

This tutorial describes how to use an existing OAuth2 provider, including OIDC, to authenticate users within Kubeapps.

## Pre-requisites

- A Kubernetes cluster that is properly configured to use an OIDC Identity Provider (IdP) to handle the authentication to your cluster.

  - Read [more information about the Kubernetes API server's configuration options for OIDC](https://kubernetes.io/docs/reference/access-authn-authz/authentication/#openid-connect-tokens). This allows the Kubernetes API server itself to trust tokens from the identity provider. Some hosted Kubernetes services are already configured to accept `access_token`s from their identity provider as bearer tokens (see GKE below).

  - Alternatively, if you do not have access to configure your cluster's API server, you can [install and configure Pinniped in your cluster to trust your identity provider and configure Kubeapps to proxy requests via Pinniped](../howto/OIDC/using-an-OIDC-provider-with-pinniped.md).

## Identity Providers (IdP)

Any OIDC-compliant Identity Provider (IdP) that can be used in a Kubernetes cluster. The steps of this guide have been validated using the following providers:

- [VMware Cloud Services](https://console.cloud.vmware.com): VMware Cloud Services as an OIDC provider.
- [Azure Active Directory](https://docs.microsoft.com/en-us/azure/active-directory/fundamentals/active-directory-whatis): Identity Provider that can be used for AKS.
- [Google OpenID Connect](https://developers.google.com/identity/protocols/OpenIDConnect): OAuth 2.0 for Google accounts.
- [Dex](https://github.com/dexidp/dex): Open Source OIDC and OAuth 2.0 Provider with Pluggable Connectors.
- [Keycloak](https://www.keycloak.org/): Open Source Identity and Access Management.

## OIDC provider configuration

Kubeapps uses [OAuth2 Proxy](https://github.com/oauth2-proxy/oauth2-proxy) to handle the OIDC authentication flow (exchange the credentials, retrieve the token, redirect back to Kubeapps, etc.)

### Configure parameters

The minimum set of parameters to use an Identity Provider for Kubeapps are the following:

- `Client ID`: Client ID of the IdP.
- `Client Secret`: (If configured) Secret used to validate the Client ID.
- `Provider name` (which can be `oidc`, in which case the OIDC Issuer URL is also required).
- `Cookie secret`: a 16, 24 or 32 byte base64 encoded seed string used to encrypt sensitive data (eg. `echo "not-good-secret" | base64`). [More information on the OAuth2 Proxy documentation](https://oauth2-proxy.github.io/oauth2-proxy/docs/configuration/overview/#generating-a-cookie-secret).

**Note**: More parameters may be necessary depending on the Identity Provider and the aimed configuration.

The following sections dive deep into each validated identity provider and describe how to locate the required configuration parameters:

- [VMware Cloud Services](../howto/OIDC/OAuth2OIDC-VMware-cloud-services.md)
- [Azure Active Directory](../howto/OIDC/OAuth2OIDC-azure-active-directory.md)
- [Google OpenID Connect](../howto/OIDC/OAuth2OIDC-google-openid-connect.md)
- [Dex](../howto/OIDC/OAuth2OIDC-dex.md)
- [Keycloak](../howto/OIDC/OAuth2OIDC-keycloak.md)

### Configure redirection

When configuring the identity provider, you will need to ensure that the redirect URL for your Kubeapps installation is configured, which is your Kubeapps URL with the absolute path `/oauth2/callback`. For example, if you are deploying Kubeapps with TLS on the domain `my-kubeapps.example.com`, then the redirect URL will be `https://my-kubeapps.example.com/oauth2/callback`.

## Deploying an auth proxy to access Kubeapps

The main difference in the authentication is that instead of accessing the Kubeapps service, you will be accessing an oauth2 proxy service that is in charge of authenticating users with the identity provider and injecting the required credentials in the requests to Kubeapps.

> Read the [OAuth2 Proxy Auth configuration page](https://oauth2-proxy.github.io/oauth2-proxy/docs/configuration/overview) for detailed information.

There are a number of available solutions for this use-case, like [keycloak-gatekeeper](https://github.com/keycloak/keycloak-gatekeeper) or [oauth2_proxy](https://github.com/oauth2-proxy/oauth2-proxy).


The next sections explain how you can deploy this proxy either using the Kubeapps chart or manually:

- [Using Kubeapps chart](../howto/OIDC/OAuth2OIDC-oauth2-proxy.md#using-the-chart)
- [Manual deployment](../howto/OIDC/OAuth2OIDC-oauth2-proxy.md#manual-deployment)

Once the proxy is accessible, you will be redirected to the identity provider to authenticate. After successfully authenticating, you will be redirected to Kubeapps and be authenticated with your user's OIDC token.

## Troubleshooting

If you find after configuring your OIDC/OAuth2 setup following the above instructions, that although you can successfully authenticate with your provider you are nonetheless unable to login to Kubeapps but instead see a 403 or 401 request in the browser's debugger, then you will need to investigate _why_ the Kubernetes cluster is not accepting your credential.

Visit the [debugging auth failures when using OIDC](../howto/OIDC/OAuth2OIDC-debugging.md) page for more information.
