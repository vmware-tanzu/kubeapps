# OCI Catalog Developer Guide

> This section is still in development.

<!-- TODO(agamez): piece of docs requiring update. Reason: new section-->

## Prerequisites

- [Git](https://git-scm.com/)
- [Rust programming language](https://www.rust-lang.org/tools/install)

## Running in development

[`cargo`](https://doc.rust-lang.org/cargo/) is the Rust package manager tool is used for most development activities. Assuming you are already in the `cmd/oci-catalog` directory, you can always compile and run the executable with:

```bash
cargo run
```

and pass command-line options to the executable after a double-dash, for example:

```bash
cargo run -- --help
   Compiling oci-catalog v0.1.0 (/mnt/c/repos/kubeapps/cmd/oci-catalog)
    Finished dev [unoptimized + debuginfo] target(s) in 1m 47s
     Running `target/debug/oci-catalog --help`
A service that returns catalog information for an OCI repository.

The OCI Catalog service uses a strategy pattern to enable listing catalog information for different OCI provider registries, until a standard is set for requesting namespaced repositories of a registry in the OCI Distribution specification.

Usage: oci-catalog [OPTIONS]

Options:
  -p, --port <PORT>
          Specify the port on which oci-catalog gRPC service listens.

          [env: OCI_CATALOG_PORT=]
          [default: 50001]

  -h, --help
          Print help (see a summary with '-h')

  -V, --version
          Print version
```
