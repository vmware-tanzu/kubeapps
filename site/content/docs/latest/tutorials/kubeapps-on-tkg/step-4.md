## Step 4: Deploy and Manage Applications with Kubeapps

Once Kubeapps is configured with one or more application repositories, it can be used to manage and deploy applications in the cluster.

The following sections discuss how to perform common tasks related to application management, including deploying an application, upgrading it, listing available applications, performing rollbacks and deleting applications.

### Deploy a New Application

To deploy a new application, follow the steps below:

1. Log in to Kubeapps to arrive at the dashboard welcome page:

   ![Kubeapps home](../../img/kubeapps-on-tkg/kubeapps-applications-empty.png)

2. Use the **Catalog** menu to search for the desired application, for instance, _MariaDB_:

   ![MariaDB chart search](../../img/kubeapps-on-tkg/kubeapps-catalog-search.png)

3. Click on the application to see the details:

   ![MariaDB chart](../../img/kubeapps-on-tkg/kubeapps-chart-mariadb.png)

4. Click on the **Deploy** button. Kubeapps prompts for the release name, cluster namespace and the values for your application deployment.

   ![MariaDB installation](../../img/kubeapps-on-tkg/kubeapps-chart-mariadb-deploy.png)

5. Click the **Deploy** button at the end of the page. Wait a while until it gets deployed.

   ![MariaDB installation wait](../../img/kubeapps-on-tkg/kubeapps-chart-mariadb-wait.png)

The application is deployed. The status of the deployment can be tracked directly from the browser. The _Notes_ section of the deployment page contains important information to help you use the application.

![MariaDB deployment](../../img/kubeapps-on-tkg/kubeapps-chart-mariadb-deployed.png)

### List All Applications

The _Applications_ page displays a list of the application deployments in your cluster.

![Deployment list](../../img/kubeapps-on-tkg/kubeapps-applications-one.png)

### Upgrade an Application

To upgrade an application with new values, follow the steps below:

1. Navigate to the **Applications** page.
2. Click the application you wish to upgrade.

   ![MariaDB chart](../../img/kubeapps-on-tkg/kubeapps-chart-mariadb-deployed.png)

3. Click the **Upgrade** button.

   ![MariaDB chart](../../img/kubeapps-on-tkg/kubeapps-chart-mariadb-deployed.png)

4. Perform the changes and click the **Changes** tab to see the differences with the installed version.

   ![MariaDB installation](../../img/kubeapps-on-tkg/kubeapps-chart-mariadb-upgrade.png)

5. Click the **Deploy** button.

The application is upgraded with the new values. The status of the deployment can be tracked directly from the browser.

### Rollback an Application (only for Helm)

To rollback an application to a previous version, follow the steps below:

1. Navigate to the **Applications** page.
2. Click the application you wish to upgrade.

   ![MariaDB chart](../../img/kubeapps-on-tkg/kubeapps-chart-mariadb-deployed.png)

3. Click the **Rollback** button.

   ![MariaDB chart](../../img/kubeapps-on-tkg/kubeapps-chart-mariadb-deployed.png)

4. Select the version to rollback to and confirm the operation by clicking the **Rollback** button in the confirmation dialog.

   ![MariaDB installation](../../img/kubeapps-on-tkg/kubeapps-chart-mariadb-rollback.png)

The application is rolled back to the desired version. The status of the deployment can be tracked directly from the browser.

### Delete an Application

Running applications can be removed from the cluster by clicking the **Delete** button on the application's status page:

![Deployment removal](../../img/kubeapps-on-tkg/kubeapps-chart-mariadb-delete.png)

At the end of this step, you should be able to use Kubeapps for common application management and deployment tasks. Continue reading for a collection of [useful links and references to help you maximize your usage of Kubeapps](./conclusion.md).

## Tutorial index

1. [Step 1: Configure an Identity Management Provider in the Cluster](./step-1.md)
2. [Step 2: Configure and Install Kubeapps](./step-2.md)
3. [Step 3: Add Application Repositories to Kubeapps](./step-3.md)
4. [Step 4: Deploy and Manage Applications with Kubeapps](./step-4.md)