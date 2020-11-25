use std::fmt;

use hyper::{Body, Request, Response};

pub fn request_log_data(req: &Request<Body>) -> LogData {
    LogData {
        target_uri: req.uri().clone(),
        method: req.method().clone(),
        // Some fields are filled in during the response handling.
        // All logging should call response_log_data to change status_code, but if not
        // switch to use Option<StatusCode> instead.
        status_code: hyper::StatusCode::IM_A_TEAPOT,
    }
}

pub fn response_log_data<'s>(resp: &Response<Body>, log_data: &'s mut LogData) -> &'s mut LogData {
    log_data.status_code = resp.status();
    log_data
}

#[derive(Debug,Clone)]
pub struct LogData {
    target_uri: hyper::Uri,
    method: hyper::Method,
    status_code: hyper::StatusCode,
}

impl fmt::Display for LogData {
    fn fmt(&self, f: &mut fmt::Formatter) -> fmt::Result {
        write!(f, "{} {} {}", self.method, self.target_uri, self.status_code)
    }
}
