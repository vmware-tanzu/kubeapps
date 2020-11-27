use std::convert::Infallible;
use std::future::Future;
use std::pin::Pin;

use anyhow::{Error, Result};
use hyper::service::{make_service_fn, service_fn};
use hyper::{Body, Client, Response, Server, StatusCode};
use log::{error, info};
use native_tls::TlsConnector;
use structopt::StructOpt;

// Ensure the root crate is aware of the child modules.
mod cli;
mod https;
mod pinniped;
mod logging;
mod proxy;

#[tokio::main]
async fn main() -> Result<()> {
    pretty_env_logger::init();
    let opt = cli::Options::from_args();

    let pinniped_executable = opt.pinniped_executable;

    let make_svc = make_service_fn(|_conn| {
        // These strings are shadowed inside the closure so that they are captured for reference
        // (without being moved/owned) within the async closure below (as generators/async blocks are stackless).
        let pinniped_executable = pinniped_executable.clone();
        async {
            // Normally the return type of the service_fn closure isn't required, but I was unable to find
            // a way to handle errors separately without explicitly defining the return type. If you
            // instead uncomment the following line to replace the existing one, the compiler complains that
            // a different generator returns for different control flows.
            // Ok::<_, Infallible>(service_fn(move |mut req| {                    
            Ok::<_, Infallible>(service_fn(move |mut req| -> Pin<Box<dyn Future<Output = std::result::Result<Response<Body>, hyper::Error>> + std::marker::Send >> {                    
                let headers = req.headers().clone();
                let k8s_api_server_url = match https::get_api_server_url(&headers) {
                    Ok(u) => u,
                    Err(e) => return handle_error(e, logging::request_log_data(&req)),
                };

                req = proxy::rewrite_request(req, k8s_api_server_url);
                let log_data = logging::request_log_data(&req);

                let k8s_api_cert_auth_data = match https::get_api_server_cert_auth_data(&headers) {
                    Ok(c) => c,
                    Err(e) => {
                        error!("{:#?}", e);
                        return handle_error(e, log_data);
                    },
                };
                let k8s_api_cert = match https::cert_for_cert_data(k8s_api_cert_auth_data.clone()) {
                    Ok(c) => c,
                    Err(e) => {
                        error!("{:#?}", e);
                        return handle_error(e, log_data);
                    },
                };

                // We need to construct the TlsConnector for each request so that we can set
                // the client cert. It'd be nice if we could do the construction once and just
                // clone to add the client cert?
                let mut tls_builder = &mut TlsConnector::builder();
                tls_builder = tls_builder.add_root_certificate(k8s_api_cert.clone());
                
                // The cert data is converted to a &str without checking the error since we already
                // know from above that it's a valid cert.
                tls_builder = match https::include_client_cert(tls_builder, req.headers().clone(), k8s_api_server_url, std::str::from_utf8(&k8s_api_cert_auth_data).unwrap(), pinniped_executable.clone()) {
                    Ok(b) => b,
                    Err(e) => {
                        error!("{:#?}", e);
                        return handle_error(e, log_data);
                    },
                };

                let conn = match https::make_https_connector(tls_builder) {
                    Ok(c) => c,
                    Err(e) => {
                        error!("{:#?}", e);
                        return handle_error(e, log_data);
                    },
                };

                let client = Client::builder().build::<_, Body>(conn);
                Box::pin(async move {
                    let mut log_data = & mut log_data.clone();
                    match client.request(req).await {
                        Ok(r) => {
                            // access log.
                            log_data = logging::response_log_data(&r, log_data);
                            info!("{}", log_data);
                            Ok(r)
                        },
                        Err(e) => {
                            error!("{:#?}", e);
                            let mut response = Response::new(Body::from(e.to_string()));
                            *response.status_mut() = StatusCode::INTERNAL_SERVER_ERROR;
                            Ok(response)
                        },
                    }
                })
            }))
        }
    });

    let addr = ([127, 0, 0, 1], opt.port).into();
    let server = Server::bind(&addr).serve(make_svc);

    info!("Listening on http://{}", addr);

    if let Err(e) = server.await {
        error!("server error: {}", e);
    }

    Ok(())
}

/// handle_error converts an error into a 500 message.
fn handle_error(e: Error, log_data: logging::LogData) -> Pin<Box<dyn Future<Output = std::result::Result<Response<Body>, hyper::Error>> + std::marker::Send>> {
    Box::pin(async move {
        let response = Response::builder()
            .status(StatusCode::INTERNAL_SERVER_ERROR)
            .body(Body::from(e.to_string())).unwrap();
        info!("{}", logging::response_log_data(&response, &mut log_data.clone()));
        Ok(response)
    })
}
