import { connect } from "react-redux";
import { Dispatch } from "redux";

import actions from "../actions";
import { AppRepoList } from "../components/Config/AppRepoList";
import { IStoreState } from "../shared/types";

function mapStateToProps({ repos }: IStoreState) {
  return {
    repos: repos.repos,
  };
}

function mapDispatchToProps(dispatch: Dispatch<IStoreState>) {
  return {
    deleteRepo: async (name: string, namespace: string = "kubeapps") => {
      return dispatch(actions.repos.deleteRepo(name, namespace));
    },
    fetchRepos: async () => {
      return dispatch(actions.repos.fetchRepos());
    },
    install: async (name: string, url: string, namespace: string = "kubeapps") => {
      return dispatch(actions.repos.installRepo(name, url, namespace));
    },
    resyncRepo: async (name: string, namespace: string = "kubeapps") => {
      return dispatch(actions.repos.resyncRepo(name, namespace));
    },
  };
}

export default connect(mapStateToProps, mapDispatchToProps)(AppRepoList);
