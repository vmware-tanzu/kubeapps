# Step 2B: Deploy a VMware Tanzu™ Community Edition managed cluster

VMware Tanzu™ Community Edition supports the following infrastructure providers for managed clusters:

- AWS
- Azure
- Docker
- vSphere

In the case of managed clusters, TCE offers a UI for setting up all the different parameters needed.

## Preparing the OpenID Connect authentication

Before creating the cluster, we will need to setup a proper OIDC provider that Kubernetes will use to authenticate requests.
This is a more secure approach than using service account tokens, specially for managed clusters, more suitable for production uses.

For this tutorial we will use a combination of DEX with a Google Identity Platform credential, but there are more options of [using an OAuth2/OIDC Provider with Kubeapps](/site/content/docs/latest/tutorials/using-an-OIDC-provider.md).

### Setting up Google credentials client

In the case of Google we can use an OAuth 2.0 client.
Create a new "Web Application" client following [this steps](https://support.google.com/cloud/answer/6158849?hl=en).

As the outcome for this section, we should have:

- Issuer URL: The IP or DNS address of the OIDC server. In our case this might be `https://accounts.google.com`.
- Client ID: The client_id value that you obtain from your OIDC provider.
- Client Secret: The secret value that you obtain from your OIDC provider.
- Scopes: A comma separated list of additional scopes to request in the token response.
- Username Claim: The name of your username claim. This is used to set a user’s username in the JSON Web Token (JWT) claim.
- Groups Claim: The name of your groups claim. This is used to set a user’s group in the JWT claim.

This information will be used to set up Kubernetes cluster and Kubeapps in order to use Google as an identity provider.

## Deploying the cluster

1. Initialize the Tanzu Community Edition installer UI to install a management cluster.

    ```bash
    tanzu management-cluster create --ui
    ```

2. Choose your infrastructure provider and follow the wizard steps in the UI. For more information, visit the [official TCE documentation on clusters deployment](https://tanzucommunityedition.io/docs/v0.12/getting-started/#deploy-clusters).

3. Once the management cluster is created, validate that it started successfully.

    ```bash
    tanzu management-cluster get
    ```

    The output will look similar to the following:

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

At this point, workload clusters could be created and be controlled from this management cluster. For simplicity reasons, we will continue the tutorial using only the management cluster, but it is not a recommended strategy for production setups.

For information on how to create TCE workload clusters please check [the official documentation](https://tanzucommunityedition.io/docs/v0.12/getting-started/#deploy-a-workload-cluster).

> Continue the tutorial by [deploying Kubeapps](./03-deploying-kubeapps.md).
