// Copyright 2018-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import {
  AvailablePackageReference,
  InstalledPackageDetail,
} from "gen/kubeappsapis/core/packages/v1alpha1/packages";
import {
  PackageRepositoryAuth_PackageRepositoryAuthType,
  PackageRepositoryDetail,
  PackageRepositoryReference,
  PackageRepositorySummary,
} from "gen/kubeappsapis/core/packages/v1alpha1/repositories";
import { Plugin } from "gen/kubeappsapis/core/plugins/v1alpha1/plugins";
import * as yaml from "js-yaml";
import { uniqBy } from "lodash";
import { ThunkAction } from "redux-thunk";
import { PackageRepositoriesService } from "shared/PackageRepositoriesService";
import PackagesService from "shared/PackagesService";
import Secret from "shared/Secret";
import { IAppRepositoryFilter, IStoreState, NotFoundError } from "shared/types";
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

export const requestRepos = createAction("REQUEST_REPOS", resolve => {
  return (namespace: string) => resolve(namespace);
});
export const receiveRepos = createAction("RECEIVE_REPOS", resolve => {
  return (repos: PackageRepositorySummary[]) => resolve(repos);
});
export const concatRepos = createAction("RECEIVE_REPOS", resolve => {
  return (repos: PackageRepositorySummary[]) => resolve(repos);
});

