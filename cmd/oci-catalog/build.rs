// Copyright 2023 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

fn main() -> Result<(), Box<dyn std::error::Error>> {
    tonic_build::compile_protos("proto/ocicatalog.proto")?;
    Ok(())
}
