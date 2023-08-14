// Copyright 2023 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

use std::env;

fn main() -> Result<(), Box<dyn std::error::Error>> {
    tonic_build::compile_protos("proto/ocicatalog/v1alpha1/ocicatalog.proto")?;

    // If the binary is built with the ENV var "OCI_CATALOG_VERSION",
    // the value will be available at buildime. Otherwise, it becomes "devel"
    // We use this ENV var to display a custom version when passing the "-- version"
    // flag
    let version = env::var("OCI_CATALOG_VERSION").unwrap_or("devel".to_string());
    println!("cargo:rustc-env=OCI_CATALOG_VERSION={}", version);

    Ok(())
}
