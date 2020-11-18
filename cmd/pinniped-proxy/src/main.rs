use std::convert::Infallible;
use std::future::Future;
use std::pin::Pin;

use anyhow::{Context, Error, Result};
use hyper::service::{make_service_fn, service_fn};
use hyper::{Body, Client, Response, Server, StatusCode};
use log::{error, info};
use native_tls::TlsConnector;
use structopt::StructOpt;

// Ensure the root crate is aware of the child modules.
mod cli;
mod https;
mod pinniped;
mod proxy;

#[tokio::main]
async fn main() -> Result<()> {
    pretty_env_logger::init();
    let opt = cli::Options::from_args();

    let out_addr = opt.api_server_url;
    
    // Use an Option for the cert authority data as it may or may not be present.
    let cert = match opt.cacert_data.len() {
        0 => None,
        _ => Some(https::decode_cert(opt.cacert_data).context("Failed to decode the provided cacert-data.")?),
    };

    let pinniped_executable = opt.pinniped_executable;

    let make_svc = make_service_fn(|_conn| {
        // These strings are shadowed inside the closure so that they are captured for reference
        // (without being moved/owned) within the async closure below (as generators/async blocks are stackless).
        let out_addr = out_addr.clone();
        let pinniped_executable = pinniped_executable.clone();
        let cert = cert.clone();
        async {
            // Normally the return type of the service_fn closure isn't required, but I was unable to find
            // a way to handle errors separately without explicitly defining the return type. If you
            // instead uncomment the following line to replace the existing one, the compiler complains that
            // a different generator returns for different control flows.
            // Ok::<_, Infallible>(service_fn(move |mut req| {                    
            Ok::<_, Infallible>(service_fn(move |mut req| -> Pin<Box<dyn Future<Output = std::result::Result<Response<Body>, hyper::Error>> + std::marker::Send >> {                    
                info!("processing request to {}", *req.uri());
                // We need to construct the TlsConnector for each request so that we can set
                // the client cert. It'd be nice if we could do the construction once and just
                // clone to add the client cert?
                let mut tls_builder = &mut TlsConnector::builder();
                tls_builder = https::include_cert_authority(tls_builder, cert.clone());
                
                tls_builder = match https::include_client_cert(tls_builder, req.headers().clone(), pinniped_executable.clone()) {
                    Ok(b) => b,
                    // Try returning an http error response here. May then mean we don't need verbosity above?
                    Err(e) => {
                        error!("{:#?}", e);
                        return handle_error(e);
                    },
                };

                let conn = match https::make_https_connector(tls_builder) {
                    Ok(c) => c,
                    Err(e) => {
                        error!("{:#?}", e);
                        return handle_error(e);
                    },
                };

                // Currently using client cert auth above (see include_client_cert), rather than an exchange
                // of header credentials (which will be needed if we switch to pinniped supervisor with OIDC),
                // so we use a no-op for the proxy_request credential_exchange fn.
                req = proxy::proxy_request(req, out_addr.clone(), |req| req);

                info!("forwarding request with exchanged credentials to {}", *req.uri());
                
                let client = Client::builder().build::<_, Body>(conn);
                Box::pin(client.request(req))
            }))
        }
    });

    let addr = ([127, 0, 0, 1], opt.port).into();
    let server = Server::bind(&addr).serve(make_svc);

    info!("Listening on http://{} and proxying to k8s api server at {}", addr, out_addr);

    if let Err(e) = server.await {
        error!("server error: {}", e);
    }

    Ok(())
}

/// handle_error converts an error into a 500 message.
fn handle_error(e: Error) -> Pin<Box<dyn Future<Output = std::result::Result<Response<Body>, hyper::Error>> + std::marker::Send>> {
    Box::pin(async move {
        Ok(Response::builder()
            .status(StatusCode::INTERNAL_SERVER_ERROR)
            .body(Body::from(e.to_string())).unwrap())
    })
}
