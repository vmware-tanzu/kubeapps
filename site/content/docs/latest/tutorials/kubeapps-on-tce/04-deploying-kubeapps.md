# Step 4: Deploying Kubeapps on a VMware Tanzu™ Community Edition cluster

In the
Once your TCE cluster is up and running, you can deploy Kubeapps into it.

1. Check that Kubeapps is an available package in the cluster with:

    ```bash
    tanzu package available list kubeapps.community.tanzu.vmware.com
    ```

    > If Kubeapps package is not present, please check that your version of TCE is v0.13 or higher.

    In case you could not get Kubeapps showing up in the list of available packages, add it manually to the catalog by running (please change version accordingly):

    ```bash
    kubectl apply \
        -f https://raw.githubusercontent.com/vmware-tanzu/package-for-kubeapps/main/metadata.yaml \
        -f https://raw.githubusercontent.com/vmware-tanzu/package-for-kubeapps/main/9.0.3/package.yaml
    ```

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

4. **Optionally** you can check with Kapp which resources have been created when deploying Kubeapps:

    ```bash
    kapp inspect -a kubeapps-ctrl -n kubeapps
    ```

5. At this point, Kubeapps should be deployed and running in the TCE cluster. Open a browser and navigate to the URL defined for Kubeapps, for example [https://tce-cluster.foo.com](https://tce-cluster.foo.com).

> Continue the tutorial by [managing applications with Kubeapps](./05-Managing-applications.md).
