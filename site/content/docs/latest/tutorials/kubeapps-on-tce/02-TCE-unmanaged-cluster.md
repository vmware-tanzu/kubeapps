# Step 2A: Deploy a VMware Tanzu™ Community Edition unmanaged cluster

In this step, the goal is to install an [unmanaged TCE cluster](https://tanzucommunityedition.io/docs/v0.12/planning/#unmanaged-cluster).

By default, unmanaged clusters run locally via kind (default) or minikube with Tanzu components installed atop.

## Spin up a TCE unmanaged cluster

1. Create a cluster named for example `kubeapps-cluster`:

   ```bash
   tanzu unmanaged-cluster create kubeapps-cluster
   ```

2. Wait for the cluster to initialise:

   ```bash
   📁 Created cluster directory

   🧲 Resolving and checking Tanzu Kubernetes release (TKr) compatibility file
   projects.registry.vmware.com/tce/compatibility
   Compatibility file exists at ~/.config/tanzu/tkg/unmanaged/compatibility/projects.registry.vmware.com_tce_compatibility_v9

   🔧 Resolving TKr
   projects.registry.vmware.com/tce/tkr:v1.22.7-2
   TKr exists at ~/.config/tanzu/tkg/unmanaged/bom/projects.registry.vmware.com_tce_tkr_v1.22.7-2
   Rendered Config: ~/.config/tanzu/tkg/unmanaged/kubeapps-cluster/config.yaml
   Bootstrap Logs: ~/.config/tanzu/tkg/unmanaged/kubeapps-cluster/bootstrap.log

   🔧 Processing Tanzu Kubernetes Release

   🎨 Selected base image
   projects.registry.vmware.com/tce/kind:v1.22.7

   📦 Selected core package repository
   projects.registry.vmware.com/tce/repo-12:0.12.0

   📦 Selected additional package repositories
   projects.registry.vmware.com/tce/main:0.12.0

   📦 Selected kapp-controller image bundle
   projects.registry.vmware.com/tce/kapp-controller-multi-pkg:v0.30.1

   🚀 Creating cluster kubeapps-cluster
   Cluster creation using kind!
   ❤️  Checkout this awesome project at https://kind.sigs.k8s.io
   Base image downloaded
   Cluster created
   To troubleshoot, use:

   kubectl ${COMMAND} --kubeconfig ~/.config/tanzu/tkg/unmanaged/kubeapps-cluster/kube.conf

   📧 Installing kapp-controller
   kapp-controller status: Running

   📧 Installing package repositories
   tkg-core-repository package repo status: Reconcile succeeded

   🌐 Installing CNI
   calico.community.tanzu.vmware.com:3.22.1

   ✅ Cluster created

   🎮 kubectl context set to kubeapps-cluster

   View available packages:
       tanzu package available list
   View running pods:
       kubectl get po -A
   Delete this cluster:
       tanzu unmanaged delete kubeapps-cluster
   ```

3. The new unmanaged cluster is automatically added to your local kubeconfig. If you have [kubectl](https://kubernetes.io/docs/tasks/tools/#kubectl) installed locally, you can now use it to interact with the cluster.

4. Once the cluster is up and running, let's check that the TCE catalog has been automatically added to it.

   ```bash
   tanzu package repository list --all-namespaces
   ```

   And the output should contain the most up-to-date TCE main catalog:

   ```bash
   NAME                                          REPOSITORY                                TAG     STATUS               DETAILS  NAMESPACE
   projects.registry.vmware.com-tce-main-0.12.0  projects.registry.vmware.com/tce/main     0.12.0  Reconcile succeeded           tanzu-package-repo-global
   tkg-core-repository                           projects.registry.vmware.com/tce/repo-12  0.12.0  Reconcile succeeded           tkg-system
   ```

## Authentication for an unmanaged cluster

Unmanaged clusters are meant for development/testing mainly, and this tutorial uses the **service account authentication**.

Please remember that for any user-facing installation you should [configure an OAuth2/OIDC provider](https://github.com/vmware-tanzu/kubeapps/blob/main/site/content/docs/latest/tutorials/using-an-OIDC-provider.md) to enable secure user authentication with Kubeapps and the cluster.

### Credentials creation

Let's continue by creating a demo credential with which to access Kubeapps and Kubernetes.
Service account named `kubeapps-operator` with a `cluster-admin` role.

```bash
kubectl create --namespace default serviceaccount kubeapps-operator
kubectl create clusterrolebinding kubeapps-operator --clusterrole=cluster-admin --serviceaccount=default:kubeapps-operator
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: Secret
metadata:
  name: kubeapps-operator-token
  namespace: default
  annotations:
    kubernetes.io/service-account.name: kubeapps-operator
type: kubernetes.io/service-account-token
EOF
```

> **NOTE** It's not recommended to assign users the `cluster-admin` role for Kubeapps production usage.
>
> Please refer to the [Access Control](https://github.com/vmware-tanzu/kubeapps/blob/main/site/content/docs/latest/howto/access-control.md) documentation to configure fine-grained access control for users.

### Credentials retrieval

In order to access Kubeapps, a token is required. Given that you are using plain service account authentication, it is straightforward to obtain a token.
Keep the obtained token for later steps of the tutorial.

#### On Linux/macOS

```bash
kubectl get --namespace default secret kubeapps-operator-token -o go-template='{{.data.token | base64decode}}'
```

#### On Windows

Open a Powershell terminal and run:

```powershell
[Text.Encoding]::UTF8.GetString([System.Convert]::FromBase64String($(kubectl get --namespace default secret kubeapps-operator-token -o jsonpath='{.data.token}')))
```

> Continue the tutorial by [preparing the Kubeapps deployment](./03-preparing-kubeapps-deployment.md).

## Tutorial index

1. [TCE cluster deployment preparation](./01-TCE-cluster-preparation.md)
2. [Deploying a managed cluster](./02-TCE-managed-cluster.md) or [Deploy an unmanaged cluster](./02-TCE-unmanaged-cluster.md)
3. [Preparing the Kubeapps deployment](./03-preparing-kubeapps-deployment.md)
4. [Deploying Kubeapps](./04-deploying-kubeapps.md)
5. [Further documentation for managing applications in Kubeapps](./05-managing-applications.md)