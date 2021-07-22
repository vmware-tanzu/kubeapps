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
  const fakeEmpyTransport = new FakeTransportBuilder().withMessages([]).build();
  const fakeErrorransport = new FakeTransportBuilder()
    .withPreTrailersError(grpc.Code.Internal, "boom")
    .build();
  const fakeUnauthenticatedTransport = new FakeTransportBuilder()
    .withPreTrailersError(grpc.Code.Unauthenticated, "you shall not pass")
    .build();

  it("it fails when an internal error is thrown", async () => {
    const kubeappsGrpcClient = new KubeappsGrpcClient(fakeErrorransport);
    const getPluginsServiceClientImpl = kubeappsGrpcClient.getPluginsServiceClientImpl();
    const getConfiguredPlugins = getPluginsServiceClientImpl.GetConfiguredPlugins({});
    expect(getConfiguredPlugins).rejects.toThrowError("boom");
  });

  it("it fails when unauthenticated", async () => {
    const kubeappsGrpcClient = new KubeappsGrpcClient(fakeUnauthenticatedTransport);
    const getPluginsServiceClientImpl = kubeappsGrpcClient.getPluginsServiceClientImpl();
    const getConfiguredPlugins = getPluginsServiceClientImpl.GetConfiguredPlugins({});
    expect(getConfiguredPlugins).rejects.toThrowError("you shall not pass");
  });

  it("it returns null when the server sends no messages", async () => {
    const kubeappsGrpcClient = new KubeappsGrpcClient(fakeEmpyTransport);
    const getPluginsServiceClientImpl = kubeappsGrpcClient.getPluginsServiceClientImpl();
    const res = await getPluginsServiceClientImpl.GetConfiguredPlugins({});
    expect(res).toBeNull();
  });

  // TODO(agamez): perform an actual check here.
  // Currently withMessages is not working with our generated sources
});
