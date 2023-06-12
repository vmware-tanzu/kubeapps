// Copyright 2023 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

use clap;

#[derive(clap::Parser, Debug)]
#[command(version = env!("OCI_CATALOG_VERSION"))]
/// A service that returns catalog information for an OCI repository.
///
/// The OCI Catalog service uses a strategy pattern to enable listing catalog
/// information for different OCI provider registries, until a standard is set
/// for requesting namespaced repositories of a registry in the OCI Distribution
/// specification.
pub struct Options {
    #[arg(
        short = 'p',
        long = "port",
        env = "OCI_CATALOG_PORT",
        default_value = "50001",
        help = "Specify the port on which oci-catalog gRPC service listens."
    )]
    pub port: u16,
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_cli() {
        use clap::CommandFactory;
        Options::command().debug_assert();
    }
}
