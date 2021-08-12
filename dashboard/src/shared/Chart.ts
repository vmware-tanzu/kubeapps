import {
  GetAvailablePackageDetailResponse,
  GetAvailablePackageSummariesResponse,
  GetAvailablePackageVersionsResponse,
} from "gen/kubeappsapis/core/packages/v1alpha1/packages";
import { KubeappsGrpcClient } from "./KubeappsGrpcClient";

export default class Chart {
  // TODO(agamez): move to the core 'PackagesServiceClientImpl' when pagination is ready there
  private static client = new KubeappsGrpcClient().getHelmPackagesServiceClientImpl();

  public static async getAvailablePackageSummaries(
    cluster: string,
    namespace: string,
    repos: string,
    page: number,
    size: number,
    query?: string,
  ): Promise<GetAvailablePackageSummariesResponse> {
    return await this.client.GetAvailablePackageSummaries({
      // TODO(agamez): add cluster when it is supported
      context: { cluster: "", namespace: namespace },
      filterOptions: {
        query: query,
        repositories: repos.split(","),
      },
      paginationOptions: { pageSize: size, pageToken: page.toString() },
    });
  }

  public static async getAvailablePackageVersions(
    cluster: string,
    namespace: string,
    id: string,
  ): Promise<GetAvailablePackageVersionsResponse> {
    return await this.client.GetAvailablePackageVersions({
      availablePackageRef: {
        // TODO(agamez): add cluster when it is supported
        context: { cluster: "", namespace: namespace },
        identifier: id,
      },
    });
  }

  public static async getAvailablePackageDetail(
    cluster: string,
    namespace: string,
    id: string,
    version?: string,
  ): Promise<GetAvailablePackageDetailResponse> {
    return await this.client.GetAvailablePackageDetail({
      pkgVersion: version,
      availablePackageRef: {
        // TODO(agamez): add cluster when it is supported
        context: { cluster: "", namespace: namespace },
        identifier: id,
      },
    });
  }
}
