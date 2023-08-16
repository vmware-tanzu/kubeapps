// Copyright 2023 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

use super::OCICatalogSender;
use super::{ListRepositoriesForRegistryRequest, ListTagsForRepositoryRequest, Repository, Tag};
use log;
use reqwest::{StatusCode, Url};
use serde::{Deserialize, Serialize};
use tokio::sync::mpsc;
use tonic::Status;

/// The default page size with which requests are sent to docker hub.
/// We fetch all results in batches of this page size.
const DEFAULT_PAGE_SIZE: u8 = 100;
pub const PROVIDER_NAME: &str = "DockerHubAPI";
pub const DOCKERHUB_URI: &str = "https://hub.docker.com";

#[derive(Serialize, Deserialize)]
struct DockerHubV2Repository {
    name: String,
    namespace: String,
    repository_type: Option<String>,
    content_types: Vec<String>,
}

#[derive(Serialize, Deserialize)]
struct DockerHubV2RepositoriesResult {
    count: u16,
    next: Option<String>,
    previous: Option<String>,
    results: Vec<DockerHubV2Repository>,
}

#[derive(Serialize, Deserialize)]
struct DockerHubV2Tag {
    name: String,
    repository_type: Option<String>,
    content_type: String,
}

#[derive(Serialize, Deserialize)]
struct DockerHubV2TagsResult {
    count: u16,
    next: Option<String>,
    previous: Option<String>,
    results: Vec<DockerHubV2Tag>,
}

#[derive(Debug, Default)]
pub struct DockerHubAPI {}

#[tonic::async_trait]
impl OCICatalogSender for DockerHubAPI {
    fn id(&self) -> &str {
        PROVIDER_NAME
    }

    async fn send_repositories(
        &self,
        tx: mpsc::Sender<Result<Repository, Status>>,
        request: &ListRepositoriesForRegistryRequest,
    ) {
        let mut url = url_for_request_repositories(request);

        let client = reqwest::Client::builder().build().unwrap();

        // We continue making the request until there is no `next` url
        // in the result.
        loop {
            log::debug!("requesting: {}", url);
            let response = match client.get(url.clone()).send().await {
                Ok(r) => r,
                Err(e) => {
                    tx.send(Err(Status::failed_precondition(e.to_string())))
                        .await
                        .unwrap();
                    return;
                }
            };

            if response.status() != StatusCode::OK {
                tx.send(Err(Status::failed_precondition(format!(
                    "unexpected status code when requesting {}: {}",
                    url,
                    response.status()
                ))))
                .await
                .unwrap();
                return;
            }

            let body = match response.text().await {
                Ok(b) => b,
                Err(e) => {
                    tx.send(Err(Status::failed_precondition(format!(
                        "unable to extract body from response: {}",
                        e.to_string()
                    ))))
                    .await
                    .unwrap();
                    return;
                }
            };
            log::trace!("response body: {}", body);

            let response: DockerHubV2RepositoriesResult = serde_json::from_str(&body).unwrap();

            for repo in response.results {
                tx.send(Ok(Repository {
                    registry: request.registry.clone(),
                    namespace: repo.namespace,
                    name: repo.name,
                }))
                .await
                .unwrap();
            }

            if response.next.is_some() {
                url = reqwest::Url::parse(&response.next.unwrap()).unwrap();
            } else {
                break;
            }
        }
    }

    async fn send_tags(
        &self,
        tx: mpsc::Sender<Result<Tag, Status>>,
        request: &ListTagsForRepositoryRequest,
    ) {
        let mut url = match url_for_request_tags(request) {
            Ok(u) => u,
            Err(e) => {
                tx.send(Err(e)).await.unwrap();
                return;
            }
        };

        let client = reqwest::Client::builder().build().unwrap();

        // We continue making the request until there is no `next` url
        // in the result.
        loop {
            log::debug!("requesting: {}", url);
            let response = match client.get(url.clone()).send().await {
                Ok(r) => r,
                Err(e) => {
                    tx.send(Err(Status::failed_precondition(e.to_string())))
                        .await
                        .unwrap();
                    return;
                }
            };

            if response.status() != StatusCode::OK {
                tx.send(Err(Status::failed_precondition(format!(
                    "unexpected status code when requesting {}: {}",
                    url,
                    response.status()
                ))))
                .await
                .unwrap();
                return;
            }

            let body = match response.text().await {
                Ok(b) => b,
                Err(e) => {
                    tx.send(Err(Status::failed_precondition(format!(
                        "unable to extract body from response: {}",
                        e.to_string()
                    ))))
                    .await
                    .unwrap();
                    return;
                }
            };
            log::trace!("response body: {}", body);

            let response: DockerHubV2TagsResult = serde_json::from_str(&body).unwrap();

            for tag in response.results {
                tx.send(Ok(Tag { name: tag.name })).await.unwrap();
            }

            if response.next.is_some() {
                url = reqwest::Url::parse(&response.next.unwrap()).unwrap();
            } else {
                break;
            }
        }
    }
}

