## Misc
The most tricky and fragile part by far was to set up authentication between flux and GCP and flux plugin and GCP. You need to make sure that both:
- helm API works without having to invoke any Google CLI code in flux source-controller pod
- ORAS API works to list the repositories. You can use curl to similate ORAS requests (see below santify checks) 

## Service Account
The following service account was set up for use by the integration tests using
Google Cloud Console
- GCP project: `vmware-kubeapps-ci`
- GCP service account:
    - Name and ID: `flux-plugin-test-sa-3`
    - Description: `Service Account for integration testing of kubeapps flux plugin`

  Roles granted to this service account:
    - Editor. 
      Verified 8/24/22 11:39pm this works. Same code failed on the previous run with
```
I0825 06:01:17.145356       1 docker_reg_v2_repo_lister.go:50] ORAS v2 Registry [oci://us-west1-docker.pkg.dev/vmware-kubeapps-ci/stefanprodan-podinfo-clone PlainHTTP=false] PING: GET "https://us-west1-docker.pkg.dev/v2/": unexpected status code 401: unauthorized: No valid credential was supplied.
```
  An intermittent issue?

  **TODO** need to reduce permissions to smallest workable set. 
  Probably some combination of these and others:
    - Artifact Registry Administrator
    - Artifact Registry Repository Administrator
    - Artifact Registry Service Agent

  Make sure you see a message "Policy Updated" at the bottom of the screen when you grant these roles.
  If you see "Failed to add project roles" or some other error message,
  create the service account with a different Name/ID
- The service account key file can be downloaded with Google Cloud Console
     Under IAM & Admin -> Service Accounts

## Sanity checks 
Here are (some of) the sanity checks I use to make sure everything is working that don't involve kubeapps:
```
$ export GOOGLE_APPLICATION_CREDENTIALS=/Users/gfichtenholt/gitlocal/kubeapps-gfichtenholt/cmd/kubeapps-apis/plugins/fluxv2/packages/v1alpha1/testdata/gcloud-kubeapps-flux-test-sa-key-file.json

$ export GCP_TOKEN=$(gcloud auth application-default print-access-token)
```

1. check PING is working 
```
$ curl -iL https://us-west1-docker.pkg.dev/v2/ -H "Authorization: Bearer $GCP_TOKEN"
HTTP/2 200 
docker-distribution-api-version: registry/2.0
date: Thu, 25 Aug 2022 06:29:59 GMT
content-length: 0
content-type: text/html; charset=UTF-8
```

2. check _catalog API is working 
```
$ curl -iL https://us-west1-docker.pkg.dev/v2/_catalog -H "Authorization: Bearer $GCP_TOKEN"
  HTTP/2 200
  content-type: application/json; charset=utf-8
  docker-distribution-api-version: registry/2.0
  date: Wed, 24 Aug 2022 16:10:18 GMT
  content-length: 75

  {"repositories":["vmware-kubeapps-ci/stefanprodan-podinfo-clone/podinfo"]}
```

TODO add example from info.txt about kubectl create helmrepository/helmrelease
