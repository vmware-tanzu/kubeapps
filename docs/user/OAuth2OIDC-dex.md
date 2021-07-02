# Dex

For Dex, you can find the parameters that you need to set in the configuration (a ConfigMap if Dex is deployed within the cluster) that the server reads the configuration from. Note that since Dex is only a connector you need to configure it with some third-party credentials that may be a client-id and client-secret as well. Don't confuse those credentials with the ones of the application that you can find under the `staticClients` section.

- Client-ID: Static client ID.
- Client-secret: Static client secret.
- OIDC Issuer URL: Dex URL (i.e. https://dex.example.com:32000).
