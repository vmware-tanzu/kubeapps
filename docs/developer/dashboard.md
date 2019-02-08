# Kubeapps Dashboard Developer Guide

The dashboard is the main UI component of the Kubeapps project. Written in Javascript, the dashboard uses the React Javascript library for the frontend.

## Prerequisites

- [Git](https://git-scm.com/)
- [Node 8.x](https://nodejs.org/)
- [Yarn](https://yarnpkg.com)
- [Kubernetes cluster (v1.8+)](https://kubernetes.io/docs/setup/pick-right-solution/)
- [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/)
- [Docker CE](https://www.docker.com/community-edition)
- [Telepresence](https://telepresence.io)

*Telepresence is not a hard requirement, but is recommended for a better developer experience*

## Environment

```bash
export GOPATH=~/gopath/
export PATH=$GOPATH/bin:$PATH
export KUBEAPPS_DIR=$GOPATH/src/github.com/kubeapps/kubeapps
```
## Download the kubeapps source code

```bash
git clone --recurse-submodules https://github.com/kubeapps/kubeapps $KUBEAPPS_DIR
```

The dashboard application source is located under the `dashboard/` directory of the repository.

```bash
cd $KUBEAPPS_DIR/dashboard
```

### Install Kubeapps in your cluster

Kubeapps is a Kubernetes-native application. To develop and test Kubeapps components we need a Kubernetes cluster with Kubeapps already installed. Follow the [Kubeapps installation guide](../../chart/kubeapps/README.md) to install Kubeapps in your cluster.

### Running the dashboard in development

[Telepresence](https://www.telepresence.io/) is a local development tool for Kubernetes microservices. As the dashboard is a service running in the Kubernetes cluster we use telepresence to proxy requests to the dashboard running in your cluster to your local development host.

First install the dashboard dependency packages:

```bash
yarn install
```

Next, create a `telepresence` shell to swap the `kubeapps-internal-dashboard` deployment in the `kubeapps` namespace, forwarding local port `3000` to port `8080` of the `kubeapps-internal-dashboard` pod.

```bash
telepresence --namespace kubeapps --method inject-tcp --swap-deployment kubeapps-internal-dashboard --expose 3000:8080 --run-shell
```

> **NOTE**: If you encounter issues getting this setup working correctly, please try switching the telepresence proxying method in the above command to `vpn-tcp`. Refer to [the telepresence docs](https://www.telepresence.io/reference/methods) to learn more about the available proxying methods and their limitations.

Finally, launch the dashboard within the telepresence shell

```bash
yarn run start
```

You can now access the local development server simply by accessing the dashboard as you usually would:

```bash
kubeapps dashboard --port=5000
```

#### Troubleshooting

In some cases, the 'Create React App' scripts keep listening on the 3000 port, even when you disconnect telepresence. If you see that `localhost:3000` is still serving the dashboard, even with your telepresence down, check if there is a 'Create React App' script process running (`ps aux | grep react`) and kill it.

### Running tests

Execute the following command within the dashboard directory to start the test runner which will watch for changes and automatically re-run the tests when changes are detected.

```bash
yarn run test
```

> **NOTE**: macOS users may need to install watchman (https://facebook.github.io/watchman/).

