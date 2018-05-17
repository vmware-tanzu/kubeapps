# Kubeapps Dashboard

Kubeapps Dashboard is the main UI for the Kubeapps project. See
[Kubeapps](https://github.com/kubeapps/kubeapps) for information on how to
install this in your cluster.

The Dashboard is written in React, see the [React Developer
Guide](docs/react-developer-guide.md) for developer documentation.

## Setting up for development

### Prerequisites

* Node 8.9+
* Yarn package manager
* Go 1.10
* A Kubernetes cluster with Kubeapps installed
* [Telepresence](https://telepresence.io)

### Install Kubeapps in your cluster

Follow the [installation guide](../docs/install.md) to install Kubeapps in your development cluster.

### Running in development

We will use Telepresence to proxy requests to the Kubeapps Dashboard running in
your cluster to your local development server. Run the following commands:

```
cd dashboard/
yarn # install any new packages
telepresence --namespace kubeapps --method inject-tcp --swap-deployment kubeapps-dashboard-ui --expose 3000:8080 --run-shell
yarn run start # when telepresence returns a shell
```

Now, to access the React app, simply run `kubeapps dashboard` as you usually
would:

```
kubeapps dashboard --port=5000
```

#### Troubleshooting

In some cases, the react processes keep listening on the 3000 port, even when you disconnect telepresence. If you see that `localhost:3000` is still serving the dashboard, even with your telepresence down, check if there is a react process running (`ps aux | grep react`) and kill it.

### Running tests

The following command will start the test runner which will watch for changes and re-run.

**NOTE**: on macOS, you may need to install watchman (https://facebook.github.io/watchman/).

```
cd dashboard/
yarn run test
```
