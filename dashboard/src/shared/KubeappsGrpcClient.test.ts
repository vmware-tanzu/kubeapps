import { grpc } from "@improbable-eng/grpc-web";
import { FakeTransportBuilder } from "@improbable-eng/grpc-web-fake-transport";
import { KubeappsGrpcClient } from "./KubeappsGrpcClient";

describe("kubeapps grpc client creation", () => {
  const fakeEmpyTransport = new FakeTransportBuilder().withMessages([]).build();

  it("should create a kubeapps grpc client", async () => {
    const kubeappsGrpcClient = new KubeappsGrpcClient(fakeEmpyTransport);
    expect(kubeappsGrpcClient).not.toBeNull();
  });

  it("should create the clients for each core service", async () => {
    const kubeappsGrpcClient = new KubeappsGrpcClient(fakeEmpyTransport);
    const serviceClients = [
      kubeappsGrpcClient.getPluginsServiceClientImpl(),
      kubeappsGrpcClient.getPackagesServiceClientImpl(),
    ];
    serviceClients.every(sc => expect(sc).not.toBeNull());
  });

  it("should create the clients for each plugin service", async () => {
    const kubeappsGrpcClient = new KubeappsGrpcClient(fakeEmpyTransport);
    const serviceClients = [
      kubeappsGrpcClient.getHelmPackagesServiceClientImpl(),
      kubeappsGrpcClient.getKappControllerPackagesServiceClientImpl(),
      kubeappsGrpcClient.getFluxv2PackagesServiceClientImpl(),
    ];
    serviceClients.every(sc => expect(sc).not.toBeNull());
  });
});

describe("kubeapps grpc core plugin service", () => {
  afterEach(() => {
    jest.restoreAllMocks();
  });

  const fakeEmpyTransport = new FakeTransportBuilder().withMessages([]).build();
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

  it("it returns null when the server sends no messages", async () => {
    const kubeappsGrpcClient = new KubeappsGrpcClient(fakeEmpyTransport);
    const getPluginsServiceClientImpl = kubeappsGrpcClient.getPluginsServiceClientImpl();
    const res = await getPluginsServiceClientImpl.GetConfiguredPlugins({});
    expect(res).toBeNull();
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

  // TODO(agamez): try to also mock the messages ussing the new FakeTransportBuilder().withMessages([])
  // More details: https://github.com/kubeapps/kubeapps/issues/3165#issuecomment-882944035
});
