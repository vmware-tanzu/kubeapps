import {
  AvailablePackageReference,
  InstalledPackageDetail,
} from "gen/kubeappsapis/core/packages/v1alpha1/packages";
import * as yaml from "js-yaml";
import { uniqBy } from "lodash";
import { ThunkAction } from "redux-thunk";
import { AppRepository } from "shared/AppRepository";
import PackagesService from "shared/PackagesService";
import Secret from "shared/Secret";
import {
  IAppRepository,
  IAppRepositoryFilter,
  IAppRepositoryKey,
  ISecret,
  IStoreState,
  NotFoundError,
} from "shared/types";
import { ActionType, deprecated } from "typesafe-actions";
import { createErrorPackage } from "./packages";

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

export const receiveReposSecret = createAction("RECEIVE_REPOS_SECRET", resolve => {
  return (secret: ISecret) => resolve(secret);
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

export const requestImagePullSecrets = createAction("REQUEST_IMAGE_PULL_SECRETS", resolve => {
  return (namespace: string) => resolve(namespace);
});
export const receiveImagePullSecrets = createAction("RECEIVE_IMAGE_PULL_SECRETS", resolve => {
  return (secrets: ISecret[]) => resolve(secrets);
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
  receiveReposSecret,
  createErrorPackage,
  requestRepo,
  redirect,
  redirected,
  requestImagePullSecrets,
  receiveImagePullSecrets,
  createImagePullSecret,
];
export type AppReposAction = ActionType<typeof allActions[number]>;

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

export const resyncRepo = (
  name: string,
  namespace: string,
): ThunkAction<Promise<void>, IStoreState, null, AppReposAction> => {
  return async (dispatch, getState) => {
    const {
      clusters: { currentCluster },
    } = getState();
    try {
      await AppRepository.resync(currentCluster, namespace, name);
    } catch (e: any) {
      dispatch(errorRepos(e, "update"));
    }
  };
};

export const resyncAllRepos = (
  repos: IAppRepositoryKey[],
): ThunkAction<Promise<void>, IStoreState, null, AppReposAction> => {
  return async dispatch => {
    repos.forEach(repo => {
      dispatch(resyncRepo(repo.name, repo.namespace));
    });
  };
};

export const fetchRepoSecret = (
  namespace: string,
  name: string,
): ThunkAction<Promise<void>, IStoreState, null, AppReposAction> => {
  return async (dispatch, getState) => {
    const {
      clusters: { currentCluster },
    } = getState();
    try {
      const secret = await Secret.get(currentCluster, namespace, name);
      dispatch(receiveReposSecret(secret));
    } catch (e: any) {
      dispatch(errorRepos(e, "fetch"));
    }
  };
};

// fetchRepos fetches the AppRepositories in a specified namespace.
export const fetchRepos = (
  namespace: string,
  listGlobal?: boolean,
): ThunkAction<Promise<void>, IStoreState, null, AppReposAction> => {
  return async (dispatch, getState) => {
    const {
      clusters: { currentCluster },
      config: { kubeappsNamespace },
    } = getState();
    try {
      dispatch(requestRepos(namespace));
      const repos = await AppRepository.list(currentCluster, namespace);
      if (!listGlobal || namespace === kubeappsNamespace) {
        dispatch(receiveRepos(repos.items));
      } else {
        let totalRepos = repos.items;
        dispatch(requestRepos(kubeappsNamespace));
        const globalRepos = await AppRepository.list(currentCluster, kubeappsNamespace);
        // Avoid adding duplicated repos: if two repos have the same uid, filter out
        totalRepos = uniqBy(totalRepos.concat(globalRepos.items), "metadata.uid");
        dispatch(receiveRepos(totalRepos));
      }
    } catch (e: any) {
      dispatch(errorRepos(e, "fetch"));
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

export const installRepo = (
  name: string,
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
  filter?: IAppRepositoryFilter,
): ThunkAction<Promise<boolean>, IStoreState, null, AppReposAction> => {
  return async (dispatch, getState) => {
    const {
      clusters: { currentCluster },
    } = getState();
    try {
      const syncJobPodTemplateObj = parsePodTemplate(syncJobPodTemplate);
      dispatch(addRepo());
      const data = await AppRepository.create(
        currentCluster,
        name,
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
        filter,
      );
      dispatch(addedRepo(data.appRepository));

      return true;
    } catch (e: any) {
      dispatch(errorRepos(e, "create"));
      return false;
    }
  };
};

export const updateRepo = (
  name: string,
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
  filter?: IAppRepositoryFilter,
): ThunkAction<Promise<boolean>, IStoreState, null, AppReposAction> => {
  return async (dispatch, getState) => {
    const {
      clusters: { currentCluster },
    } = getState();
    try {
      const syncJobPodTemplateObj = parsePodTemplate(syncJobPodTemplate);
      dispatch(requestRepoUpdate());
      const data = await AppRepository.update(
        currentCluster,
        name,
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
        filter,
      );
      dispatch(repoUpdated(data.appRepository));
      // Re-fetch the helm repo secret that could have been modified with the updated headers
      // so that if the user chooses to edit the app repo again, they will see the current value.
      if (data.appRepository.spec?.auth) {
        let secretName = "";
        if (data.appRepository.spec.auth.header) {
          secretName = data.appRepository.spec.auth.header.secretKeyRef.name;
          dispatch(fetchRepoSecret(namespace, secretName));
        }
        if (
          data.appRepository.spec.auth.customCA &&
          secretName !== data.appRepository.spec.auth.customCA.secretKeyRef.name
        ) {
          secretName = data.appRepository.spec.auth.customCA.secretKeyRef.name;
          dispatch(fetchRepoSecret(namespace, secretName));
        }
      }
      return true;
    } catch (e: any) {
      dispatch(errorRepos(e, "update"));
      return false;
    }
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

export function fetchImagePullSecrets(
  namespace: string,
): ThunkAction<Promise<void>, IStoreState, null, AppReposAction> {
  return async (dispatch, getState) => {
    const {
      clusters: { currentCluster },
    } = getState();
    try {
      dispatch(requestImagePullSecrets(namespace));
      // TODO(andresmgot): Create an endpoint for returning just the list of secret names
      // to avoid listing all the secrets with protected information
      // https://github.com/kubeapps/kubeapps/issues/1686
      const secrets = await Secret.list(
        currentCluster,
        namespace,
        "type=kubernetes.io/dockerconfigjson",
      );
      dispatch(receiveImagePullSecrets(secrets.items));
    } catch (e: any) {
      dispatch(errorRepos(e, "fetch"));
    }
  };
}

export function createDockerRegistrySecret(
  name: string,
  user: string,
  password: string,
  email: string,
  server: string,
  namespace: string,
): ThunkAction<Promise<boolean>, IStoreState, null, AppReposAction> {
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
}
