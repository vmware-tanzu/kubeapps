# Configuring Keycloak as an OIDC provider

This document explains how to configure Keycloak as an OIDC provider (check general information and pre-requisites for [using an OAuth2/OIDC Provider with Kubeapps](../using-an-OIDC-provider.md)).

In the case of Keycloak, you can find the parameters in the Keycloak admin console:

- **Client-ID**: Keycloak client ID.
- **Client-secret**: Secret associated to the client above.
- **OIDC Issuer URL**: `https://<keycloak.domain>/auth/realms/<realm>`.
