import { LocationChangeAction, LOCATION_CHANGE } from "connected-react-router";
import { IAppRepository, ISecret } from "shared/types";
import { getType } from "typesafe-actions";
import actions from "../actions";
import { AppReposAction } from "../actions/repos";

export interface IAppRepositoryState {
  addingRepo: boolean;
  errors: {
    create?: Error;
    delete?: Error;
    fetch?: Error;
    update?: Error;
    validate?: Error;
  };
  lastAdded?: IAppRepository;
  isFetching: boolean;
  isFetchingElem: {
    repositories: boolean;
    secrets: boolean;
  };
  validating: boolean;
  repo: IAppRepository;
  repos: IAppRepository[];
  form: {
    name: string;
    namespace: string;
    url: string;
    show: boolean;
  };
  imagePullSecrets: ISecret[];
  redirectTo?: string;
}

export const initialState: IAppRepositoryState = {
  addingRepo: false,
  errors: {},
  form: {
    name: "",
    namespace: "",
    show: false,
    url: "",
  },
  isFetching: false,
  isFetchingElem: {
    repositories: false,
    secrets: false,
  },
  validating: false,
  repo: {} as IAppRepository,
  repos: [],
  imagePullSecrets: [],
};

function isFetching(state: IAppRepositoryState, item: string, fetching: boolean) {
  const composedIsFetching = {
    ...state.isFetchingElem,
    [item]: fetching,
  };
  return {
    isFetching: Object.values(composedIsFetching).some(v => v),
    isFetchingElem: composedIsFetching,
  };
}

const reposReducer = (
  state: IAppRepositoryState = initialState,
  action: AppReposAction | LocationChangeAction,
): IAppRepositoryState => {
  switch (action.type) {
    case getType(actions.repos.receiveRepos):
      return {
        ...state,
        ...isFetching(state, "repositories", false),
        repos: action.payload,
        errors: {},
      };
    case getType(actions.repos.receiveRepo):
      return {
        ...state,
        ...isFetching(state, "repositories", false),
        repo: action.payload,
        errors: {},
      };
    case getType(actions.repos.requestRepos):
      return { ...state, ...isFetching(state, "repositories", true) };
    case getType(actions.repos.addRepo):
      return { ...state, addingRepo: true };
    case getType(actions.repos.addedRepo):
      return {
        ...state,
        addingRepo: false,
        lastAdded: action.payload,
        repos: [...state.repos, action.payload],
      };
    case getType(actions.repos.repoUpdated): {
      const updatedRepo = action.payload;
      const repos = state.repos.map(r =>
        r.metadata.name === updatedRepo.metadata.name &&
        r.metadata.namespace === updatedRepo.metadata.namespace
          ? updatedRepo
          : r,
      );
      return { ...state, repos };
    }
    case getType(actions.repos.repoValidating):
      return { ...state, validating: true };
    case getType(actions.repos.repoValidated):
      return { ...state, validating: false, errors: { ...state.errors, validate: undefined } };
    case getType(actions.repos.redirect):
      return { ...state, redirectTo: action.payload };
    case getType(actions.repos.redirected):
      return { ...state, redirectTo: undefined };
    case getType(actions.repos.errorRepos):
      return {
        ...state,
        // don't reset the fetch error
        errors: { fetch: state.errors.fetch, [action.payload.op]: action.payload.err },
        isFetching: false,
        isFetchingElem: {
          repositories: false,
          secrets: false,
        },
        validating: false,
      };
    case LOCATION_CHANGE:
      return {
        ...state,
        errors: {},
        isFetching: false,
        isFetchingElem: {
          repositories: false,
          secrets: false,
        },
      };
    default:
      return state;
  }
};

export default reposReducer;
