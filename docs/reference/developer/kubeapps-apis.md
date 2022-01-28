# Experimental Kubeapps APIs service

This is an experimental service that we are not yet including in our released Kubeapps chart, providing an extendable protobuf-based API service enabling Kubeapps (or other packaging UIs) to interact with various Kubernetes packaging formats in a consistent way.

## Extendable

The API definitions are using Google protocol buffers - an Interface Definition Language that was created to ensure that API messages and API services can be defined, extended and maintained with the least friction, in a language-neutral way (though we are exclusively using go here).

gRPC is an RPC implementation that uses protocol buffers, providing RPC functionality across languages. Importantly, even though we'll be writing the API service in go, applications using the service can use any language they want and get pre-built clients for that language.

If you are unfamiliar with protocol buffers and gRPC, it's worthwhile to at least read the [introduction](https://grpc.io/docs/what-is-grpc/introduction/) and [core concepts](https://www.grpc.io/docs/what-is-grpc/core-concepts/).

Though it is possible to use a JS/TypeScript client, it is also possible to use a protocol buffer extensions provided by Google that enables optionally exposing the RPCs as http requests, demo'd here. This is made possible via the [grpc-gateway project](https://github.com/grpc-ecosystem/grpc-gateway) which proxies JSON http requests through to the relevant gRPC call. As shown in the demo (and by many others examples online), both http and grpc can be served on the same port without issue.

Together, this enables the best of both worlds: a well known and used Interface Definition Language for defining extendable APIs that can also be exposed via a rest-like http interface.

Finally, we've also chosen to use the [buf](https://buf.build/) tool for generating the code from the proto files. In the past we've used `protoc` (proto buffer compiler) and its extensions directly, but `buf` allows you to specify a simple yaml config instead, and also provides a `lint` command to ensure that your choice of API structure follows best practise, as well as ensuring you're aware when you break backwards compatibility.

## Plug-able

