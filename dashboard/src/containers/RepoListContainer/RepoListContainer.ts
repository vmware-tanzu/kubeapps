import { connect } from "react-redux";
import { Action } from "redux";
import { ThunkDispatch } from "redux-thunk";
import actions from "../../actions";
import AppRepoList from "../../components/Config/AppRepoList";
import { definedNamespaces } from "../../shared/Namespace";
import { IAppRepositoryKey, IStoreState } from "../../shared/types";

function mapStateToProps({ config, clusters: { currentCluster, clusters }, repos }: IStoreState) {
  const repoNamespace = clusters[currentCluster].currentNamespace;
  let displayReposPerNamespaceMsg = false;
  if (repoNamespace !== definedNamespaces.all) {
    displayReposPerNamespaceMsg = true;
  }
  return {
    errors: repos.errors,
    namespace: repoNamespace,
    cluster: currentCluster,
    repos: repos.repos,
    displayReposPerNamespaceMsg,
    isFetching: repos.isFetching,
    repoSecrets: repos.repoSecrets,
    validating: repos.validating,
    imagePullSecrets: repos.imagePullSecrets,
    kubeappsCluster: config.kubeappsCluster,
    kubeappsNamespace: config.kubeappsNamespace,
    UI: config.featureFlags.ui,
  };
}

function mapDispatchToProps(dispatch: ThunkDispatch<IStoreState, null, Action>) {
  return {
    deleteRepo: async (name: string, namespace: string) => {
      return dispatch(actions.repos.deleteRepo(name, namespace));
    },
    fetchRepos: async (namespace: string) => {
      return dispatch(actions.repos.fetchRepos(namespace));
    },
    install: async (
      name: string,
      namespace: string,
      url: string,
      authHeader: string,
      customCA: string,
      syncJobPodTemplate: string,
      registrySecrets: string[],
    ) => {
      return dispatch(
        actions.repos.installRepo(
          name,
          namespace,
          url,
          authHeader,
          customCA,
          syncJobPodTemplate,
          registrySecrets,
        ),
      );
    },
    update: async (
      name: string,
      namespace: string,
      url: string,
      authHeader: string,
      customCA: string,
      syncJobPodTemplate: string,
      registrySecrets: string[],
    ) => {
      return dispatch(
        actions.repos.updateRepo(
          name,
          namespace,
          url,
          authHeader,
          customCA,
          syncJobPodTemplate,
          registrySecrets,
        ),
      );
    },
    validate: async (url: string, authHeader: string, customCA: string) => {
      return dispatch(actions.repos.validateRepo(url, authHeader, customCA));
    },
    fetchImagePullSecrets: async (namespace: string) => {
      return dispatch(actions.repos.fetchImagePullSecrets(namespace));
    },
    resyncRepo: async (name: string, namespace: string) => {
      return dispatch(actions.repos.resyncRepo(name, namespace));
    },
    // Update here after actions
    resyncAllRepos: async (repos: IAppRepositoryKey[]) => {
      return dispatch(actions.repos.resyncAllRepos(repos));
    },
    createDockerRegistrySecret: async (
      name: string,
      user: string,
      password: string,
      email: string,
      server: string,
      namespace: string,
    ) => {
      return dispatch(
        actions.repos.createDockerRegistrySecret(name, user, password, email, server, namespace),
      );
    },
  };
}

export default connect(mapStateToProps, mapDispatchToProps)(AppRepoList);
