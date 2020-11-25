use native_tls::{Certificate, TlsConnectorBuilder};
use anyhow::Result;
use hyper::{client::connect::HttpConnector, HeaderMap, header::HeaderValue};
use hyper_tls::HttpsConnector;
use url::Url;

use crate::pinniped;

const DEFAULT_K8S_API_SERVER_URL: &str = "https://kubernetes.local";
const HEADER_K8S_API_SERVER_URL: &str = "PINNIPED_K8S_API_SERVER_URL";
const HEADER_K8S_API_SERVER_CA_CERT: &str = "PINNIPED_K8S_API_SERVER_CA_CERT";
const INVALID_SCHEME: &'static str = "invalid scheme, https required";


/// validate_url returns a result containing the url or an error if it is not valid.
fn validate_url(u: &str) -> Result<&str> {
    let result = Url::parse(u);
    match result {
        Ok(url) => match url.scheme() {
            "https" => Ok(u),
            _ => Err(anyhow::anyhow!(INVALID_SCHEME)),
        },
        Err(e) => Err(anyhow::anyhow!(e)),
    }
}

/// include_client_cert updates a tls connection to be built with a client cert for authentication.
pub fn include_client_cert<'a>(mut tls_builder: &'a mut TlsConnectorBuilder, request_headers: HeaderMap<HeaderValue>, k8s_api_server_url: &str, k8s_api_ca_cert_data: &str, pinniped_executable: String) -> Result<&'a mut TlsConnectorBuilder> {
    if request_headers.contains_key("Authorization") {
        match pinniped::pinniped_exchange_for_identity(request_headers["Authorization"].to_str()?, k8s_api_server_url, k8s_api_ca_cert_data, pinniped_executable) {
            Ok(identity) => {
                tls_builder = tls_builder.identity(identity);
            },
            Err(e) => return Err(e),
        };
    }
    Ok(tls_builder)
}

pub fn get_api_server_url(request_headers: &HeaderMap<HeaderValue>) -> Result<&str> {
    match request_headers.get(HEADER_K8S_API_SERVER_URL) {
        Some(hv) => {
            // Header values can contain invalid chars.
            match hv.to_str() {
                Ok(hv) => validate_url(hv),
                Err(e) => Err(anyhow::anyhow!(e)),
            }
        },
        None => Ok(DEFAULT_K8S_API_SERVER_URL),
    }
}

/// make_https_connector returns the tls-configured http connector.
pub fn make_https_connector(tls_builder: &mut TlsConnectorBuilder) -> Result<HttpsConnector<HttpConnector>> {
    let tls = tls_builder.build()?;
    let tokio_tls = tokio_tls::TlsConnector::from(tls);
    let mut http = HttpConnector::new();
    http.enforce_http(false);
    let mut https = HttpsConnector::<HttpConnector>::from((http, tokio_tls));
    https.https_only(true);
    Ok(https)
}

pub fn get_api_server_cert_auth_data(request_headers: &HeaderMap<HeaderValue>) -> Result<Vec<u8>> {
    match request_headers.get(HEADER_K8S_API_SERVER_CA_CERT) {
        Some(header_value_b64) => match base64::decode(header_value_b64.as_bytes()) {
            Ok(data) => Ok(data),
            Err(e) => Err(anyhow::anyhow!(e)),
        },
        None => Err(anyhow::anyhow!("header {} required but not present", HEADER_K8S_API_SERVER_CA_CERT)),
    }
}

pub fn cert_for_cert_data(cert_data: Vec<u8>) -> Result<Certificate> {
    match Certificate::from_pem(&cert_data) {
        Ok(c) => Ok(c),
        Err(e) => Err(anyhow::anyhow!(e)),
    }
}

 
#[cfg(test)]
mod tests {
    use super::*;

