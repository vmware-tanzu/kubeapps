use std::collections::HashMap;
use std::env;
use std::process::{Command, Output};

use anyhow::{Context, Result};
use k8s_openapi::api::core::v1 as corev1;
use k8s_openapi::apimachinery::pkg::apis::meta::v1 as metav1;
use kube::{
    api::{Api, Meta, PostParams},
    Client, Config, CustomResource,
};
use log::info;
use native_tls::Identity;
use openssl::{pkcs12::Pkcs12, pkey::PKey, x509::X509};
use reqwest;
use serde::{Deserialize, Serialize};
use serde_json;
use url::Url;

/// pinniped_exchange accepts an authorization header and returns a client cert authentication Identity.
pub fn pinniped_exchange_for_identity(authorization: &str, k8s_api_server_url: &str, k8s_api_ca_cert_data: &str, pinniped_executable: String) -> Result<Identity> {
    let exchange_output = call_pinniped_exchange(authorization, k8s_api_server_url, k8s_api_ca_cert_data, &pinniped_executable)?;
    if !exchange_output.status.success() {
        let err_msg: String;
        match exchange_output.status.code() {
            Some(code) => err_msg = format!("{} exited with status code: {}", pinniped_executable, code),
            None => err_msg = format!("{} terminated by signal", pinniped_executable),
        }
        return Err(anyhow::anyhow!(format!("{}\nstdout: {}\nstderr: {}", err_msg, String::from_utf8_lossy(&exchange_output.stdout), String::from_utf8_lossy(&exchange_output.stderr))));
    }

    identity_for_exchange(exchange_output.stdout)
}

/// call_pinniped_exchange returns the output from a `pinniped exchange-credentials` CLI call with the provided authorization.
/// 
/// TODO: use the API - using the client like this is an inherent security risk
/// as the client may choos to cache credentials on the assumption that the
/// client is used by a single user.
fn call_pinniped_exchange(authorization: &str, k8s_api_server_url: &str, k8s_api_ca_cert_data: &str, pinniped_executable: &str) -> Result<Output> {
    let mut filtered_env: HashMap<String, String> = env::vars().filter(
        |&(ref k, _)|
        k == "PINNIPED_NAMESPACE" || k == "PINNIPED_AUTHENTICATOR_TYPE" || k == "PINNIPED_AUTHENTICATOR_NAME" || k == "HOME"
    ).collect();

    let auth = match authorization.to_string().strip_prefix("Bearer ") {
        Some(a) => a.to_string(),
        None => authorization.to_string(),
    };

    filtered_env.insert("PINNIPED_TOKEN".to_string(), auth);
    filtered_env.insert("PINNIPED_CA_BUNDLE".to_string(), k8s_api_ca_cert_data.to_string());
    filtered_env.insert("PINNIPED_K8S_API_ENDPOINT".to_string(), k8s_api_server_url.to_string());

    Command::new(pinniped_executable)
        .arg("exchange-credential")
        .env_clear()
        .envs(&filtered_env)
        .output()
        .with_context(|| "Extra info")
}

/// identity_for_exchange parses the JSON output of the credential exchange and returns the Identity.
/// 
/// Note: to create an identity, need to go via a pkcs12 currently.
/// https://github.com/sfackler/rust-native-tls/issues/27#issuecomment-324262673
fn identity_for_exchange(stdout: Vec<u8>) -> Result<Identity> {

    let exec_cred : ExecCredential = serde_json::from_str(&String::from_utf8_lossy(&stdout))?;
    let pkey = PKey::private_key_from_pem(exec_cred.status.client_key_data.as_bytes())?;
    let x509 = X509::from_pem(exec_cred.status.client_certificate_data.as_bytes())?;

    let pkcs_cert = Pkcs12::builder()
        .build("", "friendly-name", &pkey, &x509)?;
    let identity = Identity::from_pkcs12(&pkcs_cert.to_der()?, "")?;
    Ok(identity)
}

