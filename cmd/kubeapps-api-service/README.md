# Proof-of-concept for Kubeapps API service

This is a work in progress, but outlines the tooling involved for providing a pluggable, extensible API service enabling Kubeapps (or other K8s packaging UIs) to support various package formats in a similar way.

## Extensible

The API definitions are using Google protocol buffers - an Interface Definition Language that was created to ensure that API messages and API services can be defined, extended and maintained with the least friction, in a language-neutral way (though we are exclusively using go here).

gRPC is an RPC implementation that uses protocol buffers, providing RPC functionality across languages. Importantly, even though we'll be writing the API service in go, applications using the service can use any language they want and get pre-built clients for that language.

If you are unfamiliar with protocol buffers and gRPC, it's worthwhile to at least read the [introduction](https://grpc.io/docs/what-is-grpc/introduction/) and [core concepts](https://www.grpc.io/docs/what-is-grpc/core-concepts/).

Though it is possible to use a JS/TypeScript client, it is also possible to use a protocol buffer extensions provided by Google that enables optionally exposing the RPCs as http requests, demo'd here. This is made possible via the [grpc-gateway project](https://github.com/grpc-ecosystem/grpc-gateway) which proxies JSON http requests through to the relevant gRPC call. As shown in the demo (and by many others examples online), both http and grpc can be served on the same port without issue.

Together, this enables the best of both worlds: a well known and used Interface Definition Language for defining extensible APIs that can also be exposed via a rest-like http interface.

Finally, I've also chosen to use the [buf](https://buf.build/) tool for generating the code from the proto files. In the past we've used `protoc` (proto buffer compiler) and its extensions directly, but `buf` allows you to specify a simple yaml config instead, and also provides a `lint` command to ensure that your choice of API structure follows best practise, as well as ensuring you're aware when you break backwards compatability.

## Plug-able

