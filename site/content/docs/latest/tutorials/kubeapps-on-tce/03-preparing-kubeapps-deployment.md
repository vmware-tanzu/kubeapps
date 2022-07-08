# Step 3: Preparing Kubeapps deployment

Before Kubeapps is deployed to the TCE cluster, there are some decisions that need to be taken. This will shape the installation structure and functioning of the application.
There are topics like routing traffic into Kubeapps, TLS, or which plugins need to be enabled, that will be set up in a _configuration values file_.

A configuration values file is a Yaml file that allows you to customize the deployment of Kubeapps. TCE makes use of Carvel for installing application, and in the case of Kubeapps, the configuration file uses exactly the same parameters specified in the [Bitnami Kubeapps Helm chart.](https://github.com/bitnami/charts/tree/master/bitnami/kubeapps#parameters) It is highly recommended that you take a look at the possible parameters and adjust the values file accordingly.

In this step of the tutorial we will go over the main parameters that are required to get Kubeapps quickly up and running.

## Getting traffic using an ingress

Using an ingress is one of the most common ways for getting access to Kubeapps.
In order to do so, there needs to be a _fully qualified domain name_ assigned, and preferably a TLS certificate available so that clients, like browsers, can safely navigate the UI.

In this tutorial we will use the FQDN `kubeapps.foo.com` to access Kubeapps as an example.

And we will add a TLS certificate with the following command:

    ```bash
    kubectl -n kubeapps create secret tls kubeapps-host-tls \
        --key <YOUR_KEY>.pem \
        --cert <YOUR_CERTIFICATE>.pem
    ```

This TLS certificate can be used by any type of ingress.

### Using Contour ingress

[Contour](https://projectcontour.io/) is an open source Kubernetes Ingress controller that acts as a control plane for the Envoy edge and service proxy.

Currently it is not possible to use Contour together with OIDC authentication in Kubeapps [due to this limitation](https://github.com/projectcontour/contour/issues/4290). It is possible, though, when using the demo-only, insecure, token authentication.

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

### Nginx

- If using OauthProxy (usually with OIDC), add patch to increase proxy buffers:
<https://github.com/vmware-tanzu/kubeapps/pull/1944/files>

- Otherwise: <https://raw.githubusercontent.com/kubernetes/ingress-nginx/main/deploy/static/provider/kind/deploy.yaml>

## Using LoadBalancer

@TODO

```bash
frontend:
  service:
    type: LoadBalancer
```
