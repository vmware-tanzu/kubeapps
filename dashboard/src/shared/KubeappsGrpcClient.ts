import { grpc } from "@improbable-eng/grpc-web";
import { PackagesServiceClientImpl } from "gen/kubeappsapis/core/packages/v1alpha1/packages";
import { PluginsServiceClientImpl } from "gen/kubeappsapis/core/plugins/v1alpha1/plugins";
import { FluxV2PackagesServiceClientImpl } from "gen/kubeappsapis/plugins/fluxv2/packages/v1alpha1/fluxv2";
import { HelmPackagesServiceClientImpl } from "gen/kubeappsapis/plugins/helm/packages/v1alpha1/helm";
import { KappControllerPackagesServiceClientImpl } from "gen/kubeappsapis/plugins/kapp_controller/packages/v1alpha1/kapp_controller";
import {
  GrpcWebImpl,
  ResourcesServiceClientImpl,
} from "gen/kubeappsapis/plugins/resources/v1alpha1/resources";
import { Auth } from "./Auth";
import * as URL from "./url";

export class KubeappsGrpcClient {
  private transport: grpc.TransportFactory;

  constructor(transport?: grpc.TransportFactory) {
    this.transport = transport ?? grpc.CrossBrowserHttpTransport({});
  }

  // getClientMetadata, if using token authentication, creates grpc metadata
  // and the token in the 'authorization' field
  public getClientMetadata(token?: string) {
    if (token) {
      return new grpc.Metadata({ authorization: `Bearer ${token}` });
    }

    return Auth.getAuthToken()
      ? new grpc.Metadata({ authorization: `Bearer ${Auth.getAuthToken()}` })
      : undefined;
  }

  // getGrpcClient returns a grpc client
  public getGrpcClient = (token?: string) => {
    return new GrpcWebImpl(URL.api.kubeappsapis, {
      transport: this.transport,
      metadata: this.getClientMetadata(token),
    });
  };

  // Core APIs
  public getPackagesServiceClientImpl() {
    return new PackagesServiceClientImpl(this.getGrpcClient());
  }

  public getPluginsServiceClientImpl() {
    return new PluginsServiceClientImpl(this.getGrpcClient());
  }

  // Resources API
  //
  // The resources API client implementation takes an optional token
  // only because it is used to validate token authentication before
  // the token is stored.
  public getResourcesServiceClientImpl(token?: string) {
    return new ResourcesServiceClientImpl(this.getGrpcClient(token));
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

export default KubeappsGrpcClient;
