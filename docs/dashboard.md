# Kubeapps Dashboard

Kubeapps comes with an in-cluster dashboard that offers a web UI to easily manage the deployments created by Helm in your cluster and to manage your Kubeless functions.

## Start the Dashboard

You can easily safely access the dashboard from your system by running:

```
kubeapps dashboard
```

This will run a HTTP proxy to access the dashboard safely and will open your default browser to it.

## Deploying new applications using the Dashboard

Once you have the Dashboard up and running, you can start deploying applications into your cluster.

<img src="https://github.com/kubeapps/kubeapps/raw/master/img/dashboard.png" width="100">

Select one application for the list of charts in the official Kubernetes chart repository. In this example we will be deploying MariaDB.

<img src="https://github.com/kubeapps/kubeapps/raw/master/img/mariadb.png" width="100">

Once you click on "Install mariadb" you will be able to select which namespace of your cluster you want this to be deployed:

<img src="https://github.com/kubeapps/kubeapps/raw/master/img/namespace.png" width="100">

Now click on "Deploy" and you will be able to track your new Kubernetes Deployment directly from your browser.

<img src="https://github.com/kubeapps/kubeapps/raw/master/img/mariadb-deploy.png" width="100">

## Listing all the deployments managed by Helm

On the "Deployments" menu you can get a list of the deployments in your cluster that are managed by Helm.

<img src="https://github.com/kubeapps/kubeapps/raw/master/img/deployments.png" width="100">

## Removing existing deployments

You can remove any of the deployments that are managed by Helm, by clicking on the Remove button:

<img src="https://github.com/kubeapps/kubeapps/raw/master/img/delete-mariadb.png" width="100">

## Functions Dashboard

TBD