#[derive(Deserialize)]
struct ExecCredential {
    status: ExecCredentialStatus,
}

#[derive(Deserialize)]
#[serde(rename_all = "camelCase")]
struct ExecCredentialStatus {
    client_key_data: String,
    client_certificate_data: String,
}

/// TokenCredentialRequestSpec 
#[derive(CustomResource, Deserialize, Serialize, Clone, Debug)]
#[kube(group = "login.concierge.pinniped.dev", version = "v1alpha1", kind = "TokenCredentialRequest", namespaced)]
#[kube(status = "TokenCredentialRequestStatus")]
pub struct TokenCredentialRequestSpec {
    // Bearer token supplied with the credential request.
    token: String,

    // Reference to an authenticator which can verify this credential request.
    authenticator: corev1::TypedLocalObjectReference,
}

#[derive(Deserialize, Serialize, Clone, Debug)]
pub struct TokenCredentialRequestStatus {
    // A ClusterCredential will be returned for a successful credential request.
    credential: ClusterCredential,

    // An error message will be returned for an unsuccessful credential request.
    message: String,
}

#[derive(Deserialize, Serialize, Clone, Debug)]
/// ClusterCredential is the cluster-specific credential returned on a successful credential request. It
/// contains either a valid bearer token or a valid TLS certificate and corresponding private key for the cluster.
pub struct ClusterCredential {
    // ExpirationTimestamp indicates a time when the provided credentials expire.
    expiration_timestamp: metav1::Time,

    // Token is a bearer token used by the client for request authentication.
    token: String,

    // PEM-encoded client TLS certificates (including intermediates, if any).
    client_certificate_data: String,

    // PEM-encoded private key for the above certificate.
    client_key_data: String,
}

// If we need to support the default trait.
// impl Default for metav1::Time {
//     fn default() -> Self{ metav1::Time(DateTime::)}
// }

/// call_pinniped_exchange returns the output from a `pinniped exchange-credentials` CLI call with the provided authorization.
/// 
/// TODO: use the API - using the client like this is an inherent security risk
/// as the client may choos to cache credentials on the assumption that the
/// client is used by a single user.
async fn call_pinniped_exchange2(authorization: &str, k8s_api_server_url: &str, k8s_api_ca_cert_data: &str) -> Result<TokenCredentialRequest> {
    let pinniped_namespace = env::var("PINNIPED_NAMESPACE")?;

    // The URL has already been validated.
    let mut config = Config::new(Url::parse(k8s_api_server_url)?);
    config.default_ns = pinniped_namespace.clone();
    let cert = reqwest::Certificate::from_pem(k8s_api_ca_cert_data.as_bytes())?;
    config.root_cert = Some(vec![cert]);
    let client = Client::new(config);

    let auth_token = match authorization.to_string().strip_prefix("Bearer ") {
        Some(a) => a.to_string(),
        None => authorization.to_string(),
    };
    let token_creds: Api<TokenCredentialRequest> = Api::namespaced(client.clone(), &pinniped_namespace);
    let cred_request = TokenCredentialRequest::new("", TokenCredentialRequestSpec {
        token: auth_token,    
        authenticator: corev1::TypedLocalObjectReference {
            name: env::var("PINNIPED_AUTHENTICATOR_NAME")?,
            kind: env::var("PINNIPED_AUTHENTICATOR_TYPE")?,
            api_group: Some("authentication.concierge.pinniped.dev".into()),
        },
    }); 
    // TODO: add token and authenticator to request.
    match token_creds.create(&PostParams::default(), &cred_request).await {
        Ok(o) => {
            info!("created {} ({:?})", Meta::name(&o), o.clone().status.unwrap());
            Ok(o)
        },
        // Err(kube::Error::Api(ae)) => assert_eq!(ae.code, 409), // TODO: update to handle auth.
        Err(e) => Err(e.into()),
    }
}
