// Copyright 2021-2024 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0
import type { ServiceType } from "@bufbuild/protobuf";
import { createPromiseClient } from "@connectrpc/connect";
import { createGrpcWebTransport } from "@connectrpc/connect-web";
import { Interceptor } from "@connectrpc/connect/dist/cjs/interceptor";
import { PromiseClient } from "@connectrpc/connect/dist/cjs/promise-client";
import { Transport } from "@connectrpc/connect/dist/cjs/transport";
import { PackagesService } from "gen/kubeappsapis/core/packages/v1alpha1/packages_connect";
import { RepositoriesService } from "gen/kubeappsapis/core/packages/v1alpha1/repositories_connect";
import { PluginsService } from "gen/kubeappsapis/core/plugins/v1alpha1/plugins_connect";
import {
  FluxV2PackagesService,
  FluxV2RepositoriesService,
} from "gen/kubeappsapis/plugins/fluxv2/packages/v1alpha1/fluxv2_connect";
import {
  HelmPackagesService,
  HelmRepositoriesService,
} from "gen/kubeappsapis/plugins/helm/packages/v1alpha1/helm_connect";
import {
  KappControllerPackagesService,
  KappControllerRepositoriesService,
} from "gen/kubeappsapis/plugins/kapp_controller/packages/v1alpha1/kapp_controller_connect";
import { ResourcesService } from "gen/kubeappsapis/plugins/resources/v1alpha1/resources_connect";

import { Auth } from "./Auth";
import * as URL from "./url";

export class KubeappsGrpcClient {
  private transport: Transport;

  // Creates a client with a transport, ensuring the transport includes the auth header.
  constructor(token?: string) {
    const auth: Interceptor = next => async req => {
      const t = token ? token : Auth.getAuthToken();
      if (t) {
        req.header.set("Authorization", `Bearer ${t}`);
      }
      return await next(req);
    };
    this.transport = createGrpcWebTransport({
      baseUrl: `./${URL.api.kubeappsapis}`,
      interceptors: [auth],
    });
  }

  // getClientMetadata, if using token authentication, creates grpc metadata
  // and the token in the 'authorization' field
  public getClientMetadata(token?: string) {
    const t = token ? token : Auth.getAuthToken();
    return t ? new Headers({ Authorization: `Bearer ${t}` }) : undefined;
  }

  public getGrpcClient = <T extends ServiceType>(service: T): PromiseClient<T> => {
    return createPromiseClient(service, this.transport);
  };

  // Core APIs
  public getPackagesServiceClientImpl() {
    return this.getGrpcClient(PackagesService);
  }

  public getRepositoriesServiceClientImpl() {
    return this.getGrpcClient(RepositoriesService);
  }

  public getPluginsServiceClientImpl() {
    return this.getGrpcClient(PluginsService);
  }

  // Resources API
  //
  // The resources API client implementation takes an optional token
  // only because it is used to validate token authentication before
  // the token is stored.
  // TODO: investigate the token here.
  public getResourcesServiceClientImpl() {
    return this.getGrpcClient(ResourcesService);
  }

  // Plugins (packages/repositories) APIs
  // TODO(agamez): ideally, these clients should be loaded automatically from a list of configured plugins

  // Helm
  public getHelmPackagesServiceClientImpl() {
    return this.getGrpcClient(HelmPackagesService);
  }
  public getHelmRepositoriesServiceClientImpl() {
    return this.getGrpcClient(HelmRepositoriesService);
  }

  // KappController
  public getKappControllerPackagesServiceClientImpl() {
    return this.getGrpcClient(KappControllerPackagesService);
  }
  public getKappControllerRepositoriesServiceClientImpl() {
    return this.getGrpcClient(KappControllerRepositoriesService);
  }
  // Fluxv2
  public getFluxv2PackagesServiceClientImpl() {
    return this.getGrpcClient(FluxV2PackagesService);
  }
  public getFluxV2RepositoriesServiceClientImpl() {
    return this.getGrpcClient(FluxV2RepositoriesService);
  }
}

export default KubeappsGrpcClient;
