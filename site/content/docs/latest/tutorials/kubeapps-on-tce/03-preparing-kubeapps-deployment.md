# Step 3: Preparing Kubeapps deployment

Before Kubeapps is deployed to the TCE cluster, there are some decisions that need to be taken. This will shape the installation structure and functioning of the application.
There are topics like routing traffic into Kubeapps, TLS, or which plugins need to be enabled, that will be set up in a _configuration values file_.

A configuration values file is a Yaml file that allows you to customize the deployment of Kubeapps. TCE makes use of Carvel for installing applications, and in the case of the Kubeapps package, the configuration file uses exactly the same parameters specified in the [Bitnami Kubeapps Helm chart.](https://github.com/bitnami/charts/tree/master/bitnami/kubeapps#parameters) It is highly recommended that you take a look at the possible parameters and get familiar with them.

The outcome of this tutorial step will be:

- A configuration values file that matches your desired setup for Kubeapps
- Required packages ready to make the configuration work (for example the actual Ingress provider)

## Option A: Getting traffic using a LoadBalancer

The simplest way to expose the Kubeapps Dashboard is to assign a LoadBalancer type to the Kubeapps frontend Service. For example, you can use the following configuration value:

```yaml
frontend:
  service:
    type: LoadBalancer
```

## Option B: Getting traffic using an ingress

Using an ingress is one of the most common ways for getting access to Kubeapps.

In order to do so, you need to define a _fully qualified domain name (FQDN)_, and preferably a TLS certificate available so that clients, like browsers, can safely navigate the UI.

In this tutorial we will use the FQDN `kubeapps.foo.com` to access Kubeapps as an example.

And we will add a TLS certificate with the following command:

  ```bash
  kubectl -n kubeapps create secret tls kubeapps-host-tls \
      --key <YOUR_KEY>.pem \
      --cert <YOUR_CERTIFICATE>.pem
  ```

This TLS certificate can be used by any type of ingress.

As an alternative, you can have certificates automatically managed using [Cert-manager.](https://cert-manager.io)

Please refer to [the Kubeapps documentation covering external access with Ingress](https://github.com/vmware-tanzu/kubeapps/blob/main/chart/kubeapps/README.md#ingress) for additional information.

### Option B1: Using Contour ingress

[Contour](https://projectcontour.io/) is an open source Kubernetes Ingress controller that acts as a control plane for the Envoy edge and service proxy.

> Currently it is not possible to use Contour together with OIDC authentication in Kubeapps [due to this limitation](https://github.com/projectcontour/contour/issues/4290). It is possible, though, when using the demo-only, insecure, token authentication.

If that is your case, for example with a TCE unmanaged cluster, you can simply make use of an `HTTPProxy` to route the traffic to the Kubeapps [_frontend reverse proxy_](https://github.com/vmware-tanzu/kubeapps/blob/main/chart/kubeapps/values.yaml#L194).

1. Install Contour

    ```bash
      tanzu package install contour \
        --package-name contour.community.tanzu.vmware.com \
        --version 1.20.1
    ```

2. Create an `HTTPProxy`. Please notice how it references both the secret holding the TLS certificate and the FQDN.

    ```bash
    cat <<EOF | kubectl apply -f - 
    apiVersion: projectcontour.io/v1
    kind: HTTPProxy
    metadata:
      name: kubeapps-grpc
      namespace: kubeapps
    spec:
      virtualhost:
        fqdn: kubeapps.foo.com
        tls:
          secretName: kubeapps-host-tls
      routes:
        - conditions:
          - prefix: /apis/
          pathRewritePolicy:
            replacePrefix:
            - replacement: /
          services:
            - name: kubeapps-internal-kubeappsapis
              port: 8080
              protocol: h2c
        - services:
          - name: kubeapps
            port: 80
    EOF
    ```

3. Retrieve the external address of Contour’s Envoy load balancer

    ```bash
    kubectl get -n projectcontour service envoy -o wide
    ```

    Using the external address create a CNAME record (for example `kubeapps.foo.com`) in your DNS that maps to the LB address.

### Option B2: Using Nginx ingress

Nginx configuration comes shipped with the Kubeapps package out of the box.
In order to enable it, add the following to the configuration values:

```yaml
ingress:
  enabled: true
  hostname: kubeapps.foo.com
  tls: true
  extraTls:
    - hosts:
        - kubeapps.foo.com
      secretName: kubeapps-host-tls
  annotations:
    nginx.ingress.kubernetes.io/proxy-read-timeout: "600"
    # Required for ingress-nginx 1.0.0 for a valid ingress.
    kubernetes.io/ingress.class: nginx
```

Please notice how the configuration above references both the secret holding the TLS certificate and the FQDN.

As mentioned, Kubeapps provides the configuration handling for Nginx. But Nginx needs to be installed in the cluster.
To install it use the official resources like:

```bash
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/main/deploy/static/provider/kind/deploy.yaml
```

> When using OauthProxy for OIDC authentication, there is [an issue with the proxy buffers](https://github.com/vmware-tanzu/kubeapps/pull/1944) that needs some workaround.
>
> You can [use a modified resources file to install Nginx](/site/content/docs/latest/reference/manifests/ingress-nginx-kind-with-large-proxy-buffers.yaml) that Kubeapps provides. This is limited to the specific version shipped with Kubeapps.
> 
> Alternatively, you can make the change manually on top of your Nginx installation running `kubectl -n ingress-nginx edit cm ingress-nginx-controller` and adding the following to the `data:` section of the `ConfigMap`:
>
>  ```bash
>  proxy-buffer-size: 8k
>  proxy-buffers: 4 8k
>  ```

## Configuring OIDC

In case you selected OIDC as your authentication method, you will need to set some parameters in the configuration values file. This is needed so that the OAuth proxy used in Kubeapps can contact with the OIDC provider and exchange the tokens.

Please retrieve the values obtained in the [Setting up Google credentials client](./02-TCE-managed-cluster.md#setting-up-google-credentials-client) section and set them in your configuration values:

```yaml
authProxy:
  enabled: true
  provider: oidc
  clientID: <YOUR_CLIENT_ID>
  clientSecret: <YOUR_CLIENT_SECRET>
  ## NOTE: cookieSecret must be a particular number of bytes. It's recommended using the following
  ## script to generate a cookieSecret:
  ##   python -c 'import os,base64; print base64.urlsafe_b64encode(os.urandom(16))'
  ## ref: https://oauth2-proxy.github.io/oauth2-proxy/docs/configuration/overview#generating-a-cookie-secret
  cookieSecret: <COOKIE_SECRET>
  scope: "openid email groups"
  extraFlags:
    - --oidc-issuer-url=<YOUR_OIDC_ISSUER_URL>
    - --ssl-insecure-skip-verify=true
```

## Configuring selected plugins

Kubeapps offers three plugins for managing packages and repositories: Helm, Carvel/Kapp and Flux.
You need to define in the configuration values which plugins you want have installed, for example:

```yaml
packaging:
  helm:
    enabled: true
  carvel:
    enabled: true
  flux:
    enabled: false
```

At this point we have a proper Yaml file with configuration values.

> Continue the tutorial by [deploying Kubeapps](./04-deploying-kubeapps.md).
