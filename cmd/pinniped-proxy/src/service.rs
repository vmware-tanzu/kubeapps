use std::convert::Infallible;

use anyhow::Error;
use hyper::{Body, Request, Response, StatusCode};
use log::info;

use crate::logging;
use crate::https;

pub async fn proxy(req: Request<Body>) -> Result<Response<Body>, Infallible> {
    let log_data = logging::request_log_data(&req);

    let _ = match https::get_api_server_url(req.headers()) {
        Ok(u) => u,
        Err(e) => return handle_error(e, log_data),
    };
    // TODO: don't call this if we're using https://kubernetes.local, instead
    // grab the data from the file system.
    let cert_auth_data = match https::get_api_server_cert_auth_data(req.headers()) {
        Ok(c) => c,
        Err(e) => return handle_error(e, log_data),
    };
    let _ = match https::cert_for_cert_data(cert_auth_data) {
        Ok(c) => c,
        Err(e) => return handle_error(e, log_data),
    };
    // TODO: actual proxying to happen here.
    let response = Response::new(Body::from("pinniped-proxy stub\n"));

    info!("{}", logging::response_log_data(&response, log_data));

    Ok(response)
}

/// handle_error converts an error into a 500 message.
fn handle_error(e: Error, log_data: logging::LogData) -> Result<Response<Body>, Infallible> {
    let response = Response::builder()
        .status(StatusCode::BAD_REQUEST)
        .body(Body::from(e.to_string())).unwrap();
    // TODO: Add error to log_data so it can be included (with error context).
    info!("{}", logging::response_log_data(&response, log_data));
    Ok(response)
}