import { grpc } from "@improbable-eng/grpc-web";
import { GrpcWebImpl } from "gen/kubeappsapis/core/plugins/v1alpha1/plugins";
import * as URL from "./url";

export class GrpcClient {
  private grpcClient!: GrpcWebImpl;
  private transport: grpc.TransportFactory;

  constructor(transport?: grpc.TransportFactory) {
    this.transport = transport ?? grpc.CrossBrowserHttpTransport({});
  }

  public getGrpcClient = (transport?: grpc.TransportFactory) => {
    if (!this.grpcClient) {
      grpc.setDefaultTransport(this.transport);
      this.grpcClient = new GrpcWebImpl(URL.api.kubeappsapis, {});
    }
    return this.grpcClient;
  };
}
