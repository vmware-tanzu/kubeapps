// Copyright 2023 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

use super::OCICatalogSender;
use super::{ListRepositoriesRequest, ListTagsRequest, Repository, Tag};
use log;
use reqwest::{StatusCode, Url};
use serde::{Deserialize, Serialize};
use tokio::sync::mpsc;
use tonic::Status;

/// The default page size with which requests are sent to docker hub.
const DEFAULT_PAGE_SIZE: u8 = 100;

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

#[derive(Debug, Default)]
pub struct DockerHubAPI {}

#[tonic::async_trait]
impl OCICatalogSender for DockerHubAPI {
    // Update to return a result so errors are handled properly.
    async fn send_repositories(
        tx: mpsc::Sender<Result<Repository, Status>>,
        request: &ListRepositoriesRequest,
    ) {
        let mut url = url_for_request(request);

        let client = reqwest::Client::builder().build().unwrap();

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

    async fn send_tags(tx: mpsc::Sender<Result<Tag, Status>>, _request: &ListTagsRequest) {
        for count in 0..10 {
            tx.send(Ok(Tag {
                name: format!("tag-{}", count),
            }))
            .await
            .unwrap();
        }
    }
}

fn url_for_request(request: &ListRepositoriesRequest) -> Url {
    let mut url = reqwest::Url::parse("https://hub.docker.com/v2/repositories/").unwrap();

    if !request.namespace.is_empty() {
        url.set_path(&format!(
            "/v2/namespaces/{}/repositories/",
            request.namespace
        ));
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

#[cfg(test)]
mod tests {
    use super::*;
    use rstest::rstest;

    #[rstest]
    #[case::without_namespace(ListRepositoriesRequest{
        registry: "registry-1.dockerhub.io".to_string(),
        ..Default::default()
    }, "https://hub.docker.com/v2/repositories/?page_size=100&ordering=name")]
    #[case::with_namespace(ListRepositoriesRequest{
        registry: "registry-1.dockerhub.io".to_string(),
        namespace: "bitnamicharts".to_string(),
        ..Default::default()
    }, "https://hub.docker.com/v2/namespaces/bitnamicharts/repositories/?page_size=100&ordering=name")]
    #[case::with_content_type(ListRepositoriesRequest{
        registry: "registry-1.dockerhub.io".to_string(),
        namespace: "bitnamicharts".to_string(),
        content_types: vec!["helm".to_string()],
        ..Default::default()
    }, "https://hub.docker.com/v2/namespaces/bitnamicharts/repositories/?page_size=100&ordering=name&content_types=helm")]
    #[case::with_multiple_content_types(ListRepositoriesRequest{
        registry: "registry-1.dockerhub.io".to_string(),
        namespace: "bitnamicharts".to_string(),
        content_types: vec!["helm".to_string(), "image".to_string()],
        ..Default::default()
    }, "https://hub.docker.com/v2/namespaces/bitnamicharts/repositories/?page_size=100&ordering=name&content_types=helm&content_types=image")]
    fn test_url_for_request(#[case] request: ListRepositoriesRequest, #[case] expected_url: Url) {
        assert_eq!(url_for_request(&request), expected_url);
    }
}
