// Copyright 2018-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { JSONSchemaType } from "ajv";
import {
  AvailablePackageDetail,
  Context,
  InstalledPackageDetail,
  InstalledPackageReference,
  InstalledPackageStatus,
  InstalledPackageSummary,
  ReconciliationOptions,
  VersionReference,
} from "gen/kubeappsapis/core/packages/v1alpha1/packages";
import { ThunkAction } from "redux-thunk";
import PackagesService from "shared/PackagesService";
import {
  CreateError,
  DeleteError,
  FetchError,
  FetchWarning,
  IStoreState,
  RollbackError,
  UnprocessableEntityError,
  UpgradeError,
} from "shared/types";
import { getPluginsSupportingRollback } from "shared/utils";
import { ActionType, deprecated } from "typesafe-actions";
import { InstalledPackage } from "../shared/InstalledPackage";
import { validate } from "../shared/schema";
import { handleErrorAction } from "./auth";

const { createAction } = deprecated;

export const requestInstalledPackage = createAction("REQUEST_INSTALLED_PACKAGE");

export const requestInstalledPackageList = createAction("REQUEST_INSTALLED_PACKAGE_LIST");

export const receiveInstalledPackageList = createAction(
  "RECEIVE_INSTALLED_PACKAGE_LIST",
  resolve => {
    return (pkgs: InstalledPackageSummary[]) => resolve(pkgs);
  },
);

export const requestDeleteInstalledPackage = createAction("REQUEST_DELETE_INSTALLED_PACKAGE");

export const receiveDeleteInstalledPackage = createAction(
  "RECEIVE_DELETE_INSTALLED_PACKAGE_CONFIRMATION",
);

export const requestInstallPackage = createAction("REQUEST_INSTALL_PACKAGE");

export const receiveInstallPackage = createAction("RECEIVE_INSTALL_PACKAGE_CONFIRMATION");

export const requestUpdateInstalledPackage = createAction("REQUEST_UPDATE_INSTALLED_PACKAGE");

export const receiveUpdateInstalledPackage = createAction(
  "RECEIVE_UPDATE_INSTALLED_PACKAGE_CONFIRMATION",
);

export const requestRollbackInstalledPackage = createAction("REQUEST_ROLLBACK_INSTALLED_PACKAGE");

export const receiveRollbackInstalledPackage = createAction(
  "RECEIVE_ROLLBACK_INSTALLED_PACKAGE_CONFIRMATION",
);

export const requestInstalledPackageStatus = createAction("REQUEST_INSTALLED_PACKAGE_STATUS");

export const receiveInstalledPackageStatus = createAction(
  "RECEIVE_INSTALLED_PACKAGE_STATUS",
  resolve => {
    return (status: InstalledPackageStatus) => resolve(status);
  },
);

export const errorInstalledPackage = createAction("ERROR_INSTALLED_PACKAGE", resolve => {
  return (err: FetchError | CreateError | UpgradeError | RollbackError | DeleteError) =>
    resolve(err);
});

export const clearErrorInstalledPackage = createAction("CLEAR_ERROR_INSTALLED_PACKAGE");

export const selectInstalledPackage = createAction("SELECT_INSTALLED_PACKAGE", resolve => {
  return (pkg: InstalledPackageDetail, details?: AvailablePackageDetail) =>
    resolve({ pkg, details });
});

const allActions = [
  requestInstalledPackageList,
  requestInstalledPackage,
  receiveInstalledPackageList,
  requestInstalledPackageStatus,
  receiveInstalledPackageStatus,
  requestDeleteInstalledPackage,
  receiveDeleteInstalledPackage,
  requestInstallPackage,
  receiveInstallPackage,
  requestUpdateInstalledPackage,
  receiveUpdateInstalledPackage,
  requestRollbackInstalledPackage,
  receiveRollbackInstalledPackage,
  errorInstalledPackage,
  clearErrorInstalledPackage,
  selectInstalledPackage,
];

export type InstalledPackagesAction = ActionType<typeof allActions[number]>;

