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
  Roles: 
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
"contextual login" and that basically means using a "secured" OCI HelmRepository without
a secret ref. 

per https://cloud.google.com/kubernetes-engine/docs/how-to/workload-identity:

Setup instructions:
- Make sure your shell is logged in to GCP
```
  $ gcloud auth login
```
- Install gke-gcloud-auth-plugin for use with kubectl by following https://cloud.google.com/blog/products/containers-kubernetes/kubectl-auth-changes-in-gke
```
  $ gcloud components install gke-gcloud-auth-plugin
```
- Login to GCP console
- Enable the Google Kubernetes Engine API. 
- Ensure that you have enabled the IAM Service Account Credentials API
- Ensure that you have the following IAM roles:
  * roles/container.admin
  * roles/iam.serviceAccountAdmin
  FWIW, having the role 'Owner' will satisfy this
- Create a new GKE cluster with Workload Identity enabled
  per https://cloud.google.com/kubernetes-engine/docs/how-to/workload-identity#enable_on_cluster
  * Mode: Standard*
  * Name: cluster-flux-plugin-auto-login-test
  * Zone: us-west1-c
  * Workload Identity: Enabled 
  * Workload pool: 	vmware-kubeapps-ci.svc.id.goog (automatically set)	
  Equivalent command line:
  ```
  $ gcloud beta container --project "vmware-kubeapps-ci" clusters create "cluster-flux-plugin-auto-login-test" --zone "us-west1-c" --no-enable-basic-auth --cluster-version "1.22.12-gke.300" --release-channel "regular" --machine-type "e2-medium" --image-type "COS_CONTAINERD" --disk-type "pd-standard" --disk-size "100" --metadata disable-legacy-endpoints=true --scopes "https://www.googleapis.com/auth/devstorage.read_only","https://www.googleapis.com/auth/logging.write","https://www.googleapis.com/auth/monitoring","https://www.googleapis.com/auth/servicecontrol","https://www.googleapis.com/auth/service.management.readonly","https://www.googleapis.com/auth/trace.append" --max-pods-per-node "110" --num-nodes "3" --logging=SYSTEM,WORKLOAD --monitoring=SYSTEM --enable-ip-alias --network "projects/vmware-kubeapps-ci/global/networks/default" --subnetwork "projects/vmware-kubeapps-ci/regions/us-west1/subnetworks/default" --no-enable-intra-node-visibility --default-max-pods-per-node "110" --no-enable-master-authorized-networks --addons HorizontalPodAutoscaling,HttpLoadBalancing,GcePersistentDiskCsiDriver --enable-autoupgrade --enable-autorepair --max-surge-upgrade 1 --max-unavailable-upgrade 0 --workload-pool "vmware-kubeapps-ci.svc.id.goog" --enable-shielded-nodes --node-locations "us-west1-c"

  Creating cluster cluster-flux-plugin-auto-login-test in us-west1-c... Cluster is being health-checked (master is healthy)...done.
  
  Created [https://container.googleapis.com/v1beta1/projects/vmware-kubeapps-ci/zones/us-west1-c/clusters/cluster-flux-plugin-auto-login-test]. To inspect the contents of your cluster, go to: https://console.cloud.google.com/kubernetes/workload_/gcloud/us-west1-c/cluster-flux-plugin-auto-login-test?project=vmware-kubeapps-ci

  kubeconfig entry generated for cluster-flux-plugin-auto-login-test.
  NAME                                 LOCATION    MASTER_VERSION   MASTER_IP       MACHINE_TYPE  NODE_VERSION     NUM_NODES  STATUS
  cluster-flux-plugin-auto-login-test  us-west1-c  1.22.12-gke.300  35.233.135.157  e2-medium     1.22.12-gke.300  3          RUNNING
   ```