export const requestRepo = createAction("REQUEST_REPO");
export const receiveRepo = createAction("RECEIVE_REPO", resolve => {
  return (repo: PackageRepositoryDetail) => resolve(repo);
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

export const createImagePullSecret = createAction("CREATE_IMAGE_PULL_SECRET", resolve => {
  return (secretName: string) => resolve(secretName);
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
  createImagePullSecret,
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
      const repos = await PackageRepositoriesService.getPackageRepositorySummaries(
        currentCluster,
        namespace,
      );
      if (!listGlobal || namespace === globalReposNamespace) {
        dispatch(receiveRepos(repos.packageRepositorySummaries));
      } else {
        // Global repos need to be added
        let totalRepos = repos.packageRepositorySummaries;
        dispatch(requestRepos(globalReposNamespace));
        const globalRepos = await PackageRepositoriesService.getPackageRepositorySummaries(
          currentCluster,
        );
        // Avoid adding duplicated repos: if two repos have the same uid, filter out
        totalRepos = uniqBy(
          totalRepos.concat(globalRepos.packageRepositorySummaries),
          "packageRepoRef.identifier",
        );
        dispatch(receiveRepos(totalRepos));
      }
    } catch (e: any) {
      dispatch(errorRepos(e, "fetch"));
    }
  };
};

export const fetchRepo = (
  packageRepoRef: PackageRepositoryReference,
): ThunkAction<Promise<boolean>, IStoreState, null, AppReposAction> => {
  return async dispatch => {
    try {
      dispatch(requestRepo());
      // Check if we have enough data to retrieve the package manually (instead of using its own availablePackageRef)
      const appRepository = await PackageRepositoriesService.getPackageRepositoryDetail(
        packageRepoRef,
      );
      if (!appRepository?.detail) {
        return false;
      }
      dispatch(receiveRepo(appRepository.detail));
      return true;
    } catch (e: any) {
      dispatch(errorRepos(e, "fetch"));
      return false;
    }
  };
};

export const installRepo = (
  name: string,
  plugin: Plugin,
  namespace: string,
  repoURL: string,
  type: string,
  description: string,
  authHeader: string,
  authRegCreds: string,
  customCA: string,
  syncJobPodTemplate: string,
  registrySecrets: string[],
  ociRepositories: string[],
  skipTLS: boolean,
  passCredentials: boolean,
  authMethod: PackageRepositoryAuth_PackageRepositoryAuthType,
  interval: number,
  filter?: IAppRepositoryFilter,
): ThunkAction<Promise<boolean>, IStoreState, null, AppReposAction> => {
  return async (dispatch, getState) => {
    const {
      clusters: { currentCluster },
      config: { globalReposNamespace },
    } = getState();
    try {
      const syncJobPodTemplateObj = parsePodTemplate(syncJobPodTemplate);
      dispatch(addRepo());
      const data = await PackageRepositoriesService.addPackageRepository(
        currentCluster,
        name,
        plugin,
        namespace,
        repoURL,
        type,
        description,
        authHeader,
        authRegCreds,
        customCA,
        syncJobPodTemplateObj,
        registrySecrets,
        ociRepositories,
        skipTLS,
        passCredentials,
        namespace !== globalReposNamespace,
        authMethod,
        interval,
        filter,
      );
      // Ensure the repo have been created
      if (!data?.packageRepoRef) {
        return false;
      }
      const repo = await PackageRepositoriesService.getPackageRepositoryDetail(data.packageRepoRef);
      if (!repo?.detail) {
        return false;
      }
      dispatch(addedRepo(repo.detail));
      return true;
    } catch (e: any) {
      dispatch(errorRepos(e, "create"));
      return false;
    }
  };
};

export const updateRepo = (
  name: string,
  plugin: Plugin,
  namespace: string,
  repoURL: string,
  type: string,
  description: string,
  authHeader: string,
  authRegCreds: string,
  customCA: string,
  syncJobPodTemplate: string,
  registrySecrets: string[],
  ociRepositories: string[],
  skipTLS: boolean,
  passCredentials: boolean,
  authMethod: PackageRepositoryAuth_PackageRepositoryAuthType,
  interval: number,
  filter?: IAppRepositoryFilter,
): ThunkAction<Promise<boolean>, IStoreState, null, AppReposAction> => {
  return async (dispatch, getState) => {
    const {
      clusters: { currentCluster },
    } = getState();
    try {
      const syncJobPodTemplateObj = parsePodTemplate(syncJobPodTemplate);
      dispatch(requestRepoUpdate());
      const data = await PackageRepositoriesService.updatePackageRepository(
        currentCluster,
        name,
        plugin,
        namespace,
        repoURL,
        type,
        description,
        authHeader,
        authRegCreds,
        customCA,
        syncJobPodTemplateObj,
        registrySecrets,
        ociRepositories,
        skipTLS,
        passCredentials,
        authMethod,
        interval,
        filter,
      );

      // Ensure the repo have been updated
      if (!data?.packageRepoRef) {
        return false;
      }
      const repo = await PackageRepositoriesService.getPackageRepositoryDetail(data.packageRepoRef);
      if (!repo?.detail) {
        return false;
      }
      dispatch(repoUpdated(repo.detail));
      return true;
    } catch (e: any) {
      dispatch(errorRepos(e, "update"));
      return false;
    }
  };
};

export const deleteRepo = (
  packageRepoRef: PackageRepositoryReference,
): ThunkAction<Promise<boolean>, IStoreState, null, AppReposAction> => {
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
): ThunkAction<Promise<boolean>, IStoreState, null, AppReposAction> => {
  return async dispatch => {
    dispatch(requestRepo());
    // Check if we have enough data to retrieve the package manually (instead of using its own availablePackageRef)
    if (app?.availablePackageRef?.identifier && app?.availablePackageRef?.plugin) {
      const appRepository = await PackageRepositoriesService.getPackageRepositoryDetail({
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
        if (!appRepository?.detail) {
          return false;
        }
        dispatch(receiveRepo(appRepository.detail));
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

// ............................... DEPRECATED ...............................

export const resyncRepo = (
  name: string,
  namespace: string,
): ThunkAction<Promise<void>, IStoreState, null, AppReposAction> => {
  return async (dispatch, getState) => {
    const {
      clusters: { currentCluster },
    } = getState();
    try {
      await PackageRepositoriesService.resync(currentCluster, namespace, name);
    } catch (e: any) {
      dispatch(errorRepos(e, "update"));
    }
  };
};

export const resyncAllRepos = (
  packageRepositoryReferences: (PackageRepositoryReference | undefined)[],
): ThunkAction<Promise<void>, IStoreState, null, AppReposAction> => {
  return async dispatch => {
    packageRepositoryReferences.forEach(ref => {
      if (ref) {
        dispatch(resyncRepo(ref.identifier, ref.context?.namespace || ""));
      }
    });
  };
};

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
      const data = await PackageRepositoriesService.validate(
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

export const createDockerRegistrySecret = (
  name: string,
  user: string,
  password: string,
  email: string,
  server: string,
  namespace: string,
): ThunkAction<Promise<boolean>, IStoreState, null, AppReposAction> => {
  return async (dispatch, getState) => {
    const {
      clusters: { currentCluster },
    } = getState();
    try {
      await Secret.createPullSecret(currentCluster, name, user, password, email, server, namespace);
      dispatch(createImagePullSecret(name));
      return true;
    } catch (e: any) {
      dispatch(errorRepos(e, "fetch"));
      return false;
    }
  };
};

function parsePodTemplate(syncJobPodTemplate: string) {
  let syncJobPodTemplateObj: any = {};
  if (syncJobPodTemplate.length) {
    syncJobPodTemplateObj = yaml.load(syncJobPodTemplate);
  }
  return syncJobPodTemplateObj;
}
