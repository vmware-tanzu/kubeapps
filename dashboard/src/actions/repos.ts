import { ActionType, createActionDeprecated } from "typesafe-actions";

import { Dispatch } from "redux";
import { ThunkDispatch } from "redux-thunk";

import { AppRepository } from "../shared/AppRepository";
import Secret from "../shared/Secret";
import * as url from "../shared/url";

import { IAppRepository, IOwnerReference, IStoreState, NotFoundError } from "../shared/types";

export const addRepo = createActionDeprecated("ADD_REPO");
export const addedRepo = createActionDeprecated("ADDED_REPO", (added: IAppRepository) => ({
  added,
  type: "ADDED_REPO",
}));
export const requestRepos = createActionDeprecated("REQUEST_REPOS");
export const receiveRepos = createActionDeprecated("RECEIVE_REPOS", (repos: IAppRepository[]) => {
  return {
    repos,
    type: "RECEIVE_REPOS",
  };
});
export const requestRepo = createActionDeprecated("REQUEST_REPO");
export const receiveRepo = createActionDeprecated("RECEIVE_REPO", (repo: IAppRepository) => ({
  repo,
  type: "RECEIVE_REPO",
}));
export const errorChart = createActionDeprecated("ERROR_CHART", (err: Error) => ({
  err,
  type: "ERROR_CHART",
}));
export const showForm = createActionDeprecated("SHOW_FORM");
export const hideForm = createActionDeprecated("HIDE_FORM");
export const resetForm = createActionDeprecated("RESET_FORM");
export const submitForm = createActionDeprecated("SUBMIT_FROM");
export const updateForm = createActionDeprecated(
  "UPDATE_FORM",
  (values: { name?: string; namespace?: string; url?: string }) => {
    return {
      type: "UPDATE_FORM",
      values,
    };
  },
);
export const redirect = createActionDeprecated("REDIRECT", (path: string) => ({
  type: "REDIRECT",
  path,
}));
export const redirected = createActionDeprecated("REDIRECTED");
export const errorRepos = createActionDeprecated(
  "ERROR_REPOS",
  (err: Error, op: "create" | "update" | "fetch" | "delete") => ({
    err,
    op,
    type: "ERROR_REPOS",
  }),
);

const allActions = [
  addRepo,
  addedRepo,
  errorChart,
  errorRepos,
  requestRepos,
  receiveRepo,
  receiveRepos,
  resetForm,
  submitForm,
  updateForm,
  showForm,
  hideForm,
  redirect,
  redirected,
];
export type AppReposAction = ActionType<typeof allActions[number]>;

export const deleteRepo = (name: string) => {
  return async (
    dispatch: ThunkDispatch<IStoreState, null, AppReposAction>,
    getState: () => IStoreState,
  ) => {
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

export const resyncRepo = (name: string) => {
  return async (dispatch: Dispatch, getState: () => IStoreState) => {
    try {
      const {
        config: { namespace },
      } = getState();
      const repo = await AppRepository.get(name, namespace);
      repo.spec.resyncRequests = repo.spec.resyncRequests || 0;
      repo.spec.resyncRequests++;
      await AppRepository.update(name, namespace, repo);
      // TODO: Do something to show progress
      dispatch(requestRepos());
      const repos = await AppRepository.list(namespace);
      dispatch(receiveRepos(repos.items));
    } catch (e) {
      dispatch(errorRepos(e, "update"));
    }
  };
};

export const fetchRepos = () => {
  return async (dispatch: Dispatch, getState: () => IStoreState) => {
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

export const installRepo = (name: string, repoURL: string, authHeader: string) => {
  return async (dispatch: Dispatch, getState: () => IStoreState) => {
    try {
      const {
        config: { namespace },
      } = getState();
      let auth;
      const secretName = `apprepo-${name}-secrets`;
      if (authHeader.length) {
        // ensure we can create secrets in the kubeapps namespace
        auth = {
          header: {
            secretKeyRef: {
              key: "authorizationHeader",
              name: secretName,
            },
          },
        };
      }
      dispatch(addRepo());
      const apprepo = await AppRepository.create(name, namespace, repoURL, auth);
      dispatch(addedRepo(apprepo));

      if (authHeader.length) {
        await Secret.create(
          secretName,
          { authorizationHeader: btoa(authHeader) },
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

export function checkChart(repo: string, chartName: string) {
  return async (dispatch: Dispatch, getState: () => IStoreState) => {
    const {
      config: { namespace },
    } = getState();
    dispatch(requestRepo());
    const appRepository = await AppRepository.get(repo, namespace);
    const res = await fetch(url.api.charts.listVersions(`${repo}/${chartName}`));
    if (res.ok) {
      dispatch(receiveRepo(appRepository));
    } else {
      dispatch(
        errorChart(new NotFoundError(`Chart ${chartName} not found in the repository ${repo}.`)),
      );
    }
  };
}

export function clearRepo() {
  return async (dispatch: Dispatch) => {
    dispatch(receiveRepo({} as IAppRepository));
  };
}
