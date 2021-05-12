# Experimental Kubeapps APIs service

This is an experimental service that we are not yet including in our released Kubeapps chart, providing an extensible protobuf-based API service enabling Kubeapps (or other packaging UIs) to interact with various Kubernetes packaging formats in a consistent way.

## Extensible

The API definitions are using Google protocol buffers - an Interface Definition Language that was created to ensure that API messages and API services can be defined, extended and maintained with the least friction, in a language-neutral way (though we are exclusively using go here).

gRPC is an RPC implementation that uses protocol buffers, providing RPC functionality across languages. Importantly, even though we'll be writing the API service in go, applications using the service can use any language they want and get pre-built clients for that language.

If you are unfamiliar with protocol buffers and gRPC, it's worthwhile to at least read the [introduction](https://grpc.io/docs/what-is-grpc/introduction/) and [core concepts](https://www.grpc.io/docs/what-is-grpc/core-concepts/).

Though it is possible to use a JS/TypeScript client, it is also possible to use a protocol buffer extensions provided by Google that enables optionally exposing the RPCs as http requests, demo'd here. This is made possible via the [grpc-gateway project](https://github.com/grpc-ecosystem/grpc-gateway) which proxies JSON http requests through to the relevant gRPC call. As shown in the demo (and by many others examples online), both http and grpc can be served on the same port without issue.

Together, this enables the best of both worlds: a well known and used Interface Definition Language for defining extensible APIs that can also be exposed via a rest-like http interface.

Finally, we've also chosen to use the [buf](https://buf.build/) tool for generating the code from the proto files. In the past we've used `protoc` (proto buffer compiler) and its extensions directly, but `buf` allows you to specify a simple yaml config instead, and also provides a `lint` command to ensure that your choice of API structure follows best practise, as well as ensuring you're aware when you break backwards compatability.

## CLI

Similar to most go commands, we've used [Cobra](https://github.com/spf13/cobra) for the CLI interface. Currently there is only a root command to run server, but we may later add a `version` subcommand or a `new-plugin` subcommand, but even without these it provides a lot of useful defaults for config, env var support etc.

## Trying it out

You can run `make run` to run the currently stubbed service.

```bash
make run

I0511 11:39:56.444553 4116647 server.go:25] Starting server on :50051
```

You can then verify the (currently stubbed) registered plugins endpoint via http:


```bash
curl http://localhost:50051/core/v1/registered-plugins

{"plugins":["foobar.package.v1"]}
```

or via gRPC (using the [grpcurl tool](https://github.com/fullstorydev/grpcurl)):

```bash
grpcurl -plaintext localhost:50051 kubeappsapis.core.v1.CoreService.RegisteredPlugins
{
  "plugins": [
    "foobar.package.v1"
  ]
}
```


## Hacking

A few extra tools will be needed to contribute to the development of this service.

### Install go cli deps

You should be able to install the exact versions of the various go CLI dependencies into your $GOPATH/bin with the following, after ensuring `$GOPATH/bin` is included in your `$PATH`:

```bash
make cli-dependencies
```

This will ensure that the cobra command is available should you need to add a sub-command.

### Install buf

Grab the latest binary from the [buf releases](https://github.com/bufbuild/buf/releases).

You can now try changing the url in the proto file (such as in `proto/kubeappsapis/core/v1/core.proto`) and then run:

```
buf generate
make run
```

and then verify that the RegisteredPlugins RPC call is exposed via HTTP at the new URL path that you specified.

You can also use `buf lint` to ensure that the proto IDLs are sane (ie. extensible, no backwards incompatible changes etc.)
