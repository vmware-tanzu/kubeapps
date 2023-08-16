// Copyright 2023 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

use clap::Parser;
use log;
use tokio::sync::mpsc;
use tokio_stream::wrappers::ReceiverStream;
use tonic::{transport::Server, Request, Response, Status};

// Ensure that the compiled proto API is available within a module
// before importing the required items.
pub mod oci_catalog {
    tonic::include_proto!("ocicatalog.v1alpha1");
}
use oci_catalog::oci_catalog_service_server::{OciCatalogService, OciCatalogServiceServer};
use oci_catalog::{
    ListRepositoriesForRegistryRequest, ListTagsForRepositoryRequest, Repository, Tag,
};

mod cli;
mod providers;

#[derive(Debug, Default)]
pub struct KubeappsOCICatalog {}

#[tonic::async_trait]
impl OciCatalogService for KubeappsOCICatalog {
    type ListRepositoriesForRegistryStream = ReceiverStream<Result<Repository, Status>>;
    type ListTagsForRepositoryStream = ReceiverStream<Result<Tag, Status>>;

    async fn list_repositories_for_registry(
        &self,
        request: Request<ListRepositoriesForRegistryRequest>,
    ) -> Result<Response<Self::ListRepositoriesForRegistryStream>, Status> {
        // The provider for request strategy provides the registry-specific
        // implementation.
        let provider = providers::provider_for_request(
            request.get_ref().registry.clone(),
            request.get_ref().registry_provider(),
        )
        .map_err(|_e| Status::failed_precondition("support for registry not found"))?;

        // Initially for prototype, just implement support for
        // docker's registry-1.docker.io. Later split out relevant
        // functionality to a trait that can be implemented separately
        // by different services (harbor, gcr etc.)
        let (tx, rx) = mpsc::channel(4);

        tokio::spawn(async move {
            provider.send_repositories(tx, request.get_ref()).await;
        });

        Ok(Response::new(ReceiverStream::new(rx)))
    }

    async fn list_tags_for_repository(
        &self,
        request: Request<ListTagsForRepositoryRequest>,
    ) -> Result<Response<Self::ListTagsForRepositoryStream>, Status> {
        // The provider for request strategy provides the registry-specific
        // implementation.
        let provider = providers::provider_for_request(
            request.get_ref().repository.clone().unwrap().registry,
            request.get_ref().registry_provider(),
        )
        .map_err(|_e| Status::failed_precondition("support for registry not found"))?;

        let (tx, rx) = mpsc::channel(4);

        tokio::spawn(async move {
            // Possibly just use generic OCI API for listing tags.
            provider.send_tags(tx, request.get_ref()).await;
        });
        Ok(Response::new(ReceiverStream::new(rx)))
    }
}

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    env_logger::init();
    let opt = cli::Options::parse();
    let addr = ([0, 0, 0, 0], opt.port).into();
    let kubeapps_oci_catalog = KubeappsOCICatalog::default();

    let (mut _health_reporter, health_service) = tonic_health::server::health_reporter();
    // TODO(absoludity): Need to implement a decent check for the actual service
    // that won't kill us with request quotas.  See
    // https://github.com/hyperium/tonic/blob/master/examples/src/health/server.rs
    // for an example setup.

    let server = Server::builder()
        .add_service(OciCatalogServiceServer::new(kubeapps_oci_catalog))
        .add_service(health_service)
        .serve(addr);
    log::info!("listening for gRPC requests at {}", addr);
    server.await.expect("unexpected error while serving");
    Ok(())
}
