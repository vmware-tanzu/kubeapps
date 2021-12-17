use std::convert::TryFrom;
use std::env;

use anyhow::{Context, Result};
use http::Uri;
use k8s_openapi::api::core::v1 as corev1;
use k8s_openapi::apimachinery::pkg::apis::meta::v1 as metav1;
use kube::{
    api::{Api, ApiResource, DynamicObject, PostParams},
    core::GroupVersionKind,
    Client, Config,
};
use log::debug;
use native_tls::Identity;
use openssl::{pkcs12::Pkcs12, pkey::PKey, x509::X509};
use serde::{Deserialize, Serialize};
use serde_json;
use thiserror::Error;

const DEFAULT_PINNIPED_API_SUFFIX: &str = "DEFAULT_PINNIPED_API_SUFFIX";
const DEFAULT_PINNIPED_NAMESPACE: &str = "DEFAULT_PINNIPED_NAMESPACE";
const DEFAULT_PINNIPED_AUTHENTICATOR_NAME: &str = "DEFAULT_PINNIPED_AUTHENTICATOR_NAME";
const DEFAULT_PINNIPED_AUTHENTICATOR_TYPE: &str = "DEFAULT_PINNIPED_AUTHENTICATOR_TYPE";

const TOKEN_REQUEST_VERSION: &str = "v1alpha1";
const TOKEN_REQUEST_KIND: &str = "TokenCredentialRequest";

const DEFAULT_API_SUFFIX: &str = "pinniped.dev";

#[derive(Error, Debug)]
pub enum PinnipedError {
    #[error("Unauthorized by pinniped: {0}")]
    UnsuccessfulAuthentication(String),
}

/// TokenCredentialRequest
///
/// Request, Spec and the Status including the returned cluster credential
/// are structs based on the corresponding structs in the pinniped code at:
/// https://github.com/vmware-tanzu/pinniped/blob/main/generated/1.19/apis/concierge/login/v1alpha1/types_token.go#L11
#[derive(Deserialize, Serialize, Clone, Debug)]
pub struct TokenCredentialRequest {
    spec: TokenCredentialRequestSpec,
    status: Option<TokenCredentialRequestStatus>,
}

#[derive(Deserialize, Serialize, Clone, Debug)]
pub struct TokenCredentialRequestSpec {
    // Bearer token supplied with the credential request.
    token: Option<String>,

    // Reference to an authenticator which can verify this credential request.
    authenticator: corev1::TypedLocalObjectReference,
}

#[derive(Deserialize, Serialize, Clone, Debug)]
pub struct TokenCredentialRequestStatus {
    // A ClusterCredential will be returned for a successful credential request.
    credential: Option<ClusterCredential>,

    // An error message will be returned for an unsuccessful credential request.
    message: Option<String>,
}

#[derive(Deserialize, Serialize, Clone, Debug)]
/// ClusterCredential is the cluster-specific credential returned on a successful credential request. It
/// contains either a valid bearer token or a valid TLS certificate and corresponding private key for the cluster.
#[serde(rename_all = "camelCase")]
pub struct ClusterCredential {
    // ExpirationTimestamp indicates a time when the provided credentials expire.
    expiration_timestamp: metav1::Time,

    // Token is a bearer token used by the client for request authentication.
    token: Option<String>,

    // PEM-encoded client TLS certificates (including intermediates, if any).
    client_certificate_data: String,

    // PEM-encoded private key for the above certificate.
    client_key_data: String,
}

/// exchange_token_for_identity accepts an authorization header and returns a client cert authentication Identity in exchange.
///
/// The token is exchanged with pinniped concierge API running on the identified kubernetes api server.
pub async fn exchange_token_for_identity(
    authorization: &str,
    k8s_api_server_url: &str,
    k8s_api_ca_cert_data: &[u8],
) -> Result<Identity> {
    let credential_request =
        call_pinniped_exchange(authorization, k8s_api_server_url, k8s_api_ca_cert_data)
            .await
            .context("Failed to exchange credentials")?;
    match credential_request.status {
        Some(s) => {
            match s.credential {
                Some(c) => return identity_for_exchange(&c),
                None => match s.message {
                    // A returned status without a credential is unsuccessful authentication so
                    // add context to identify this.
                    Some(m) => {
                        return Err(anyhow::anyhow!(m.clone())
                            .context(PinnipedError::UnsuccessfulAuthentication(m)))
                    }
                    None => {
                        return Err(anyhow::anyhow!(
                            "response status neither an error msg or a credential: {:#?}",
                            s
                        ))
                    }
                },
            }
        }
        None => {
            return Err(anyhow::anyhow!(
                "pinniped credential request did not include status: {:#?}",
                credential_request
            ))
        }
    }
}

