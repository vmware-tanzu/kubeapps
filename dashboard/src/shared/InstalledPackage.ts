// Copyright 2018-2023 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import {
  AvailablePackageReference,
  Context,
  CreateInstalledPackageRequest,
  DeleteInstalledPackageRequest,
  InstalledPackageReference,
  ReconciliationOptions,
  UpdateInstalledPackageRequest,
  VersionReference,
} from "gen/kubeappsapis/core/packages/v1alpha1/packages_pb";
import {
  RollbackInstalledPackageRequest,
  RollbackInstalledPackageResponse,
} from "gen/kubeappsapis/plugins/helm/packages/v1alpha1/helm_pb";
import { KubeappsGrpcClient } from "./KubeappsGrpcClient";
import { convertGrpcAuthError, getPluginsSupportingRollback } from "./utils";

export class InstalledPackage {
  public static packagesServiceClient = () =>
    new KubeappsGrpcClient().getPackagesServiceClientImpl();
  public static helmPackagesServiceClient = () =>
    new KubeappsGrpcClient().getHelmPackagesServiceClientImpl();

  public static async GetInstalledPackageSummaries(
    cluster: string,
    namespace?: string,
    pageToken?: string,
    size?: number,
  ) {
    return await this.packagesServiceClient()
      .getInstalledPackageSummaries({
        context: { cluster: cluster, namespace: namespace },
        paginationOptions: { pageSize: size || 0, pageToken: pageToken || "" },
      })
      .catch((e: any) => {
        throw convertGrpcAuthError(e);
      });
  }

  public static async GetInstalledPackageDetail(installedPackageRef?: InstalledPackageReference) {
    return await this.packagesServiceClient()
      .getInstalledPackageDetail({
        installedPackageRef: installedPackageRef,
      })
      .catch((e: any) => {
        throw convertGrpcAuthError(e);
      });
  }

  public static async GetInstalledPackageResourceRefs(
    installedPackageRef?: InstalledPackageReference,
  ) {
    return await this.packagesServiceClient()
      .getInstalledPackageResourceRefs({ installedPackageRef })
      .catch((e: any) => {
        throw convertGrpcAuthError(e);
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
    return await this.packagesServiceClient()
      .createInstalledPackage({
        name,
        values,
        targetContext,
        availablePackageRef,
        pkgVersionReference,
        reconciliationOptions,
      } as CreateInstalledPackageRequest)
      .catch((e: any) => {
        throw convertGrpcAuthError(e);
      });
  }

  public static async UpdateInstalledPackage(
    installedPackageRef: InstalledPackageReference,
    pkgVersionReference: VersionReference,
    values?: string,
    reconciliationOptions?: ReconciliationOptions,
  ) {
    return await this.packagesServiceClient()
      .updateInstalledPackage({
        installedPackageRef,
        pkgVersionReference,
        values,
        reconciliationOptions,
      } as UpdateInstalledPackageRequest)
      .catch((e: any) => {
        throw convertGrpcAuthError(e);
      });
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
      return await this.helmPackagesServiceClient()
        .rollbackInstalledPackage({
          installedPackageRef,
          releaseRevision,
        } as RollbackInstalledPackageRequest)
        .catch((e: any) => {
          throw convertGrpcAuthError(e);
        });
    } else {
      return {} as RollbackInstalledPackageResponse;
    }
  }

  public static async DeleteInstalledPackage(installedPackageRef: InstalledPackageReference) {
    return await this.packagesServiceClient()
      .deleteInstalledPackage({
        installedPackageRef,
      } as DeleteInstalledPackageRequest)
      .catch((e: any) => {
        throw convertGrpcAuthError(e);
      });
  }
}
