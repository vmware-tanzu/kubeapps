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
[Docker Registry HTTP API V2](https://github.com/distribution/distribution/blob/main/docs/spec/api.md#listing-repositories) for listing of repostories

---
Here is probably the most confusing part of the whole document:
  1. Assume we have a Flux OCI HelmRepository CRD with URL `"oci://ghcr.io/gfichtenholt/helm-charts"` 
  2. Assume the remote OCI registry contains a single chart `"podinfo"` with version `"6.1.5"`
  3. ORAS go library will return repository list `["gfichtenholt/helm-charts/podinfo"]`
  4. kubeapps flux plugin will call `RegistryClient.Tags()` with respect to OCI reference `"ghcr.io/gfichtenholt/helm-charts/podinfo"` which will return `["6.1.5"]`
  5. kubeapps flux plugin will call `RegistryClient.DownloadChart()` with respect to a chart with version `"6.1.5"` a URL `"ghcr.io/gfichtenholt/helm-charts/podinfo:6.1.5"`. Here, the identifier `"podinfo"` refers **BOTH to repository basename AND the chart name!**
---
