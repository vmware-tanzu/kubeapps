import { grpc } from "@improbable-eng/grpc-web";
import { GrpcWebImpl } from "gen/kubeappsapis/core/plugins/v1alpha1/plugins";
import * as URL from "./url";

export class GrpcClient {
  private baseUrl = URL.api.kubeappsapis;
  private grpcClient!: GrpcWebImpl;

  public getGrpcClient = () => {
    if (!this.grpcClient) {
      grpc.setDefaultTransport(grpc.CrossBrowserHttpTransport({}));
      this.grpcClient = new GrpcWebImpl(this.baseUrl, {});
    }
    return this.grpcClient;
  };
}
