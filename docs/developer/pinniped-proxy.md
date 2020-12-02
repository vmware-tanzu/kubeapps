# Kubeapps Pinniped-Proxy Developer Guide

`pinniped-proxy` proxies incoming requests with an `Authorization: Bearer
token` header, exchanging the token via the pinniped aggregate API for x509
short-lived client certificates, before forwarding the request onwards
to the destination k8s API server.

`pinniped-proxy` can be used by our Kubeapps frontend to ensure OIDC requests for the Kubernetes API server are forwarded through only after exchanging the OIDC id token for client certificates used by the Kubernetes API server, for situations where the Kubernetes API server is not configured for OIDC.

You can read more in the [investigation and POC design document for `pinniped-proxy`](https://docs.google.com/document/d/1Sqhq_JIfb7M3K5RloV4T2itznu56EDJm_PEz2yrPk1E/edit#).

## Prerequisites

- [Git](https://git-scm.com/)
- [Rust programming language](https://www.rust-lang.org/tools/install)
- (more to come)

## Running in development

[`cargo`](https://doc.rust-lang.org/cargo/) is the Rust package manager tool is used for most development activities. Assuming you are already in the `cmd/pinniped-proxy` directory, you can always compile and run the executable with:

```bash
cargo run
```

and pass command-line options to the executable after a double-dash, for example:

```bash
cargo run -- -h
    Finished dev [unoptimized + debuginfo] target(s) in 0.05s
     Running `target/debug/pinniped-proxy -h`
pinniped-proxy 0.1.0
A proxy server which converts k8s API server requests with bearer tokens to requests with short-lived X509 certs
exchanged by pinniped.

pinniped-proxy proxies incoming requests with an `Authorization: Bearer token` header, exchanging the token via the
pinniped aggregate API for x509 short-lived client certificates, before forwarding the request onwards to the
destination k8s API server.

USAGE:
    pinniped-proxy [OPTIONS]

FLAGS:
    -h, --help       Prints help information
    -V, --version    Prints version information

OPTIONS:
    -p, --port <port>    Specify the port on which pinniped-proxy listens. [default: 3333]
```

## Running tests

Similarly, tests can be run with the cargo tool:

```bash
cargo test
```
