use std::collections::HashMap;
use std::env;
use std::process::{Command, Output};

use anyhow::{Context, Result};
use native_tls::Identity;
use openssl::{pkcs12::Pkcs12, pkey::PKey, x509::X509};
use serde::Deserialize;
use serde_json;

/// pinniped_exchange accepts an authorization header and returns a client cert authentication Identity.
pub fn pinniped_exchange_for_identity(authorization: &str, pinniped_executable: String) -> Result<Identity> {
    let exchange_output = call_pinniped_exchange(authorization, &pinniped_executable)?;
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
fn call_pinniped_exchange(authorization: &str, pinniped_executable: &str) -> Result<Output> {
    let mut filtered_env: HashMap<String, String> = env::vars().filter(
        |&(ref k, _)|
        k == "PINNIPED_NAMESPACE" || k == "PINNIPED_K8S_API_ENDPOINT" || k == "PINNIPED_AUTHENTICATOR_TYPE" || k == "PINNIPED_AUTHENTICATOR_NAME" ||
        k == "PINNIPED_CA_BUNDLE" || k == "HOME"
    ).collect();

    let auth = match authorization.to_string().strip_prefix("Bearer ") {
        Some(a) => a.to_string(),
        None => authorization.to_string(),
    };

    filtered_env.insert("PINNIPED_TOKEN".to_string(), auth);
    
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
