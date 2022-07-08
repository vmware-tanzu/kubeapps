// Copyright 2018-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { LocationChangeAction, LOCATION_CHANGE } from "connected-react-router";
import {
  PackageRepositoryDetail,
  PackageRepositorySummary,
} from "gen/kubeappsapis/core/packages/v1alpha1/repositories";
import { RepositoryCustomDetails } from "gen/kubeappsapis/plugins/helm/packages/v1alpha1/helm";
import { ISecret } from "shared/types";
import { PluginNames } from "shared/utils";
import { getType } from "typesafe-actions";
import actions from "../actions";
import { PkgReposAction } from "../actions/repos";

export interface IPackageRepositoryState {
  addingRepo: boolean;
  errors: {
    create?: Error;
    delete?: Error;
    fetch?: Error;
    update?: Error;
    validate?: Error;
  };
  lastAdded?: PackageRepositoryDetail;
  isFetching: boolean;
  isFetchingElem: {
    repositories: boolean;
    secrets: boolean;
  };
  validating: boolean;
  repo: PackageRepositoryDetail;
  repos: PackageRepositorySummary[];
  form: {
    name: string;
    namespace: string;
    url: string;
    show: boolean;
  };
  imagePullSecrets: ISecret[];
  redirectTo?: string;
}

export const initialState: IPackageRepositoryState = {
  addingRepo: false,
  errors: {},
  form: {
    name: "",
    namespace: "",
    show: false,
    url: "",
  },
  isFetching: false,
  isFetchingElem: {
    repositories: false,
    secrets: false,
  },
  validating: false,
  repo: {} as PackageRepositoryDetail,
  repos: [] as PackageRepositorySummary[],
  imagePullSecrets: [],
};

function isFetching(state: IPackageRepositoryState, item: string, fetching: boolean) {
  const composedIsFetching = {
    ...state.isFetchingElem,
    [item]: fetching,
  };
  return {
    isFetching: Object.values(composedIsFetching).some(v => v),
    isFetchingElem: composedIsFetching,
  };
}

const reposReducer = (
  state: IPackageRepositoryState = initialState,
  action: PkgReposAction | LocationChangeAction,
): IPackageRepositoryState => {
  switch (action.type) {
    case getType(actions.repos.receiveRepoSummaries):
      return {
        ...state,
        ...isFetching(state, "repositories", false),
        repos: action.payload,
        errors: {},
      };
    case getType(actions.repos.receiveRepoDetail):
      // eslint-disable-next-line no-case-declarations
      let customDetail: any;

      // TODO(agamez): decoding customDetail just for the helm plugin
      if (action.payload.packageRepoRef?.plugin?.name === PluginNames.PACKAGES_HELM) {
        customDetail = {
          dockerRegistrySecrets: [],
          ociRepositories: [],
          performValidation: false,
        } as RepositoryCustomDetails;

        try {
          if (action.payload?.customDetail?.value) {
            // TODO(agamez): verify why the field is not automatically decoded.
            customDetail = RepositoryCustomDetails.decode(
              action.payload.customDetail.value as unknown as Uint8Array,
            );
          }
          // eslint-disable-next-line no-empty
        } catch (error) {}
      }

      return {
        ...state,
        ...isFetching(state, "repositories", false),
        repo: { ...action.payload, customDetail: customDetail },
        errors: {},
      };
    case getType(actions.repos.requestRepoSummaries):
      return { ...state, ...isFetching(state, "repositories", true) };
    case getType(actions.repos.requestRepoDetail):
      return { ...state, repo: initialState.repo, errors: {} };
    case getType(actions.repos.addRepo):
      return { ...state, addingRepo: true };
    case getType(actions.repos.addedRepo):
      return {
        ...state,
        addingRepo: false,
        repos: [...state.repos, action.payload].sort((a, b) =>
          a.name.toLowerCase() > b.name.toLowerCase()
            ? 1
            : b.name.toLowerCase() > a.name.toLowerCase()
            ? -1
            : 0,
        ),
      };
    case getType(actions.repos.repoUpdated): {
      const updatedRepo = action.payload;
      const repos = state.repos.map(r =>
        r.name === updatedRepo.name &&
        r.packageRepoRef?.context?.namespace === updatedRepo.packageRepoRef?.context?.namespace
          ? updatedRepo
          : r,
      );
      return { ...state, repos };
    }
    case getType(actions.repos.redirect):
      return { ...state, redirectTo: action.payload };
    case getType(actions.repos.redirected):
      return { ...state, redirectTo: undefined };
    case getType(actions.repos.errorRepos):
      return {
        ...state,
        // don't reset the fetch error
        errors: { fetch: state.errors.fetch, [action.payload.op]: action.payload.err },
        isFetching: false,
        isFetchingElem: {
          repositories: false,
          secrets: false,
        },
        validating: false,
      };
    case LOCATION_CHANGE:
      return {
        ...state,
        errors: {},
        isFetching: false,
        isFetchingElem: {
          repositories: false,
          secrets: false,
        },
      };
    default:
      return state;
  }
};

export default reposReducer;
