import { connect } from "react-redux";
import { Action } from "redux";
import { ThunkDispatch } from "redux-thunk";

import actions from "../../actions";
import AppRepoList from "../../components/Config/AppRepoList";
import { definedNamespaces } from "../../shared/Namespace";
import { IStoreState } from "../../shared/types";

function mapStateToProps({ config, namespace, repos }: IStoreState) {
  let repoNamespace = config.namespace;
  if (config.featureFlags.reposPerNamespace && namespace.current !== definedNamespaces.all) {
    repoNamespace = namespace.current;
  }
  return {
    errors: repos.errors,
    namespace: repoNamespace,
    repos: repos.repos,
  };
}

function mapDispatchToProps(dispatch: ThunkDispatch<IStoreState, null, Action>) {
  return {
    deleteRepo: async (name: string) => {
      return dispatch(actions.repos.deleteRepo(name));
    },
    fetchRepos: async (namespace: string) => {
      return dispatch(actions.repos.fetchRepos(namespace));
    },
    install: async (
      name: string,
      url: string,
      authHeader: string,
      customCA: string,
      syncJobPodTemplate: string,
    ) => {
      return dispatch(
        actions.repos.installRepo(name, url, authHeader, customCA, syncJobPodTemplate),
      );
    },
    resyncRepo: async (name: string) => {
      return dispatch(actions.repos.resyncRepo(name));
    },
    resyncAllRepos: async (names: string[]) => {
      return dispatch(actions.repos.resyncAllRepos(names));
    },
  };
}

export default connect(mapStateToProps, mapDispatchToProps)(AppRepoList);
