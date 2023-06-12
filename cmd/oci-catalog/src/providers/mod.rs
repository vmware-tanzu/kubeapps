// Copyright 2023 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

use super::{ListRepositoriesRequest, ListTagsRequest, Repository, Tag};
use tokio::sync::mpsc;
use tonic::Status;

pub mod dockerhub;

// OCICatalogSender is a trait that can be implemented by different providers.
//
// Initially we'll provide a DockerHub implementation followed by Harbor and
// others.
#[tonic::async_trait]
pub trait OCICatalogSender {
    async fn send_repositories(
        tx: mpsc::Sender<Result<Repository, Status>>,
        request: &ListRepositoriesRequest,
    );
    async fn send_tags(tx: mpsc::Sender<Result<Tag, Status>>, request: &ListTagsRequest);
}
