// Copyright 2018-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { LocationChangeAction, LOCATION_CHANGE } from "connected-react-router";
import { InstalledPackageDetailCustomDataHelm } from "gen/kubeappsapis/plugins/helm/packages/v1alpha1/helm";
import { IInstalledPackageState } from "shared/types";
import { getType } from "typesafe-actions";
import actions from "../actions";
import { InstalledPackagesAction } from "../actions/installedpackages";

export const initialState: IInstalledPackageState = {
  isFetching: false,
  items: [],
};

const installedPackagesReducer = (
  state: IInstalledPackageState = initialState,
  action: InstalledPackagesAction | LocationChangeAction,
): IInstalledPackageState => {
  switch (action.type) {
    case getType(actions.installedpackages.requestInstalledPackage):
      return {
        ...state,
        isFetching: true,
        selected: undefined,
        selectedDetails: undefined,
        resourceRefs: undefined,
      };
    case getType(actions.installedpackages.errorInstalledPackage):
      return { ...state, isFetching: false, error: action.payload };
    case getType(actions.installedpackages.clearErrorInstalledPackage):
      return { ...state, error: undefined };
    case getType(actions.installedpackages.selectInstalledPackage): {
      let revision: number;
      try {
        // TODO(agamez): verify why the field is not automatically decoded.
        revision = InstalledPackageDetailCustomDataHelm.decode(
          action.payload.pkg?.customDetail?.value as unknown as Uint8Array,
        ).releaseRevision;
      } catch (error) {
        // If the decoding fails, ignore it and just fall back to "no revisions"
        revision = 0;
      }
      return {
        ...state,
        isFetching: false,
        selected: {
          ...action.payload.pkg,
          // TODO(agamez): remove it once we have a core mechanism for rolling back
          revision: revision,
        },
        selectedDetails: action.payload.details,
      };
    }
    case getType(actions.installedpackages.receiveInstalledPackageStatus):
      if (state.selected) {
        return {
          ...state,
          selected: {
            ...state.selected!,
            revision: state.selected!.revision,
            status: action.payload,
          },
        };
      }
      return state;
    case getType(actions.installedpackages.requestInstalledPackageList):
      return { ...state, isFetching: true };
    case getType(actions.installedpackages.receiveInstalledPackageList):
      return { ...state, isFetching: false, listOverview: action.payload };
    case getType(actions.installedpackages.requestDeleteInstalledPackage):
      return { ...state, isFetching: true };
    case getType(actions.installedpackages.receiveDeleteInstalledPackage):
      return { ...state, isFetching: false };
    case getType(actions.installedpackages.requestInstallPackage):
      return { ...state, isFetching: true };
    case getType(actions.installedpackages.receiveInstallPackage):
      return { ...state, isFetching: false };
    case getType(actions.installedpackages.requestRollbackInstalledPackage):
      return { ...state, isFetching: true };
    case getType(actions.installedpackages.receiveRollbackInstalledPackage):
      return { ...state, isFetching: false };
    case LOCATION_CHANGE:
      return {
        ...state,
        error: undefined,
        isFetching: false,
        selected: undefined,
      };
    default:
  }
  return state;
};

export default installedPackagesReducer;