- After create finishes, make sure that Enable GKE Metadata Server is set for the default-pool
  per https://cloud.google.com/kubernetes-engine/docs/how-to/workload-identity#option_2_node_pool_modification
  ```
  $ gcloud container node-pools update default-pool --cluster=cluster-flux-plugin-auto-login-test --workload-metadata=GKE_METADATA  --zone="us-west1-c"

  Default change: During creation of nodepools or autoscaling configuration changes for cluster versions greater than 1.24.1-gke.800 a default location policy is applied. For Spot and PVM it defaults to ANY, and for all other VM kinds a BALANCED policy is used. To change the default values use the `--location-policy` flag.

  Updating node pool default-pool...done.                                                     
  Updated [https://container.googleapis.com/v1/projects/vmware-kubeapps-ci/zones/us-west1-c/clusters/cluster-flux-plugin-auto-login-test/nodePools/default-pool].
  ```
- Prepare cluster to use use Workload Identity
  per https://cloud.google.com/kubernetes-engine/docs/how-to/workload-identity#authenticating_to
```
  $ gcloud container clusters get-credentials cluster-flux-plugin-auto-login-test --zone us-west1-c
    Fetching cluster endpoint and auth data.
    kubeconfig entry generated for cluster-flux-plugin-auto-login-test.
  $ kubectl config get-contexts
    CURRENT   NAME CLUSTER AUTHINFO NAMESPACE
    *  gke_vmware-kubeapps-ci_us-west1-c_cluster-flux-plugin-auto-login-test   gke_vmware-kubeapps-ci_us-west1-c_cluster-flux-plugin-auto-login-test   gke_vmware-kubeapps-ci_us-west1-c_cluster-flux-plugin-auto-login-test   
      kind-kubeapps kind-kubeapps kind-kubeapps

  $ gcloud iam service-accounts create flux-plugin-test-wi-sa --project=vmware-kubeapps-ci
  Created service account [flux-plugin-test-wi-sa].

  $ gcloud projects add-iam-policy-binding vmware-kubeapps-ci \
    --member "serviceAccount:flux-plugin-test-wi-sa@vmware-kubeapps-ci.iam.gserviceaccount.com" \
    --role "roles/artifactregistry.reader"
  Updated IAM policy for project [vmware-kubeapps-ci].
  bindings:
  - members:
    - serviceAccount:flux-plugin-test-wi-sa@vmware-kubeapps-ci.iam.gserviceaccount.com
    role: roles/artifactregistry.reader

  Also need roles:
    - Artifact Registry Repository Administrator
    - Viewer
```
- install flux on GKE cluster
```
$ make deploy-flux-controllers
kubectl apply -f https://github.com/fluxcd/flux2/releases/download/v0.34.0/install.yaml
...
$ kubectl get pods -n flux-system
NAME                                           READY   STATUS    RESTARTS   AGE
helm-controller-7d658687cb-xm5mc               1/1     Running   0          2m14s
image-automation-controller-7c77759c96-8hnzz   1/1     Running   0          2m13s
image-reflector-controller-76c455d887-kn4n8    1/1     Running   0          2m12s
kustomize-controller-85b8994c7d-d7wj9          1/1     Running   0          2m11s
notification-controller-78bb45df6c-ptskm       1/1     Running   0          2m10s
source-controller-54c7c7c777-7qhx8             1/1     Running   0          2m9s

$ flux version --context=gke_vmware-kubeapps-ci_us-west1-c_cluster-flux-plugin-auto-login-test
flux: v0.34.0
helm-controller: v0.24.0
image-automation-controller: v0.25.0
image-reflector-controller: v0.21.0
kustomize-controller: v0.28.0
notification-controller: v0.26.0
source-controller: v0.29.0
```

