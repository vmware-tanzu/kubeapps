# Step 3: Preparing Kubeapps deployment

Before Kubeapps is deployed to the TCE cluster, there are some decisions to take in order to shape the installation structure and functioning of the application.

Some relevant topics like routing traffic into Kubeapps, TLS, or which plugins need to be enabled, are set up in a _configuration values file_.
A configuration values file is a Yaml file that allows you to customize the deployment of Kubeapps. TCE makes use of [Carvel](https://carvel.dev/) for installing applications, and in the case of the Kubeapps package, the configuration file uses exactly the same parameters specified in the [Bitnami Kubeapps Helm chart](https://github.com/bitnami/charts/tree/master/bitnami/kubeapps#parameters). It is highly recommended that you take a look at the possible parameters and get familiar with them.

The outcome of this step is:

- A configuration values file that matches your desired setup for Kubeapps.
- Required packages ready to make the configuration work (for example, installing the actual Ingress provider).

## Option A: Getting traffic into Kubeapps using a LoadBalancer

The simplest way to expose the Kubeapps Dashboard is to assign a [_LoadBalancer_ service](https://kubernetes.io/docs/concepts/services-networking/service/#loadbalancer) type to the Kubeapps frontend Service. For example, you can use the following configuration value:

```yaml
frontend:
  service:
    type: LoadBalancer
```

## Option B: Getting traffic into Kubeapps using an ingress

Using an [ingress](https://kubernetes.io/docs/concepts/services-networking/ingress/) is one of the most common ways for getting access to Kubeapps.

In order to do so, you need to define a _fully qualified domain name_ (FQDN), and preferably a TLS certificate available so that clients, like browsers, can safely navigate the UI.

This tutorial uses the FQDN `kubeapps.foo.com` to access Kubeapps as an example.

Add a TLS certificate with the following command:

```bash
kubectl -n kubeapps create secret tls kubeapps-host-tls \
    --key <YOUR_KEY>.pem \
    --cert <YOUR_CERTIFICATE>.pem
```

The public/private key pair must exist before hand. For more information, please check the [Kubernetes documentation for TLS secrets](https://kubernetes.io/docs/concepts/configuration/secret/#tls-secrets).

This TLS certificate can be used by any type of ingress.

As an alternative, you can have certificates automatically managed using [Cert-manager](https://cert-manager.io).

Please refer to [the Kubeapps documentation covering external access with Ingress](https://github.com/vmware-tanzu/kubeapps/blob/main/chart/kubeapps/README.md#ingress) for additional information.

### Option B1: Using Contour ingress

[Contour](https://projectcontour.io/) is an open source Kubernetes Ingress controller that acts as a control plane for the Envoy edge and service proxy.

> Currently it is not possible to use Contour together with OIDC authentication in Kubeapps [due to this limitation](https://github.com/projectcontour/contour/issues/4290). It is possible, though, when using the demo-only, insecure, token authentication.

In order to use Contour with the token authentication, for example with a TCE unmanaged cluster, you can make use of an `HTTPProxy` to route the traffic to the Kubeapps [_frontend reverse proxy_](https://github.com/vmware-tanzu/kubeapps/blob/main/chart/kubeapps/values.yaml#L194).

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

   Using the external address create a CNAME record (for example `kubeapps.foo.com`) in your DNS that maps to the load balancer's address.

### Option B2: Using Nginx ingress

Nginx configuration comes shipped with the Kubeapps package out of the box.
In order to enable it, add the following to the configuration values:

```yaml
ingress:
  enabled: true
  ingressClassName: "nginx"
  hostname: kubeapps.foo.com
  tls: true
  extraTls:
    - hosts:
        - kubeapps.foo.com
      secretName: kubeapps-host-tls
  annotations:
    nginx.ingress.kubernetes.io/proxy-read-timeout: "600"
```

Please note how the configuration above references both the secret holding the TLS certificate and the FQDN.

As mentioned, Kubeapps provides the configuration handling for Nginx. But Nginx needs to be installed in the cluster.
To install it use the official resources like:

```bash
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/main/deploy/static/provider/kind/deploy.yaml
```

> When using OauthProxy for OIDC authentication, there is [an issue with the proxy buffers](https://github.com/vmware-tanzu/kubeapps/pull/1944) that needs some workaround.
>
> You can [use a modified resources file to install Nginx](https://github.com/vmware-tanzu/kubeapps/blob/main/site/content/docs/latest/reference/manifests/ingress-nginx-kind-with-large-proxy-buffers.yaml) that Kubeapps provides. This is limited to the specific version shipped with Kubeapps.
>
> Alternatively, you can make the change manually on top of your Nginx installation running `kubectl -n ingress-nginx edit cm ingress-nginx-controller` and adding the following to the `data:` section of the `ConfigMap`:
>
> ```bash
> proxy-buffer-size: 8k
> proxy-buffers: 4 8k
> ```

## Configuring OIDC

In case you selected OIDC as your authentication method, you need to set some parameters in the configuration values file. This is needed so that the OAuth proxy used in Kubeapps can contact the OIDC provider and exchange the tokens.

Please retrieve the values obtained in the [Setting up Google credentials client](./02-TCE-managed-cluster.md#setting-up-the-google-credentials-client) section and set them in your configuration values:

```yaml
authProxy:
  enabled: true
  provider: oidc
  clientID: <YOUR_CLIENT_ID>
  clientSecret: <YOUR_CLIENT_SECRET>
  ## NOTE: cookieSecret must be a particular number of bytes. It's recommended using the following
  ## script to generate a cookieSecret:
  ##   openssl rand -base64 32 | tr -- '+/' '-_'
  ## ref: https://oauth2-proxy.github.io/oauth2-proxy/docs/configuration/overview#generating-a-cookie-secret
  cookieSecret: <COOKIE_SECRET>
  scope: "openid email groups"
  extraFlags:
    - --oidc-issuer-url=<YOUR_OIDC_ISSUER_URL>
```

## Configuring selected plugins

Kubeapps offers three plugins for managing packages and repositories: [Helm](https://helm.sh/docs/topics/chart_repository/), [Carvel](https://carvel.dev/kapp-controller/docs/develop/packaging/#package-repository) and [Helm via Flux](https://fluxcd.io/docs/components/source/helmrepositories/).
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

At this point, you should have a proper Yaml file with configuration values.

> Continue the tutorial by [deploying Kubeapps](./04-deploying-kubeapps.md).

## Tutorial index

1. [TCE cluster deployment preparation](./01-TCE-cluster-preparation.md)
2. [Deploying a managed cluster](./02-TCE-managed-cluster.md) or [Deploy an unmanaged cluster](./02-TCE-unmanaged-cluster.md)
3. [Preparing the Kubeapps deployment](./03-preparing-kubeapps-deployment.md)
4. [Deploying Kubeapps](./04-deploying-kubeapps.md)
5. [Further documentation for managing applications in Kubeapps](./05-managing-applications.md)