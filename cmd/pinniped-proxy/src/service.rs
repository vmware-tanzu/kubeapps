use std::convert::Infallible;

use anyhow::Error;
use hyper::{Body, Request, Response, StatusCode};
use log::{error, info};
use native_tls::TlsConnector;

use crate::logging;
use crate::https;

pub async fn proxy(mut req: Request<Body>) -> Result<Response<Body>, Infallible> {

    let k8s_api_server_url = match https::get_api_server_url(req.headers()) {
        Ok(u) => u,
        Err(e) => return handle_error(e, logging::request_log_data(&req)),
    };
    req = https::rewrite_request(req, k8s_api_server_url).unwrap();
    let log_data = logging::request_log_data(&req);

    // TODO: don't call this if we're using https://kubernetes.local, instead
    // grab the data from the file system.
    let cert_auth_data = match https::get_api_server_cert_auth_data(req.headers()) {
        Ok(c) => c,
        Err(e) => return handle_error(e, log_data),
    };
    let k8s_api_cert = match https::cert_for_cert_data(cert_auth_data) {
        Ok(c) => c,
        Err(e) => return handle_error(e, log_data),
    };

    // Create an https client with which to proxy the request.
    // We need to construct the TlsConnector for each request so that we can set
    // the client cert. It'd be nice if we could do the construction once and just
    // clone to add the client cert?
    let mut tls_builder = &mut TlsConnector::builder();
    tls_builder = tls_builder.add_root_certificate(k8s_api_cert.clone());

    // TODO: Call credential exchange fn to set the client auth certs.

    let client = match https::make_https_client(tls_builder) {
        Ok(c) => c,
        Err(e) => {
            error!("{:#?}", e);
            let response = Response::builder()
            .status(StatusCode::INTERNAL_SERVER_ERROR)
            .body(Body::from(e.to_string())).unwrap();
            return Ok(response);
        },
    };

    match client.request(req).await {
        Ok(r) => {
            info!("{}", logging::response_log_data(&r, log_data));
            Ok(r)
        },
        Err(e) => {
            error!("{:#?}", e);
            let response = Response::builder()
            .status(StatusCode::INTERNAL_SERVER_ERROR)
            .body(Body::from(e.to_string())).unwrap();
            Ok(response)
        },
    }
}

/// handle_error converts an error into a BAD_REQUEST response.
/// 
/// We may need to expand this to give different responses for different errors.
fn handle_error(e: Error, log_data: logging::LogData) -> Result<Response<Body>, Infallible> {
    let response = Response::builder()
        .status(StatusCode::BAD_REQUEST)
        .body(Body::from(e.to_string())).unwrap();
    // TODO: Add error to log_data so it can be included (with error context).
    info!("{}", logging::response_log_data(&response, log_data));
    Ok(response)
}
