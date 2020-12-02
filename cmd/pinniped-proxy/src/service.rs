use std::convert::Infallible;

use hyper::{Body, Request, Response};
use log::info;

use crate::logging;
use crate::https;

pub async fn proxy(req: Request<Body>) -> Result<Response<Body>, Infallible> {
    let log_data = logging::request_log_data(&req);

    // TODO: actual proxying to happen here.
    let _ = https::get_api_server_url(req.headers());
    let _ = https::get_api_server_cert_auth_data(req.headers());
    let response = Response::new(Body::from("pinniped-proxy stub\n"));

    info!("{}", logging::response_log_data(&response, log_data));

    Ok(response)
}
