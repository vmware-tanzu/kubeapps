## Step 4: Deploy and Manage Applications with Kubeapps

Once Kubeapps has been configured with one or more application repositories, it can be used to manage and deploy applications in the cluster.

The following sections discuss how to perform common tasks related to application management, including deploying an application, listing available applications and deleting applications.

### Deploy a new application

To deploy a new application, follow the steps below:

1. Log in to Kubeapps to arrive at the Dashboard welcome page:

  ![Dashboard main page](./img/step-4-1.png)

2. Use the _Catalog_ menu to select an application from the list of applications available. This example assumes you want to deploy MariaDB.

  ![MariaDB chart](./img/step-4-2.png)

3. Click the _Deploy_ button. You will be prompted for the release name, cluster namespace and values for your application deployment.

  ![MariaDB installation](./img/step-4-3.png)

4. Click the _Submit_ button.

The application is deployed. The status of the deployment can be tracked directly from the browser. The _Notes_ section of the deployment page contains important information to help you use the application.

  ![MariaDB deployment](./img/step-4-4.png)

### List all the applications running in your cluster

The _Applications_ page displays a list of the application deployments in your cluster.

![Deployment list](./img/step-4-5.png)

### Remove existing application deployments

Running applications can be removed from the cluster by clicking the _Delete_ button on the application's status page:

![Deployment removal](./img/step-4-6.png)

At the end of this step, you should be able to use Kubeapps for common application management and deployment tasks. Continue reading for a collection of [useful links and references to help you maximize your usage of Kubeapps](./conclusion.md).