The kubeapps-apis service uses the standard [go plugin package](https://golang.org/pkg/plugin/) to be able to load api plugins at run-time.

Each plugin consists of 2 source files (and some generated files):

- A `.proto` file defining the service that uses the messages defined in relevant part of kubeappsapis.core, in `./proto/kubeappsapis/plugins/<plugin-name>`,
- A `main.go` that compiles to an .so file for that plugin. This `main.go` has two public functions: one to register the plugin with a GRPC server and one to register the plugin for the http handler as well as the implementation for the server. This may be split into further modules as the complexity of the plugin grows. This file is under `./plugins/<plugin-name>`

With this structure, the kubeapps-apis' main.go simply loads the `.so` files from the specified plugin dirs and register them when starting. You can see this in the [kubeapps-apis/server/server.go](server/server.go) file.

## Aggregated

When plugins are registered, they are also checked to see if they implement a core API (currently the only one is core.packages.v1alpha1). If they do, they are registered for use by the corresponding core API for aggregating results across plugins. See below for an example.

## CLI

Similar to most go commands, we've used [Cobra](https://github.com/spf13/cobra) for the CLI interface. Currently there is only a root command to run server, but we may later add a `version` subcommand or a `new-plugin` subcommand, but even without these it provides a lot of useful defaults for config, env var support etc.

## Trying it out

You can run `make run` to run the currently stubbed service. If passing the `KUBECONFIG` env var it will try to use your local configuration instead of relying on the usual inCluster config. This behavior is just intended for development purposes and it will get eventually removed.

```bash
export KUBECONFIG=/home/user/.kube/config # replace it with your desired kube config file
make run

I0514 14:14:52.969498 1932386 server.go:129] Successfully registered plugin "/home/michael/dev/vmware/kubeapps/cmd/kubeapps-apis/devel/fluxv2-packages-v1alpha1-plugin.so"
I0514 14:14:52.975884 1932386 server.go:129] Successfully registered plugin "/home/michael/dev/vmware/kubeapps/cmd/kubeapps-apis/devel/kapp-controller-packages-v1alpha1-plugin.so"
I0511 11:39:56.444553 4116647 server.go:25] Starting server on :50051
```

You can then verify the configured plugins endpoint via http:

```bash
curl -s http://localhost:50051/core/plugins/v1alpha1/configured-plugins | jq .
{
  "plugins": [
    {
      "name": "fluxv2.packages",
      "version": "v1alpha1"
    },
    {
      "name": "kapp_controller.packages",
      "version": "v1alpha1"
    }
  ]
}
```

or via gRPC (using the [grpcurl tool](https://github.com/fullstorydev/grpcurl)):

```bash
grpcurl -plaintext localhost:50051 kubeappsapis.core.plugins.v1alpha1.PluginsService.GetConfiguredPlugins
{
  "plugins": [
    {
      "name": "fluxv2.packages",
      "version": "v1alpha1"
    },
    {
      "name": "kapp_controller.packages",
      "version": "v1alpha1"
    }
  ]
}
```

To test the packages endpoints for the fluxv2 or kapp_controller plugins, you will currently need to build the image from the kubeapps root directory with:

```bash
IMAGE_TAG=dev1 make kubeapps/kubeapps-apis
```

and make that image available on your cluster somehow. If using kind, you can simply do:

```bash
kind load docker-image kubeapps/kubeapps-apis:dev1 --name kubeapps
```

You can edit the values file to change the `kubeappsapis.image.tag` field to match the tag above, or edit the deployment once deployed to match, such as:

```bash
kubectl set image deployment/kubeapps-internal-kubeappsapis -n kubeapps kubeappsapis=kubeapps/kubeapps-apis:dev1 --record
```

With the kubeapps-apis service running, you can then test the packages endpoints in cluster by port-forwarding the service in one terminal:

```bash
kubectl -n kubeapps port-forward svc/kubeapps-internal-kubeappsapis 8080:8080
```

and then curling or grpcurling in another:

```bash
# TOKEN value comes from the k8s secret associated with the service account of the user on behalf of which the call below is made
$ export TOKEN="Bearer eyJhbGciO..."
$ curl -s http://localhost:8080/plugins/fluxv2/packages/v1alpha1/packagerepositories?context.cluster=default -H "Authorization: $TOKEN" | jq .
{
  "repositories": [
    {
      "name": "bitnami",
      "namespace": "flux-system",
      "url": "https://charts.bitnami.com/bitnami"
    }
  ]
}

$ curl -s http://localhost:8080/plugins/kapp_controller/packages/v1alpha1/packagerepositories | jq .
{
  "repositories": [
    {
      "name": "repo-name.example.com",
      "url": "foo.registry.example.com/repo-name/main@sha256:cecd9b51b1f29a773a5228fe04faec121c9fbd2969de55b0c3804269a1d57aa5"
    }
  ]
}
```

Here is an example that shows how to use grpcurl to get the details on package "bitnami/apache" from helm plugin

```bash
$ export token="Bearer eyJhbGciO..."
$ grpcurl -plaintext -d '{"available_package_ref": {"context": {"cluster": "default", "namespace": "kubeapps"}, "plugin": {"name": "helm.packages", "version": "v1alpha1"}, "identifier": "bitnami/apache"}}' -H "Authorization: $token" localhost:8080 kubeappsapis.core.packages.v1alpha1.PackagesService.GetAvailablePackageDetail
{
  "availablePackageDetail": {
    "availablePackageRef": {
      "context": {
        "namespace": "kubeapps"
      },
      "identifier": "bitnami/apache",
      "plugin": {
        "name": "helm.packages",
        "version": "v1alpha1"
      }
    },
    "name": "apache",
    "version": {
      "pkgVersion": "8.8.6",
      "appVersion": "2.4.51"
    },
    "repoUrl": "https://charts.bitnami.com/bitnami",
    "homeUrl": "https://github.com/bitnami/charts/tree/master/bitnami/apache",
    "iconUrl": "https://bitnami.com/assets/stacks/apache/img/apache-stack-220x234.png",
    "displayName": "apache",
    "shortDescription": "Chart for Apache HTTP Server",
    "readme": "...",
    "sourceUrls": [
      "https://github.com/bitnami/bitnami-docker-apache",
      "https://httpd.apache.org"
    ],
    "maintainers": [
      {
        "name": "Bitnami",
        "email": "containers@bitnami.com"
      }
    ],
    "categories": [
      "Infrastructure"
    ]
  }
}
```

Or you can query the core API to get an aggregation of all package repositories (or packages) across the relevant plugins. This output will include the additional plugin field for each item:

```bash
curl -s http://localhost:8080/core/packages/v1alpha1/packagerepositories | jq .
{
  "repositories": [
    {
      "name": "bitnami",
      "namespace": "flux-system",
      "url": "https://charts.bitnami.com/bitnami",
      "plugin": {
        "name": "fluxv2.packages",
        "version": "v1alpha1"
      }
    },
    {
      "name": "demo-package-repository",
      "url": "k8slt/corp-com-pkg-repo:1.0.0",
      "plugin": {
        "name": "kapp_controller.packages",
        "version": "v1alpha1"
      }
    }
  ]
}
```

Of course, you will need to have the appropriate Flux HelmRepository or Carvel PackageRepository available ([example](https://github.com/vmware-tanzu/carvel-kapp-controller/tree/develop/examples/packaging-with-repo)) in your cluster.

## Hacking

A few extra tools will be needed to contribute to the development of this service.

### GOPATH env variable

Make sure your GOPATH environment variable is set.
You can use the value of command

```bash
go env GOPATH
```

### Install go cli deps

You should be able to install the exact versions of the various go CLI dependencies into your $GOPATH/bin with the following, after ensuring `$GOPATH/bin`is included in your`$PATH`:

```bash
make cli-dependencies
```

This will ensure that the cobra command is available should you need to add a sub-command.

### Install buf

Grab the latest binary from the [buf releases](https://github.com/bufbuild/buf/releases).

You can now try changing the url in the proto file (such as in `proto/kubeappsapis/core/v1/core.proto`) and then run:

```bash
buf generate
export KUBECONFIG=/home/user/.kube/config # replace it with your desired kube config file
make run
```

and then verify that the RegisteredPlugins RPC call is exposed via HTTP at the new URL path that you specified.

You can also use `buf lint` to ensure that the proto IDLs are valid (ie. extendable, no backwards incompatible changes etc.)
