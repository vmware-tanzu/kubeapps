import * as yaml from "js-yaml";
import { ThunkAction } from "redux-thunk";
import { ActionType, createAction } from "typesafe-actions";
import { AppRepository } from "../shared/AppRepository";
import Chart from "../shared/Chart";
import { definedNamespaces } from "../shared/Namespace";
import Secret from "../shared/Secret";
import {
  IAppRepository,
  IAppRepositoryKey,
  ISecret,
  IStoreState,
  NotFoundError,
} from "../shared/types";
import { errorChart } from "./charts";

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

export const receiveReposSecrets = createAction("RECEIVE_REPOS_SECRETS", resolve => {
  return (secrets: ISecret[]) => resolve(secrets);
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

// Clear repo is basically receiving an empty repo
export const clearRepo = createAction("RECEIVE_REPO", resolve => {
  return () => resolve({} as IAppRepository);
});

export const showForm = createAction("SHOW_FORM");
export const hideForm = createAction("HIDE_FORM");
export const resetForm = createAction("RESET_FORM");
export const submitForm = createAction("SUBMIT_FROM");

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
  return (secret: ISecret) => resolve(secret);
});

const allActions = [
  addRepo,
  addedRepo,
  requestRepoUpdate,
  repoUpdated,
  repoValidating,
  repoValidated,
  clearRepo,
  errorRepos,
  requestRepos,
  receiveRepo,
  receiveRepos,
  receiveReposSecrets,
  receiveReposSecret,
  resetForm,
  errorChart,
  requestRepo,
  submitForm,
  showForm,
  hideForm,
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
      clusters: {
        currentCluster,
        clusters: {
          [currentCluster]: { currentNamespace },
        },
      },
    } = getState();
    try {
      await AppRepository.delete(name, namespace);
      dispatch(fetchRepos(currentNamespace));
      return true;
    } catch (e) {
      dispatch(errorRepos(e, "delete"));
      return false;
    }
  };
};

export const resyncRepo = (
  name: string,
  namespace: string,
): ThunkAction<Promise<void>, IStoreState, null, AppReposAction> => {
  return async dispatch => {
    try {
      await AppRepository.resync(name, namespace);
    } catch (e) {
      dispatch(errorRepos(e, "update"));
    }
  };
};

export const resyncAllRepos = (
  repos: IAppRepositoryKey[],
): ThunkAction<Promise<void>, IStoreState, null, AppReposAction> => {
  return async (dispatch, getState) => {
    repos.forEach(repo => {
      dispatch(resyncRepo(repo.name, repo.namespace));
    });
  };
};

export const fetchRepoSecrets = (
  namespace: string,
): ThunkAction<Promise<void>, IStoreState, null, AppReposAction> => {
  return async (dispatch, getState) => {
    const {
      clusters: { currentCluster },
    } = getState();
    // TODO(andresmgot): Create an endpoint for returning credentials related to an AppRepository
    // to avoid listing secrets
    // https://github.com/kubeapps/kubeapps/issues/1686
    const secrets = await Secret.list(currentCluster, namespace);
    const repoSecrets = secrets.items?.filter(s =>
      s.metadata.ownerReferences?.some(ownerRef => ownerRef.kind === "AppRepository"),
    );
    dispatch(receiveReposSecrets(repoSecrets));
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
    const secret = await Secret.get(currentCluster, name, namespace);
    dispatch(receiveReposSecret(secret));
  };
};

// fetchRepos fetches the AppRepositories in a specified namespace.
export const fetchRepos = (
  namespace: string,
  ...otherNamespaces: string[]
): ThunkAction<Promise<void>, IStoreState, null, AppReposAction> => {
  return async (dispatch, getState) => {
    try {
      dispatch(requestRepos(namespace));
      const repos = await AppRepository.list(namespace);
      dispatch(fetchRepoSecrets(namespace));
      if (!otherNamespaces || !otherNamespaces.length) {
        dispatch(receiveRepos(repos.items));
      } else {
        let totalRepos = repos.items;
        await Promise.all(
          otherNamespaces.map(async otherNamespace => {
            dispatch(requestRepos(otherNamespace));
            const otherRepos = await AppRepository.list(otherNamespace);
            totalRepos = totalRepos.concat(otherRepos.items);
          }),
        );
        dispatch(receiveRepos(totalRepos));
      }
    } catch (e) {
      dispatch(errorRepos(e, "fetch"));
    }
  };
};

