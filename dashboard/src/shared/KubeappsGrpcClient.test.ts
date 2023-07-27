// Copyright 2021-2023 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { KubeappsGrpcClient } from "./KubeappsGrpcClient";

describe("kubeapps grpc client creation", () => {
  it("should create a kubeapps grpc client", async () => {
    const kubeappsGrpcClient = new KubeappsGrpcClient();
    expect(kubeappsGrpcClient).not.toBeNull();
  });

  it("should create the clients for each core service", async () => {
    const kubeappsGrpcClient = new KubeappsGrpcClient();
    const serviceClients = [
      kubeappsGrpcClient.getPluginsServiceClientImpl(),
      kubeappsGrpcClient.getPackagesServiceClientImpl(),
      kubeappsGrpcClient.getRepositoriesServiceClientImpl(),
      kubeappsGrpcClient.getResourcesServiceClientImpl(),
    ];
    serviceClients.every(sc => expect(sc).not.toBeNull());
  });

  it("should create the clients for each plugin service", async () => {
    const kubeappsGrpcClient = new KubeappsGrpcClient();
    const packagesServiceClients = [
      kubeappsGrpcClient.getHelmPackagesServiceClientImpl(),
      kubeappsGrpcClient.getKappControllerPackagesServiceClientImpl(),
      kubeappsGrpcClient.getFluxv2PackagesServiceClientImpl(),
    ];
    const repositoriesServiceClients = [
      kubeappsGrpcClient.getHelmRepositoriesServiceClientImpl(),
      kubeappsGrpcClient.getKappControllerRepositoriesServiceClientImpl(),
      kubeappsGrpcClient.getFluxV2RepositoriesServiceClientImpl(),
    ];
    packagesServiceClients.every(sc => expect(sc).not.toBeNull());
    repositoriesServiceClients.every(sc => expect(sc).not.toBeNull());
  });
});