export function getInstalledPackage(
  installedPackageRef?: InstalledPackageReference,
): ThunkAction<Promise<void>, IStoreState, null, InstalledPackagesAction> {
  return async dispatch => {
    dispatch(requestInstalledPackage());
    try {
      // Get the details of an installed package
      const { installedPackageDetail } = await InstalledPackage.GetInstalledPackageDetail(
        installedPackageRef,
      );

      // For local packages with no references to any available packages (eg.a local package for development)
      // we aren't able to get the details, but still want to display the available data so far
      let availablePackageDetail;
      try {
        // Get the details of the available package that corresponds to the installed package
        const resp = await PackagesService.getAvailablePackageDetail(
          installedPackageDetail?.availablePackageRef,
          installedPackageDetail?.currentVersion?.pkgVersion,
        );
        availablePackageDetail = resp.availablePackageDetail;
      } catch (e: any) {
        dispatch(
          handleErrorAction(
            e,
            errorInstalledPackage(
              new FetchWarning(
                "this package has missing information, some actions might not be available.",
              ),
            ),
          ),
        );
      }
      dispatch(selectInstalledPackage(installedPackageDetail!, availablePackageDetail));
    } catch (e: any) {
      dispatch(
        handleErrorAction(
          e,
          errorInstalledPackage(new FetchError("Unable to get installed package", [e])),
        ),
      );
    }
  };
}

export function getInstalledPkgStatus(
  installedPackageRef?: InstalledPackageReference,
): ThunkAction<Promise<void>, IStoreState, null, InstalledPackagesAction> {
  return async (dispatch, getState) => {
    const {
      kube: { subscriptions },
    } = getState();
    const subscriptionId = `${installedPackageRef?.context?.cluster}/${installedPackageRef?.context?.namespace}/${installedPackageRef?.identifier}`;

    // only get the status if the subscription is active, otherwise the component would have been unmounted
    if (subscriptions[subscriptionId]) {
      dispatch(requestInstalledPackageStatus());
      try {
        // Get the details of an installed package for the status.
        const { installedPackageDetail } = await InstalledPackage.GetInstalledPackageDetail(
          installedPackageRef,
        );
        dispatch(receiveInstalledPackageStatus(installedPackageDetail!.status!));
      } catch (e: any) {
        dispatch(
          handleErrorAction(
            e,
            errorInstalledPackage(new FetchError("Unable to refresh installed package", [e])),
          ),
        );
      }
    }
  };
}

export function deleteInstalledPackage(
  installedPackageRef: InstalledPackageReference,
): ThunkAction<Promise<boolean>, IStoreState, null, InstalledPackagesAction> {
  return async dispatch => {
    dispatch(requestDeleteInstalledPackage());
    try {
      await InstalledPackage.DeleteInstalledPackage(installedPackageRef);
      dispatch(receiveDeleteInstalledPackage());
      return true;
    } catch (e: any) {
      dispatch(handleErrorAction(e, errorInstalledPackage(new DeleteError(e.message))));
      return false;
    }
  };
}

// fetchInstalledPackages returns a list of apps for other actions to compose on top of it
export function fetchInstalledPackages(
  cluster: string,
  namespace?: string,
): ThunkAction<Promise<InstalledPackageSummary[]>, IStoreState, null, InstalledPackagesAction> {
  return async dispatch => {
    dispatch(requestInstalledPackageList());
    let installedPackageSummaries: InstalledPackageSummary[];
    try {
      const res = await InstalledPackage.GetInstalledPackageSummaries(cluster, namespace);
      installedPackageSummaries = res?.installedPackageSummaries;

      dispatch(receiveInstalledPackageList(installedPackageSummaries));
      return installedPackageSummaries;
    } catch (e: any) {
      dispatch(
        handleErrorAction(e, errorInstalledPackage(new FetchError("Unable to list apps", [e]))),
      );
      return [];
    }
  };
}

