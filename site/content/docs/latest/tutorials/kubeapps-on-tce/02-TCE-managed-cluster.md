# Step 2B: Deploy a VMware Tanzu™ Community Edition managed cluster

In this step, the goal is to install a [managed TCE cluster](https://tanzucommunityedition.io/docs/v0.12/planning/#managed-cluster).

VMware Tanzu™ Community Edition supports the following infrastructure providers for managed clusters:

- AWS
- Azure
- Docker
- vSphere

TCE offers a UI for setting up all the different parameters needed.

## Preparing the OpenID Connect (OIDC) authentication

Before creating the cluster, it is necessary to set up a proper OIDC provider that Kubernetes use to authenticate requests.

This tutorial addresses the configuration of the _Google Identity Platform_ OIDC provider, but there are more options of [using an OAuth2/OIDC Provider with Kubeapps](https://github.com/vmware-tanzu/kubeapps/blob/main/site/content/docs/latest/tutorials/using-an-OIDC-provider.md).

### Setting up the Google credentials client

You can use an OAuth 2.0 client.
Create a new "Web Application" client following [these steps](https://support.google.com/cloud/answer/6158849?hl=en).

At the end of this section, you should have:

- _Issuer URL_: The IP or DNS address of the OIDC server. In this case, it might be `https://accounts.google.com`.
- _Client ID_: The client_id value that you obtain from your OIDC provider.
- _Client Secret_: The secret value that you obtain from your OIDC provider.
- _Scopes_: A comma-separated list of additional scopes to request in the token response.
- _Username Claim_: The name of your username claim. This is used to set a user’s username in the JSON Web Token (JWT) claim.
- _Groups Claim_: The name of your group's claim. This is used to set a user’s group in the JWT claim.

This information is used to both set up the Kubernetes cluster, and in a further step, configure Kubeapps so that both use Google as an identity provider.

## Deploying the cluster

1. Initialise the Tanzu Community Edition installer UI to spin up a management cluster.

   ```bash
   tanzu management-cluster create --ui
   ```

2. Choose your infrastructure provider and follow the wizard steps in the UI. For more information, visit the [official TCE documentation on clusters deployment](https://tanzucommunityedition.io/docs/v0.12/getting-started/#deploy-clusters).

   > OIDC data obtained in the previous section needs to be set in the UI during this step.

3. Once the management cluster is created, validate that it started successfully.

   ```bash
   tanzu management-cluster get
   ```

   The output should be similar to the following:

   ```bash
   NAME          NAMESPACE   STATUS   CONTROLPLANE  WORKERS  KUBERNETES        ROLES       PLAN
   kubeapps-tce  tkg-system  running  1/1           1/1      v1.22.8+vmware.1  management  dev


   Details:

   NAME                                                             READY  SEVERITY  REASON  SINCE  MESSAGE
   /kubeapps-tce                                                    True                     3m37s
   ├─ClusterInfrastructure - AWSCluster/kubeapps-tce                True                     3m42s
   ├─ControlPlane - KubeadmControlPlane/kubeapps-tce-control-plane  True                     3m37s
   │ └─Machine/kubeapps-tce-control-plane-n9bbs                     True                     3m42s
   └─Workers
   └─MachineDeployment/kubeapps-tce-md-0                            True                     3m54s
       └─Machine/kubeapps-tce-md-0-95787cb65-gfnkz                  True                     3m42s


   Providers:

   NAMESPACE                          NAME                   TYPE                    PROVIDERNAME  VERSION  WATCHNAMESPACE
   capa-system                        infrastructure-aws     InfrastructureProvider  aws           v1.2.0
   capi-kubeadm-bootstrap-system      bootstrap-kubeadm      BootstrapProvider       kubeadm       v1.0.1
   capi-kubeadm-control-plane-system  control-plane-kubeadm  ControlPlaneProvider    kubeadm       v1.0.1
   capi-system                        cluster-api            CoreProvider            cluster-api   v1.0.1
   ```

4. Capture the management cluster’s kubeconfig:

   ```bash
   tanzu management-cluster kubeconfig get kubeapps-tce --admin
   ```

   and select the cluster context to be used with `kubectl`:

   ```bash
   kubectl config use-context kubeapps-tce-admin@kubeapps-tce
   ```

5. Unlike unmanaged clusters, the TCE packages catalog is not added by default to managed clusters. In order to add it, run:

   ```bash
   tanzu package repository add tce-repo --url projects.registry.vmware.com/tce/main:0.12.0 --namespace tanzu-package-repo-global
   ```

The outcome of the actions above is a _managed_ TCE cluster running on your preferred infrastructure provider. From this _management_ cluster, new _workload_ clusters can be spun up and attached to it.

For the sake of simplicity, this tutorial only uses a single _management_ cluster with no attached _workload: clusters, but note that it is not a recommended strategy for production setups.

For further information on how to create TCE workload clusters please check [the official documentation](https://tanzucommunityedition.io/docs/v0.12/getting-started/#deploy-a-workload-cluster).

> Continue the tutorial by [preparing the Kubeapps deployment](./03-preparing-kubeapps-deployment.md).

## Tutorial index

1. [TCE cluster deployment preparation](./01-TCE-cluster-preparation.md)
2. [Deploying a managed cluster](./02-TCE-managed-cluster.md) or [Deploy an unmanaged cluster](./02-TCE-unmanaged-cluster.md)
3. [Preparing the Kubeapps deployment](./03-preparing-kubeapps-deployment.md)
4. [Deploying Kubeapps](./04-deploying-kubeapps.md)
5. [Further documentation for managing applications in Kubeapps](./05-managing-applications.md)