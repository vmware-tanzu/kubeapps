// Copyright 2018-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import {
  AvailablePackageReference,
  InstalledPackageDetail,
} from "gen/kubeappsapis/core/packages/v1alpha1/packages";
import {
  PackageRepositoryDetail,
  PackageRepositoryReference,
  PackageRepositorySummary,
} from "gen/kubeappsapis/core/packages/v1alpha1/repositories";
import { uniqBy } from "lodash";
import { ThunkAction } from "redux-thunk";
import { PackageRepositoriesService } from "shared/PackageRepositoriesService";
import PackagesService from "shared/PackagesService";
import { IPkgRepoFormData, IStoreState, NotFoundError } from "shared/types";
import { PluginNames } from "shared/utils";
import { ActionType, deprecated } from "typesafe-actions";
import { createErrorPackage } from "./availablepackages";

const { createAction } = deprecated;
export const addRepo = createAction("ADD_REPO");
export const addedRepo = createAction("ADDED_REPO", resolve => {
  return (added: PackageRepositoryDetail) => resolve(added);
});

export const requestRepoUpdate = createAction("REQUEST_REPO_UPDATE");
export const repoUpdated = createAction("REPO_UPDATED", resolve => {
  return (updated: PackageRepositoryDetail) => resolve(updated);
});

export const requestRepoSummaries = createAction("REQUEST_REPOS", resolve => {
  return (namespace: string) => resolve(namespace);
});
export const receiveRepoSummaries = createAction("RECEIVE_REPOS", resolve => {
  return (repos: PackageRepositorySummary[]) => resolve(repos);
});
export const concatRepos = createAction("RECEIVE_REPOS", resolve => {
  return (repos: PackageRepositorySummary[]) => resolve(repos);
});

export const requestRepoDetail = createAction("REQUEST_REPO");
export const receiveRepoDetail = createAction("RECEIVE_REPO", resolve => {
  return (repo: PackageRepositoryDetail) => resolve(repo);
});

export const redirect = createAction("REDIRECT", resolve => {
  return (path: string) => resolve(path);
});

export const redirected = createAction("REDIRECTED");
export const errorRepos = createAction("ERROR_REPOS", resolve => {
  return (err: Error, op: "create" | "update" | "fetch" | "delete" | "validate") =>
    resolve({ err, op });
});

const allActions = [
  addRepo,
  addedRepo,
  requestRepoUpdate,
  repoUpdated,
  errorRepos,
  requestRepoSummaries,
  receiveRepoDetail,
  receiveRepoSummaries,
  createErrorPackage,
  requestRepoDetail,
  redirect,
  redirected,
];
export type PkgReposAction = ActionType<typeof allActions[number]>;

// fetchRepos fetches the PackageRepositories in a specified namespace.
export const fetchRepoSummaries = (
  namespace: string,
  listGlobal?: boolean,
): ThunkAction<Promise<void>, IStoreState, null, PkgReposAction> => {
  return async (dispatch, getState) => {
    const {
      clusters: { currentCluster },
      config: { globalReposNamespace },
    } = getState();
    try {
      dispatch(requestRepoSummaries(namespace));
      const repos = await PackageRepositoriesService.getPackageRepositorySummaries({
        cluster: currentCluster,
        namespace: namespace,
      });
      if (!listGlobal || namespace === globalReposNamespace) {
        dispatch(receiveRepoSummaries(repos.packageRepositorySummaries));
      } else {
        // Global repos need to be added
        let totalRepos = repos.packageRepositorySummaries;
        dispatch(requestRepoSummaries(globalReposNamespace));
        const globalRepos = await PackageRepositoriesService.getPackageRepositorySummaries({
          cluster: currentCluster,
          namespace: "",
        });
        // Avoid adding duplicated repos: if two repos have the same uid, filter out
        totalRepos = uniqBy(
          totalRepos.concat(globalRepos.packageRepositorySummaries),
          "packageRepoRef.identifier",
        );
        dispatch(receiveRepoSummaries(totalRepos));
      }
    } catch (e: any) {
      dispatch(errorRepos(e, "fetch"));
    }
  };
};

export const fetchRepo = (
  packageRepoRef: PackageRepositoryReference,
): ThunkAction<Promise<boolean>, IStoreState, null, PkgReposAction> => {
  return async dispatch => {
    try {
      dispatch(requestRepoDetail());
      // Check if we have enough data to retrieve the package manually (instead of using its own availablePackageRef)
      const getPackageRepositoryDetailResponse =
        await PackageRepositoriesService.getPackageRepositoryDetail(packageRepoRef);
      if (!getPackageRepositoryDetailResponse?.detail) {
        dispatch(
          errorRepos(
            new Error(
              `Can't get the repository: ${JSON.stringify(getPackageRepositoryDetailResponse)}`,
            ),
            "fetch",
          ),
        );
        return false;
      }
      dispatch(receiveRepoDetail(getPackageRepositoryDetailResponse.detail));
      return true;
    } catch (e: any) {
      dispatch(errorRepos(e, "fetch"));
      return false;
    }
  };
};

