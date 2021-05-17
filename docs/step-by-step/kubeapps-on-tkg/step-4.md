# Step 4 - Manage your Applications

In this step, once you have installed Kubeapps and configured some Application Repositories, you can use the Dashboard to start managing and deploying applications in your cluster.

The following sections walk you through some common tasks, namely: i) deploy an application; ii) listing all the applications; iii) deleting an application.

## Deploying a new application

- Start with the Dashboard welcome page:

  ![Dashboard main page](./img/step-4-1.png)

- Use the "Catalog" menu to select an application from the list of applications available. This example assumes you want to deploy MariaDB.

  ![MariaDB chart](./img/step-4-2.png)

- Click the "Deploy" button. You will be prompted for the release name, cluster namespace and values for your application deployment.

  ![MariaDB installation](./img/step-4-3.png)

- Click the "Submit" button. The application will be deployed. You will be able to track the new Kubernetes deployment directly from the browser. The "Notes" section of the deployment page contains important information to help you use the application.

  ![MariaDB deployment](./img/step-4-4.png)

### List all the applications running in your cluster

The "Applications" page displays a list of the application deployments in your cluster.

![Deployment list](./img/step-4-5.png)

### Remove existing application deployments

You can remove any of the applications from your cluster by clicking the "Delete" button on the application's status page:

![Deployment removal](./img/step-4-6.png)

## What to Do Next?

At this point, you have successfully installed Kuebapps allowing your users to log in to Kubeapps using your custom OIDC provider. Then, you have configured a public and a private Application Repository and, finally, you have installed and removed applications using Kubeapps.

Reach the developers at [#kubeapps on Kubernetes Slack](https://kubernetes.slack.com/messages/kubeapps) (click [here](http://slack.k8s.io) to sign up to the Kubernetes Slack). Please, feel free to [drop us an issue](https://github.com/kubeapps/kubeapps/issues/new) if you face any problems or run into a bug.

## Additional References

- [Using the Dashboard](https://github.com/kubeapps/kubeapps/blob/master/docs/user/dashboard.md)
