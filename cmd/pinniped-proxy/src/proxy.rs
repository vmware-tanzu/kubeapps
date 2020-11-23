use hyper::Request;

/// proxy_request directs the specified request to the backend service,
/// potentially exchanging credentials on the request for HTTP-header based
/// authorization (not used for client cert authentication).
pub fn proxy_request(mut req: Request<hyper::Body>,
                    k8s_api_server_url: &str,
                    credential_exchange: fn(Request<hyper::Body>) -> Request<hyper::Body>) -> Request<hyper::Body> {
    // TODO: Move to separate re-write URL.
    let uri_string = format!(
        "{}{}",
        k8s_api_server_url,
        req.uri().path_and_query().map(|x| x.as_str()).unwrap_or("")
    );
    let uri = uri_string.parse().unwrap();
    *req.uri_mut() = uri;

    req = credential_exchange(req);

    req
}

#[cfg(test)]
mod tests {
    use super::*;
    use hyper::Body;
    use hyper::header::HeaderValue;

    const API_URL: &str = "https://127.0.0.1:12345";

    /// exchange_credentials_dummy is a dummy function which adds an `X-Dummy-Auth-Check` header to the
    /// request with a static value.
    fn exchange_credentials_dummy(mut req: Request<hyper::Body>) -> Request<hyper::Body> {
        req.headers_mut().insert("X-Dummy-Auth-Check", HeaderValue::from_static("exchanged_credential"));
        req
    }

    #[test]
    fn test_pinniped_proxy_routes_request_to_k8s_api_server() {
        let mut req = Request::get("http://local.pinniped:9876")
            .body(Body::empty())
            .unwrap();

        req = proxy_request(req, API_URL, exchange_credentials_dummy);

        assert_eq!(req.uri().authority().unwrap(), "127.0.0.1:12345");
        assert_eq!(req.uri().scheme().unwrap(), "https");
    }

    #[test]
    fn test_proxy_request_with_credential_exchange() {
        let mut req = Request::get("http://local.pinniped:9876")
            .header("Authorization", "Bearer foo")
            .body(Body::empty())
            .unwrap();

        req = proxy_request(req, API_URL, exchange_credentials_dummy);

        assert_eq!(req.uri().authority().unwrap(), "127.0.0.1:12345");
        assert_eq!(req.uri().scheme().unwrap(), "https");
        assert_eq!(req.headers()["X-Dummy-Auth-Check"], "exchanged_credential");
    }
}
