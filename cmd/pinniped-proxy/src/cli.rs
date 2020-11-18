use structopt::StructOpt;
use url::Url;

const INVALID_SCHEME: &'static str = "invalid scheme, https required";

/// valid_url is a clap-rs validator ensuring the passed url is a valid one.main()
fn valid_url(u: String) -> Result<(), String> {
    let result = Url::parse(&u);
    match result {
        Ok(url) => match url.scheme() {
            "https" => Ok(()),
            _ => Err(INVALID_SCHEME.into()),
        },
        Err(e) => Err(e.to_string()),
    }
}

#[derive(StructOpt)]
/// Converts requests with bearer tokens to requests with short-lived certs.
///
/// pinniped-proxy proxies incoming requests with an `Authorization: Bearer
/// token` header upstream as requests with short-lived client certificates,
/// where the bearer token has been exchanged for the client certs using the
/// pinniped aggregate API.
pub struct Options {
    #[structopt(
        short = "s", 
        long = "api-server-url", 
        default_value = "https://kubernetes.default",
        help = "Specify the Kubernetes API server URL with which the credentials will be exchanged (using the pinniped aggregate API).",
        validator = valid_url
    )]
    pub api_server_url: String,

    #[structopt(
        short = "p",
        long = "port",
        default_value = "3333",
        help = "Specify the port on which pinniped-proxy listens."
    )]
    pub port: u16, 

    #[structopt(
        long = "cacert-data",
        default_value = "",
        help = "Use the specified base64-encoded certificate authorization data to verify the peer."
    )]
    pub cacert_data: String,

    #[structopt(
        long = "pinniped-executable",
        short = "x",
        default_value = "pinniped",
        help = "The name of the executable, including the full path if required",
    )]
    pub pinniped_executable: String,
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_valid_url_success() {
        assert_eq!(valid_url(String::from("https://example.com:8443")), Ok(()))
    }

    #[test]
    fn test_valid_url_failure() {
        assert_eq!(valid_url(String::from("https://example space com")), Err(String::from("invalid domain character")))
    }

    #[test]
    fn test_valid_protocol() {
        assert_eq!(valid_url(String::from("ftp://example.com")), Err(String::from(INVALID_SCHEME)))
    }
}
