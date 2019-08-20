import * as yaml from "js-yaml";
import { ThunkAction } from "redux-thunk";
import { ActionType, createAction } from "typesafe-actions";
import { AppRepository } from "../shared/AppRepository";
import { axios } from "../shared/AxiosInstance";
import Secret from "../shared/Secret";
import * as url from "../shared/url";
import { errorChart } from "./charts";

import { IAppRepository, IOwnerReference, IStoreState, NotFoundError } from "../shared/types";

export const addRepo = createAction("ADD_REPO");
export const addedRepo = createAction("ADDED_REPO", resolve => {
  return (added: IAppRepository) => resolve(added);
});

export const requestRepos = createAction("REQUEST_REPOS");
export const receiveRepos = createAction("RECEIVE_REPOS", resolve => {
  return (repos: IAppRepository[]) => resolve(repos);
});

export const requestRepo = createAction("REQUEST_REPO");
export const receiveRepo = createAction("RECEIVE_REPO", resolve => {
  return (repo: IAppRepository) => resolve(repo);
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
  return (err: Error, op: "create" | "update" | "fetch" | "delete") => resolve({ err, op });
});

const allActions = [
  addRepo,
  addedRepo,
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
): ThunkAction<Promise<boolean>, IStoreState, null, AppReposAction> => {
  return async (dispatch, getState) => {
    try {
      const {
        config: { namespace },
      } = getState();
      await AppRepository.delete(name, namespace);
      dispatch(fetchRepos());
      return true;
    } catch (e) {
      dispatch(errorRepos(e, "delete"));
      return false;
    }
  };
};

export const resyncRepo = (
  name: string,
): ThunkAction<Promise<void>, IStoreState, null, AppReposAction> => {
  return async (dispatch, getState) => {
    try {
      const {
        config: { namespace },
      } = getState();
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
  repoNames: string[],
): ThunkAction<Promise<void>, IStoreState, null, AppReposAction> => {
  return async (dispatch, getState) => {
    repoNames.forEach(name => {
      dispatch(resyncRepo(name));
    });
  };
};

export const fetchRepos = (): ThunkAction<Promise<void>, IStoreState, null, AppReposAction> => {
  return async (dispatch, getState) => {
    dispatch(requestRepos());
    try {
      const {
        config: { namespace },
      } = getState();
      const repos = await AppRepository.list(namespace);
      dispatch(receiveRepos(repos.items));
    } catch (e) {
      dispatch(errorRepos(e, "fetch"));
    }
  };
};

export const installRepo = (
  name: string,
  repoURL: string,
  authHeader: string,
  customCA: string,
  syncJobPodTemplate: string,
): ThunkAction<Promise<boolean>, IStoreState, null, AppReposAction> => {
  return async (dispatch, getState) => {
    try {
      const {
        config: { namespace },
      } = getState();
      interface ISecretKeyRef {
        key: string;
        name: string;
      }
      const auth: {
        header?: { secretKeyRef: ISecretKeyRef };
        customCA?: { secretKeyRef: ISecretKeyRef };
      } = {};
      const secrets: { [s: string]: string } = {};
      const secretName = `apprepo-${name}-secrets`;
      if (authHeader.length || customCA.length) {
        // ensure we can create secrets in the kubeapps namespace
        if (authHeader.length) {
          auth.header = {
            secretKeyRef: {
              key: "authorizationHeader",
              name: secretName,
            },
          };
          secrets.authorizationHeader = btoa(authHeader);
        }
        if (customCA.length) {
          auth.customCA = {
            secretKeyRef: {
              key: "ca.crt",
              name: secretName,
            },
          };
          secrets["ca.crt"] = btoa(customCA);
        }
      }
      let syncJobPodTemplateObj = {};
      if (syncJobPodTemplate.length) {
        syncJobPodTemplateObj = yaml.safeLoad(syncJobPodTemplate);
      }
      dispatch(addRepo());
      const apprepo = await AppRepository.create(
        name,
        namespace,
        repoURL,
        auth,
        syncJobPodTemplateObj,
      );
      dispatch(addedRepo(apprepo));

      if (authHeader.length || customCA.length) {
        await Secret.create(
          secretName,
          secrets,
          {
            apiVersion: apprepo.apiVersion,
            blockOwnerDeletion: true,
            kind: apprepo.kind,
            name: apprepo.metadata.name,
            uid: apprepo.metadata.uid,
          } as IOwnerReference,
          namespace,
        );
      }
      return true;
    } catch (e) {
      dispatch(errorRepos(e, "create"));
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
      await axios.get(url.api.charts.listVersions(`${repo}/${chartName}`));
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
