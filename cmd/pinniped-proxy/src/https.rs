use anyhow::Result;
use hyper::{HeaderMap, header::HeaderValue};
use url::Url;

const DEFAULT_K8S_API_SERVER_URL: &str = "https://kubernetes.local";
const HEADER_K8S_API_SERVER_URL: &str = "PINNIPED_PROXY_API_SERVER_URL";
const INVALID_SCHEME_ERROR: &'static str = "invalid scheme, https required";

/// validate_url returns a result containing the validated url or an error if it is invalid.
fn validate_url(u: String) -> Result<String> {
    let result = Url::parse(&u);
    match result {
        Ok(url) => match url.scheme() {
            "https" => Ok(u),
            _ => Err(anyhow::anyhow!(INVALID_SCHEME_ERROR)),
        },
        Err(e) => Err(anyhow::anyhow!(e)),
    }
}

/// get_api_server_url returns a string result from the specified header.
///
/// If none is specified we default to the in-cluster K8S API server URL.
pub fn get_api_server_url(request_headers: &HeaderMap<HeaderValue>) -> Result<String> {
    match request_headers.get(HEADER_K8S_API_SERVER_URL) {
        Some(hv) => {
            // Header values can contain invalid chars.
            match hv.to_str() {
                Ok(hv) => validate_url(hv.to_string()),
                Err(e) => Err(anyhow::anyhow!(e)),
            }
        },
        None => Ok(DEFAULT_K8S_API_SERVER_URL.to_string()),
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    const VALID_API_SERVER_URL: &str = "https://172.1.18.4";

    #[test]
    fn test_valid_url_success() -> Result<()> {
        let valid_url = "https://example.com:8443".to_string();
        match validate_url(valid_url.clone()) {
            Ok(u) => {
                assert_eq!(valid_url, u);
                Ok(())
            },
            Err(e) => anyhow::bail!("got: {:#?}, want: {}", e, valid_url)
        }
    }

    #[test]
    fn test_valid_url_failure() -> Result<()> {
        let bad_url = "https://example space com".to_string();
        match validate_url(bad_url) {
            Ok(u) => anyhow::bail!("got: {}, want: error", u),
            Err(e) => {
                assert!(e.is::<url::ParseError>(), "got: {:#?}, want: {}", e, url::ParseError::InvalidDomainCharacter);
                Ok(())
            }
        }
    }

    #[test]
    fn test_invalid_protocol() -> Result<()>{
        let invalid_proto = "ftp://example.com".to_string();
        match validate_url(invalid_proto) {
            Ok(u) => anyhow::bail!("got: {}, want: error", u),
            Err(e) => {
                assert_eq!(INVALID_SCHEME_ERROR, e.to_string(), "got: {:#?}, want: {}", e, INVALID_SCHEME_ERROR);
                Ok(())
            }
        }
    }

    #[test]
    fn get_api_server_url_success() -> Result<()> {
        let mut headers = HeaderMap::new();
        headers.insert(HEADER_K8S_API_SERVER_URL, HeaderValue::from_static(VALID_API_SERVER_URL));

        assert_eq!(get_api_server_url(&headers)?, VALID_API_SERVER_URL.to_string());
        Ok(())
    }

    #[test]
    fn get_api_server_url_invalid() -> Result<()> {
        let mut headers = HeaderMap::new();
        headers.insert(HEADER_K8S_API_SERVER_URL, HeaderValue::from_static("not a url"));

        let want = url::ParseError::InvalidDomainCharacter;
        match get_api_server_url(&headers) {
            Ok(got) => anyhow::bail!("got: {}, want: {}", got, want),
            Err(got) => {
                assert!(got.is::<url::ParseError>(), "got: {:#?}, want: {}", got, want);
                Ok(())
            },
        }
    }

    #[test]
    fn get_api_server_url_wrong_scheme() -> Result<()> {
        let mut headers = HeaderMap::new();
        headers.insert(HEADER_K8S_API_SERVER_URL, HeaderValue::from_static("http://172.1.2.18"));

        match get_api_server_url(&headers) {
            Ok(got) => anyhow::bail!("got: {}, want: Err({})", got, INVALID_SCHEME_ERROR),
            Err(got) => {
                assert_eq!(got.to_string(), INVALID_SCHEME_ERROR, "got: {:#?}, want: Err({})", got, INVALID_SCHEME_ERROR);
                Ok(())
            },
        }
    }

    #[test]
    fn get_api_server_url_default() -> Result<()> {
        let headers = HeaderMap::new();

        assert_eq!(get_api_server_url(&headers)?, DEFAULT_K8S_API_SERVER_URL.to_string());
        Ok(())
    }
}