After trying a number of methods and weighing those up with the pros and cons outlined in the design doc (I'll make that public next week), I've used the standard [go plugin package](https://golang.org/pkg/plugin/).

Each plugin consists of 3 source files (and some generated files):

* A `.proto` file defining the service that uses the messages defined in our core,
* A `server.go` that implements the actual function(s), and
* A `plugin/main.go` that is used to build the `.so` file for that plugin. This `main.go` has two public functions: one to register the plugin with a GRPC server and one to register the plugin for the http handler.

This means that the actual go command just needs to load all the `.so` files from the specified plugin dirs and register them to start. You can see this in the `kubeapps-api-service/server/server.go` file.

**Note**: I've chosen to demo both simple RPC calls (in the `kubeappsapis/core/v1/core.proto` you'll see the `PluginsAvailable` which returns a simple `PluginsAvailableResponse`) while the `GetAvailablePackages` call in `kubeappsapis/plugins/packagerepositories/helm/v1/helm.proto` returns a streaming response. For more info about these two types of RPC calls, see [core-concepts](https://grpc.io/docs/what-is-grpc/core-concepts/).

## CLI

As most projects do these days, I've used [Cobra](https://github.com/spf13/cobra) for the CLI interface. Currently there is only a root command to run server, but we may later add a `version` subcommand or a `new-plugin` subcommand, but even without these it provides a lot of useful defaults for config, env var support etc.

## Trying it out

You can run `make run` to build the plugins (it'll place the `.so` files in the devel dir) and run the api service and loading the plugins:

```bash
make run

go build -o devel/helm-packagerepositories-v1-plugin.so -buildmode=plugin kubeappsapis/plugins/packagerepositories/helm/v1/plugin/main.go
go build -o devel/helm-packages-v1-plugin.so -buildmode=plugin kubeappsapis/plugins/packages/helm/v1/plugin/main.go
go build -o devel/kapp-controller-packagerepositories-v1-plugin.so -buildmode=plugin kubeappsapis/plugins/packagerepositories/kapp_controller/v1/plugin/main.go

go run main.go --plugin-dir devel/

I0326 22:55:10.413477   57046 server.go:136] Successfully registered plugin "/home/michael/dev/vmware/kubeapps/cmd/kubeapps-api-service/devel/helm-packagerepositories-v1-plugin.so"
I0326 22:55:10.419568   57046 server.go:136] Successfully registered plugin "/home/michael/dev/vmware/kubeapps/cmd/kubeapps-api-service/devel/helm-packages-v1-plugin.so"
I0326 22:55:10.425436   57046 server.go:136] Successfully registered plugin "/home/michael/dev/vmware/kubeapps/cmd/kubeapps-api-service/devel/kapp-controller-packagerepositories-v1-plugin.so"
I0326 22:55:10.425592   57046 server.go:105] Starting server on :50051
```

You can then verify that you can access the endpoints via http:

```bash
curl http://localhost:50051/helm/v1/available-packages

{"result":{"name":"package-a","repository":{"name":"bitnami","namespace":"kubeapps"},"latestVersion":"1.2.0","iconUrl":"http://example.com/package-a.jpg"}}
{"result":{"name":"package-b","repository":{"name":"bitnami","namespace":"kubeapps"},"latestVersion":"1.4.0","iconUrl":"http://example.com/package-b.jpg"}}

curl http://localhost:50051/helm/v1/installed-packages

{"result":{"name":"Apache","namespace":"user1","version":"6.8.0","iconUrl":"http://example.com/apache.jpg"}}
{"result":{"name":"nginx","namespace":"user1","version":"3.4.0","iconUrl":"http://example.com/nginx.jpg"}}
```

Or alternatively, you can access the same endpoint via gRPC (using the [grpcurl tool](https://github.com/fullstorydev/grpcurl))

```bash
grpcurl -vv -plaintext localhost:50051 kubeappsapis.plugins.packagerepositories.helm.v1.PackageRepositoriesService.GetAvailablePackages

Resolved method descriptor:
rpc GetAvailablePackages ( .kubeappsapis.core.packagerepositories.v1.GetAvailablePackagesRequest ) returns ( stream .kubeappsapis.core.packagerepositories.v1.AvailablePackage ) {
  option (.google.api.http) = { get:"/helm/v1/available-packages"  };
}

Request metadata to send:
(empty)

Response headers received:
content-type: application/grpc

Estimated response size: 73 bytes

Response contents:
{
  "name": "package-a",
  "repository": {
    "name": "bitnami",
    "namespace": "kubeapps"
  },
  "latestVersion": "1.2.0",
  "iconUrl": "http://example.com/package-a.jpg"
}

Estimated response size: 73 bytes

Response contents:
{
  "name": "package-b",
  "repository": {
    "name": "bitnami",
    "namespace": "kubeapps"
  },
  "latestVersion": "1.4.0",
  "iconUrl": "http://example.com/package-b.jpg"
}

Response trailers received:
(empty)
Sent 0 requests and received 2 responses
```

### Go version 1.16

Note that the tests that I've written are using a new feature in Go 1.16.x that allows defining a Map file-system for tests, so they won't run without 1.16.

## Hacking

To do some development, a few extra tools will be needed.
### Install go cli deps.

You should be able to install the exact versions of the various go CLI dependencies into your $GOPATH/bin with:

```bash
make cli-dependencies
```

Ensure $GOPATH/bin is on path.

### Install buf

Grab the latest binary from the [buf releases](https://github.com/bufbuild/buf/releases).

You can now try changing the url in the proto file (such as in `kubeappsapis/plugins/packages/helm/v1/helm.proto`) and then run:

```
buf generate
make run
```

and then verify that the GetInstalledPackages RPC call is exposed the new URL path that you specified.

## TODO

* Re-org so generated files are all in a separate directory.
* Enforce required methods for each api.
* Add required authz (will need for actual calls anyway).
* Actually serve generated swagger files
* Other cmds to demo
