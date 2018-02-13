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

In order for the new Dashboard to function correctly, we need to install a
version of Kubeapps with some new services. This is currently being developed
out of the `0.3.0` branch in the Kubeapps repository.

If you haven't already developed with Kubeapps before, first checkout the
sources we'll need:

```
go get -d github.com/kubeapps/kubeapps
go get -d github.com/kubeapps/dashboard
```

Now checkout the `0.3.0` branch of the Kubeapps repository and build the
`kubeapps` binary:

```
cd $GOPATH/src/github.com/kubeapps/kubeapps
git checkout 0.3.0
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
cd ../dashboard
git checkout 2.0
yarn # install any new packages
telepresence --namespace kubeapps --method inject-tcp --swap-deployment kubeapps-dashboard-ui --expose 3000:8080 --run-shell
yarn run start # when telepresence returns a shell
```

Now, to access the React app, simply run `kubeapps dashboard` as you usually
would:

```
cd ../kubeapps
./kubeapps dashboard
```

#### Troubleshooting

In some cases, the react processes keep listening on the 3000 port, even when you disconnect telepresence. If you see that `localhost:3000` is still serving the dashboard, even with your telepresence down, check if there is a react process running (`ps aux | grep react`) and kill it.
