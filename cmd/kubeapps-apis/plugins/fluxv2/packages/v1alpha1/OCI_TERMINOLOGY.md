# OCI TERMINOLOGY

Note the differences between various technologies:

## Flux

Given an OCI HelmRepository CRD with URL like `"oci://ghcr.io/stefanprodan/charts"` or `"oci://ghcr.io/stefanprodan/charts/"` (URL ending with a slash), then:
- `oci://ghcr.io/stefanprodan/charts` is the *OCI chart repository* URL with the following components:
  - `oci://` - URL scheme, indicating this is an *OCI chart repository*, as opposed to an *HTTP chart repository*
  - `ghcr.io` - OCI registry host
  - `/stefanprodan/charts` - registry path
- That OCI registry may contain multiple helm chart repositories, such as `"podinfo"` and `"nginx"`. The associated OCI references would be: 
  - `oci://ghcr.io/stefanprodan/charts/podinfo`
  - `oci://ghcr.io/stefanprodan/charts/nginx`
- Each of the repositories may only contain a single chart, whose name matches that of the repository basename. For example, repository with the basename (the last segment of the URL path) `"podinfo"` may only contain a single chart also called `"podinfo"`. Also see helm section below.
- Each of the charts may have multiple versions a.k.a. tags, e.g. "`6.1.5"`, `"6.1.4"`, etc.

References:
  - https://fluxcd.io/docs/components/source/helmrepositories/
  - https://github.com/fluxcd/flux2/tree/main/rfcs/0002-helm-oci
  - https://github.com/fluxcd/source-controller/blob/main/controllers/helmrepository_controller_oci.go

## ORAS v2 go libraries
Given a remote OCI registry, such as `ghcr.io`, will list all repository names hosted in the format `"{REGISTRY_PATH}/{NAME}"`. Unlike the flux section, REGISTRY_PATH does not begin with a slash. For example, assuming the remote registry with the URL `"oci://ghcr.io/stefanprodan/charts"` contains 2 repositories, `"podinfo"` and `"podinfo-2"`, then the following list is returned from ORAS `Registry.Repositories()` API:
  1. `"stefanprodan/charts/podinfo"`
  2. `"stefanprodan/charts/podinfo-2"`

References: 
  - https://github.com/distribution/distribution/blob/main/docs/spec/api.md#listing-repositories
  - https://oras.land/ 
  - https://github.com/oras-project/oras-go/blob/14422086e41897a44cb706726e687d39dc728805/registry/remote/url.go#L43

## Helm

One can login to or logout from a registry host, such as
```
$ helm registry login ghcr.io -u $GITHUB_USER -p $GITHUB_TOKEN
```
and 
```
$ helm registry logout ghcr.io
Logout succeeded
```
Here `ghcr.io` is the registry host

From [helm docs](https://helm.sh/docs/topics/registries/):
```
When using helm push to upload a chart to an OCI registry, the reference must be prefixed with oci:// and must not contain the basename or tag

The registry reference basename is inferred from the chart's name, and the tag is inferred from the chart's semantic version

Certain registries require the repository and/or namespace (if specified) to be created beforehand
```
From [helm HIPS spec](https://github.com/helm/community/blob/main/hips/hip-0006.md#4-chart-names--oci-reference-basenames):
```
To keep things simple, the basename (the last segment of the URL path) on a registry reference should be equivalent to the chart name.

For example, given a chart with the name pepper and the version 1.2.3, users may run a command such as the following:

$ helm push pepper-1.2.3.tgz oci://r.myreg.io/mycharts

which would result in the following reference:

oci://r.myreg.io/mycharts/pepper:1.2.3

By placing such restrictions on registry URLs Helm users are less likely to do "strange things" with charts in registries
```

In this case:
    - a single repository named `"mycharts/pepper"` will be created if one does not exist
    - the repository contains a chart named `"pepper"`
    - the chart `"pepper"` has a version `"1.2.3"`   

You can use the command ```helm show all``` to see (some) information about the `"pepper"` chart:
```
$ helm show all oci://r.myreg.io/mycharts/pepper | head -9 
apiVersion: v1
appVersion: 1.2.3
description: ...
home: ...
kubeVersion: '>=1.19.0-0'
maintainers:
- email: stefanprodan@users.noreply.github.com
  name: stefanprodan
