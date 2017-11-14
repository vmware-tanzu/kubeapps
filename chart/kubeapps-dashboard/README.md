# Monocular

[Monocular](https://github.com/kubernetes-helm/monocular) is web-based UI for managing Kubernetes applications packaged as Helm Charts. It allows you to search and discover available charts from multiple repositories, and install them in your cluster with one click.

## TL;DR;

```console
$ helm repo add monocular https://helm.github.io/monocular
$ helm install monocular/monocular
```

## Introduction

This chart bootstraps a [Monocular](https://github.com/kubernetes-helm/monocular) deployment on a [Kubernetes](http://kubernetes.io) cluster using the [Helm](https://helm.sh) package manager.

## Prerequisites

### [Nginx Ingress controller](https://github.com/kubernetes/ingress)

To avoid issues with Cross-Origin Resource Sharing, the Monocular chart sets up an Ingress resource to serve the frontend and the API on the same domain. This is used to route requests made to `<host>:<port>/` to the frontend pods, and `<host>:<port>/api` to the backend pods.

It is possible to run Monocular on separate domains and without the Nginx Ingress controller, see the [configuration](#serve-monocular-frontend-and-api-on-different-domains) section on how to do this.

## Installing the Chart

First, ensure you have added the Monocular chart repository:

```console
$ helm repo add monocular https://kubernetes-helm.github.io/monocular
```

To install the chart with the release name `my-release`:

```console
$ helm install --name my-release monocular
```

The command deploys Monocular on the Kubernetes cluster in the default configuration. The [configuration](#configuration) section lists the parameters that can be configured during installation.

> **Tip**: List all releases using `helm list`

## Uninstalling the Chart

To uninstall/delete the `my-release` deployment:

```console
$ helm delete my-release
```

The command removes all the Kubernetes components associated with the chart and deletes the release.

## Configuration

See the [values](values.yaml) for the full list of configurable values.

### Disabling Helm releases (deployment) management

If you want to run Monocular without giving the option to install and manage charts in your cluster, similar to [KubeApps](https://kubeapps.com) you can configure `api.config.releasesEnabled`:

```console
$ helm install monocular/monocular --set api.config.releasedEnabled=false
```

### Configuring chart repositories

You can configure the chart repositories you want to see in Monocular with the `api.config.repos` value, for example:

```console
$ cat > custom-repos.yaml <<<EOF
api:
  config:
    repos:
      - name: stable
        url: http://storage.googleapis.com/kubernetes-charts
        source: https://github.com/kubernetes/charts/tree/master/stable
      - name: incubator
        url: http://storage.googleapis.com/kubernetes-charts-incubator
        source: https://github.com/kubernetes/charts/tree/master/incubator
      - name: monocular
        url: https://kubernetes-helm.github.io/monocular
        source: https://github.com/kubernetes-helm/monocular/tree/master/charts
EOF

$ helm install monocular/monocular -f custom-repos.yaml
```

### Serve Monocular on a single domain

You can configure the Ingress object with the hostnames you wish to serve Monocular on:

```console
$ cat > custom-domains.yaml <<<EOF
ingress:
  hosts:
  - monocular.local
EOF

$ helm install monocular/monocular -f custom-domains.yaml
```

### Serve Monocular frontend and API on different domains

In order to serve the frontend and the API on different domains, you need to configure the frontend with the API location and configure CORS correctly for the API to accept requests from the frontend.

To do this, you can use the `ui.backendHostname` and `api.config.cors.allowed_origins` values. You should also disable the Ingress resource and manually configure each hostname to point to the pods.

```console
$ cat > separate-domains.yaml <<<EOF
ingress:
  enabled: false
api:
  config:
    cors:
      allowed_origins:
        - http://$FRONTEND_HOSTNAME
ui:
  backendHostname: http://$API_HOSTNAME
EOF

$ helm install monocular/monocular -f separate-domains.yaml
```

Ensure that you replace `$FRONTEND_HOSTNAME` and `$API_HOSTNAME` with the hostnames you want to use.

### Other configuration options

| Value                                   | Description                                                                                 | Default                                                                         |
|-----------------------------------------|---------------------------------------------------------------------------------------------|---------------------------------------------------------------------------------|
| `api.livenessProbe.initialDelaySeconds` | Increase this if the API pods are crashing due to the chart repository sync taking too long | `180`                                                                           |
| `api.config.releasesEnabled`            | Enable installing and managing charts in the cluster                                        | `3600`                                                                          |
| `api.config.cacheRefreshInterval`       | How often to sync with chart repositories                                                   | `3600`                                                                          |
| `ui.googleAnalyticsId`                  | Google Analytics ID                                                                         | `UA-XXXXXX-X` (unset)                                                           |
| `ingress.enabled`                       | If enabled, create an Ingress object                                                        | `true`                                                                          |
| `ingress.annotations`                   | Ingress annotations                                                                         | `{ingress.kubernetes.io/rewrite-target: /, kubernetes.io/ingress.class: nginx}` |
| `ingress.tls`                           | TLS configuration for the Ingress object                                                    | `nil`                                                                           |
