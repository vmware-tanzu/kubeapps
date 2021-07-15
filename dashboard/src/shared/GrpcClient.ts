import { grpc } from "@improbable-eng/grpc-web";
import { GrpcWebImpl } from "gen/kubeappsapis/core/plugins/v1alpha1/plugins";
import * as URL from "./url";

export class GrpcClient {
  private baseUrl = URL.api.kubeappsapis;
  private grpcClient!: GrpcWebImpl;

  private GetClientConfig = () => {
    return {
      // https://github.com/improbable-eng/grpc-web/blob/master/client/grpc-web/docs/transport.md#specifying-transports.
      transport: grpc.CrossBrowserHttpTransport({ withCredentials: true }),
    };
  };

  public getGrpcClient = () => {
    if (!this.grpcClient) {
      this.grpcClient = new GrpcWebImpl(this.baseUrl, this.GetClientConfig());
    }
    return this.grpcClient;
  };
}
