// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import {
  AvailablePackageReference,
  GetAvailablePackageDetailResponse,
  GetAvailablePackageSummariesResponse,
  GetAvailablePackageVersionsResponse,
} from "gen/kubeappsapis/core/packages/v1alpha1/packages";
import { KubeappsGrpcClient } from "./KubeappsGrpcClient";
import { convertGrpcAuthError } from "./utils";

export default class PackagesService {
  public static packagesServiceClient = () =>
    new KubeappsGrpcClient().getPackagesServiceClientImpl();
  public static pluginsServiceClientImpl = () =>
    new KubeappsGrpcClient().getPluginsServiceClientImpl();

  public static async getAvailablePackageSummaries(
    cluster: string,
    namespace: string,
    repos: string,
    paginationToken: string,
    size: number,
    query?: string,
  ): Promise<GetAvailablePackageSummariesResponse> {
    return await this.packagesServiceClient()
      .GetAvailablePackageSummaries({
        context: { cluster: cluster, namespace: namespace },
        filterOptions: {
          query: query,
          repositories: repos ? repos.split(",") : [],
        },
        paginationOptions: { pageSize: size, pageToken: paginationToken },
      })
      .catch((e: any) => {
        throw convertGrpcAuthError(e);
      });
  }

  public static async getAvailablePackageVersions(
    availablePackageReference?: AvailablePackageReference,
  ): Promise<GetAvailablePackageVersionsResponse> {
    return await this.packagesServiceClient()
      .GetAvailablePackageVersions({
        availablePackageRef: availablePackageReference,
      })
      .catch((e: any) => {
        throw convertGrpcAuthError(e);
      });
  }

  public static async getAvailablePackageDetail(
    availablePackageReference?: AvailablePackageReference,
    version?: string,
  ): Promise<GetAvailablePackageDetailResponse> {
    return await this.packagesServiceClient()
      .GetAvailablePackageDetail({
        pkgVersion: version,
        availablePackageRef: availablePackageReference,
      })
      .catch((e: any) => {
        throw convertGrpcAuthError(e);
      });
  }
}