- Allow the Kubernetes service account to impersonate the IAM service account by adding an IAM policy binding 
between the two service accounts (IAM service account and flux source-controller GKE service account). This binding allows the Kubernetes service account to act as the IAM service account.
```
$ gcloud iam service-accounts add-iam-policy-binding flux-plugin-test-wi-sa@vmware-kubeapps-ci.iam.gserviceaccount.com \
  --role roles/iam.workloadIdentityUser \
  --member "serviceAccount:vmware-kubeapps-ci.svc.id.goog[flux-system/source-controller]"

Updated IAM policy for serviceAccount [flux-plugin-test-wi-sa@vmware-kubeapps-ci.iam.gserviceaccount.com].
bindings:
- members:
  - serviceAccount:vmware-kubeapps-ci.svc.id.goog[flux-system/source-controller]
  role: roles/iam.workloadIdentityUser
etag: BwXo6qTIstw=
version: 1

```
- Update your flux source-controller pod spec to schedule the workloads on nodes that use Workload Identity and to use the annotated Kubernetes service account. The recommended way to do this is to use flux CLI to bootstrap the flux source-controller running in GKE cluster using a Kustomize-er in a new Git repo created for the purpose of having a Kustomizer. See https://fluxcd.io/flux/components/source/helmrepositories/#gcp and https://github.com/fluxcd/flux2-kustomize-helm-example. I did test this way successfully, but I here I am skipping it because it would overly complicate an already complicated scenario.  For the purposes of testing out just a single scenario in a cluster created for that purpose only, I am going to workaround by doing:
```
  $ kubectl annotate serviceaccount source-controller --namespace flux-system --context=gke_vmware-kubeapps-ci_us-west1-c_cluster-flux-plugin-auto-login-test iam.gke.io/gcp-service-account=flux-plugin-test-wi-sa@vmware-kubeapps-ci.iam.gserviceaccount.com
  serviceaccount/source-controller annotated

  $ kubectl get serviceaccount/source-controller --namespace flux-system --context=gke_vmware-kubeapps-ci_us-west1-c_cluster-flux-plugin-auto-login-test -o yaml
  apiVersion: v1
  kind: ServiceAccount
  metadata:
    annotations:
      iam.gke.io/gcp-service-account: flux-plugin-test-wi-sa@vmware-kubeapps-ci.iam.gserviceaccount.com
  ...
```
- Now test with flux. flux CLI currently does not support spec.provider option so we use kubectl apply to work-around:
apply the following YAML
```
$ kubectl create --context=gke_vmware-kubeapps-ci_us-west1-c_cluster-flux-plugin-auto-login-test -f - <<EOF
apiVersion: source.toolkit.fluxcd.io/v1beta2
kind: HelmRepository
metadata:
  name: podinfo-1
  namespace: default
spec:
  type: "oci"
  provider: "gcp"
  interval: 45s
  url: oci://us-west1-docker.pkg.dev/vmware-kubeapps-ci/stefanprodan-podinfo-clone
EOF

$ kubectl get helmrepositories --context=gke_vmware-kubeapps-ci_us-west1-c_cluster-flux-plugin-auto-login-test
NAME        URL                                                                           AGE   READY   STATUS
podinfo-1   oci://us-west1-docker.pkg.dev/vmware-kubeapps-ci/stefanprodan-podinfo-clone   7s    True    Helm repository is ready

$ flux create hr podinfo-1 --source=HelmRepository/podinfo-1.default --chart=podinfo
✚ generating HelmRelease
► applying HelmRelease
✔ HelmRelease created
◎ waiting for HelmRelease reconciliation
✔ HelmRelease podinfo-1 is ready
✔ applied revision 6.1.8
```
- install kubeapps and verify scenario with kubeapps:
```
$ helm upgrade kubeapps bitnami/kubeapps --install --namespace kubeapps --create-namespace \
  --set "packaging.helm.enabled=false" \
  --set "packaging.flux.enabled=true" \
  --kube-context gke_vmware-kubeapps-ci_us-west1-c_cluster-flux-plugin-auto-login-test \
  --version 10.3.3 \
  --wait

  NAME: kubeapps
  LAST DEPLOYED: Mon Sep 19 18:14:56 2022
  NAMESPACE: kubeapps
  STATUS: deployed
  REVISION: 1
  TEST SUITE: None
  NOTES:
  CHART NAME: kubeapps
  CHART VERSION: 10.3.3
  APP VERSION: 2.5.1** Please be patient while the chart is being deployed **

$ kubectl get pods -n kubeapps 
  NAME                                             READY   STATUS    RESTARTS        AGE
  kubeapps-5d59d586d4-8jbwk                        1/1     Running   0               8m49s
  kubeapps-5d59d586d4-9dtjz                        1/1     Running   0               8m50s
  kubeapps-internal-dashboard-f8b6f4865-zmn5n      1/1     Running   0               8m50s
  kubeapps-internal-dashboard-f8b6f4865-zt66l      1/1     Running   0               8m50s
  kubeapps-internal-kubeappsapis-ff5fc986f-krj9d   1/1     Running   1 (8m13s ago)   8m50s
  kubeapps-internal-kubeappsapis-ff5fc986f-vxzvx   1/1     Running   2 (8m2s ago)    8m50s
  kubeapps-internal-kubeops-5bbbf9c6bd-ms6k6       1/1     Running   0               8m50s
  kubeapps-internal-kubeops-5bbbf9c6bd-r7hrv       1/1     Running   0               8m50s
  kubeapps-redis-master-0                          1/1     Running   0               8m49s
  kubeapps-redis-replicas-0                        1/1     Running   0               8m49s
```
- Update your flux source-controller pod spec to schedule the workloads on nodes that use Workload Identity and to use the annotated Kubernetes service account. Same way we did it for flux.
```
$ gcloud iam service-accounts add-iam-policy-binding flux-plugin-test-wi-sa@vmware-kubeapps-ci.iam.gserviceaccount.com \
  --role roles/iam.workloadIdentityUser \
  --member "serviceAccount:vmware-kubeapps-ci.svc.id.goog[kubeapps/kubeapps-internal-kubeappsapis]"
  Updated IAM policy for serviceAccount [flux-plugin-test-wi-sa@vmware-kubeapps-ci.iam.gserviceaccount.com].
  bindings:
  - members:
    - serviceAccount:vmware-kubeapps-ci.svc.id.goog[flux-system/source-controller]
    - serviceAccount:vmware-kubeapps-ci.svc.id.goog[kubeapps/kubeapps-internal-kubeappsapis]
    role: roles/iam.workloadIdentityUser
  etag: BwXpLQ8cSyY=
  version: 1

$ kubectl annotate serviceaccount kubeapps-internal-kubeappsapis --namespace kubeapps --context=gke_vmware-kubeapps-ci_us-west1-c_cluster-flux-plugin-auto-login-test iam.gke.io/gcp-service-account=flux-plugin-test-wi-sa@vmware-kubeapps-ci.iam.gserviceaccount.com 

$ kubectl get serviceaccount/kubeapps-internal-kubeappsapis --namespace kubeapps --context=gke_vmware-kubeapps-ci_us-west1-c_cluster-flux-plugin-auto-login-test -o yaml
apiVersion: v1
automountServiceAccountToken: true
kind: ServiceAccount
metadata:
  annotations:
    iam.gke.io/gcp-service-account: flux-plugin-test-wi-sa@vmware-kubeapps-ci.iam.gserviceaccount.com
...
```

