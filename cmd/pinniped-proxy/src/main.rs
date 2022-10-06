// Copyright 2020-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

use std::io::Write;
use chrono::Local;
use std::convert::Infallible;
use std::fs;

use anyhow::{Context, Result};
use hyper::{
    server::conn::AddrIncoming,
    service::{make_service_fn, service_fn},
    Server,
};
use log::{info, LevelFilter};
use structopt::StructOpt;
use tls_listener::TlsListener;

// Ensure the root crate is aware of the child modules.
mod cli;
mod https;
mod logging;
mod pinniped;
mod service;
mod tls_config;

#[tokio::main]
async fn main() -> Result<()> {
    pretty_env_logger::formatted_timed_builder()
    .format(|buf, record| {
        writeln!(buf,
            "{} [{}] - {}",
            Local::now().format("%Y-%m-%dT%H:%M:%S%.6f"),
            record.level(),
            record.args()
        )
    })
    .filter(None, LevelFilter::Info)
    .init();
    let opt = cli::Options::from_args();

    // Load the default certificate authority data on startup once.
    let default_ca_data = fs::read_to_string(opt.default_ca_cert.clone())
        .with_context(|| format!("unable to load default-ca-cert at {}", opt.default_ca_cert))?;

    let addr = ([0, 0, 0, 0], opt.port).into();

    let with_tls = opt.proxy_tls_cert != "" && opt.proxy_tls_cert_key != "";
    if !with_tls && (opt.proxy_tls_cert != "" || opt.proxy_tls_cert_key != "") {
        panic!("If configuring TLS support, you must set both --proxy-tls-cert and --proxy-tls-cert-key");
    }

    // Run the server for ever. If it returns with an error, return the
    // result, otherwise, if it completes, we return Ok.
    if with_tls {
        info!(
            "Configuring with TLS cert filepath {} and key filepath {}",
            opt.proxy_tls_cert, opt.proxy_tls_cert_key
        );
        // For every incoming connection, we make a new hyper `Service` to handle
        // all incoming HTTP requests on that connection. This is done by passing a
        // closure to the hyper `make_service_fn` which returns our custom `make_svc`
        // function that can be used for each connection.
        // The closure just returns an async block that runs a service to handle
        // all requests for the connection.
        // `service_fn` is a helper from the hyper crate which converts a
        // function that returns a Response into a `Service`. We pass a closure
        // to service_fn here so we can pass the default certificate authority
        // data.
        let make_svc = make_service_fn(|_conn| {
            let default_ca_data = default_ca_data.clone();
            async {
                Ok::<_, Infallible>(service_fn(move |req| {
                    service::proxy(req, default_ca_data.clone().into_bytes())
                }))
            }
        });

        let incoming = TlsListener::new(
            tls_config::tls_acceptor(opt.proxy_tls_cert, opt.proxy_tls_cert_key)
                .expect("unable to create tls acceptor"),
            AddrIncoming::bind(&addr).expect("unable to bind to address"),
        );
        let server = Server::builder(incoming).serve(make_svc);
        info!("Listening on https://{}", addr);
        server.await.expect("unexpected error while serving");
    } else {
        // The make_svc function needs to be defined in this scope as the
        // compiler uses a different type for the non-tls version.
        let make_svc = make_service_fn(|_conn| {
            let default_ca_data = default_ca_data.clone();
            async {
                Ok::<_, Infallible>(service_fn(move |req| {
                    service::proxy(req, default_ca_data.clone().into_bytes())
                }))
            }
        });
        let server = Server::bind(&addr).serve(make_svc);
        info!("Listening on http://{}", addr);
        server.await.expect("unexpected error while serving");
    }

    Ok(())
}
