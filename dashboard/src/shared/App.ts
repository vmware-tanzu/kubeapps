import {
  AvailablePackageDetail,
  InstalledPackageReference,
} from "gen/kubeappsapis/core/packages/v1alpha1/packages";
import * as url from "shared/url";
import { axiosWithAuth } from "./AxiosInstance";
import { KubeappsGrpcClient } from "./KubeappsGrpcClient";

export const KUBEOPS_ROOT_URL = "api/kubeops/v1";
export class App {
  private static client = () => new KubeappsGrpcClient().getPackagesServiceClientImpl();

  public static async GetInstalledPackageSummaries(
    cluster: string,
    namespace?: string,
    page?: number,
    size?: number,
  ) {
    return await this.client().GetInstalledPackageSummaries({
      context: { cluster: cluster, namespace: namespace },
      paginationOptions: { pageSize: size || 0, pageToken: page?.toString() || "0" },
    });
  }

  public static async GetInstalledPackageDetail(installedPackageRef?: InstalledPackageReference) {
    return await this.client().GetInstalledPackageDetail({
      installedPackageRef: installedPackageRef,
    });
  }

  public static async create(
    cluster: string,
    namespace: string,
    releaseName: string,
    availablePackageDetail: AvailablePackageDetail,
    values?: string,
  ) {
    const endpoint = url.kubeops.releases.list(cluster, namespace);
    const { data } = await axiosWithAuth.post(endpoint, {
      // TODO(agamez): get the repo name once available
      // https://github.com/kubeapps/kubeapps/issues/3165#issuecomment-884574732
      appRepositoryResourceName:
        availablePackageDetail.availablePackageRef?.identifier.split("/")[0],
      appRepositoryResourceNamespace:
        availablePackageDetail.availablePackageRef?.context?.namespace,
      chartName: decodeURIComponent(availablePackageDetail.name),
      releaseName,
      values,
      version: availablePackageDetail.version?.pkgVersion,
    });
    return data;
  }

  public static async upgrade(
    installedPackageRef: InstalledPackageReference,
    packageNamespace: string,
    availablePackageDetail: AvailablePackageDetail,
    values?: string,
  ) {
    const endpoint = url.kubeops.releases.get(
      installedPackageRef.context?.cluster ?? "",
      installedPackageRef.context?.namespace ?? "",
      installedPackageRef.identifier,
    );
    const { data } = await axiosWithAuth.put(endpoint, {
      appRepositoryResourceName:
        availablePackageDetail.availablePackageRef?.identifier.split("/")[0],
      appRepositoryResourceNamespace: packageNamespace,
      chartName: decodeURIComponent(availablePackageDetail.name),
      releaseName: installedPackageRef.identifier,
      values,
      version: availablePackageDetail.version?.pkgVersion,
    });
    return data;
  }

  public static async rollback(installedPackageRef: InstalledPackageReference, revision: number) {
    const endpoint = url.kubeops.releases.get(
      installedPackageRef.context?.cluster ?? "",
      installedPackageRef.context?.namespace ?? "",
      installedPackageRef.identifier,
    );
    const { data } = await axiosWithAuth.put(
      endpoint,
      {},
      {
        params: {
          action: "rollback",
          revision,
        },
      },
    );
    return data;
  }

  public static async delete(installedPackageRef: InstalledPackageReference, purge: boolean) {
    let endpoint = url.kubeops.releases.get(
      installedPackageRef.context?.cluster ?? "",
      installedPackageRef.context?.namespace ?? "",
      installedPackageRef.identifier,
    );
    if (purge) {
      endpoint += "?purge=true";
    }
    const { data } = await axiosWithAuth.delete(endpoint);
    return data;
  }

  // TODO(agamez): remove it once we return the generated resources as part of the InstalledPackageDetail.
  public static async getRelease(installedPackageRef?: InstalledPackageReference) {
    const { data } = await axiosWithAuth.get<{ data: { manifest: any } }>(
      url.kubeops.releases.get(
        installedPackageRef?.context?.cluster ?? "",
        installedPackageRef?.context?.namespace ?? "",
        installedPackageRef?.identifier ?? "",
      ),
    );
    return data.data;
  }
}
