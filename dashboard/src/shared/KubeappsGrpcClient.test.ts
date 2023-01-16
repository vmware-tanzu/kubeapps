// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { grpc } from "@improbable-eng/grpc-web";
import { FakeTransportBuilder } from "@improbable-eng/grpc-web-fake-transport";
import { KubeappsGrpcClient } from "./KubeappsGrpcClient";

describe("kubeapps grpc client creation", () => {
  const fakeEmptyTransport = new FakeTransportBuilder().withMessages([]).build();

  it("should create a kubeapps grpc client", async () => {
    const kubeappsGrpcClient = new KubeappsGrpcClient(fakeEmptyTransport);
    expect(kubeappsGrpcClient).not.toBeNull();
  });

  it("should create the clients for each core service", async () => {
    const kubeappsGrpcClient = new KubeappsGrpcClient(fakeEmptyTransport);
    const serviceClients = [
      kubeappsGrpcClient.getPluginsServiceClientImpl(),
      kubeappsGrpcClient.getPackagesServiceClientImpl(),
      kubeappsGrpcClient.getRepositoriesServiceClientImpl(),
      kubeappsGrpcClient.getResourcesServiceClientImpl(),
    ];
    serviceClients.every(sc => expect(sc).not.toBeNull());
  });

  it("should create the clients for each plugin service", async () => {
    const kubeappsGrpcClient = new KubeappsGrpcClient(fakeEmptyTransport);
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

describe("kubeapps grpc core plugin service", () => {
  afterEach(() => {
    jest.restoreAllMocks();
  });

  const fakeErrorransport = new FakeTransportBuilder()
    .withPreTrailersError(grpc.Code.Internal, "boom")
    .build();
  const fakeUnauthenticatedTransport = new FakeTransportBuilder()
    .withPreTrailersError(grpc.Code.Unauthenticated, "you shall not pass")
    .build();
  const fakeAuthTransport = new FakeTransportBuilder()
    .withHeaders(new grpc.Metadata({ authorization: "Bearer topsecret" }))
    .build();

  it("it fails when an internal error is thrown", async () => {
    const kubeappsGrpcClient = new KubeappsGrpcClient(fakeErrorransport);
    const getPluginsServiceClientImpl = kubeappsGrpcClient.getPluginsServiceClientImpl();
    const getConfiguredPlugins = getPluginsServiceClientImpl.GetConfiguredPlugins({});
    await expect(getConfiguredPlugins).rejects.toThrowError("boom");
  });

  it("it fails when unauthenticated", async () => {
    const kubeappsGrpcClient = new KubeappsGrpcClient(fakeUnauthenticatedTransport);
    const getPluginsServiceClientImpl = kubeappsGrpcClient.getPluginsServiceClientImpl();
    const getConfiguredPlugins = getPluginsServiceClientImpl.GetConfiguredPlugins({});
    await expect(getConfiguredPlugins).rejects.toThrowError("you shall not pass");
  });

  it("it set the metadata if using token auth", async () => {
    const kubeappsGrpcClient = new KubeappsGrpcClient(fakeAuthTransport);
    jest.spyOn(window.localStorage.__proto__, "getItem").mockReturnValue("topsecret");

    const getClientMetadataMock = jest.spyOn(KubeappsGrpcClient.prototype, "getClientMetadata");
    kubeappsGrpcClient.getPluginsServiceClientImpl();
    const expectedMetadata = new grpc.Metadata({ authorization: "Bearer topsecret" });
    expect(getClientMetadataMock.mock.results[0].value).toEqual(expectedMetadata);
  });

  it("it doesn't set the metadata if not using token auth", async () => {
    const kubeappsGrpcClient = new KubeappsGrpcClient(fakeAuthTransport);
    jest.spyOn(window.localStorage.__proto__, "getItem").mockReturnValue(null);

    const getClientMetadataMock = jest.spyOn(KubeappsGrpcClient.prototype, "getClientMetadata");
    kubeappsGrpcClient.getPluginsServiceClientImpl();
    expect(getClientMetadataMock.mock.results[0].value).toBeUndefined();
  });

  // TODO(agamez): try to also mock the messages using the new FakeTransportBuilder().withMessages([])
  // More details: https://github.com/vmware-tanzu/kubeapps/issues/3165#issuecomment-882944035
});

describe("kubeapps grpc resources plugin service", () => {
  afterEach(() => {
    jest.restoreAllMocks();
  });

  const fakeAuthTransport = new FakeTransportBuilder()
    .withHeaders(new grpc.Metadata({ authorization: "Bearer topsecret" }))
    .build();

  it("it set the metadata if using token auth", async () => {
    const kubeappsGrpcClient = new KubeappsGrpcClient(fakeAuthTransport);
    jest.spyOn(window.localStorage.__proto__, "getItem").mockReturnValue("topsecret");

    const getClientMetadataMock = jest.spyOn(KubeappsGrpcClient.prototype, "getClientMetadata");
    kubeappsGrpcClient.getResourcesServiceClientImpl();

    const expectedMetadata = new grpc.Metadata({ authorization: "Bearer topsecret" });
    expect(getClientMetadataMock.mock.results[0].value).toEqual(expectedMetadata);
  });

  it("it doesn't set the metadata if not using token auth", async () => {
    const kubeappsGrpcClient = new KubeappsGrpcClient(fakeAuthTransport);
    jest.spyOn(window.localStorage.__proto__, "getItem").mockReturnValue(null);
    const getClientMetadataMock = jest.spyOn(KubeappsGrpcClient.prototype, "getClientMetadata");

    kubeappsGrpcClient.getResourcesServiceClientImpl();

    expect(getClientMetadataMock.mock.results[0].value).toBeUndefined();
  });

  it("it sets the metadata if passed an explicit token", async () => {
    const kubeappsGrpcClient = new KubeappsGrpcClient(fakeAuthTransport);
    jest.spyOn(window.localStorage.__proto__, "getItem").mockReturnValue(null);
    const getClientMetadataMock = jest.spyOn(KubeappsGrpcClient.prototype, "getClientMetadata");

    kubeappsGrpcClient.getResourcesServiceClientImpl("topsecret");

    const expectedMetadata = new grpc.Metadata({ authorization: "Bearer topsecret" });
    expect(getClientMetadataMock.mock.results[0].value).toEqual(expectedMetadata);
  });
});
