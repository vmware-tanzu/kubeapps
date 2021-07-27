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

  // core apis
  private packagesServiceClientImpl!: PackagesServiceClientImpl;
  private pluginsServiceClientImpl!: PluginsServiceClientImpl;

  // plugins package apis
  private helmPackagesServiceClientImpl!: HelmPackagesServiceClientImpl;
  private kappControllerPackagesServiceClientImpl!: KappControllerPackagesServiceClientImpl;
  private fluxv2PackagesServiceClientImpl!: FluxV2PackagesServiceClientImpl;

  constructor(transport?: grpc.TransportFactory) {
    this.transport = transport ?? grpc.CrossBrowserHttpTransport({ withCredentials: true });
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
    return this.packagesServiceClientImpl || new PackagesServiceClientImpl(this.getGrpcClient());
  }

  public getPluginsServiceClientImpl() {
    return this.pluginsServiceClientImpl || new PluginsServiceClientImpl(this.getGrpcClient());
  }

  // Plugins (packages) APIs
  public getHelmPackagesServiceClientImpl() {
    return (
      this.helmPackagesServiceClientImpl || new HelmPackagesServiceClientImpl(this.getGrpcClient())
    );
  }

  public getKappControllerPackagesServiceClientImpl() {
    return (
      this.kappControllerPackagesServiceClientImpl ||
      new KappControllerPackagesServiceClientImpl(this.getGrpcClient())
    );
  }

  public getFluxv2PackagesServiceClientImpl() {
    return (
      this.fluxv2PackagesServiceClientImpl ||
      new FluxV2PackagesServiceClientImpl(this.getGrpcClient())
    );
  }
}

export default new KubeappsGrpcClient();