function parsePodTemplate(syncJobPodTemplate: string) {
  let syncJobPodTemplateObj = {};
  if (syncJobPodTemplate.length) {
    syncJobPodTemplateObj = yaml.safeLoad(syncJobPodTemplate);
  }
  return syncJobPodTemplateObj;
}

function getTargetNS(getState: () => IStoreState, namespace: string) {
  let target = namespace;
  const {
    config: { kubeappsNamespace },
  } = getState();
  if (namespace === definedNamespaces.all) {
    target = kubeappsNamespace;
  }
  return target;
}

export const installRepo = (
  name: string,
  namespace: string,
  repoURL: string,
  authHeader: string,
  customCA: string,
  syncJobPodTemplate: string,
  registrySecrets: string[],
): ThunkAction<Promise<boolean>, IStoreState, null, AppReposAction> => {
  return async (dispatch, getState) => {
    try {
      const syncJobPodTemplateObj = parsePodTemplate(syncJobPodTemplate);
      const ns = getTargetNS(getState, namespace);
      dispatch(addRepo());
      const data = await AppRepository.create(
        name,
        ns,
        repoURL,
        authHeader,
        customCA,
        syncJobPodTemplateObj,
        registrySecrets,
      );
      dispatch(addedRepo(data.appRepository));

      return true;
    } catch (e) {
      dispatch(errorRepos(e, "create"));
      return false;
    }
  };
};

export const updateRepo = (
  name: string,
  namespace: string,
  repoURL: string,
  authHeader: string,
  customCA: string,
  syncJobPodTemplate: string,
  registrySecrets: string[],
): ThunkAction<Promise<boolean>, IStoreState, null, AppReposAction> => {
  return async (dispatch, getState) => {
    try {
      const syncJobPodTemplateObj = parsePodTemplate(syncJobPodTemplate);
      const ns = getTargetNS(getState, namespace);
      dispatch(requestRepoUpdate());
      const data = await AppRepository.update(
        name,
        ns,
        repoURL,
        authHeader,
        customCA,
        syncJobPodTemplateObj,
        registrySecrets,
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
    } catch (e) {
      dispatch(errorRepos(e, "update"));
      return false;
    }
  };
};

export const validateRepo = (
  repoURL: string,
  authHeader: string,
  customCA: string,
): ThunkAction<Promise<boolean>, IStoreState, null, AppReposAction> => {
  return async dispatch => {
    try {
      dispatch(repoValidating());
      const data = await AppRepository.validate(repoURL, authHeader, customCA);
      if (data.code === 200) {
        dispatch(repoValidated(data));
        return true;
      } else {
        dispatch(errorRepos(new Error(JSON.stringify(data)), "validate"));
        return false;
      }
    } catch (e) {
      dispatch(errorRepos(e, "validate"));
      return false;
    }
  };
};

export function checkChart(
  repoNamespace: string,
  repo: string,
  chartName: string,
): ThunkAction<Promise<boolean>, IStoreState, null, AppReposAction> {
  return async (dispatch, getState) => {
    dispatch(requestRepo());
    const appRepository = await AppRepository.get(repo, repoNamespace);
    try {
      await Chart.fetchChartVersions(repoNamespace, `${repo}/${chartName}`);
      dispatch(receiveRepo(appRepository));
      return true;
    } catch (e) {
      dispatch(
        errorChart(new NotFoundError(`Chart ${chartName} not found in the repository ${repo}.`)),
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
      const secrets = await Secret.list(currentCluster, namespace);
      const imgPullSecrets = secrets.items?.filter(
        s => s.type === "kubernetes.io/dockerconfigjson",
      );
      dispatch(receiveImagePullSecrets(imgPullSecrets));
    } catch (e) {
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
      const secret = await Secret.createPullSecret(
        currentCluster,
        name,
        user,
        password,
        email,
        server,
        namespace,
      );
      dispatch(createImagePullSecret(secret));
      return true;
    } catch (e) {
      dispatch(errorRepos(e, "fetch"));
      return false;
    }
  };
}
