// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

use std::env;

fn main() {
    // If the binary is built with the ENV var "PINNIPED_PROXY_VERSION",
    // the value will be available at buildime. Otherwise, it becomes "devel"
    // We use this ENV var to display a custom version when passing the "-- version" flag
    let version = env::var("PINNIPED_PROXY_VERSION").unwrap_or("devel".to_string());
    println!("cargo:rustc-env=PINNIPED_PROXY_VERSION={}", version);
}
