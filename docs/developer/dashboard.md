# Kubeapps Dashboard Developer Guide

The dashboard is the main UI component of the Kubeapps project. Written in Javascript, the dashboard uses the React Javascript library for the frontend.

Written in Javascript using the React javascript library, the dashboard is the main UI for the Kubeapps project.

## Prerequisites

- [Git](https://git-scm.com/)
- [Node 8.x](https://nodejs.org/)
- [Yarn](https://yarnpkg.com)
- [Docker CE](https://www.docker.com/community-edition)
- [Kubernetes cluster](https://kubernetes.io/docs/setup/pick-right-solution/)
- [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/)
- [Telepresence](https://telepresence.io)

## Environment setup

```bash
export GOPATH=~/gopath/
export PATH=$GOPATH/bin:$PATH
export KUBEAPPS_DIR=$GOPATH/src/github.com/kubeapps/kubeapps
```
## Download kubeapps source code

The dashboard application code is located under the `dashboard/` directory of the Kubeapps project repository.

```bash
git clone --recurse-submodules https://github.com/kubeapps/kubeapps $KUBEAPPS_DIR
```

### Install Kubeapps in your cluster

Kubeapps is a Kubernetes-native application and to develop Kubeapps components we need a Kubernetes cluster with Kubeapps already installed. Follow the [Kubeapps installation guide](../user/install.md) to install Kubeapps in your cluster.

### Running in development

[Telepresence](https://www.telepresence.io/) is a local development tool for Kubernetes microservices. As the dashboard is a service running in the Kubernetes cluster we use telepresence to proxy requests to the Kubeapps Dashboard running in your cluster to your local development server.

First install the dashboard dependency packages:

```bash
yarn install
```

Next, create a `telepresence` shell to swap the `kubeapps-dashboard-ui` deployment in the `kubeapps` namespace, forwarding local port `3000` to port `8080` of the `kubeapps-dashboard-ui` pod.

```bash
telepresence --namespace kubeapps --method inject-tcp --swap-deployment kubeapps-dashboard-ui --expose 3000:8080 --run-shell
```

> **NOTE**: If you are having issues getting this setup working correctly, please try switching the telepresence proxying method in the above command to `vpn-tcp`. Refer to [the telepresence docs](https://www.telepresence.io/reference/methods) to learn more about the available proxying methods and their limitations.

Next, launch the dashboard within the telepresence shell

```bash
yarn run start
```

You can now access the local development server simply by accessing the dashboard as you usually would:

```bash
kubeapps dashboard --port=5000
```

#### Troubleshooting

In some cases, the react processes keep listening on the 3000 port, even when you disconnect telepresence. If you see that `localhost:3000` is still serving the dashboard, even with your telepresence down, check if there is a react process running (`ps aux | grep react`) and kill it.

### Running tests

Execute the following command within the dashboard directory to start the test runner which will watch for changes and automatically re-run the tests when changes are detected.

```bash
yarn run test
```

> **NOTE**: macOS users may need to install watchman (https://facebook.github.io/watchman/).

