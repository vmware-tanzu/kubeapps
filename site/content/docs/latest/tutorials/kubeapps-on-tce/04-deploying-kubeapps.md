# Step 4: Deploying Kubeapps on a VMware Tanzu™ Community Edition cluster

Once your TCE cluster is up and running and you have a valid configuration values file, you can proceed to deploy Kubeapps.

One of the key features of Tanzu is its use of _Carvel_.
[Carvel](https://carvel.dev/) is a project that provides a set of reliable, single-purpose, composable tools that aid in your application building, configuration, and deployment to Kubernetes. This tutorial uses the Carvel packaging format to install Kubeapps.

1. Check that Kubeapps is an available package in the cluster with:

   ```bash
   tanzu package available list kubeapps.community.tanzu.vmware.com
   ```

   > In case you could not get Kubeapps showing up in the list of available packages, add it manually to the catalog by running (please change [package version accordingly](https://github.com/vmware-tanzu/package-for-kubeapps/tags)):
   >
   > ```bash
   > kubectl apply \
   >    -f https://raw.githubusercontent.com/vmware-tanzu/package-for-kubeapps/main/metadata.yaml \
   >    -f https://raw.githubusercontent.com/vmware-tanzu/package-for-kubeapps/main/8.1.7/package.yaml
   > ```

2. Install Kubeapps with the _configuration values file_ created in the previous step of the tutorial

   ```bash
   tanzu package install kubeapps --create-namespace -n kubeapps \
      --package-name kubeapps.community.tanzu.vmware.com \
      --version 8.1.7 \
      --values-file your-values-file.yaml
   ```

3. Check that Kubeapps has been successfully reconciled by running

   ```bash
   tanzu package installed list -n kubeapps
   ```

   And output should look like this:

   ```bash
   NAME      PACKAGE-NAME                         PACKAGE-VERSION  STATUS
   kubeapps  kubeapps.community.tanzu.vmware.com  8.1.7            Reconcile succeeded
   ```

   If Kubeapps could not be reconciled, you can get more information by running:

   ```bash
   tanzu package installed get kubeapps -n kubeapps
   ```

4. **Optionally** you can check with [Kapp](https://carvel.dev/kapp/) (part of Carvel tools) which resources have been created when deploying Kubeapps:

   ```bash
   kapp inspect -a kubeapps-ctrl -n kubeapps
   ```

5. At this point, Kubeapps is deployed and running in the TCE cluster.

   If you chose a **LoadBalancer** to access Kubeapps: wait for your cluster to assign a `LoadBalancer` IP or Hostname to the kubeapps Service and access it on that address:

   ```bash
   kubectl get service kubeapps --namespace kubeapps --watch
   ```

   If you chose an **Ingress** to access Kubeapps: open a browser and navigate to the FQDN defined for Kubeapps, for example [https://tce-cluster.foo.com](https://tce-cluster.foo.com).

   > When using OIDC, you need to configure your OAuth2 client to admit the `LoadBalancer` IP/Host or the `Ingress` FQDN as authorized origins and redirects. Please add the suffix `/oauth2/callback` to the redirect URLs in your OIDC provider setup.

> Continue the tutorial by [managing applications with Kubeapps](./05-managing-applications.md).

## Tutorial index

1. [TCE cluster deployment preparation](./01-TCE-cluster-preparation.md)
2. [Deploying a managed cluster](./02-TCE-managed-cluster.md) or [Deploy an unmanaged cluster](./02-TCE-unmanaged-cluster.md)
3. [Preparing the Kubeapps deployment](./03-preparing-kubeapps-deployment.md)
4. [Deploying Kubeapps](./04-deploying-kubeapps.md)
5. [Further documentation for managing applications in Kubeapps](./05-managing-applications.md)