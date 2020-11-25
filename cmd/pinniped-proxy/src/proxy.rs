use hyper::{header::HeaderValue, Request};

/// rewrite_request directs the specified request to the backend service,
/// potentially exchanging credentials on the request for HTTP-header based
/// authorization (not used for client cert authentication).
pub fn rewrite_request(mut req: Request<hyper::Body>, k8s_api_server_url: &str) -> Request<hyper::Body> {
    // Update the request URI and the Host header.
    let uri_string = format!(
        "{}{}",
        k8s_api_server_url,
        req.uri().path_and_query().map(|x| x.as_str()).unwrap_or("")
    );
    let uri: hyper::Uri = uri_string.parse().unwrap();
    let host: Option<HeaderValue> = match uri.authority() {
        Some(auth) => {
            let mut host_header = auth.host().to_string();
            host_header = match auth.port() {
                Some(p) => format!("{}:{}", host_header, p),
                None => host_header,
            };
            match HeaderValue::from_str(&host_header) {
                Ok(hv) => Some(hv),
                Err(_) => None,
            }
        },
        None => None,
    };
    *req.uri_mut() = uri;

    // Update the host (TODO: also set Forwarded)
    match host {
        Some(h) => req.headers_mut().insert("Host", h),
        None => None,
    };

    req
}

#[cfg(test)]
mod tests {
    use super::*;
    use hyper::Body;

    const API_URL: &str = "https://127.0.0.1:12345";

    #[test]
    fn test_pinniped_proxy_routes_request_to_k8s_api_server() {
        let mut req = Request::get("http://local.pinniped:9876")
            .body(Body::empty())
            .unwrap();

        req = rewrite_request(req, API_URL);

        assert_eq!(req.uri().authority().unwrap(), "127.0.0.1:12345");
        assert_eq!(req.uri().scheme().unwrap(), "https");
    }
}
