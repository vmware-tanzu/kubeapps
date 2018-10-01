import { LOCATION_CHANGE, LocationChangeAction } from "connected-react-router";
import { getType } from "typesafe-actions";

import actions from "../actions";
import { AppReposAction } from "../actions/repos";
import { IAppRepository } from "../shared/types";

export interface IAppRepositoryState {
  addingRepo: boolean;
  errors: {
    create?: Error;
    delete?: Error;
    fetch?: Error;
    update?: Error;
  };
  lastAdded?: IAppRepository;
  isFetching: boolean;
  repo: IAppRepository;
  repos: IAppRepository[];
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
  errors: {},
  form: {
    name: "",
    namespace: "",
    show: false,
    url: "",
  },
  isFetching: false,
  repo: {} as IAppRepository,
  repos: [],
};

const reposReducer = (
  state: IAppRepositoryState = initialState,
  action: AppReposAction | LocationChangeAction,
): IAppRepositoryState => {
  switch (action.type) {
    case getType(actions.repos.receiveRepos):
      return { ...state, isFetching: false, repos: action.payload, errors: {} };
    case getType(actions.repos.receiveRepo):
      return { ...state, isFetching: false, repo: action.payload, errors: {} };
    case getType(actions.repos.requestRepos):
      return { ...state, isFetching: true };
    case getType(actions.repos.addRepo):
      return { ...state, addingRepo: true };
    case getType(actions.repos.addedRepo):
      return {
        ...state,
        addingRepo: false,
        lastAdded: action.payload,
        repos: [...state.repos, action.payload],
      };
    case getType(actions.repos.resetForm):
      return { ...state, form: { ...state.form, name: "", namespace: "", url: "" } };
    case getType(actions.repos.showForm):
      return { ...state, form: { ...state.form, show: true } };
    case getType(actions.repos.hideForm):
      return { ...state, form: { ...state.form, show: false } };
    case getType(actions.repos.redirect):
      return { ...state, redirectTo: action.payload };
    case getType(actions.repos.redirected):
      return { ...state, redirectTo: undefined };
    case getType(actions.charts.errorChart):
      return {
        ...state,
        errors: { fetch: action.payload },
      };
    case getType(actions.repos.errorRepos):
      return {
        ...state,
        // don't reset the fetch error
        errors: { fetch: state.errors.fetch, [action.payload.op]: action.payload.err },
        isFetching: false,
      };
    case LOCATION_CHANGE:
      return {
        ...state,
        errors: {},
        isFetching: false,
      };
    default:
      return state;
  }
};

export default reposReducer;
