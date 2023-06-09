// Copyright 2023 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

use super::OCICatalogSender;
use super::{ListRepositoriesRequest, ListTagsRequest, Repository, Tag};
use tokio::sync::mpsc;
use tonic::Status;

// TODO: Move to separate files once multiple implementation are available.
#[derive(Debug, Default)]
pub struct DockerHubAPI {}

// Create trait
#[tonic::async_trait]
impl OCICatalogSender for DockerHubAPI {
    // Update to return a result so errors are handled properly.
    async fn send_repositories(
        tx: mpsc::Sender<Result<Repository, Status>>,
        request: &ListRepositoriesRequest,
    ) {
        // TODO: Stubbed with dockerhub API.
        // let mut url = reqwest::Url::parse("https://hub.docker.com/v2/repositories/").unwrap();
        // if !request.namespace.is_empty() {
        //     url.set_path(&format!("/v2/namespace/{}/repositories/",
        // request.namespace)); }
        // let client = reqwest::Client::builder().build().unwrap();
        // let body = client.get(url).send().await.unwrap().text().await.unwrap();

        // parse json and send.

        // While still more pages, request and send.
        for count in 0..10 {
            tx.send(Ok(Repository {
                registry: request.registry.clone(),
                name: format!("repo-{}", count),
            }))
            .await
            .unwrap();
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
