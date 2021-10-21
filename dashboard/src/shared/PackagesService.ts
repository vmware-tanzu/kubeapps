import {
  AvailablePackageReference,
  GetAvailablePackageDetailResponse,
  GetAvailablePackageSummariesResponse,
  GetAvailablePackageVersionsResponse,
} from "gen/kubeappsapis/core/packages/v1alpha1/packages";
import { KubeappsGrpcClient } from "./KubeappsGrpcClient";

export default class PackagesService {
  public static client = () => new KubeappsGrpcClient().getPackagesServiceClientImpl();

  public static async getAvailablePackageSummaries(
    cluster: string,
    namespace: string,
    repos: string,
    page: number,
    size: number,
    query?: string,
  ): Promise<GetAvailablePackageSummariesResponse> {
    return await this.client().GetAvailablePackageSummaries({
      context: { cluster: cluster, namespace: namespace },
      filterOptions: {
        query: query,
        repositories: repos ? repos.split(",") : [],
      },
      paginationOptions: { pageSize: size, pageToken: page.toString() },
    });
  }

  public static async getAvailablePackageVersions(
    availablePackageReference?: AvailablePackageReference,
  ): Promise<GetAvailablePackageVersionsResponse> {
    return await this.client().GetAvailablePackageVersions({
      availablePackageRef: availablePackageReference,
    });
  }

  public static async getAvailablePackageDetail(
    availablePackageReference?: AvailablePackageReference,
    version?: string,
  ): Promise<GetAvailablePackageDetailResponse> {
    return await this.client().GetAvailablePackageDetail({
      pkgVersion: version,
      availablePackageRef: availablePackageReference,
    });
  }
}
