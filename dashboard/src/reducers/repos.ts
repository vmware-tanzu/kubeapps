import { getType } from "typesafe-actions";

import actions from "../actions";
import { AppReposAction } from "../actions/repos";
import { IAppRepository } from "../shared/types";

export interface IAppRepositoryState {
  addingRepo: boolean;
  lastAdded?: IAppRepository;
  isFetching: boolean;
  repos: IAppRepository[];
  selected?: IAppRepository;
  form: {
    name: string;
    namespace: string;
    url: string;
    show: boolean;
  };
  redirectTo?: string;
}

const initialState: IAppRepositoryState = {
  addingRepo: false,
  form: {
    name: "",
    namespace: "",
    show: false,
    url: "",
  },
  isFetching: false,
  repos: [],
};

const reposReducer = (
  state: IAppRepositoryState = initialState,
  action: AppReposAction,
): IAppRepositoryState => {
  switch (action.type) {
    case getType(actions.repos.receiveRepos):
      const { repos } = action;
      return { ...state, isFetching: false, repos };
    case getType(actions.repos.requestRepos):
      return { ...state, isFetching: true };
    case getType(actions.repos.addRepo):
      return { ...state, addingRepo: true };
    case getType(actions.repos.addedRepo):
      const { added } = action;
      return { ...state, addingRepo: false, lastAdded: added, repos: [...state.repos, added] };
    case getType(actions.repos.resetForm):
      return { ...state, form: { ...state.form, name: "", namespace: "", url: "" } };
    case getType(actions.repos.updateForm):
      const { values } = action;
      return { ...state, form: { ...state.form, ...values } };
    case getType(actions.repos.showForm):
      return { ...state, form: { ...state.form, show: true } };
    case getType(actions.repos.hideForm):
      return { ...state, form: { ...state.form, show: false } };
    case getType(actions.repos.redirect):
      return { ...state, redirectTo: action.path };
    case getType(actions.repos.redirected):
      return { ...state, redirectTo: undefined };
    default:
      return state;
  }
};

export default reposReducer;
