# Experimental Kubeapps APIs service

This is an experimental service that we are not yet including in our released Kubeapps chart, providing an extensible protobuf-based API service enabling Kubeapps (or other packaging UIs) to interact with various Kubernetes packaging formats in a consistent way.

## CLI

Similar to most go commands, we've used [Cobra](https://github.com/spf13/cobra) for the CLI interface. Currently there is only a root command to run server, but we may later add a `version` subcommand or a `new-plugin` subcommand, but even without these it provides a lot of useful defaults for config, env var support etc.

## Trying it out

You can run `make run` to run the currently stubbed service.

```bash
make run

I0511 11:39:56.444553 4116647 server.go:25] Starting server on :50051
```

You can then verify the (currently echo) service via http:


```bash
curl http://localhost:50051/helm/v1/available-packages

Hello, "/helm/v1/available-packages"
```

## Hacking

A few exra tools will be needed to contribute to the development of this service.

### Install go cli deps

You should be able to install the exact versions of the various go CLI dependencies into your $GOPATH/bin with the following, after ensuring `$GOPATH/bin` is included in your `$PATH`:

```bash
make cli-dependencies
```

This will ensure that the cobra command is available should you need to add a sub-command.
