import { grpc } from "@improbable-eng/grpc-web";
import { PackagesServiceClientImpl } from "gen/kubeappsapis/core/packages/v1alpha1/packages";
import {
  GrpcWebImpl,
  PluginsServiceClientImpl,
} from "gen/kubeappsapis/core/plugins/v1alpha1/plugins";
import { FluxV2PackagesServiceClientImpl } from "gen/kubeappsapis/plugins/fluxv2/packages/v1alpha1/fluxv2";
import { HelmPackagesServiceClientImpl } from "gen/kubeappsapis/plugins/helm/packages/v1alpha1/helm";
import { KappControllerPackagesServiceClientImpl } from "gen/kubeappsapis/plugins/kapp_controller/packages/v1alpha1/kapp_controller";
import { Auth } from "./Auth";
import * as URL from "./url";

export class KubeappsGrpcClient {
  private grpcWebImpl!: GrpcWebImpl;
  private transport: grpc.TransportFactory;

  constructor(transport?: grpc.TransportFactory) {
    this.transport = transport ?? grpc.CrossBrowserHttpTransport({});
  }

  // getClientMetadata, if using token authentication, creates grpc metadata
  // and the token in the 'authorization' field
  public getClientMetadata() {
    return Auth.getAuthToken()
      ? new grpc.Metadata({ authorization: `Bearer ${Auth.getAuthToken()}` })
      : undefined;
  }

  // getGrpcClient returns the already configured grpcWebImpl
  // if uncreated, it is created and returned
  public getGrpcClient = () => {
    return (
      this.grpcWebImpl ||
      new GrpcWebImpl(URL.api.kubeappsapis, {
        transport: this.transport,
        metadata: this.getClientMetadata(),
      })
    );
  };

  // Core APIs
  public getPackagesServiceClientImpl() {
    return new PackagesServiceClientImpl(this.getGrpcClient());
  }

  public getPluginsServiceClientImpl() {
    return new PluginsServiceClientImpl(this.getGrpcClient());
  }

  // Plugins (packages) APIs
  // TODO(agamez): ideally, these clients should be loaded automatically from a list of configured plugins
  public getHelmPackagesServiceClientImpl() {
    return new HelmPackagesServiceClientImpl(this.getGrpcClient());
  }

  public getKappControllerPackagesServiceClientImpl() {
    return new KappControllerPackagesServiceClientImpl(this.getGrpcClient());
  }

  public getFluxv2PackagesServiceClientImpl() {
    return new FluxV2PackagesServiceClientImpl(this.getGrpcClient());
  }
}

export default new KubeappsGrpcClient();
