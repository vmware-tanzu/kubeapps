// Copyright 2020-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

use std::cmp::PartialEq;
use std::fmt;

use hyper::{Body, Request, Response};

#[derive(Clone, Debug, Default, PartialEq)]
/// LogData encapsulates all the data we want to collect during a proxied request.
///
/// This includes data from the initial request as well as the returned response.
pub struct LogData {
    target_uri: hyper::Uri,
    method: hyper::Method,
    // status_code is optional since we don't know it's value until the upstream
    // response is returned.
    status_code: Option<hyper::StatusCode>,
}

/// We implement the Display trait for our LogData struct so that it
/// knows how to display itself as a string when requested.
impl fmt::Display for LogData {
    fn fmt(&self, f: &mut fmt::Formatter) -> fmt::Result {
        // If we have a status, use the string version of that,
        // otherwise we'll just display a hyphen.
        let status: String = match self.status_code {
            Some(sc) => sc.to_string(),
            None => "-".into(),
        };
        write!(f, "{} {} {}", self.method, self.target_uri, status)
    }
}

/// request_log_data returns a new LogData with the relevant data from the request.
pub fn request_log_data(req: &Request<Body>) -> LogData {
    LogData {
        target_uri: req.uri().clone(),
        method: req.method().clone(),
        status_code: None,
    }
}

/// response_log_data returns a new LogData including the relevant data from the response.
pub fn response_log_data(resp: &Response<Body>, log_data: LogData) -> LogData {
    let mut updated_log_data = log_data.clone();
    updated_log_data.status_code = Some(resp.status());
    updated_log_data
}

#[cfg(test)]
mod tests {
    use super::*;
    use hyper::{Method, Request, StatusCode, Uri};

    const URI: &str = "https://example.com/foo/bar?baz";

    #[test]
    fn display_including_status() {
        let log_data = LogData {
            target_uri: URI.parse::<Uri>().unwrap(),
            method: Method::PUT,
            status_code: Some(StatusCode::OK),
        };

        assert_eq!(format!("{}", log_data), format!("PUT {} 200 OK", URI));
    }

    #[test]
    fn display_without_status() {
        let log_data = LogData {
            target_uri: URI.parse::<Uri>().unwrap(),
            method: Method::PUT,
            status_code: None,
        };

        assert_eq!(format!("{}", log_data), format!("PUT {} -", URI));
    }

    #[test]
    fn populate_from_request() {
        let request = Request::builder()
            .uri(URI)
            .method(Method::GET)
            .body(Body::empty())
            .unwrap();

        let log_data = request_log_data(&request);

        assert_eq!(
            log_data,
            LogData {
                target_uri: URI.parse::<Uri>().unwrap(),
                method: Method::GET,
                status_code: None,
            }
        )
    }

    #[test]
    fn populate_from_response() {
        let response = Response::builder()
            .status(StatusCode::NOT_FOUND)
            .body(Body::empty())
            .unwrap();

        let log_data = response_log_data(&response, Default::default());

        assert_eq!(
            log_data,
            LogData {
                status_code: Some(StatusCode::NOT_FOUND),
                ..Default::default()
            }
        )
    }
}
