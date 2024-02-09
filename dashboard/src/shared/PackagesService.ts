// Copyright 2021-2024 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import {
  AvailablePackageReference,
  GetAvailablePackageDetailResponse,
  GetAvailablePackageMetadatasResponse,
  GetAvailablePackageSummariesResponse,
  GetAvailablePackageVersionsResponse,
} from "gen/kubeappsapis/core/packages/v1alpha1/packages_pb";
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
      .getAvailablePackageSummaries({
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
      .getAvailablePackageVersions({
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
      .getAvailablePackageDetail({
        pkgVersion: version,
        availablePackageRef: availablePackageReference,
      })
      .catch((e: any) => {
        throw convertGrpcAuthError(e);
      });
  }

  public static async getAvailablePackageMetadatas(
    availablePackageReference: AvailablePackageReference,
    pkgVersion: string,
  ): Promise<GetAvailablePackageMetadatasResponse> {
    return await this.packagesServiceClient().getAvailablePackageMetadatas({
      availablePackageRef: availablePackageReference,
      pkgVersion,
    });
  }
}
