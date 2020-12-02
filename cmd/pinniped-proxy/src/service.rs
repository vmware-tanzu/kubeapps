use std::convert::Infallible;

use hyper::{Body, Request, Response};

pub async fn proxy(_: Request<Body>) -> Result<Response<Body>, Infallible> {
    Ok(Response::new(Body::from("pinniped-proxy stub\n")))
}
