// Copyright 2018-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import {
  AvailablePackageReference,
  InstalledPackageDetail,
} from "gen/kubeappsapis/core/packages/v1alpha1/packages";
import { uniqBy } from "lodash";
import { ThunkAction } from "redux-thunk";
import { AppRepository } from "shared/AppRepository";
import PackagesService from "shared/PackagesService";
import { IAppRepository, IPkgRepoFormData, IStoreState, NotFoundError } from "shared/types";
import { ActionType, deprecated } from "typesafe-actions";
import { createErrorPackage } from "./availablepackages";

const { createAction } = deprecated;
export const addRepo = createAction("ADD_REPO");
export const addedRepo = createAction("ADDED_REPO", resolve => {
  return (added: IAppRepository) => resolve(added);
});

export const requestRepoUpdate = createAction("REQUEST_REPO_UPDATE");
export const repoUpdated = createAction("REPO_UPDATED", resolve => {
  return (updated: IAppRepository) => resolve(updated);
});

export const requestRepos = createAction("REQUEST_REPOS", resolve => {
  return (namespace: string) => resolve(namespace);
});
export const receiveRepos = createAction("RECEIVE_REPOS", resolve => {
  return (repos: IAppRepository[]) => resolve(repos);
});
export const concatRepos = createAction("RECEIVE_REPOS", resolve => {
  return (repos: IAppRepository[]) => resolve(repos);
});

export const requestRepo = createAction("REQUEST_REPO");
export const receiveRepo = createAction("RECEIVE_REPO", resolve => {
  return (repo: IAppRepository) => resolve(repo);
});

export const repoValidating = createAction("REPO_VALIDATING");
export const repoValidated = createAction("REPO_VALIDATED", resolve => {
  return (data: any) => resolve(data);
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
  repoValidating,
  repoValidated,
  errorRepos,
  requestRepos,
  receiveRepo,
  receiveRepos,
  createErrorPackage,
  requestRepo,
  redirect,
  redirected,
];
export type AppReposAction = ActionType<typeof allActions[number]>;

// fetchRepos fetches the AppRepositories in a specified namespace.
export const fetchRepos = (
  namespace: string,
  listGlobal?: boolean,
): ThunkAction<Promise<void>, IStoreState, null, AppReposAction> => {
  return async (dispatch, getState) => {
    const {
      clusters: { currentCluster },
      config: { globalReposNamespace },
    } = getState();
    try {
      dispatch(requestRepos(namespace));
      const repos = await AppRepository.list(currentCluster, namespace);
      if (!listGlobal || namespace === globalReposNamespace) {
        dispatch(receiveRepos(repos.items));
      } else {
        // Global repos need to be added
        let totalRepos = repos.items;
        dispatch(requestRepos(globalReposNamespace));
        const globalRepos = await AppRepository.list(currentCluster, globalReposNamespace);
        // Avoid adding duplicated repos: if two repos have the same uid, filter out
        totalRepos = uniqBy(totalRepos.concat(globalRepos.items), "metadata.uid");
        dispatch(receiveRepos(totalRepos));
      }
    } catch (e: any) {
      dispatch(errorRepos(e, "fetch"));
    }
  };
};

export const fetchRepo = (
  cluster: string,
  repoNamespace: string,
  repoName: string,
): ThunkAction<Promise<boolean>, IStoreState, null, AppReposAction> => {
  return async dispatch => {
    try {
      dispatch(requestRepo());
      const appRepository = await AppRepository.get(cluster, repoNamespace, repoName);
      if (!appRepository) {
        dispatch(
          errorRepos(
            new Error(`Can't get the repository: ${JSON.stringify(appRepository)}`),
            "fetch",
          ),
        );
        return false;
      }
      dispatch(receiveRepo(appRepository.data));
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
): ThunkAction<Promise<boolean>, IStoreState, null, AppReposAction> => {
  return async (dispatch, getState) => {
    const {
      clusters: { currentCluster },
    } = getState();
    try {
      dispatch(addRepo());
      const data = await AppRepository.create(currentCluster, namespace, request);
      dispatch(addedRepo(data.appRepository));

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
): ThunkAction<Promise<boolean>, IStoreState, null, AppReposAction> => {
  return async (dispatch, getState) => {
    const {
      clusters: { currentCluster },
    } = getState();
    try {
      dispatch(requestRepoUpdate());
      const data = await AppRepository.update(currentCluster, namespace, request);
      dispatch(repoUpdated(data.appRepository));
      return true;
    } catch (e: any) {
      dispatch(errorRepos(e, "update"));
      return false;
    }
  };
};

export const deleteRepo = (
  name: string,
  namespace: string,
): ThunkAction<Promise<boolean>, IStoreState, null, AppReposAction> => {
  return async (dispatch, getState) => {
    const {
      clusters: { currentCluster },
    } = getState();
    try {
      await AppRepository.delete(currentCluster, namespace, name);
      return true;
    } catch (e: any) {
      dispatch(errorRepos(e, "delete"));
      return false;
    }
  };
};

export function findPackageInRepo(
  cluster: string,
  repoNamespace: string,
  repoName: string,
  app?: InstalledPackageDetail,
): ThunkAction<Promise<boolean>, IStoreState, null, AppReposAction> {
  return async dispatch => {
    dispatch(requestRepo());
    // Check if we have enough data to retrieve the package manually (instead of using its own availablePackageRef)
    if (app?.availablePackageRef?.identifier && app?.availablePackageRef?.plugin) {
      const appRepository = await AppRepository.get(cluster, repoNamespace, repoName);
      try {
        await PackagesService.getAvailablePackageVersions({
          context: { cluster: cluster, namespace: repoNamespace },
          plugin: app.availablePackageRef.plugin,
          identifier: app.availablePackageRef.identifier,
        } as AvailablePackageReference);
        dispatch(receiveRepo(appRepository));
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
}

// to be deprecated
export const validateRepo = (
  repoURL: string,
  type: string,
  authHeader: string,
  authRegCreds: string,
  customCA: string,
  ociRepositories: string[],
  skipTLS: boolean,
  passCredentials: boolean,
): ThunkAction<Promise<boolean>, IStoreState, null, AppReposAction> => {
  return async (dispatch, getState) => {
    const {
      clusters: { currentCluster, clusters },
    } = getState();
    const namespace = clusters[currentCluster].currentNamespace;
    try {
      dispatch(repoValidating());
      const data = await AppRepository.validate(
        currentCluster,
        namespace,
        repoURL,
        type,
        authHeader,
        authRegCreds,
        customCA,
        ociRepositories,
        skipTLS,
        passCredentials,
      );
      if (data.code === 200) {
        dispatch(repoValidated(data));
        return true;
      } else {
        dispatch(errorRepos(new Error(JSON.stringify(data)), "validate"));
        return false;
      }
    } catch (e: any) {
      dispatch(errorRepos(e, "validate"));
      return false;
    }
  };
};
