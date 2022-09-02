// Copyright 2020-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0
use std::convert::TryFrom;
use std::env;
use std::hash::{Hash, Hasher};

use anyhow::{Context, Result};
use cached::{
    proc_macro::cached,
    stores::{CanExpire, ExpiringValueCache},
};
use chrono::Utc;
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
use openssl::x509::X509;
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
#[derive(Deserialize, Serialize, Clone, Debug, PartialEq, Eq)]
pub struct TokenCredentialRequest {
    spec: TokenCredentialRequestSpec,
    status: Option<TokenCredentialRequestStatus>,
}

// We specifically do not include the request status in the hash generation,
// considering only the request spec for our caching.
impl Hash for TokenCredentialRequest {
    fn hash<H: Hasher>(&self, state: &mut H) {
        self.spec.hash(state);
    }
}

// CanExpire is implemented for the request so that the cache can determine
// if the request has expired.
impl CanExpire for TokenCredentialRequest {
    fn is_expired(&self) -> bool {
        match self.status.clone() {
            Some(s) => match s.credential {
                Some(c) => c.expiration_timestamp < metav1::Time(Utc::now()),
                None => true,
            },
            None => true,
        }
    }
}

#[derive(Deserialize, Serialize, Clone, Debug)]
pub struct TokenCredentialRequestSpec {
    // Bearer token supplied with the credential request.
    token: Option<String>,

    // Reference to an authenticator which can verify this credential request.
    authenticator: corev1::TypedLocalObjectReference,
}

// We need to implement equality explicitly as it's not implemented for
// TypedLocalObjectReference.
impl PartialEq for TokenCredentialRequestSpec {
    fn eq(&self, other: &Self) -> bool {
        self.token == other.token && self.authenticator.name == other.authenticator.name
    }
}
impl Eq for TokenCredentialRequestSpec {}

// Since the TypedLocalObjectReference doesn't implement the Hash trait
// we cannot simple derive the Hash trait, instead calculating it manually
// for the token and authenticator fields.
impl Hash for TokenCredentialRequestSpec {
    fn hash<H: Hasher>(&self, state: &mut H) {
        self.token.hash(state);
        self.authenticator.api_group.hash(state);
        self.authenticator.kind.hash(state);
        self.authenticator.name.hash(state);
    }
}

#[derive(Deserialize, Serialize, Clone, Debug, PartialEq, Eq)]
pub struct TokenCredentialRequestStatus {
    // A ClusterCredential will be returned for a successful credential request.
    credential: Option<ClusterCredential>,

    // An error message will be returned for an unsuccessful credential request.
    message: Option<String>,
}

#[derive(Deserialize, Serialize, Clone, Debug, PartialEq, Eq)]
/// ClusterCredential is the cluster-specific credential returned on a
/// successful credential request. It contains either a valid bearer token or a
/// valid TLS certificate and corresponding private key for the cluster.
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

