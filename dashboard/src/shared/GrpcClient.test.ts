import { FakeTransportBuilder } from "@improbable-eng/grpc-web-fake-transport";
import { PluginsServiceClientImpl } from "gen/kubeappsapis/core/plugins/v1alpha1/plugins";
import { Message } from "google-protobuf";
import { GrpcClient } from "./GrpcClient";

describe("grpc client creation", () => {
  it("should create a grpc client", async () => {
    const fakeTransport = new FakeTransportBuilder().withMessages([{} as Message]).build();
    const grpcClient = new GrpcClient(fakeTransport).getGrpcClient();
    const client = new PluginsServiceClientImpl(grpcClient);

    expect(grpcClient).not.toBeNull();
    expect(client).not.toBeNull();
    // TODO(agamez): perform an actual check here.
    // Currently withMessages is not working with our generated sources,
    // const res = await client.GetConfiguredPlugins({});
    // expect(res.plugins).toBe({});
  });
});
