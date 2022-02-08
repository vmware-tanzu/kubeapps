// Copyright 2018-2022 the Kubeapps contributors.
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
} from "gen/kubeappsapis/core/packages/v1alpha1/packages";
import {
  RollbackInstalledPackageRequest,
  RollbackInstalledPackageResponse,
} from "gen/kubeappsapis/plugins/helm/packages/v1alpha1/helm";
import { KubeappsGrpcClient } from "./KubeappsGrpcClient";
import { getPluginsSupportingRollback } from "./utils";

export class InstalledPackage {
  public static coreClient = () => new KubeappsGrpcClient().getPackagesServiceClientImpl();
  public static helmPluginClient = () =>
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
    initialWait = 500,
  ) {
    // TODO(minelson): The backend plugin may take care of waiting
    // for the required data to become available in which case this
    // can be removed.
    // See https://github.com/kubeapps/kubeapps/issues/4213
    // Note: initialWait is set with a default value so it can be
    // tested with a value of 0 (because I couldn't get jest's mock
    // timers to work with promises here.)
    const fn = async () =>
      this.coreClient().GetInstalledPackageResourceRefs({
        installedPackageRef: installedPackageRef,
      });
    return await callWithRetry(fn, initialWait);
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

// Helpers to be able to call with a retry backing off exponentially
// with 500ms, 1000ms, 2000ms, 4000ms etc.
const wait = (ms: number) => new Promise(res => setTimeout(res, ms));

const callWithRetry = async (fn: any, initialWait: number, depth = 0): Promise<any> => {
  try {
    return await fn();
  } catch (e) {
    if (depth >= 4) {
      throw e;
    }
    // Wait for initialWait ms first time, doubling each time.
    await wait(initialWait * 2 ** depth);

    return callWithRetry(fn, initialWait, depth + 1);
  }
};
