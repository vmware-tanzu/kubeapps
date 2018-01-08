# Kubeapps Dashboard

Kubeapps Dashboard is the main UI for the Kubeapps project. See
[Kubeapps](https://github.com/kubeapps/kubeapps) for information on how to
install this in your cluster.

The Dashboard is written in React, see the [React Developer
Guide](docs/react-developer-guide.md) for developer documentation.

## Setting up for development

### Prerequisites

- Node 9.3.0
- Yarn package manager
- Go 1.9
- A Kubernetes cluster with Kubeapps installed
- [Telepresence](https://telepresence.io)

### Setting up Kubeapps with additional services

In order for the new Dashboard to function correctly, we need to install an
in-development version of Kubeapps with some new services. These are currently
being developed in a pull request in the [manifests
repository](https://github.com/kubeapps/manifest/pull/36).

If you haven't already developed with Kubeapps before, first checkout the
sources we'll need:

```
go get -d github.com/kubeapps/manifest
go get -d github.com/kubeapps/kubeapps
```

Now checkout the changes from the PR above:

```
cd $GOPATH/src/github.com/kubeapps/manifest
git checkout -b prydonius-apprepos master
git pull https://github.com/prydonius/manifest.git apprepos
```

Now build the `kubeapps` binary with these changes:

```
cd ../kubeapps
./scripts/sync-manifests.sh
make
```

Finally, run `kubeapps up` to install/update Kubeapps in your cluster:

```
./kubeapps up
```

### Running in development

We will use Telepresence to proxy requests to the Kubeapps Dashboard running in
your cluster to your local development server. Run the following commands:

```
yarn # install any new packages
telepresence --namespace kubeapps --method inject-tcp --swap-deployment kubeapps-dashboard-ui --expose 3000:8080 --run yarn run start
```