export const installRepo = (
  namespace: string,
  request: IPkgRepoFormData,
): ThunkAction<Promise<boolean>, IStoreState, null, PkgReposAction> => {
  return async (dispatch, getState) => {
    const {
      clusters: { currentCluster },
      config: { globalReposNamespace },
    } = getState();
    try {
      dispatch(addRepo());

      let namespaceScoped = namespace !== globalReposNamespace;
      // TODO(agamez): currently, flux doesn't support this value to be true
      if (request.plugin?.name === PluginNames.PACKAGES_FLUX) {
        namespaceScoped = false;
      }

      const addPackageRepositoryResponse = await PackageRepositoriesService.addPackageRepository(
        currentCluster,
        namespace,
        request,
        namespaceScoped,
      );
      // Ensure the repo have been created
      if (!addPackageRepositoryResponse?.packageRepoRef) {
        dispatch(
          errorRepos(
            new Error(
              `Can't create the repository: ${JSON.stringify(addPackageRepositoryResponse)}`,
            ),
            "create",
          ),
        );
        return false;
      }
      const getPackageRepositoryDetailResponse =
        await PackageRepositoriesService.getPackageRepositoryDetail(
          addPackageRepositoryResponse.packageRepoRef,
        );
      if (!getPackageRepositoryDetailResponse?.detail) {
        dispatch(
          errorRepos(
            new Error(
              `The repo wasn't created: ${JSON.stringify(getPackageRepositoryDetailResponse)}`,
            ),
            "create",
          ),
        );
        return false;
      }
      dispatch(addedRepo(getPackageRepositoryDetailResponse.detail));
      return true;
    } catch (e: any) {
      dispatch(errorRepos(e, "create"));
      return false;
    }
  };
};

export const updateRepo = (
  namespace: string,
  request: IPkgRepoFormData,
): ThunkAction<Promise<boolean>, IStoreState, null, PkgReposAction> => {
  return async (dispatch, getState) => {
    const {
      clusters: { currentCluster },
    } = getState();
    try {
      dispatch(requestRepoUpdate());
      const updatePackageRepositoryResponse =
        await PackageRepositoriesService.updatePackageRepository(
          currentCluster,
          namespace,
          request,
        );

      // Ensure the repo have been updated
      if (!updatePackageRepositoryResponse?.packageRepoRef) {
        dispatch(
          errorRepos(
            new Error(
              `Can't update the repository: ${JSON.stringify(updatePackageRepositoryResponse)}`,
            ),
            "update",
          ),
        );
        return false;
      }
      const getPackageRepositoryDetailResponse =
        await PackageRepositoriesService.getPackageRepositoryDetail(
          updatePackageRepositoryResponse.packageRepoRef,
        );
      if (!getPackageRepositoryDetailResponse?.detail) {
        dispatch(
          errorRepos(
            new Error(
              `The repo wasn't updated: ${JSON.stringify(getPackageRepositoryDetailResponse)}`,
            ),
            "update",
          ),
        );
        return false;
      }
      dispatch(repoUpdated(getPackageRepositoryDetailResponse.detail));
      return true;
    } catch (e: any) {
      dispatch(errorRepos(e, "update"));
      return false;
    }
  };
};

export const deleteRepo = (
  packageRepoRef: PackageRepositoryReference,
): ThunkAction<Promise<boolean>, IStoreState, null, PkgReposAction> => {
  return async dispatch => {
    try {
      await PackageRepositoriesService.deletePackageRepository(packageRepoRef);
      return true;
    } catch (e: any) {
      dispatch(errorRepos(e, "delete"));
      return false;
    }
  };
};

export const findPackageInRepo = (
  cluster: string,
  repoNamespace: string,
  repoName: string,
  app?: InstalledPackageDetail,
): ThunkAction<Promise<boolean>, IStoreState, null, PkgReposAction> => {
  return async dispatch => {
    dispatch(requestRepoDetail());
    // Check if we have enough data to retrieve the package manually (instead of using its own availablePackageRef)
    if (app?.availablePackageRef?.identifier && app?.availablePackageRef?.plugin) {
      const getPackageRepositoryDetailResponse =
        await PackageRepositoriesService.getPackageRepositoryDetail({
          identifier: repoName,
          context: { cluster, namespace: repoNamespace },
          plugin: app.availablePackageRef.plugin,
        });
      try {
        await PackagesService.getAvailablePackageVersions({
          context: { cluster: cluster, namespace: repoNamespace },
          plugin: app.availablePackageRef.plugin,
          identifier: app.availablePackageRef.identifier,
        } as AvailablePackageReference);
        if (!getPackageRepositoryDetailResponse?.detail) {
          dispatch(
            createErrorPackage(
              new NotFoundError(
                `Package ${app.availablePackageRef.identifier} not found in the repository ${repoNamespace}.`,
              ),
            ),
          );
          return false;
        }
        dispatch(receiveRepoDetail(getPackageRepositoryDetailResponse.detail));
        return true;
      } catch (e: any) {
        dispatch(
          createErrorPackage(
            new NotFoundError(
              `Package ${app.availablePackageRef.identifier} not found in the repository ${repoNamespace}.`,
            ),
          ),
        );
        return false;
      }
    } else {
      dispatch(
        createErrorPackage(
          new NotFoundError(
            `The installed application '${app?.name}' does not have any matching package in the repository '${repoName}'. Are you sure you installed this application from a repository?`,
          ),
        ),
      );
      return false;
    }
  };
};
