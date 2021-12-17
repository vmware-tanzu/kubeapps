use std::convert::Infallible;
use std::fs;

use anyhow::{Context, Result};
use hyper::{
    service::{make_service_fn, service_fn},
    Server,
};
use log::info;
use structopt::StructOpt;

// Ensure the root crate is aware of the child modules.
mod cli;
mod https;
mod logging;
mod pinniped;
mod service;
#[tokio::main]
async fn main() -> Result<()> {
    pretty_env_logger::init();
    let opt = cli::Options::from_args();

    // Load the default certificate authority data on startup once.
    let default_ca_data = fs::read_to_string(opt.default_ca_cert.clone())
        .with_context(|| format!("unable to load default-ca-cert at {}", opt.default_ca_cert))?;

    // For every incoming connection, we make a new hyper `Service` to handle
    // all incoming HTTP requests on that connection. This is done by passing a
    // closure to the hyper `make_service_fn` which returns our custom `make_svc`
    // function that can be used for each connection.
    let make_svc = make_service_fn(|_conn| {
        let default_ca_data = default_ca_data.clone();
        // The closure just returns an async block that runs a service to handle
        // all requests for the connection.
        async {
            // `service_fn` is a helper from the hyper crate which converts a
            // function that returns a Response into a `Service`. We pass a
            // closure to service_fn here so we can pass the default certificate
            // authority data.
            Ok::<_, Infallible>(service_fn(move |req| {
                service::proxy(req, default_ca_data.clone().into_bytes())
            }))
        }
    });

    let addr = ([0, 0, 0, 0], opt.port).into();

    let server = Server::bind(&addr).serve(make_svc);

    info!("Listening on http://{}", addr);

    // Run the server for ever. If it returns with an error, return the
    // result, otherwise, if it completes, we return Ok.
    server.await?;

    Ok(())
}
