// Copyright 2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

use tokio_native_tls::native_tls::{Identity, TlsAcceptor};
use anyhow::Result;
use std::fs;

pub fn tls_acceptor(cert_file: String, key_file: String) -> Result<tokio_native_tls::TlsAcceptor> {
    let identity = Identity::from_pkcs8(
        fs::read_to_string(cert_file).expect("Unable to read cert file as PKCS#8 PEM").as_bytes(),
        fs::read_to_string(key_file).expect("Unable to read key file as PKCS#8 PEM").as_bytes()
    )?;
    Ok(TlsAcceptor::builder(identity).build()?.into())
}