/// exchange_token_for_identity accepts an authorization header and returns a
/// client cert authentication Identity in exchange.
///
/// The token is exchanged with pinniped concierge API running on the identified
/// kubernetes api server.
pub async fn exchange_token_for_identity(
    authorization: &str,
    k8s_api_server_url: &str,
    k8s_api_ca_cert_data: &[u8],
) -> Result<Identity> {
    let credential_request =
        prepare_and_call_pinniped_exchange(authorization, k8s_api_server_url, k8s_api_ca_cert_data)
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

/// identity_for_exchange parses the JSON output of the credential exchange and
/// returns the Identity.
fn identity_for_exchange(cred: &ClusterCredential) -> Result<Identity> {
    let identity = Identity::from_pkcs8(
        cred.client_certificate_data.as_bytes(),
        cred.client_key_data.as_bytes(),
    )
    .context("error creating identity from x509")?;
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

/// prepare_and_call_pinniped_exchange returns the resulting
/// TokenCredentialRequest with Status after requesting a token credential
/// exchange.
async fn prepare_and_call_pinniped_exchange(
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

    call_pinniped(pinniped_namespace, client, cred_data).await
}

// More details about the cached macro options at
// https://docs.rs/cached/latest/cached/proc_macro/index.html
#[cached(
    // Creates an ExpiringValueCache with key and value both being TokenCredentialRequests.
    type = "ExpiringValueCache<TokenCredentialRequest, TokenCredentialRequest>",
    create = "{ ExpiringValueCache::with_size(5) }",
    key = "TokenCredentialRequest",
    // We convert the arguments so that we're only using the cred_data as the key.
    convert = r#"{ cred_data.clone() }"#,
    result = true,
    // If two requests with the same key arrive, only one is sent through on a miss
    // with the others returning once the result is cached.
    sync_writes = true,
)]
async fn call_pinniped(
    pinniped_namespace: String,
    client: kube::Client,
    cred_data: TokenCredentialRequest,
) -> Result<TokenCredentialRequest> {
    // define the Api Resource dynamically
    let gvk = GroupVersionKind::gvk(
        &get_pinniped_login_api_group(),
        TOKEN_REQUEST_VERSION,
        TOKEN_REQUEST_KIND,
    );
    let ar = ApiResource::from_gvk(&gvk);
    let cred_request = DynamicObject::new("", &ar)
        .within(&pinniped_namespace)
        .data(serde_json::to_value(cred_data)?);
    debug!("{}", serde_json::to_string(&cred_request).unwrap());
    // token credential request invocation
    // we start first as a cluster-based call. if this call fails with a NotFound
    // error, it is an indication we are on an old version which was
    // namespace-based. we thus fallback to a namespace-based call.
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
    use chrono::{Duration, TimeZone, Utc};
    use serial_test::serial;
    use std::collections::hash_map::DefaultHasher;

    const VALID_CERT_BASE64: &'static str = "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUN5RENDQWJDZ0F3SUJBZ0lCQURBTkJna3Foa2lHOXcwQkFRc0ZBREFWTVJNd0VRWURWUVFERXdwcmRXSmwKY201bGRHVnpNQjRYRFRJd01UQXlOakl6TXpBME5Wb1hEVE13TVRBeU5ESXpNekEwTlZvd0ZURVRNQkVHQTFVRQpBeE1LYTNWaVpYSnVaWFJsY3pDQ0FTSXdEUVlKS29aSWh2Y05BUUVCQlFBRGdnRVBBRENDQVFvQ2dnRUJBT1ZKCnFuOVBFZUp3UDRQYnI0cFo1ZjZKUmliOFZ5a2tOYjV2K1hzTVZER01aWGZLb293Y29IYjFwRWh5d0pzeDFiME4Kd2YvZ1JURi9maEgzT0drRnNQMlV2a0lHVytzNUlBd0sxMFRXYkN5VzAwT3lzVkdLcnl5bHNWcEhCWXBZRGJBcQpkdnQzc0FkcFJZaGlLZSs2NkVTL3dQNTdLV3g0SVdwZko0UGpyejh2NkJBWlptZ3o5ZzRCSFNMQkhpbTVFbTdYClBJTmpKL1RJTXFzVW1PR1ppUUNHR0ptRnQxZ21jQTd3eHZ0ZXg2ckkxSWdFNkh5NW10UzJ3NDZaMCtlVU1RSzgKSE9UdnI5aGFETnhJenVjbkduaFlCT2Z2U2VVaXNCR0pOUm5QbENydWx4b2NSZGI3N20rQUdzWW52QitNd2prVQpEbXNQTWZBelpSRHEwekhzcGEwQ0F3RUFBYU1qTUNFd0RnWURWUjBQQVFIL0JBUURBZ0trTUE4R0ExVWRFd0VCCi93UUZNQU1CQWY4d0RRWUpLb1pJaHZjTkFRRUxCUUFEZ2dFQkFBWndybXJLa3FVaDJUYld2VHdwSWlOd0o1NzAKaU9lTVl2WWhNakZxTmt6Tk9OUW55c3lPd1laRGJFMDRrV3AxclRLNHVZaUh3NTJUc0cyelJsZ0QzMzNKaEtvUQpIVloyV1hUT3Z5U2RJaWl5bVpKM2N3d0p2T0lhMW5zZnhYY1NJakJnYnNzYXowMndpRCtlazRPdmlRZktjcXJpCnFQbWZabDZDSkk0NU1rd3JwTExFaTZkNVhGbkhDb3d4eklxQjBrUDhwOFlOaGJYWTNYY2JaNElvY2lMemRBamUKQ1l6NXFVSlBlSDJCcHNaM0JXNXRDbjcycGZYazVQUjlYOFRUTHh6aTA4SU9yYjgvRDB4Tnk3emQyMnVjNXM1bwoveXZIeEt6cXBiczVuRXJkT0JFVXNGWnBpUEhaVGc1dExmWlZ4TG00VjNTZzQwRWUyNFd6d09zaDNIOD0KLS0tLS1FTkQgQ0VSVElGSUNBVEUtLS0tLQo=";

    // By default cargo will run rust unit tests in parallel. The serial macro
    // ensures that a specific test (or group of tests) runs serially.
    #[test]
    #[serial(envtest)]
    fn test_call_pinniped_exchange_no_env() -> Result<()> {
        env::remove_var(DEFAULT_PINNIPED_NAMESPACE);
        match tokio_test::block_on(prepare_and_call_pinniped_exchange(
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
        match tokio_test::block_on(prepare_and_call_pinniped_exchange(
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
        match tokio_test::block_on(prepare_and_call_pinniped_exchange(
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

    // RSA cert and key generated using
    // https://support.google.com/cloudidentity/answer/6342198?hl=en
    const VALID_CERT_PEM: &'static str = "-----BEGIN CERTIFICATE-----
MIIDazCCAlOgAwIBAgIUBQl5lApAIWapfksb8y6nGWrvmWQwDQYJKoZIhvcNAQEL
BQAwRTELMAkGA1UEBhMCQVUxEzARBgNVBAgMClNvbWUtU3RhdGUxITAfBgNVBAoM
GEludGVybmV0IFdpZGdpdHMgUHR5IEx0ZDAeFw0yMjA2MDgwMjQ0NTdaFw0yMjA3
MDgwMjQ0NTdaMEUxCzAJBgNVBAYTAkFVMRMwEQYDVQQIDApTb21lLVN0YXRlMSEw
HwYDVQQKDBhJbnRlcm5ldCBXaWRnaXRzIFB0eSBMdGQwggEiMA0GCSqGSIb3DQEB
AQUAA4IBDwAwggEKAoIBAQDKCpsDtZDgG7sgoMkTHFXfPEjJlfnWoDCWrC9DT3d7
o36gOQrnXzfJvf8ySg6YWX0qbbMLL1kwfRTo2xlaQIed2N+JPLhKVsdAtl8wh0wx
1ItgZQPzx5fdvjcBlcxWgxZovNwIv45Bsu3T2jh2qJmCXmq6Z/3Y8Kk+IGZvj8rT
qZygRF9/UupVWFNUCTivQwB4mkjNxOvYYfJ1T0NidlfxIswXecn7JQzRApyTvwrh
F4/pVXK4ZNy5U+ZrxJ5CLLMlng7FSeB2dBGTi1knaq9vtxwpXKM3ukvjoSq9lMu8
Kfjs/W+V2g81ClvquOACel8D5+hgItpwLJRCkYuFvRSVAgMBAAGjUzBRMB0GA1Ud
DgQWBBSDayGbQJejUul4VD0cHmNEFYKO7TAfBgNVHSMEGDAWgBSDayGbQJejUul4
VD0cHmNEFYKO7TAPBgNVHRMBAf8EBTADAQH/MA0GCSqGSIb3DQEBCwUAA4IBAQCQ
Ypsr6ad4c0TBJsiCEaGCcWVtHAL2KqlGguhL+Ao69/VEI6Lg7UuMoBBGlAI6FAfQ
XU+ndWsdY/3U8H2ZCpCP5muStqcYbj3ZNHxQhuNLQl/BL5wkr62nZ6M0sRrPDOXH
xzv/gz4FwPlgO2Pxb1IwqkW1ZoSmU7lK+OJSFbaSsRVjClkoFT5gFxmKSauqzY+0
BzItnYcwueDkUdgBvKPk4vDQ0NQEpGTOUHTXU6vh3Ioho38sQphNeNgEkTmz2UDk
MOpW248fauBDkVOQi5AEWYiPG0KMGICk+Pg+gEdDCXrKMwGXN1ys9sxifOqIBFrz
tGsw5ToC/b5hvjv96OYb
-----END CERTIFICATE-----";

    const VALID_KEY_PEM: &'static str = "-----BEGIN PRIVATE KEY-----
MIIEvgIBADANBgkqhkiG9w0BAQEFAASCBKgwggSkAgEAAoIBAQDKCpsDtZDgG7sg
oMkTHFXfPEjJlfnWoDCWrC9DT3d7o36gOQrnXzfJvf8ySg6YWX0qbbMLL1kwfRTo
2xlaQIed2N+JPLhKVsdAtl8wh0wx1ItgZQPzx5fdvjcBlcxWgxZovNwIv45Bsu3T
2jh2qJmCXmq6Z/3Y8Kk+IGZvj8rTqZygRF9/UupVWFNUCTivQwB4mkjNxOvYYfJ1
T0NidlfxIswXecn7JQzRApyTvwrhF4/pVXK4ZNy5U+ZrxJ5CLLMlng7FSeB2dBGT
i1knaq9vtxwpXKM3ukvjoSq9lMu8Kfjs/W+V2g81ClvquOACel8D5+hgItpwLJRC
kYuFvRSVAgMBAAECggEANZ0w24Af7MiPFK52DUM0qmOF8TCCNukVW7ZfaF47F60g
GgZpFVLYLAnmIYMzckw1AcBQhcRPx6U5mj0h8igzlLiLQRDC2r9CarK6edc9ae+7
+J11uggaDba/RAVrTv3EQZD0VsH2TwrbP5+l4h8FdWn2qnaUDzB1yM2yQSKIMThV
UcnDCkDmlNw+o6rTYqi89uyagPOSdmXo+ebuvp5pu+fmQcRlCBGN3P40pS8t5EAe
7wehYgOuJvbLawj8n/B7jVOmCY2dRtjBzeTLsrysvqtwoFq9m4qWsx2UwnfOMU4B
XUY8EgA1meNjWZgZWLk2zEH/nqL7LIsqWd2Qr7x0dQKBgQDs0cwWxnxQq+Ku0tSy
9HrNcNnk0+OULFDXEwkSfdIOJB5m2oeii1dgJ6TR7PGYfj1j9kIgQ1njmTVL8XgK
QtjQgIyM0mVndnEW2T6U1eq4vwvLniCj/V73KGILjev2uuAP3VTXTEpZy9929Jkn
7Lm46eAoxLJUrmKgaXYCUXMbtwKBgQDaZ7ikwLqtp9QBvn5UN996RD5Q57qjBov4
sOV77cuhXfGJUoQnDBRryAcUdAUtNZZWiHLG4ovCGzuzIvzBUndTEPNAizGBmsep
cpaysLwD450AjpmTAB7bSy1Xcj7hcSjl0Y9JIhATPH3iOUVnYFx8Cfk051BJtJCp
r9zVkYoqEwKBgQDCM51IhAZH5Vyj/rJ7+i6GMGgO1Y/H37t/U9XZuyI5hHcF42jc
66WAbaIkoEjSw5s2USiS6ohZMzdYirDkwUKpYPFhPdv4R1Gf6hD+3pl4XPqgRJEB
yfJJfm1AimaZU1AQ0nETiTVjg+NB2n2KFv+KWwf+hqay+LpaT4F9jyt06wKBgDxR
sRkvcV9Mnqzso4827y2hc2R823ry7+17TaXwgvDKNU8rzvvJxkoOMIZhlJxr1F2J
yclMADVXuCE9ZHkwAWybndMRnlahHMubrisjzIl2b4Ib4CZNPjhqhtdD4kH5MsZm
HiCgm7f0WQAFuTlXz7MiPgVybSYuDFYRD/ib/YCpAoGBAI4kCCy9yy3CCt8xKDes
A1hAfFSMLKYFQgB5NGbJk6l/5qyUBmXUO/WJ/BzMi/RIGzp9w6HT3e2FAiS3jnMo
fbMIp6gTs4+dKJM2o5YtCOQ4nAarWgNT5pRwo18C3/1FEhqRejBob2g9Kx9ymKwR
mZu9A/ivt37pOQXm/HOX6tHB
-----END PRIVATE KEY-----";

    #[test]
    fn test_identity_for_exchange() -> Result<()> {
        let valid_credential: ClusterCredential = ClusterCredential {
            client_certificate_data: String::from(VALID_CERT_PEM),
            client_key_data: String::from(VALID_KEY_PEM),
            expiration_timestamp: metav1::Time(Utc.timestamp(0, 0)),
            token: Some(String::from("")),
        };

        let result = identity_for_exchange(&valid_credential);
        match result {
            Err(e) => anyhow::bail!("expected Ok, got {:#?}", e),
            Ok(_) => {}
        }

        let cred_with_bad_cert = ClusterCredential {
            client_certificate_data: String::from("foo"),
            ..valid_credential.clone()
        };

        if identity_for_exchange(&cred_with_bad_cert).is_ok() {
            anyhow::bail!("expected error");
        }

        let cred_with_bad_key = ClusterCredential {
            client_key_data: String::from("foo"),
            ..valid_credential.clone()
        };

        match identity_for_exchange(&cred_with_bad_key) {
            Ok(_) => anyhow::bail!("expected error"),
            Err(_) => Ok(()),
        }
    }

    // Without any changes, the make_token_credential_request function
    // returns a request with the following hash.
    const DEFAULT_TOKEN_CREDENTIAL_REQUEST_HASH: u64 = 2363471629413450951;

    fn make_token_credential_request() -> TokenCredentialRequest {
        TokenCredentialRequest {
            spec: TokenCredentialRequestSpec {
                token: Some(String::from("fake-token")),
                authenticator: corev1::TypedLocalObjectReference {
                    name: String::from("fake-authenticator-name"),
                    kind: String::from("fake-authenticator-kind"),
                    api_group: Some(get_pinniped_authenticator_api_group().into()),
                },
            },
            status: None,
        }
    }

    // Disabling these hash tests as they occasionally fail with a specific
    // other hash.
    #[ignore]
    #[test]
    fn test_token_credential_request_hash_default() -> Result<()> {
        let cred_data = make_token_credential_request();

        let mut hasher = DefaultHasher::new();
        cred_data.hash(&mut hasher);

        assert_eq!(hasher.finish(), DEFAULT_TOKEN_CREDENTIAL_REQUEST_HASH);
        Ok(())
    }

    #[ignore]
    #[test]
    fn test_token_credential_request_hash_differs_with_token() -> Result<()> {
        let mut cred_data = make_token_credential_request();
        cred_data.spec.token = Some(String::from("another-token"));

        let mut hasher = DefaultHasher::new();
        cred_data.hash(&mut hasher);

        assert!(hasher.finish() != DEFAULT_TOKEN_CREDENTIAL_REQUEST_HASH);
        Ok(())
    }

    #[ignore]
    #[test]
    fn test_token_credential_request_hash_differs_with_authenticator() -> Result<()> {
        let mut cred_data = make_token_credential_request();
        cred_data.spec.authenticator.name = String::from("another-authenticator-name");

        let mut hasher = DefaultHasher::new();
        cred_data.hash(&mut hasher);

        assert!(hasher.finish() != DEFAULT_TOKEN_CREDENTIAL_REQUEST_HASH);
        Ok(())
    }

    #[ignore]
    #[test]
    fn test_token_credential_request_hash_identical_with_status_change() -> Result<()> {
        let mut cred_data = make_token_credential_request();
        cred_data.status = Some(TokenCredentialRequestStatus {
            credential: Some(ClusterCredential {
                token: Some(String::from("returned token")),
                client_certificate_data: String::from("cert-data"),
                client_key_data: String::from("key-data"),
                expiration_timestamp: metav1::Time(Utc.timestamp(0, 0)),
            }),
            message: Some(String::from("some status message")),
        });

        let mut hasher = DefaultHasher::new();
        cred_data.hash(&mut hasher);

        assert_eq!(hasher.finish(), DEFAULT_TOKEN_CREDENTIAL_REQUEST_HASH);
        Ok(())
    }

    #[test]
    fn test_token_credential_request_without_status_is_expired() -> Result<()> {
        let cred_data = make_token_credential_request();

        assert!(cred_data.is_expired());
        Ok(())
    }

    #[test]
    fn test_token_credential_request_without_credential_is_expired() -> Result<()> {
        let mut cred_data = make_token_credential_request();
        cred_data.status = Some(TokenCredentialRequestStatus {
            credential: None,
            message: Some(String::from("some message")),
        });

        assert!(cred_data.is_expired());
        Ok(())
    }

    #[test]
    fn test_token_credential_request_with_old_expiry_is_expired() -> Result<()> {
        let mut cred_data = make_token_credential_request();
        cred_data.status = Some(TokenCredentialRequestStatus {
            credential: Some(ClusterCredential {
                expiration_timestamp: metav1::Time(Utc.timestamp(0, 0)),
                client_certificate_data: String::from("cert data"),
                client_key_data: String::from("key data"),
                token: Some(String::from("token")),
            }),
            message: None,
        });

        assert!(cred_data.is_expired());
        Ok(())
    }

    #[test]
    fn test_token_credential_request_with_future_expiry_is_not_expired() -> Result<()> {
        let mut cred_data = make_token_credential_request();
        cred_data.status = Some(TokenCredentialRequestStatus {
            credential: Some(ClusterCredential {
                expiration_timestamp: metav1::Time(Utc::now() + Duration::days(1)),
                client_certificate_data: String::from("cert data"),
                client_key_data: String::from("key data"),
                token: Some(String::from("token")),
            }),
            message: None,
        });

        assert!(!cred_data.is_expired());
        Ok(())
    }
}