/// identity_for_exchange parses the JSON output of the credential exchange and returns the Identity.
///
/// Note: to create an identity, need to go via a pkcs12 currently.
/// https://github.com/sfackler/rust-native-tls/issues/27#issuecomment-324262673
fn identity_for_exchange(cred: &ClusterCredential) -> Result<Identity> {
    let pkey = PKey::private_key_from_pem(cred.client_key_data.as_bytes())
        .context("error creating private key from pem")?;
    let x509 = X509::from_pem(cred.client_certificate_data.as_bytes())
        .context("error creating x509 from pem")?;

    let pkcs_cert = Pkcs12::builder()
        .build("", "friendly-name", &pkey, &x509)
        .context("Error building Pkcs12 from private key and x509")?;
    let identity = Identity::from_pkcs12(
        &pkcs_cert
            .to_der()
            .context("error creating der from pkcs12")?,
        "",
    )
    .context("error creating identity from der-formatted pkcs12")?;
    Ok(identity)
}

fn get_client_config(
    k8s_api_server_url: &str,
    k8s_api_ca_cert_data: &[u8],
    pinniped_namespace: String,
) -> Result<kube::Client> {
    let mut config = Config::new(
        k8s_api_server_url
            .parse::<Uri>()
            .context("Failed parsing url for exchange")?,
    );
    config.default_namespace = pinniped_namespace.clone();
    let x509 = X509::from_pem(k8s_api_ca_cert_data).context("error creating x509 from pem")?;
    let der = x509.to_der().context("error creating der from x509")?;
    config.root_cert = Some(vec![der]);

    Ok(Client::try_from(config)?)
}

/// call_pinniped_exchange returns the resulting TokenCredentialRequest with Status after requesting a token credential exchange.
async fn call_pinniped_exchange(
    authorization: &str,
    k8s_api_server_url: &str,
    k8s_api_ca_cert_data: &[u8],
) -> Result<TokenCredentialRequest> {
    // context data
    let pinniped_namespace: String = env::var(DEFAULT_PINNIPED_NAMESPACE)?;
    let pinniped_auth_type: String = env::var(DEFAULT_PINNIPED_AUTHENTICATOR_TYPE)?;
    let pinniped_auth_name: String = env::var(DEFAULT_PINNIPED_AUTHENTICATOR_NAME)?;

    // kube client
    let client = get_client_config(
        k8s_api_server_url,
        k8s_api_ca_cert_data,
        pinniped_namespace.clone(),
    )?;

    // extract token
    let auth_token = match authorization.to_string().strip_prefix("Bearer ") {
        Some(a) => a.to_string(),
        None => authorization.to_string(),
    };

    // define the Api Resource dynamically
    let gvk = GroupVersionKind::gvk(
        &get_pinniped_login_api_group(),
        TOKEN_REQUEST_VERSION,
        TOKEN_REQUEST_KIND,
    );
    let ar = ApiResource::from_gvk(&gvk);

    // create request
    let cred_data = TokenCredentialRequest {
        spec: TokenCredentialRequestSpec {
            token: Some(auth_token),
            authenticator: corev1::TypedLocalObjectReference {
                name: pinniped_auth_name.clone(),
                kind: pinniped_auth_type.clone(),
                api_group: Some(get_pinniped_authenticator_api_group().into()),
            },
        },
        status: None,
    };
    let cred_request = DynamicObject::new("", &ar)
        .within(&pinniped_namespace)
        .data(serde_json::to_value(cred_data)?);
    debug!("{}", serde_json::to_string(&cred_request).unwrap());

    // token credential request invocation
    // we start first as a cluster-based call. if this call fails with a NotFound error, it is
    // an indication we are on an old version which was namespace-based. we thus fallback to
    // a namespace-based call.
    let token_creds_all: Api<DynamicObject> = Api::all_with(client.clone(), &ar);
    match token_creds_all
        .create(&PostParams::default(), &cred_request)
        .await
    {
        Ok(o) => Ok(serde_json::from_value(o.data)?),
        Err(kube::Error::Api(e)) => {
            if e.reason == "NotFound" {
                let token_creds_ns: Api<DynamicObject> =
                    Api::namespaced_with(client.clone(), &pinniped_namespace, &ar);
                match token_creds_ns
                    .create(&PostParams::default(), &cred_request)
                    .await
                {
                    Ok(o) => Ok(serde_json::from_value(o.data)?),
                    Err(e) => Err(anyhow::anyhow!(
                        "err creating token exchange: {:#?}\n{}",
                        serde_json::to_string(&cred_request).unwrap(),
                        e
                    )),
                }
            } else {
                Err(anyhow::anyhow!(
                    "err creating token exchange: {:#?}\n{}",
                    serde_json::to_string(&cred_request).unwrap(),
                    e
                ))
            }
        }
        Err(e) => Err(anyhow::anyhow!(
            "err creating token exchange: {:#?}\n{}",
            serde_json::to_string(&cred_request).unwrap(),
            e
        )),
    }
}