export function installPackage(
  targetCluster: string,
  targetNamespace: string,
  availablePackageDetail: AvailablePackageDetail,
  releaseName: string,
  values?: string,
  schema?: JSONSchemaType<any>,
  reconciliationOptions?: ReconciliationOptions,
): ThunkAction<Promise<boolean>, IStoreState, null, InstalledPackagesAction> {
  return async dispatch => {
    dispatch(requestInstallPackage());
    try {
      if (values && schema) {
        const validation = validate(values, schema);
        if (!validation.valid) {
          const errorText =
            validation.errors &&
            validation.errors.map(e => `  - ${e.instancePath}: ${e.message}`).join("\n");
          throw new UnprocessableEntityError(
            `The given values don't match the required format. The following errors were found:\n${errorText}`,
          );
        }
      }
      if (
        availablePackageDetail?.availablePackageRef &&
        availablePackageDetail?.version?.pkgVersion
      ) {
        await InstalledPackage.CreateInstalledPackage(
          { cluster: targetCluster, namespace: targetNamespace } as Context,
          releaseName,
          availablePackageDetail.availablePackageRef,
          { version: availablePackageDetail.version.pkgVersion } as VersionReference,
          values,
          reconciliationOptions as ReconciliationOptions,
        );
        dispatch(receiveInstallPackage());
        return true;
      } else {
        dispatch(
          errorInstalledPackage(
            new CreateError("This package does not contain enough information to be installed"),
          ),
        );
        return false;
      }
    } catch (e: any) {
      dispatch(handleErrorAction(e, errorInstalledPackage(new CreateError(e.message))));
      return false;
    }
  };
}

export function updateInstalledPackage(
  installedPackageRef: InstalledPackageReference,
  availablePackageDetail: AvailablePackageDetail,
  values?: string,
  schema?: JSONSchemaType<any>,
): ThunkAction<Promise<boolean>, IStoreState, null, InstalledPackagesAction> {
  return async dispatch => {
    dispatch(requestUpdateInstalledPackage());
    try {
      if (values && schema) {
        const validation = validate(values, schema);
        if (!validation.valid) {
          const errorText =
            validation.errors &&
            validation.errors.map(e => `  - ${e.instancePath}: ${e.message}`).join("\n");
          throw new UnprocessableEntityError(
            `The given values don't match the required format. The following errors were found:\n${errorText}`,
          );
        }
      }
      if (availablePackageDetail?.version?.pkgVersion) {
        await InstalledPackage.UpdateInstalledPackage(
          installedPackageRef,
          { version: availablePackageDetail.version.pkgVersion } as VersionReference,
          values,
        );
        dispatch(receiveUpdateInstalledPackage());
        return true;
      } else {
        dispatch(
          errorInstalledPackage(
            new UpgradeError("This package does not contain enough information to be installed"),
          ),
        );
        return false;
      }
    } catch (e: any) {
      dispatch(handleErrorAction(e, errorInstalledPackage(new UpgradeError(e.message))));
      return false;
    }
  };
}

export function rollbackInstalledPackage(
  installedPackageRef: InstalledPackageReference,
  revision: number,
): ThunkAction<Promise<boolean>, IStoreState, null, InstalledPackagesAction> {
  return async dispatch => {
    // rollbackInstalledPackage is currently only available for Helm packages
    if (
      installedPackageRef?.plugin?.name &&
      getPluginsSupportingRollback().includes(installedPackageRef.plugin.name)
    ) {
      dispatch(requestRollbackInstalledPackage());
      try {
        await InstalledPackage.RollbackInstalledPackage(installedPackageRef, revision);
        dispatch(receiveRollbackInstalledPackage());
        dispatch(getInstalledPackage(installedPackageRef));
        return true;
      } catch (e: any) {
        dispatch(handleErrorAction(e, errorInstalledPackage(new RollbackError(e.message))));
        return false;
      }
    } else {
      dispatch(
        errorInstalledPackage(
          new RollbackError(
            "This package cannot be rolled back; this operation is only available for Helm packages",
          ),
        ),
      );
      return false;
    }
  };
}
