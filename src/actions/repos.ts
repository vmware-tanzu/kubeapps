import { createAction, getReturnOfExpression } from "typesafe-actions";

import { Dispatch } from "react-redux";
import { AppRepository } from "../shared/AppRepository";
import { IAppRepository, IStoreState } from "../shared/types";

export const addRepo = createAction("ADD_REPO");
export const addedRepo = createAction("ADDED_REPO", (added: IAppRepository) => ({
  added,
  type: "ADDED_REPO",
}));
export const requestRepos = createAction("REQUEST_REPOS");
export const receiveRepos = createAction("RECEIVE_REPOS", (repos: IAppRepository[]) => {
  return {
    repos,
    type: "RECEIVE_REPOS",
  };
});
export const showForm = createAction("SHOW_FORM");
export const hideForm = createAction("HIDE_FORM");
export const resetForm = createAction("RESET_FORM");
export const submitForm = createAction("SUBMIT_FROM");
export const updateForm = createAction(
  "UPDATE_FORM",
  (values: { name?: string; namespace?: string; url?: string }) => {
    return {
      type: "UPDATE_FORM",
      values,
    };
  },
);
export const redirect = createAction("REDIRECT", (path: string) => ({ type: "REDIRECT", path }));
export const redirected = createAction("REDIRECTED");

const allActions = [
  addRepo,
  addedRepo,
  requestRepos,
  receiveRepos,
  resetForm,
  submitForm,
  updateForm,
  showForm,
  hideForm,
  redirect,
  redirected,
].map(getReturnOfExpression);
export type AppReposAction = typeof allActions[number];

export const deleteRepo = (name: string) => {
  return async (dispatch: Dispatch<IStoreState>) => {
    await AppRepository.delete(name);
    dispatch(requestRepos());
    const repos = await AppRepository.list();
    dispatch(receiveRepos(repos.items));
    return repos;
  };
};

export const resyncRepo = (name: string) => {
  return async (dispatch: Dispatch<IStoreState>) => {
    const repo = await AppRepository.get(name);
    repo.spec.resyncRequests = repo.spec.resyncRequests || 0;
    repo.spec.resyncRequests++;
    await AppRepository.update(name, repo);
    // TODO: Do something to show progress
    dispatch(requestRepos());
    const repos = await AppRepository.list();
    dispatch(receiveRepos(repos.items));
    return repos;
  };
};

export const fetchRepos = () => {
  return async (dispatch: Dispatch<IStoreState>) => {
    dispatch(requestRepos());
    const repos = await AppRepository.list();
    dispatch(receiveRepos(repos.items));
    return repos;
  };
};

export const installRepo = (name: string, url: string) => {
  return async (dispatch: Dispatch<IStoreState>) => {
    dispatch(addRepo());
    const added = await AppRepository.create(name, url);
    dispatch(addedRepo(added));
    return added;
  };
};