fn get_pinniped_authenticator_api_group() -> String {
    let api_suffix = env::var(DEFAULT_PINNIPED_API_SUFFIX).unwrap_or(DEFAULT_API_SUFFIX.into());
    return format!("{}.{}", "authentication.concierge", &api_suffix).to_string();
}

fn get_pinniped_login_api_group() -> String {
    let api_suffix = env::var(DEFAULT_PINNIPED_API_SUFFIX).unwrap_or(DEFAULT_API_SUFFIX.into());
    return format!("{}.{}", "login.concierge", &api_suffix).to_string();
}

#[macro_use]
#[cfg(test)]
mod tests {
    use super::*;
    use serial_test::serial;

    const VALID_CERT_BASE64: &'static str = "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUN5RENDQWJDZ0F3SUJBZ0lCQURBTkJna3Foa2lHOXcwQkFRc0ZBREFWTVJNd0VRWURWUVFERXdwcmRXSmwKY201bGRHVnpNQjRYRFRJd01UQXlOakl6TXpBME5Wb1hEVE13TVRBeU5ESXpNekEwTlZvd0ZURVRNQkVHQTFVRQpBeE1LYTNWaVpYSnVaWFJsY3pDQ0FTSXdEUVlKS29aSWh2Y05BUUVCQlFBRGdnRVBBRENDQVFvQ2dnRUJBT1ZKCnFuOVBFZUp3UDRQYnI0cFo1ZjZKUmliOFZ5a2tOYjV2K1hzTVZER01aWGZLb293Y29IYjFwRWh5d0pzeDFiME4Kd2YvZ1JURi9maEgzT0drRnNQMlV2a0lHVytzNUlBd0sxMFRXYkN5VzAwT3lzVkdLcnl5bHNWcEhCWXBZRGJBcQpkdnQzc0FkcFJZaGlLZSs2NkVTL3dQNTdLV3g0SVdwZko0UGpyejh2NkJBWlptZ3o5ZzRCSFNMQkhpbTVFbTdYClBJTmpKL1RJTXFzVW1PR1ppUUNHR0ptRnQxZ21jQTd3eHZ0ZXg2ckkxSWdFNkh5NW10UzJ3NDZaMCtlVU1RSzgKSE9UdnI5aGFETnhJenVjbkduaFlCT2Z2U2VVaXNCR0pOUm5QbENydWx4b2NSZGI3N20rQUdzWW52QitNd2prVQpEbXNQTWZBelpSRHEwekhzcGEwQ0F3RUFBYU1qTUNFd0RnWURWUjBQQVFIL0JBUURBZ0trTUE4R0ExVWRFd0VCCi93UUZNQU1CQWY4d0RRWUpLb1pJaHZjTkFRRUxCUUFEZ2dFQkFBWndybXJLa3FVaDJUYld2VHdwSWlOd0o1NzAKaU9lTVl2WWhNakZxTmt6Tk9OUW55c3lPd1laRGJFMDRrV3AxclRLNHVZaUh3NTJUc0cyelJsZ0QzMzNKaEtvUQpIVloyV1hUT3Z5U2RJaWl5bVpKM2N3d0p2T0lhMW5zZnhYY1NJakJnYnNzYXowMndpRCtlazRPdmlRZktjcXJpCnFQbWZabDZDSkk0NU1rd3JwTExFaTZkNVhGbkhDb3d4eklxQjBrUDhwOFlOaGJYWTNYY2JaNElvY2lMemRBamUKQ1l6NXFVSlBlSDJCcHNaM0JXNXRDbjcycGZYazVQUjlYOFRUTHh6aTA4SU9yYjgvRDB4Tnk3emQyMnVjNXM1bwoveXZIeEt6cXBiczVuRXJkT0JFVXNGWnBpUEhaVGc1dExmWlZ4TG00VjNTZzQwRWUyNFd6d09zaDNIOD0KLS0tLS1FTkQgQ0VSVElGSUNBVEUtLS0tLQo=";

