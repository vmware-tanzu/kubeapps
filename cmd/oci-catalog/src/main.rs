// Copyright 2023 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

use tokio_stream::wrappers::ReceiverStream;
use tokio::sync::mpsc;
use tonic::{transport::Server, Request, Response, Status};

// Ensure that the compiled proto API is available within a module
// before importing the required items.
pub mod oci_catalog {
    tonic::include_proto!("ocicatalog");
}
use oci_catalog::oci_catalog_server::{OciCatalog, OciCatalogServer};
use oci_catalog::{ListRepositoriesRequest, ListTagsRequest, Repository, Tag};

#[derive(Debug, Default)]
pub struct KubeappsOCICatalog {}

// OCICatalogAPI is a trait that can be implemented by different providers.
// Initially we'll provide a DockerHub implementation followed by Harbor and others.
#[tonic::async_trait]
pub trait OCICatalogAPI {
    async fn send_repositories(tx: mpsc::Sender<Result<Repository, Status>>, request: &ListRepositoriesRequest);
    async fn send_tags(tx: mpsc::Sender<Result<Tag, Status>>, request: &ListTagsRequest);
}

// TODO: Move to separate files once multiple implementation are available.
#[derive(Debug, Default)]
pub struct DockerHubAPI{}

// Create trait
#[tonic::async_trait]
impl OCICatalogAPI for DockerHubAPI {
    // Update to return a result so errors are handled properly.
    async fn send_repositories(tx: mpsc::Sender<Result<Repository, Status>>, request: &ListRepositoriesRequest) {
        // TODO: Stubbed with dockerhub API.
        // let mut url = reqwest::Url::parse("https://hub.docker.com/v2/repositories/").unwrap();
        // if !request.namespace.is_empty() {
        //     url.set_path(&format!("/v2/namespace/{}/repositories/", request.namespace));
        // }
        // let client = reqwest::Client::builder().build().unwrap();
        // let body = client.get(url).send().await.unwrap().text().await.unwrap();

        // parse json and send.

        // While still more pages, request and send.
        for count in 0..10 {
            tx.send(Ok(Repository {
                registry: request.registry.clone(),
                name: format!("repo-{}", count),
            })).await.unwrap();
        }
    }

    async fn send_tags(tx: mpsc::Sender<Result<Tag, Status>>, _request: &ListTagsRequest) {
        for count in 0..10 {
            tx.send(Ok(Tag {
                name: format!("tag-{}", count),
            })).await.unwrap();
        }
    }
}

#[tonic::async_trait]
impl OciCatalog for KubeappsOCICatalog {
    type ListRepositoriesForRegistryStream = ReceiverStream<Result<Repository, Status>>;
    type ListTagsForRepositoryStream = ReceiverStream<Result<Tag, Status>>;

    async fn list_repositories_for_registry(
        &self,
        request: Request<ListRepositoriesRequest>,
    ) -> Result<Response<Self::ListRepositoriesForRegistryStream>, Status> {
        // Initially for prototype, just implement support for
        // docker's registry-1.docker.io. Later split out relevant
        // functionality to a trait that can be implemented separately
        // by different services (harbor, gcr etc.)
        let (tx, rx) = mpsc::channel(4);

        tokio::spawn(async move {
            // Have a trait which each registry plugin implements for matching a request.
            match request.get_ref().registry.as_str() {
                "registry-1.docker.io" => {
                    DockerHubAPI::send_repositories(tx, request.get_ref()).await;
                },
                _ => {
                    unimplemented!()
                }
            }
        });
        Ok(Response::new(ReceiverStream::new(rx)))
    }

    async fn list_tags_for_repository(&self, request: Request<ListTagsRequest>) -> Result<Response<Self::ListTagsForRepositoryStream>, Status> {
        let (tx, rx) = mpsc::channel(4);

        tokio::spawn(async move {
            // Possibly just use generic OCI API for listing tags.
            DockerHubAPI::send_tags(tx, request.get_ref()).await;
        });
        Ok(Response::new(ReceiverStream::new(rx)))
    }
}

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    let addr = "0.0.0.0:50051".parse()?;
    let kubeapps_oci_catalog = KubeappsOCICatalog::default();

    Server::builder()
        .add_service(OciCatalogServer::new(kubeapps_oci_catalog))
        .serve(addr)
        .await?;

    Ok(())
}
