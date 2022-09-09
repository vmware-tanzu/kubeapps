## Misc
The most tricky and fragile part by far was to set up authentication between flux and GCP and flux plugin and GCP. You need to make sure that both:
- helm API works without having to invoke any Google CLI code in flux source-controller pod
- ORAS API works to list the repositories. You can use curl to similate ORAS requests (see below santify checks) 

## Service Account
The following service account was set up for use by the integration tests using
Google Cloud Console
- GCP project: `vmware-kubeapps-ci`
- GCP service account:
    - Name and ID: `flux-plugin-test-sa-4`
    - Description: `Service Account for integration testing of kubeapps flux plugin`

  Roles granted to this service account:
    - Artifact Registry Administrator
    - Artifact Registry Repository Administrator
    - Viewer

  Make sure you see a message "Policy Updated" at the bottom of the screen when you grant these roles. If you see "Failed to add project roles" or some other error message,
  repeat the process w.r.t. a service account with a different Name/ID and remember to clean up unused service accounts afterwards.
  Also, you can see the permissions/roles for the given service account and project here:
  https://console.cloud.google.com/iam-admin/iam?project=vmware-kubeapps-ci. 
  You should see something like
  Principal:
    flux-plugin-test-sa-4@vmware-kubeapps-ci.iam.gserviceaccount.com 
  Name:
  	flux-plugin-test-sa-4
  Role: 
    Artifact Registry Administrator
    Artifact Registry Repository Administrator
    Viewer 
- The service account key file can be downloaded with Google Cloud Console
     Under IAM & Admin -> Service Accounts

## Sanity checks 
Here are (some of) the sanity checks I use to make sure everything is working that don't involve kubeapps:
```
$ export GOOGLE_APPLICATION_CREDENTIALS=..../gcloud-kubeapps-flux-test-sa-key-file.json

$ export GCP_TOKEN=$(gcloud auth application-default print-access-token)
```
FYI: GCP access token expires 1 hour after it's issued

1. check PING is working 
  * with service account access token 
```
$ curl -iL https://us-west1-docker.pkg.dev/v2/ -H "Authorization: Bearer $GCP_TOKEN"
HTTP/2 200 
docker-distribution-api-version: registry/2.0
date: Thu, 25 Aug 2022 06:29:59 GMT
content-length: 0
content-type: text/html; charset=UTF-8
```

  * with JSON key file
  You will need to install [oauth2l tool](https://github.com/google/oauth2l) locally
```
$ curl -iL https://us-west1-docker.pkg.dev/v2/ -H "$(oauth2l header --scope=cloud-platform)"
HTTP/2 200 
docker-distribution-api-version: registry/2.0
date: Fri, 26 Aug 2022 02:18:15 GMT
content-length: 0
content-type: text/html; charset=UTF-8
```

2. check [catalog API](https://github.com/opencontainers/distribution-spec/blob/main/spec.md#listing-repositories) is working 
  * with service account access token
```
$ curl -iL https://us-west1-docker.pkg.dev/v2/_catalog -H "Authorization: Bearer $GCP_TOKEN"
  HTTP/2 200
  content-type: application/json; charset=utf-8
  docker-distribution-api-version: registry/2.0
  date: Wed, 24 Aug 2022 16:10:18 GMT
  content-length: 75

  {"repositories":["vmware-kubeapps-ci/stefanprodan-podinfo-clone/podinfo"]}
```
  * with JSON key file
```
$ curl -iL https://us-west1-docker.pkg.dev/v2/_catalog -H "$(oauth2l header --scope=cloud-platform)"
HTTP/2 200 
content-type: application/json; charset=utf-8
docker-distribution-api-version: registry/2.0
date: Fri, 26 Aug 2022 02:22:52 GMT
content-length: 75

{"repositories":["vmware-kubeapps-ci/stefanprodan-podinfo-clone/podinfo"]}
```

3. check flux is able to reconcile a `HelmRepository` and a `HelmRelease`:
  * with service account access tokens (short-lived)
```
$ kubectl create secret docker-registry gcp-repo-auth-9 \
  --docker-server=us-west1-docker.pkg.dev \
  --docker-username=oauth2accesstoken \
  --docker-password="$(gcloud auth application-default print-access-token)"

$ flux create source helm podinfo-9 \
       --url=oci://us-west1-docker.pkg.dev/vmware-kubeapps-ci/stefanprodan-podinfo-clone \
       --namespace=default \
       --secret-ref=gcp-repo-auth-9

$ flux create hr podinfo-9 \
      --source=HelmRepository/podinfo-9.default \
      --chart=podinfo
✚ generating HelmRelease
► applying HelmRelease
✔ HelmRelease created
◎ waiting for HelmRelease reconciliation
✔ HelmRelease podinfo-9 is ready
✔ applied revision 6.1.8
```
  * with JSON key file (long-lived)
```
kubectl create secret docker-registry gcp-repo-auth \
  --docker-server=us-west1-docker.pkg.dev \
  --docker-username=_json_key \
  --docker-password="$(cat $GOOGLE_APPLICATION_CREDENTIALS)"

$ flux create source helm podinfo-7 \
       --url=oci://us-west1-docker.pkg.dev/vmware-kubeapps-ci/stefanprodan-podinfo-clone \
       --namespace=default \
       --secret-ref=gcp-repo-auth
✚ generating HelmRepository source
► applying HelmRepository source
✔ source created
◎ waiting for HelmRepository source reconciliation
✔ HelmRepository source reconciliation completed

$ flux create hr podinfo-7 \
       --source=HelmRepository/podinfo-7.default \
       --chart=podinfo
✚ generating HelmRelease
► applying HelmRelease
✔ HelmRelease updated
◎ waiting for HelmRelease reconciliation
✔ HelmRelease podinfo-7 is ready
✔ applied revision 6.1.8
```

## Environment setup for GCP auto-login scenario
Motivation: https://github.com/vmware-tanzu/kubeapps/issues/5007#issuecomment-1233691352
This feature is for something flux authors call "auto-login" or
"contextual login" and that basically means using a "secured" HelmRepository without
a secretRef

per https://cloud.google.com/kubernetes-engine/docs/how-to/workload-identity:

Setup instructions:
- Login to GCP console
- Enable the Google Kubernetes Engine API. 
- Ensure that you have enabled the IAM Service Account Credentials API
- Ensure that you have the following IAM roles:
  * roles/container.admin
  * roles/iam.serviceAccountAdmin
  FWIW, having the role 'Owner' will satisfy this
- Create a new GKE cluster with Workload Identity enabled
  per https://cloud.google.com/kubernetes-engine/docs/how-to/workload-identity#enable_on_cluster
  * Name: cluster-flux-plugin-auto-login-test
  * Zone: us-west1-c  