fn url_for_request_repositories(request: &ListRepositoriesForRegistryRequest) -> Url {
    let mut url = reqwest::Url::parse(DOCKERHUB_URI).unwrap();

    if !request.namespace.is_empty() {
        url.set_path(&format!(
            "/v2/namespaces/{}/repositories/",
            request.namespace
        ));
    } else {
        url.set_path("/v2/repositories/");
    }
    // For now we use a default page size and default ordering.
    url.query_pairs_mut()
        .append_pair("page_size", &format!("{}", DEFAULT_PAGE_SIZE))
        .append_pair("ordering", "name");

    // Append any content types from the query.
    for ct in request.content_types.iter() {
        url.query_pairs_mut().append_pair("content_types", ct);
    }
    url
}

fn url_for_request_tags(request: &ListTagsForRepositoryRequest) -> Result<Url, Status> {
    let mut url = reqwest::Url::parse(DOCKERHUB_URI).unwrap();

    let repo = match request.repository.clone() {
        Some(r) => r,
        None => {
            return Err(Status::invalid_argument(format!(
                "repository not set in request"
            )))
        }
    };

    url.set_path(&format!(
        "/v2/repositories/{}/{}/tags",
        repo.namespace, repo.name
    ));

    // For now we use a default page size.
    url.query_pairs_mut()
        .append_pair("page_size", &format!("{}", DEFAULT_PAGE_SIZE));

    Ok(url)
}

#[cfg(test)]
mod tests {
    use super::*;
    use rstest::rstest;

    #[rstest]
    #[case::without_namespace(ListRepositoriesForRegistryRequest{
        registry: "registry-1.dockerhub.io".to_string(),
        ..Default::default()
    }, "https://hub.docker.com/v2/repositories/?page_size=100&ordering=name")]
    #[case::with_namespace(ListRepositoriesForRegistryRequest{
        registry: "registry-1.dockerhub.io".to_string(),
        namespace: "bitnamicharts".to_string(),
        ..Default::default()
    }, "https://hub.docker.com/v2/namespaces/bitnamicharts/repositories/?page_size=100&ordering=name")]
    #[case::with_content_type(ListRepositoriesForRegistryRequest{
        registry: "registry-1.dockerhub.io".to_string(),
        namespace: "bitnamicharts".to_string(),
        content_types: vec!["helm".to_string()],
        ..Default::default()
    }, "https://hub.docker.com/v2/namespaces/bitnamicharts/repositories/?page_size=100&ordering=name&content_types=helm")]
    #[case::with_multiple_content_types(ListRepositoriesForRegistryRequest{
        registry: "registry-1.dockerhub.io".to_string(),
        namespace: "bitnamicharts".to_string(),
        content_types: vec!["helm".to_string(), "image".to_string()],
        ..Default::default()
    }, "https://hub.docker.com/v2/namespaces/bitnamicharts/repositories/?page_size=100&ordering=name&content_types=helm&content_types=image")]
    fn test_url_for_request(
        #[case] request: ListRepositoriesForRegistryRequest,
        #[case] expected_url: Url,
    ) {
        assert_eq!(url_for_request_repositories(&request), expected_url);
    }

    #[rstest]
    #[case::without_repository(ListTagsForRepositoryRequest{
        ..Default::default()
    }, Err(Status::invalid_argument("bang")))]
    #[case::with_repository(ListTagsForRepositoryRequest{
        repository: Some(Repository{
            namespace: "bitnamicharts".to_string(),
            name: "apache".to_string(),
            ..Default::default()
        }),
        ..Default::default()
    }, Ok(reqwest::Url::parse("https://hub.docker.com/v2/repositories/bitnamicharts/apache/tags?page_size=100").unwrap()))]
    fn test_url_for_request_tags(
        #[case] request: ListTagsForRepositoryRequest,
        #[case] expected_result: Result<Url, Status>,
    ) {
        match expected_result {
            Ok(url) => {
                assert_eq!(url_for_request_tags(&request).unwrap(), url);
            }
            Err(e) => {
                assert_eq!(
                    url_for_request_tags(&request).err().unwrap().code(),
                    e.code()
                )
            }
        }
    }
}
