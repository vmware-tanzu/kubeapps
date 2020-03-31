import * as yaml from "js-yaml";
import { ThunkAction } from "redux-thunk";
import { ActionType, createAction } from "typesafe-actions";
import { AppRepository } from "../shared/AppRepository";
import Chart from "../shared/Chart";
import { definedNamespaces } from "../shared/Namespace";
import { errorChart } from "./charts";

import { IAppRepository, IAppRepositoryKey, IStoreState, NotFoundError } from "../shared/types";

export const addRepo = createAction("ADD_REPO");
export const addedRepo = createAction("ADDED_REPO", resolve => {
  return (added: IAppRepository) => resolve(added);
});

export const requestRepos = createAction("REQUEST_REPOS", resolve => {
  return (namespace: string) => resolve(namespace);
});
export const receiveRepos = createAction("RECEIVE_REPOS", resolve => {
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

const allActions = [
  addRepo,
  addedRepo,
  repoValidating,
  repoValidated,
  clearRepo,
  errorRepos,
  requestRepos,
  receiveRepo,
  receiveRepos,
  resetForm,
  errorChart,
  requestRepo,
  submitForm,
  showForm,
  hideForm,
  redirect,
  redirected,
];
export type AppReposAction = ActionType<typeof allActions[number]>;

export const deleteRepo = (
  name: string,
  namespace: string,
): ThunkAction<Promise<boolean>, IStoreState, null, AppReposAction> => {
  return async (dispatch, getState) => {
    const {
      namespace: { current },
      config: { namespace: kubeappsNamespace, featureFlags },
    } = getState();
    try {
      await AppRepository.delete(name, namespace);
      const fetchFromNamespace: string = featureFlags.reposPerNamespace
        ? current
        : kubeappsNamespace;
      dispatch(fetchRepos(fetchFromNamespace));
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
      const repo = await AppRepository.get(name, namespace);
      repo.spec.resyncRequests = repo.spec.resyncRequests || 0;
      repo.spec.resyncRequests++;
      await AppRepository.update(name, namespace, repo);
      // TODO: Do something to show progress
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

// fetchRepos fetches the AppRepositories in a specified namespace.
export const fetchRepos = (
  namespace: string,
): ThunkAction<Promise<void>, IStoreState, null, AppReposAction> => {
  return async (dispatch, getState) => {
    try {
      dispatch(requestRepos(namespace));
      const repos = await AppRepository.list(namespace);
      dispatch(receiveRepos(repos.items));
    } catch (e) {
      dispatch(errorRepos(e, "fetch"));
    }
  };
};

export const installRepo = (
  name: string,
  namespace: string,
  repoURL: string,
  authHeader: string,
  customCA: string,
  syncJobPodTemplate: string,
): ThunkAction<Promise<boolean>, IStoreState, null, AppReposAction> => {
  return async (dispatch, getState) => {
    let syncJobPodTemplateObj = {};
    try {
      if (syncJobPodTemplate.length) {
        syncJobPodTemplateObj = yaml.safeLoad(syncJobPodTemplate);
      }
      const {
        config: { namespace: kubeappsNamespace },
      } = getState();
      if (namespace === definedNamespaces.all) {
        namespace = kubeappsNamespace;
      }
      dispatch(addRepo());
      const data = await AppRepository.create(
        name,
        namespace,
        repoURL,
        authHeader,
        customCA,
        syncJobPodTemplateObj,
      );
      dispatch(addedRepo(data.appRepository));

      return true;
    } catch (e) {
      dispatch(errorRepos(e, "create"));
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
      if (data === "OK") {
        dispatch(repoValidated(data));
        return true;
      } else {
        // Unexpected error
        dispatch(
          errorRepos(
            new Error(`Unable to parse validation response, got: ${JSON.stringify(data)}`),
            "validate",
          ),
        );
        return false;
      }
    } catch (e) {
      dispatch(errorRepos(e, "validate"));
      return false;
    }
  };
};

export function checkChart(
  repo: string,
  chartName: string,
): ThunkAction<Promise<boolean>, IStoreState, null, AppReposAction> {
  return async (dispatch, getState) => {
    const {
      config: { namespace },
    } = getState();
    dispatch(requestRepo());
    const appRepository = await AppRepository.get(repo, namespace);
    try {
      await Chart.fetchChartVersions(namespace, `${repo}/${chartName}`);
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