    // By default cargo will run rust unit tests in parallel. The serial macro ensures that a specific test
    // (or group of tests) runs serially.
    #[test]
    #[serial(envtest)]
    fn test_call_pinniped_exchange_no_env() -> Result<()> {
        env::remove_var(DEFAULT_PINNIPED_NAMESPACE);
        match tokio_test::block_on(call_pinniped_exchange(
            "authorization",
            "https://example.com",
            VALID_CERT_BASE64.as_bytes(),
        )) {
            Ok(_) => anyhow::bail!("expected error"),
            Err(e) => {
                assert!(
                    e.is::<env::VarError>(),
                    "got: {:#?}, want: {}",
                    e,
                    env::VarError::NotPresent
                );
                Ok(())
            }
        }
    }

    #[test]
    #[serial(envtest)]
    fn test_call_pinniped_exchange_bad_url() -> Result<()> {
        env::set_var(DEFAULT_PINNIPED_NAMESPACE, "pinniped-concierge");
        env::set_var(DEFAULT_PINNIPED_AUTHENTICATOR_TYPE, "JWTAuthenticator");
        env::set_var(DEFAULT_PINNIPED_AUTHENTICATOR_NAME, "oidc-authenticator");
        match tokio_test::block_on(call_pinniped_exchange(
            "authorization",
            "not a url",
            VALID_CERT_BASE64.as_bytes(),
        )) {
            Ok(_) => anyhow::bail!("expected error"),
            Err(e) => {
                assert!(
                    e.is::<http::uri::InvalidUri>(),
                    "got: {:#?}, want: {}",
                    e,
                    "InvalidUri.InvalidUriChar"
                );
                Ok(())
            }
        }
    }

    #[test]
    #[serial(envtest)]
    fn test_call_pinniped_exchange_bad_cert() -> Result<()> {
        env::set_var(DEFAULT_PINNIPED_NAMESPACE, "pinniped-concierge");
        env::set_var(DEFAULT_PINNIPED_AUTHENTICATOR_TYPE, "JWTAuthenticator");
        env::set_var(DEFAULT_PINNIPED_AUTHENTICATOR_NAME, "oidc-authenticator");
        match tokio_test::block_on(call_pinniped_exchange(
            "authorization",
            "https://example.com",
            "not a cert".as_bytes(),
        )) {
            Ok(_) => anyhow::bail!("expected error"),
            Err(e) => {
                assert!(
                    e.is::<openssl::error::ErrorStack>(),
                    "got: {:#?}, want: openssl::error::ErrorStack",
                    e
                );
                Ok(())
            }
        }
    }

    #[test]
    #[serial(envtest)]
    fn test_get_api_group_getters() -> Result<()> {
        env::remove_var("DEFAULT_PINNIPED_API_SUFFIX");
        let authenticator_api_group = get_pinniped_authenticator_api_group();
        assert_eq!(
            authenticator_api_group,
            "authentication.concierge.pinniped.dev"
        );

        let login_api_group = get_pinniped_login_api_group();
        assert_eq!(login_api_group, "login.concierge.pinniped.dev");

        env::set_var(DEFAULT_PINNIPED_API_SUFFIX, "foo.bar");
        let authenticator_api_group = get_pinniped_authenticator_api_group();
        assert_eq!(authenticator_api_group, "authentication.concierge.foo.bar");

        let login_api_group = get_pinniped_login_api_group();
        assert_eq!(login_api_group, "login.concierge.foo.bar");
        Ok(())
    }
}