- Optional: to make a little easier to collect logs from kubeappsapis pod
```
$ kubectl scale --replicas=1 deployment/kubeapps-internal-kubeappsapis -n kubeapps
```

- Optional: to get the latest kubeapps-apis bits on GKE, I pushed the locally built image to ghcr.io: 
```
$ docker login ghcr.io -u $GITHUB_USER -p $GITHUB_TOKEN 
$ export TAG=dev308
$ docker tag kubeapps/kubeapps-apis:$TAG ghcr.io/gfichtenholt/kubeapps/kubeapps-apis:$TAG
$ docker push ghcr.io/gfichtenholt/kubeapps/kubeapps-apis:$TAG
$ docker logout ghcr.io
# Then change visibility of package repository 'kubeapps/kubeapps-apis'
# from private to public using github portal

$ kubectl set image deployment/kubeapps-internal-kubeappsapis -n kubeapps kubeappsapis=ghcr.io/gfichtenholt/kubeapps/kubeapps-apis:$TAG --record --context=gke_vmware-kubeapps-ci_us-west1-c_cluster-flux-plugin-auto-login-test

# to verify:
$ kubectl get pods -n kubeapps --context gke_vmware-kubeapps-ci_us-west1-c_cluster-flux-plugin-auto-login-test
```

- The following is needed in order to obtain an auth token for grpc calls:
```
$ kubectl create -f - <<EOF
apiVersion: v1
kind: Secret
type: kubernetes.io/service-account-token
metadata:
  name: kubeapps-operator-token-bz7x5
  annotations:
    kubernetes.io/service-account.name: "kubeapps-operator"
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: kubeapps-operator
  namespace: default
secrets:
  - name: kubeapps-operator-token-bz7x5
EOF

$ kubectl create clusterrolebinding kubeapps-operator \
    --clusterrole=cluster-admin \
    --serviceaccount=default:kubeapps-operator
clusterrolebinding.rbac.authorization.k8s.io/kubeapps-operator created

$ export token=$(kubectl get -n default secret $(kubectl get -n default serviceaccount kubeapps-operator -o jsonpath='{.secrets[].name}') -o go-template='{{.data.token | base64decode}}')

# assuming local port 8080 is already used by something unrelated
$ kubectl -n kubeapps --context=gke_vmware-kubeapps-ci_us-west1-c_cluster-flux-plugin-auto-login-test port-forward svc/kubeapps-internal-kubeappsapis 8081:8080

# sanity check 1
$ grpcurl -plaintext localhost:8081 kubeappsapis.core.plugins.v1alpha1.PluginsService.GetConfiguredPlugins
{
  "plugins": [
    {
      "name": "fluxv2.packages",
      "version": "v1alpha1"
    },
    {
      "name": "resources",
      "version": "v1alpha1"
    }
  ]
}

$ grpcurl -plaintext -d '{"context": {"cluster": "default", "namespace": "default"}}' -H "Authorization: Bearer $token" localhost:8081 kubeappsapis.core.packages.v1alpha1.PackagesService.GetAvailablePackageSummaries
  {
    "availablePackageSummaries": [
      {
        "availablePackageRef": {
          "context": {
            "cluster": "default",
            "namespace": "default"
          },
          "identifier": "podinfo-1/podinfo",
          "plugin": {
            "name": "fluxv2.packages",
            "version": "v1alpha1"
          }
        },
        "name": "podinfo",
        "latestVersion": {
          "pkgVersion": "6.1.8"
        },
        "displayName": "podinfo",
        "shortDescription": "Podinfo Helm chart for Kubernetes",
        "categories": [
          ""
        ]
      }
    ]
  }

$ grpcurl -plaintext -d '{"available_package_ref": {"context": {"cluster": "default", "namespace": "default"}, "plugin": {"name": "fluxv2.packages", "version": "v1alpha1"}, "identifier": "podinfo-1/podinfo"}, "target_context": {"cluster": "default", "namespace": "default"}, "name": "podinfo-1"}' -H "Authorization: Bearer $token" localhost:8081 kubeappsapis.core.packages.v1alpha1.PackagesService.CreateInstalledPackage
{
  "installedPackageRef": {
    "context": {
      "cluster": "default",
      "namespace": "default"
    },
    "identifier": "podinfo-1",
    "plugin": {
      "name": "fluxv2.packages",
      "version": "v1alpha1"
    }
  }
}

$ $ grpcurl -plaintext -d '{"installed_package_ref": {"context": {"cluster": "default", "namespace": "default"}, "plugin": {"name": "fluxv2.packages", "version": "v1alpha1"}, "identifier": "podinfo-1"}}' -H "Authorization: Bearer $token" localhost:8081 kubeappsapis.core.packages.v1alpha1.PackagesService.GetInstalledPackageDetail
{
  "installedPackageDetail": {
    "installedPackageRef": {
      "context": {
        "cluster": "default",
        "namespace": "default"
      },
      "identifier": "podinfo-1",
      "plugin": {
        "name": "fluxv2.packages",
        "version": "v1alpha1"
      }
    },
    "pkgVersionReference": {
      "version": "*"
    },
    "name": "podinfo-1",
    "currentVersion": {
      "pkgVersion": "6.1.8",
      "appVersion": "6.1.8"
    },
    "reconciliationOptions": {
      "interval": "1m"
    },
    "status": {
      "ready": true,
      "reason": "STATUS_REASON_INSTALLED",
      "userReason": "ReconciliationSucceeded: Release reconciliation succeeded"
    },
    "postInstallationNotes": "1. Get the application URL by running these commands:\n  echo \"Visit http://127.0.0.1:8080 to use your application\"\n  kubectl -n default port-forward deploy/podinfo-1 8080:9898\n",
    "availablePackageRef": {
      "context": {
        "cluster": "default",
        "namespace": "default"
      },
      "identifier": "podinfo-1/podinfo",
      "plugin": {
        "name": "fluxv2.packages",
        "version": "v1alpha1"
      }
    }
  }
}
```

