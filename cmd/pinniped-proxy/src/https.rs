use native_tls::{Certificate, TlsConnectorBuilder};
use anyhow::Result;
use hyper::{client::connect::HttpConnector, HeaderMap, header::HeaderValue};
use hyper_tls::HttpsConnector;

use crate::pinniped;

/// decode_cert returns a resulting Certificate for a provided b64 string.
pub fn decode_cert(cert_b64: String) -> Result<Certificate> {
    let cert_data = base64::decode(cert_b64.as_bytes())?;

    let cert = Certificate::from_pem(&cert_data)?;
    Ok(cert)
}

/// include_cert enables a TLS connection to be built using a custom certificate authority cert.
pub fn include_cert_authority(tls_builder: &mut TlsConnectorBuilder, cert: Option<Certificate>) -> &mut TlsConnectorBuilder {
    match cert {
        None => tls_builder,
        Some(c) => tls_builder.add_root_certificate(c),
    }
}

/// include_client_cert updates a tls connection to be built with a client cert for authentication.
pub fn include_client_cert(mut tls_builder: &mut TlsConnectorBuilder, request_headers: HeaderMap<HeaderValue>, pinniped_executable: String) -> Result<&mut TlsConnectorBuilder> {
    if request_headers.contains_key("Authorization") {
        match pinniped::pinniped_exchange_for_identity(request_headers["Authorization"].to_str()?, pinniped_executable) {
            Ok(identity) => {
                tls_builder = tls_builder.identity(identity);
            },
            // TODO update include_*_cert to return Results to handle errors.
            Err(e) => return Err(e),
        };
    }
    Ok(tls_builder)
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
 
#[cfg(test)]
mod tests {
    use super::*;

    const VALID_CERT_BASE64: &'static str = "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUN5RENDQWJDZ0F3SUJBZ0lCQURBTkJna3Foa2lHOXcwQkFRc0ZBREFWTVJNd0VRWURWUVFERXdwcmRXSmwKY201bGRHVnpNQjRYRFRJd01UQXlOakl6TXpBME5Wb1hEVE13TVRBeU5ESXpNekEwTlZvd0ZURVRNQkVHQTFVRQpBeE1LYTNWaVpYSnVaWFJsY3pDQ0FTSXdEUVlKS29aSWh2Y05BUUVCQlFBRGdnRVBBRENDQVFvQ2dnRUJBT1ZKCnFuOVBFZUp3UDRQYnI0cFo1ZjZKUmliOFZ5a2tOYjV2K1hzTVZER01aWGZLb293Y29IYjFwRWh5d0pzeDFiME4Kd2YvZ1JURi9maEgzT0drRnNQMlV2a0lHVytzNUlBd0sxMFRXYkN5VzAwT3lzVkdLcnl5bHNWcEhCWXBZRGJBcQpkdnQzc0FkcFJZaGlLZSs2NkVTL3dQNTdLV3g0SVdwZko0UGpyejh2NkJBWlptZ3o5ZzRCSFNMQkhpbTVFbTdYClBJTmpKL1RJTXFzVW1PR1ppUUNHR0ptRnQxZ21jQTd3eHZ0ZXg2ckkxSWdFNkh5NW10UzJ3NDZaMCtlVU1RSzgKSE9UdnI5aGFETnhJenVjbkduaFlCT2Z2U2VVaXNCR0pOUm5QbENydWx4b2NSZGI3N20rQUdzWW52QitNd2prVQpEbXNQTWZBelpSRHEwekhzcGEwQ0F3RUFBYU1qTUNFd0RnWURWUjBQQVFIL0JBUURBZ0trTUE4R0ExVWRFd0VCCi93UUZNQU1CQWY4d0RRWUpLb1pJaHZjTkFRRUxCUUFEZ2dFQkFBWndybXJLa3FVaDJUYld2VHdwSWlOd0o1NzAKaU9lTVl2WWhNakZxTmt6Tk9OUW55c3lPd1laRGJFMDRrV3AxclRLNHVZaUh3NTJUc0cyelJsZ0QzMzNKaEtvUQpIVloyV1hUT3Z5U2RJaWl5bVpKM2N3d0p2T0lhMW5zZnhYY1NJakJnYnNzYXowMndpRCtlazRPdmlRZktjcXJpCnFQbWZabDZDSkk0NU1rd3JwTExFaTZkNVhGbkhDb3d4eklxQjBrUDhwOFlOaGJYWTNYY2JaNElvY2lMemRBamUKQ1l6NXFVSlBlSDJCcHNaM0JXNXRDbjcycGZYazVQUjlYOFRUTHh6aTA4SU9yYjgvRDB4Tnk3emQyMnVjNXM1bwoveXZIeEt6cXBiczVuRXJkT0JFVXNGWnBpUEhaVGc1dExmWlZ4TG00VjNTZzQwRWUyNFd6d09zaDNIOD0KLS0tLS1FTkQgQ0VSVElGSUNBVEUtLS0tLQo=";

    #[test]
    fn test_decode_cert_success() -> Result<()> {
        decode_cert(String::from(VALID_CERT_BASE64))?;
        Ok(())
    }

    #[test]
    fn test_decode_cert_non_base64() -> Result<()> {
        let result = decode_cert(String::from("not base64 encoded"));
        match result {
            Err(e) => {
                assert!(e.is::<base64::DecodeError>(), "got: {:#?}, want: base64::DecodeError", e);
                Ok(())
            },
            _ => anyhow::bail!("got: valid cert, wanted base64::DecodeError"),
        }
    }

    #[test]
    fn test_decode_cert_invalid_pem_cert() -> Result<()> {
        let result = decode_cert(String::from("bm90IGEgY2VydAo="));
        match result {
            Err(e) => {
                assert!(e.is::<native_tls::Error>(), "got: {:#?}, want: native_tls::Error", e);
                Ok(())
            },
            _ => anyhow::bail!("got: valid cert, wanted native_tls::Error"),
        }
    }
}