name: pepper
...
```

References:
  - https://helm.sh/blog/storing-charts-in-oci/
  - https://helm.sh/docs/topics/registries/
  - https://github.com/helm/community/blob/main/hips/hip-0006.md#specification

## GitHub Container Registry `ghcr.io`
Take an OCI registry URL like `"oci://ghcr.io/gfichtenholt/charts/podinfo:6.1.5"`
GitHub Container Registry WebPortal, CLI and API do not use the term *"OCI registry"* and *"OCI repository"*. Instead, the following terms are used.
  - **Host** - always `ghcr.io`
  - **Owner** - may be an organization or a indiviual account,e.g. `stefanprodan`
  - **Package** - with package type `container`, e.g. `charts/podinfo`
  - **Package Version** - package version a.k.a. tag, e.g. `"6.1.5"`

The term `package` seems to correspond to be a concatenation of an last segment of an OCI repository path and chart name, `charts/podinfo` in the example above. The following list shows a URL used with `helm push` command and the resulting package name on `ghcr.io`:
  - oci://ghcr.io/gfichtenholt - podinfo
  - oci://ghcr.io/gfichtenholt/charts - charts/podinfo
  - oci://ghcr.io/gfichtenholt/charts/podinfo - charts/podinfo/podinfo

A given owner may have mutiple packages, e.g. `"nginx/nginx"`, `"charts/podinfo"`, etc. A given package may have multiple versions. The use case with multiple charts in the same repository doesn't really apply (TODO double check if there is a workaround)

References:
  - https://docs.github.com/en/rest/packages
  - https://docs.github.com/en/packages/working-with-a-github-packages-registry/working-with-the-container-registry
  - https://github.com/gfichtenholt?tab=packages


## Harbor Container Registry `demo.goharbor.io` v2.5.0
Like ghcr.io, appears to support 
[Docker Registry HTTP API V2](https://github.com/distribution/distribution/blob/main/docs/spec/api.md#listing-repositories) for listing of repostories with one caveat:
  - suppose I am looking for list of repository names that have prefix `'stefanprodan-podinfo-clone'`.
    To do this, I pass in a parameter startAt `'stefanprodan-podinfo-clone'` to ORAS v2 libraries (see `docker_reg_v2_repo_lister.go`) and I get back this list:
    ```
    [1uoih/ap-alertmanager 1uoih/ap-astro-ui 1uoih/ap-base 1uoih/ap-cli-install 1uoih/ap-commander 1uoih/ap-configmap-reloader 1uoih/ap-curator 1uoih/ap-db-bootstrapper 1uoih/ap-default-backend 1uoih/ap-elasticsearch 1uoih/ap-elasticsearch-exporter 1uoih/ap-grafana 1uoih/ap-houston-api 1uoih/ap-kibana 1uoih/ap-kube-state 1uoih/ap-nats-exporter 1uoih/ap-nats-server 1uoih/ap-nats-streaming 1uoih/ap-nginx 1uoih/ap-nginx-es 1uoih/ap-node-exporter 1uoih/ap-postgresql 1uoih/ap-prometheus 1uoih/ap-registry aaa/bb/ccv.a-prod al/aaa al/aalllaa al/ahaha al/alllaa al/blabla al/ddd al/dsdsd al/fff al/hehe al/mammamia al/pupeczka al/siup fs8v0/ap-alertmanager fs8v0/ap-astro-ui fs8v0/ap-base fs8v0/ap-cli-install fs8v0/ap-commander fs8v0/ap-configmap-reloader fs8v0/ap-curator fs8v0/ap-db-bootstrapper fs8v0/ap-default-backend fs8v0/ap-elasticsearch fs8v0/ap-elasticsearch-exporter fs8v0/ap-grafana fs8v0/ap-houston-api fs8v0/ap-kibana fs8v0/ap-kube-state fs8v0/ap-nats-exporter fs8v0/ap-nats-server fs8v0/ap-nats-streaming fs8v0/ap-nginx fs8v0/ap-nginx-es fs8v0/ap-node-exporter fs8v0/ap-postgresql fs8v0/ap-prometheus fs8v0/ap-registry hello/awesome-redis library/harbor-portal ocis/ocis oioi/test-image oioi/test-image2 privatetest/linuxserver/emby secure_project/nginx sriniharbor/ap-airflow-dev stefanprodan-podinfo-clone/podinfo tce/harbor tce/main techedu/alpine test/test test/test1 test/test2 wrj/redis-image wrj/ubuntu-image]
    ```
    Notice that list does not really start with `'stefanprodan-podinfo-clone'`. The response does contain `stefanprodan-podinfo-clone/podinfo` but I am having some concerns why it works the way it does. For comparison, same scenario w.r.t. ghcr.io gives the following response: 
    ```
    [stefanprodan-podinfo-clone/podinfo gfoidl/datacompression/dotnet-gnuplot5 gfoidl-Tests/Vue_Server_Test/web-app gforestier2000/simple-mariadb-ci-cd gforien/dockeragent gforien/reas gforien/webapp-nodejs gfw404/scripts/jd_scripts gg-martins091/forgottenserver/forgottenserver 
    ... 
    ghcr-library/golang]

## Google Cloud Platform Artifact Repository 
Like ghcr.io, appears to support 
[Docker Registry HTTP API V2](https://github.com/distribution/distribution/blob/main/docs/spec/api.md#listing-repositories) for listing of repostories:
  - suppose I am looking for list of repository names that have prefix `'stefanprodan-podinfo-clone'`.
    To do this, I pass in a parameter startAt `'stefanprodan-podinfo-clone'` to ORAS v2 libraries (see `docker_reg_v2_repo_lister.go`) and I get back this list:
    ```
    $ curl -iL https://us-west1-docker.pkg.dev/v2/_catalog?last=stefanprodan-podinfo-clone -H "Authorization: Bearer $GCP_TOKEN"
    HTTP/2 200
    content-type: application/json; charset=utf-8
    docker-distribution-api-version: registry/2.0
    
    {"repositories":["vmware-kubeapps-ci/stefanprodan-podinfo-clone/podinfo"]}
    ```
    
---
Here is probably the most confusing part of the whole document:
  1. Assume we have a Flux OCI HelmRepository CRD with URL `"oci://ghcr.io/gfichtenholt/helm-charts"` 
  2. Assume the remote OCI registry contains a single chart `"podinfo"` with version `"6.1.5"`
  3. ORAS go library will return repository list `["gfichtenholt/helm-charts/podinfo"]`
  4. kubeapps flux plugin will call `RegistryClient.Tags()` with respect to OCI reference `"ghcr.io/gfichtenholt/helm-charts/podinfo"` which will return `["6.1.5"]`
  5. kubeapps flux plugin will call `RegistryClient.DownloadChart()` with respect to a chart with version `"6.1.5"` a URL `"ghcr.io/gfichtenholt/helm-charts/podinfo:6.1.5"`. Here, the identifier `"podinfo"` refers **BOTH to repository basename AND the chart name!**
---