## Misc Hacks
- To debug curl commands in a pod running in GKE cluster with Workload Identity enabled:
```
$ kubectl create --context=gke_vmware-kubeapps-ci_us-west1-c_cluster-flux-plugin-auto-login-test -f - <<EOF
apiVersion: v1
kind: Pod
metadata:
  name: workload-identity-test
  namespace: flux-system
spec:
  containers:
  - image: google/cloud-sdk:slim
    name: workload-identity-test
    command: ["sleep","infinity"]
  serviceAccountName: source-controller
  nodeSelector:
    iam.gke.io/gke-metadata-server-enabled: "true"
EOF

$ kubectl get pods -n flux-system --context=gke_vmware-kubeapps-ci_us-west1-c_cluster-flux-plugin-auto-login-test
workload-identity-test                         1/1     Running   0          41s

$ kubectl exec -it workload-identity-test --context=gke_vmware-kubeapps-ci_us-west1-c_cluster-flux-plugin-auto-login-test --namespace flux-system -- /bin/bash

root@workload-identity-test:/# curl -H "Metadata-Flavor: Google" http://metadata.google.internal/computeMetadata/v1/instance/service-accounts/default/email
flux-plugin-test-wi-sa@vmware-kubeapps-ci.iam.gserviceaccount.comroot@workload-identity-test:/# 

root@workload-identity-test:/# curl -H "Metadata-Flavor: Google" http://metadata.google.internal/computeMetadata/v1/instance/service-accounts/default/token
{"access_token":"ya29...."expires_in":3599,"token_type":"Bearer"}


```
- To "pause" GKE cluster, so that it does not incur billing charges:
```
$ gcloud container clusters resize cluster-flux-plugin-auto-login-test --num-nodes=0 --zone us-west1-c
```
- To "resume" GKE cluster:
```
$ gcloud container clusters resize cluster-flux-plugin-auto-login-test --num-nodes=3 --zone us-west1-c
```
- To delete GKE cluster:
```
$ gcloud container clusters delete gke_vmware-kubeapps-ci_us-west1-c_cluster-flux-plugin-auto-login-test
```
- To Join Google Cloud Slack open https://cloud.google.com/developers. You can click the slack icon toward the bottom of the page.
ref https://github.com/jenkinsci/google-compute-engine-plugin/issues/139#issuecomment-941327025