    const VALID_CERT_BASE64: &'static str = "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUN5RENDQWJDZ0F3SUJBZ0lCQURBTkJna3Foa2lHOXcwQkFRc0ZBREFWTVJNd0VRWURWUVFERXdwcmRXSmwKY201bGRHVnpNQjRYRFRJd01UQXlOakl6TXpBME5Wb1hEVE13TVRBeU5ESXpNekEwTlZvd0ZURVRNQkVHQTFVRQpBeE1LYTNWaVpYSnVaWFJsY3pDQ0FTSXdEUVlKS29aSWh2Y05BUUVCQlFBRGdnRVBBRENDQVFvQ2dnRUJBT1ZKCnFuOVBFZUp3UDRQYnI0cFo1ZjZKUmliOFZ5a2tOYjV2K1hzTVZER01aWGZLb293Y29IYjFwRWh5d0pzeDFiME4Kd2YvZ1JURi9maEgzT0drRnNQMlV2a0lHVytzNUlBd0sxMFRXYkN5VzAwT3lzVkdLcnl5bHNWcEhCWXBZRGJBcQpkdnQzc0FkcFJZaGlLZSs2NkVTL3dQNTdLV3g0SVdwZko0UGpyejh2NkJBWlptZ3o5ZzRCSFNMQkhpbTVFbTdYClBJTmpKL1RJTXFzVW1PR1ppUUNHR0ptRnQxZ21jQTd3eHZ0ZXg2ckkxSWdFNkh5NW10UzJ3NDZaMCtlVU1RSzgKSE9UdnI5aGFETnhJenVjbkduaFlCT2Z2U2VVaXNCR0pOUm5QbENydWx4b2NSZGI3N20rQUdzWW52QitNd2prVQpEbXNQTWZBelpSRHEwekhzcGEwQ0F3RUFBYU1qTUNFd0RnWURWUjBQQVFIL0JBUURBZ0trTUE4R0ExVWRFd0VCCi93UUZNQU1CQWY4d0RRWUpLb1pJaHZjTkFRRUxCUUFEZ2dFQkFBWndybXJLa3FVaDJUYld2VHdwSWlOd0o1NzAKaU9lTVl2WWhNakZxTmt6Tk9OUW55c3lPd1laRGJFMDRrV3AxclRLNHVZaUh3NTJUc0cyelJsZ0QzMzNKaEtvUQpIVloyV1hUT3Z5U2RJaWl5bVpKM2N3d0p2T0lhMW5zZnhYY1NJakJnYnNzYXowMndpRCtlazRPdmlRZktjcXJpCnFQbWZabDZDSkk0NU1rd3JwTExFaTZkNVhGbkhDb3d4eklxQjBrUDhwOFlOaGJYWTNYY2JaNElvY2lMemRBamUKQ1l6NXFVSlBlSDJCcHNaM0JXNXRDbjcycGZYazVQUjlYOFRUTHh6aTA4SU9yYjgvRDB4Tnk3emQyMnVjNXM1bwoveXZIeEt6cXBiczVuRXJkT0JFVXNGWnBpUEhaVGc1dExmWlZ4TG00VjNTZzQwRWUyNFd6d09zaDNIOD0KLS0tLS1FTkQgQ0VSVElGSUNBVEUtLS0tLQo=";

    // TODO: Requires refactoring back into unit tests after splitting some functionality into two
    // separate functions.
    #[test]
    fn test_get_cert_data_valid() -> Result<()> {
        let mut headers = HeaderMap::new();
        headers.insert(HEADER_K8S_API_SERVER_CA_CERT, HeaderValue::from_static(VALID_CERT_BASE64));

        match get_api_server_cert_auth_data(&headers) {
            Ok(data) => match cert_for_cert_data(data) {
                Ok(_) => Ok(()),
                Err(e) => anyhow::bail!("got {}, want: valid cert", e),
            },
            Err(e) => anyhow::bail!("got {}, want: valid cert", e),
        }
    }

    #[test]
    fn test_decode_cert_non_base64() -> Result<()> {
        let mut headers = HeaderMap::new();
        headers.insert(HEADER_K8S_API_SERVER_CA_CERT, HeaderValue::from_static("not base64 data"));

        match get_api_server_cert_auth_data(&headers) {
            Err(e) => {
                assert!(e.is::<base64::DecodeError>(), "got: {:#?}, want: base64::DecodeError", e);
                Ok(())
            },
            _ => anyhow::bail!("got: valid cert, wanted base64::DecodeError"),
        }
    }

    #[test]
    fn test_decode_cert_invalid_pem_cert() -> Result<()> {
        let mut headers = HeaderMap::new();
        headers.insert(HEADER_K8S_API_SERVER_CA_CERT, HeaderValue::from_static("bm90IGEgY2VydAo="));

        match get_api_server_cert_auth_data(&headers) {
            Ok(data) => match cert_for_cert_data(data) {
                Err(e) => {
                    assert!(e.is::<native_tls::Error>(), "got: {:#?}, want: native_tls::Error", e);
                    Ok(())
                },
                _ => anyhow::bail!("got: valid cert, wanted native_tls::Error"),
            },
            Err(e) => anyhow::bail!("got: {}, wanted decoded value", e),
        }
    }

    #[test]
    fn test_valid_url_success() -> Result<()> {
        let valid_url = "https://example.com:8443";
        match validate_url(valid_url) {
            Ok(u) => {
                assert_eq!(valid_url, u);
                Ok(())
            },
            Err(e) => anyhow::bail!("got: {:#?}, want: {}", e, valid_url)
        }
    }

    #[test]
    fn test_valid_url_failure() -> Result<()> {
        let bad_url = "https://example space com";
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
        let invalid_proto = "ftp://example.com";
        match validate_url(invalid_proto) {
            Ok(u) => anyhow::bail!("got: {}, want: error", u),
            Err(e) => {
                assert_eq!(INVALID_SCHEME, e.to_string(), "got: {:#?}, want: {}", e, INVALID_SCHEME);
                Ok(())
            }
        }
    }
}
