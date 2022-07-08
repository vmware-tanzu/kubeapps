# Deploying Kubeapps on a VMware Tanzu™ Community Edition cluster

Once your TCE cluster is up and running, you can deploy Kubeapps into it.

1. Check that Kubeapps is an available package in the cluster with:

    ```bash
    tanzu package available list kubeapps.community.tanzu.vmware.com
    ```

    > If Kubeapps is not present, please check that your version of TCE is v0.13 or higher.

2. Install Kubeapps with an optional _configuration values file_

    ```bash
    tanzu package install kubeapps --create-namespace -n kubeapps \
       --package-name kubeapps.community.tanzu.vmware.com \
       --version 8.1.7 \
       --values-file your-values-file.yaml
    ```

    Configuration values file is optional and allows you to customize the deployment of Kubeapps.

    > For your configuration values file, you can use exactly the same parameters specified in the [Bitnami Kubeapps chart.](https://github.com/bitnami/charts/tree/master/bitnami/kubeapps#parameters)

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

> Continue the tutorial by [bringing traffic into Kubeapps](./04-ingress-traffic.md).
