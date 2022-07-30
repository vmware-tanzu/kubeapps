# OCI TERMINOLOGY

Note the differences between various technologies:

## Flux

Given an OCI HelmRepository CRD with URL like `"oci://ghcr.io/stefanprodan/charts"` or `"oci://ghcr.io/stefanprodan/charts/"` (URL ending with a slash), then:
- `oci://ghcr.io/stefanprodan/charts` is the *OCI chart repository* URL with the following components:
  - `oci://` - URL scheme, indicating this is an *OCI chart repository*, as opposed to an *HTTP chart repository*
  - `ghcr.io` - OCI registry host
  - `/stefanprodan/charts` - repository path
- That repository may contain multiple charts, e.g. `"podinfo"` and `"nginx"`. Chart names are not part of the URL and must be specified by user when creating HelmChart or HelmRelease CRD
- Each of the charts may have multiple versions a.k.a. tags, e.g. "`6.1.5"`, `"6.1.4"`, etc.

References:
  - https://fluxcd.io/docs/components/source/helmrepositories/
  - https://github.com/fluxcd/flux2/tree/main/rfcs/0002-helm-oci
  - https://github.com/fluxcd/source-controller/blob/main/controllers/helmrepository_controller_oci.go

## ORAS v2 go libraries
Given a remote OCI registry, such as `ghcr.io`, will list all repository names hosted. The names are listed in the format
`"stefanprodan/charts/podinfo"`, `"stefanprodan/charts/podinfo-2"`, etc.

References: 
  - https://github.com/distribution/distribution/blob/main/docs/spec/api.md#listing-repositories

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

For example, once logged in, you may use the command ```"helm push"```, with repository URL like this:
```
$ helm push podinfo-6.1.5.tgz oci://ghcr.io/gfichtenholt
Pushed: ghcr.io/gfichtenholt/podinfo:6.1.5
Digest: sha256:80e6d2e7f6d21800621530fc4c5b70d0e458f11b2c05386ea5d058c4e86d6e93
```
In this case:
    - a single repository named `"gfichtenholt/podinfo"` will be created if one does not exist
    - the repository contains a chart named `"podinfo"`
    - the chart `"podinfo"` has a version `"6.1.5"`   

Or you may use ```"helm push"```, with repository URL like this:
```
$ helm push podinfo-6.1.5.tgz oci://ghcr.io/gfichtenholt/charts
Pushed: ghcr.io/gfichtenholt/charts/podinfo:6.1.5
Digest: sha256:80e6d2e7f6d21800621530fc4c5b70d0e458f11b2c05386ea5d058c4e86d6e93
```
In this case the repository name is `"gfichtenholt/charts/podinfo"` while the rest is as above.  You can use the command ```helm show all``` to see (some) information about the `"podinfo"` chart:
```
$ helm show all oci://ghcr.io/gfichtenholt/charts/podinfo | head -9 
apiVersion: v1
appVersion: 6.1.5
description: Podinfo Helm chart for Kubernetes
home: https://github.com/stefanprodan/podinfo
kubeVersion: '>=1.19.0-0'
maintainers:
- email: stefanprodan@users.noreply.github.com
  name: stefanprodan
name: podinfo
...
```
Or you may use 
```
helm push podinfo-6.1.5.tgz oci://ghcr.io/gfichtenholt/charts/podinfo
Pushed: ghcr.io/gfichtenholt/charts/podinfo/podinfo:6.1.5
Digest: sha256:80e6d2e7f6d21800621530fc4c5b70d0e458f11b2c05386ea5d058c4e86d6e93
```
In this case, the repository name is `"gfichtenholt/charts/podinfo/podinfo"`, while the rest is the same. And so on and so forth

References:
  - https://helm.sh/blog/storing-charts-in-oci/
  - https://helm.sh/docs/topics/registries/

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
