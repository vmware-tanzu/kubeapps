// Copyright 2023 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

use crate::oci_catalog::RegistryProvider;

use super::{ListRepositoriesForRegistryRequest, ListTagsForRepositoryRequest, Repository, Tag};
use tokio::sync::mpsc;
use tonic::Status;

pub mod dockerhub;

/// OCICatalogSender is a trait that can be implemented by different providers.
///
/// Initially we'll provide a DockerHub implementation followed by Harbor and
/// others.
#[tonic::async_trait]
pub trait OCICatalogSender {
    /// send_repositories requests repositories from the provider and sends
    /// them down a channel for our API to return.
    async fn send_repositories(
        &self,
        tx: mpsc::Sender<Result<Repository, Status>>,
        request: &ListRepositoriesForRegistryRequest,
    );

    /// send_tags requests tags for a repository of a provider and sends
    /// them down a channel for our API to return.
    async fn send_tags(
        &self,
        tx: mpsc::Sender<Result<Tag, Status>>,
        request: &ListTagsForRepositoryRequest,
    );

    // The id simply gives a way in tests to determine which provider
    // has been selected.
    fn id(&self) -> &str;
}

#[derive(Debug, Clone, PartialEq)]
pub struct ProviderError;

pub fn provider_for_request(
    registry: String,
    provider: RegistryProvider,
) -> Result<Box<dyn OCICatalogSender + Send + Sync>, ProviderError> {
    match provider {
        RegistryProvider::DockerHub => Ok(Box::new(dockerhub::DockerHubAPI::default())),
        _ => match registry {
            r if r.ends_with("docker.io") => Ok(Box::new(dockerhub::DockerHubAPI::default())),
            _ => Err(ProviderError),
        },
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use rstest::rstest;

    #[rstest]
    #[case::with_registry_provider(
        "https://example.com",
        RegistryProvider::DockerHub,
        dockerhub::PROVIDER_NAME
    )]
    #[case::without_provider_but_matching_url(
        "https://registry-x.docker.io",
        RegistryProvider::Unspecified,
        dockerhub::PROVIDER_NAME
    )]
    #[case::without_provider_or_matching_registry(
        "https://registry-x.dockers.io",
        RegistryProvider::Unspecified,
        ""
    )]
    fn test_provider_for_request(
        #[case] registry: String,
        #[case] provider: RegistryProvider,
        #[case] expected_id: &str,
    ) {
        let result = provider_for_request(registry, provider);

        if expected_id == "" {
            assert!(result.is_err());
            return;
        } else {
            assert!(result.is_ok());
        }

        let provider = result.unwrap();

        assert_eq!(provider.id(), expected_id);
    }
}
