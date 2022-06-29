// Copyright 2020-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

use structopt::StructOpt;

#[derive(StructOpt, Debug)]
#[structopt(version = env!("PINNIPED_PROXY_VERSION"))]
/// A proxy server which converts k8s API server requests with bearer tokens to
/// requests with short-lived X509 certs exchanged by pinniped.
///
/// pinniped-proxy proxies incoming requests with an `Authorization: Bearer
/// token` header, exchanging the token via the pinniped aggregate API for x509
/// short-lived client certificates, before forwarding the request onwards
/// to the destination k8s API server.
pub struct Options {
    #[structopt(
        short = "p",
        long = "port",
        env = "PINNIPED_PROXY_PORT",
        default_value = "3333",
        help = "Specify the port on which pinniped-proxy listens."
    )]
    pub port: u16,
    #[structopt(
        long = "default-ca-cert",
        env = "PINNIPED_PROXY_DEFAULT_CA_CERT",
        default_value = "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt",
        help = "Specify the file path to the cert authority for the default api server https://kubernetes.default"
    )]
    pub default_ca_cert: String,
    #[structopt(
        long = "proxy-tls-cert",
        env = "PINNIPED_PROXY_TLS_CERT",
        default_value = "",
        help = "Specify the file path to a PEM encoded TLS certificate. Providing the cert and key implies listening for TLS requests."
    )]
    pub proxy_tls_cert: String,
    #[structopt(
        long = "proxy-tls-cert-key",
        env = "PINNIPED_PROXY_TLS_CERT_KEY",
        default_value = "",
        help = "Specify the file path to a PEM encoded TLS certificate key. Providing the cert and key implies listening for TLS requests."
    )]
    pub proxy_tls_cert_key: String,
}
