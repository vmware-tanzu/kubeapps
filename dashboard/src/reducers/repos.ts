// Copyright 2018-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { LocationChangeAction, LOCATION_CHANGE } from "connected-react-router";
import {
  PackageRepositoryDetail,
  PackageRepositorySummary,
} from "gen/kubeappsapis/core/packages/v1alpha1/repositories";
import { HelmPackageRepositoryCustomDetail } from "gen/kubeappsapis/plugins/helm/packages/v1alpha1/helm";
import {
  KappControllerPackageRepositoryCustomDetail,
  PackageRepositoryFetch,
} from "gen/kubeappsapis/plugins/kapp_controller/packages/v1alpha1/kapp_controller";
import { PluginNames } from "shared/types";
import { getType } from "typesafe-actions";
import actions from "../actions";
import { PkgReposAction } from "../actions/repos";

export interface IPackageRepositoryState {
  errors: {
    create?: Error;
    delete?: Error;
    fetch?: Error;
    update?: Error;
  };
  isFetching: boolean;
  repoDetail: PackageRepositoryDetail;
  reposSummaries: PackageRepositorySummary[];
}

export const initialState: IPackageRepositoryState = {
  errors: {},
  isFetching: false,
  repoDetail: {} as PackageRepositoryDetail,
  reposSummaries: [] as PackageRepositorySummary[],
};

const helmPackageRepositoryCustomDetail = {
  imagesPullSecret: {
    secretRef: "",
  },
  ociRepositories: [],
  performValidation: false,
} as HelmPackageRepositoryCustomDetail;

const kappPackageRepositoryCustomDetail = {
  fetch: {} as PackageRepositoryFetch,
} as KappControllerPackageRepositoryCustomDetail;

const reposReducer = (
  state: IPackageRepositoryState = initialState,
  action: PkgReposAction | LocationChangeAction,
): IPackageRepositoryState => {
  switch (action.type) {
    case getType(actions.repos.receiveRepoSummaries):
      return {
        ...state,
        isFetching: false,
        reposSummaries: action.payload,
        errors: {},
      };
    case getType(actions.repos.receiveRepoDetail): {
      let customDetail: any;
      let repoWithCustomDetail = { ...action.payload };

      if (action.payload?.customDetail?.value) {
        switch (action.payload.packageRepoRef?.plugin?.name) {
          // handle the decoding of each plugin's customDetail
          case PluginNames.PACKAGES_HELM:
            customDetail = helmPackageRepositoryCustomDetail;
            try {
              customDetail = HelmPackageRepositoryCustomDetail.decode(
                action.payload.customDetail.value,
              );
              repoWithCustomDetail = { ...action.payload, customDetail };
            } catch (error) {
              repoWithCustomDetail = { ...action.payload };
            }
            break;
          case PluginNames.PACKAGES_KAPP:
            customDetail = kappPackageRepositoryCustomDetail;
            try {
              customDetail = KappControllerPackageRepositoryCustomDetail.decode(
                action.payload.customDetail.value,
              );
              repoWithCustomDetail = { ...action.payload, customDetail };
            } catch (error) {
              repoWithCustomDetail = { ...action.payload };
            }
            break;
          default:
            repoWithCustomDetail = { ...action.payload };
            break;
        }
      }
      return {
        ...state,
        isFetching: false,
        repoDetail: repoWithCustomDetail,
        errors: {},
      };
    }
    case getType(actions.repos.requestRepoSummaries):
    case getType(actions.repos.addOrUpdateRepo):
      return { ...state, isFetching: true };
    case getType(actions.repos.requestRepoDetail):
      return { ...state, repoDetail: initialState.repoDetail, errors: {} };
    case getType(actions.repos.addedRepo):
      return {
        ...state,
        isFetching: false,
        reposSummaries: [...state.reposSummaries, action.payload].sort((a, b) =>
          a.name.toLowerCase() > b.name.toLowerCase()
            ? 1
            : b.name.toLowerCase() > a.name.toLowerCase()
            ? -1
            : 0,
        ),
      };
    case getType(actions.repos.repoUpdated): {
      const updatedRepo = action.payload;
      const repos = state.reposSummaries.map(r =>
        r.name === updatedRepo.name &&
        r.packageRepoRef?.context?.namespace === updatedRepo.packageRepoRef?.context?.namespace
          ? updatedRepo
          : r,
      );
      return { ...state, isFetching: false, reposSummaries: repos };
    }
    case getType(actions.repos.errorRepos):
      return {
        ...state,
        // don't reset the fetch error
        errors: { fetch: state.errors.fetch, [action.payload.op]: action.payload.err },
        isFetching: false,
      };
    case LOCATION_CHANGE:
      return {
        ...state,
        errors: {},
        isFetching: false,
      };
    default:
      return state;
  }
};

export default reposReducer;
