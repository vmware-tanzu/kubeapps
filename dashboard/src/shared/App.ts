import {
  AvailablePackageReference,
  Context,
  CreateInstalledPackageRequest,
  DeleteInstalledPackageRequest,
  InstalledPackageReference,
  ReconciliationOptions,
  UpdateInstalledPackageRequest,
  VersionReference,
} from "gen/kubeappsapis/core/packages/v1alpha1/packages";
import {
  RollbackInstalledPackageRequest,
  RollbackInstalledPackageResponse,
} from "gen/kubeappsapis/plugins/helm/packages/v1alpha1/helm";
import { KubeappsGrpcClient } from "./KubeappsGrpcClient";
import { getPluginsSupportingRollback } from "./utils";

export class App {
  private static coreClient = () => new KubeappsGrpcClient().getPackagesServiceClientImpl();
  private static helmPluginClient = () =>
    new KubeappsGrpcClient().getHelmPackagesServiceClientImpl();

  public static async GetInstalledPackageSummaries(
    cluster: string,
    namespace?: string,
    page?: number,
    size?: number,
  ) {
    return await this.coreClient().GetInstalledPackageSummaries({
      context: { cluster: cluster, namespace: namespace },
      paginationOptions: { pageSize: size || 0, pageToken: page?.toString() || "0" },
    });
  }

  public static async GetInstalledPackageDetail(installedPackageRef?: InstalledPackageReference) {
    return await this.coreClient().GetInstalledPackageDetail({
      installedPackageRef: installedPackageRef,
    });
  }

  public static async GetInstalledPackageResourceRefs(
    installedPackageRef?: InstalledPackageReference,
  ) {
    return await this.coreClient().GetInstalledPackageResourceRefs({
      installedPackageRef: installedPackageRef,
    });
  }

  public static async CreateInstalledPackage(
    targetContext: Context,
    name: string,
    availablePackageRef: AvailablePackageReference,
    pkgVersionReference: VersionReference,
    values?: string,
    reconciliationOptions?: ReconciliationOptions,
  ) {
    return await this.coreClient().CreateInstalledPackage({
      name,
      values,
      targetContext,
      availablePackageRef,
      pkgVersionReference,
      reconciliationOptions,
    } as CreateInstalledPackageRequest);
  }

  public static async UpdateInstalledPackage(
    installedPackageRef: InstalledPackageReference,
    pkgVersionReference: VersionReference,
    values?: string,
    reconciliationOptions?: ReconciliationOptions,
  ) {
    return await this.coreClient().UpdateInstalledPackage({
      installedPackageRef,
      pkgVersionReference,
      values,
      reconciliationOptions,
    } as UpdateInstalledPackageRequest);
  }

  public static async RollbackInstalledPackage(
    installedPackageRef: InstalledPackageReference,
    releaseRevision: number,
  ) {
    // rollbackInstalledPackage is currently only available for Helm packages
    if (
      installedPackageRef?.plugin?.name &&
      getPluginsSupportingRollback().includes(installedPackageRef.plugin.name)
    ) {
      return await this.helmPluginClient().RollbackInstalledPackage({
        installedPackageRef,
        releaseRevision,
      } as RollbackInstalledPackageRequest);
    } else {
      return {} as RollbackInstalledPackageResponse;
    }
  }

  public static async DeleteInstalledPackage(installedPackageRef: InstalledPackageReference) {
    return await this.coreClient().DeleteInstalledPackage({
      installedPackageRef,
    } as DeleteInstalledPackageRequest);
  }
}
