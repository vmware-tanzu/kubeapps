import * as React from "react";
import { Link } from "react-router-dom";

import * as url from "shared/url";
import { definedNamespaces } from "../../../shared/Namespace";
import { IAppRepository, IAppRepositoryKey, IRBACRole, ISecret } from "../../../shared/types";
import { ErrorSelector, MessageAlert } from "../../ErrorAlert";
import LoadingWrapper from "../../LoadingWrapper";
import { AppRepoAddButton } from "./AppRepoButton";
import { AppRepoListItem } from "./AppRepoListItem";
import { AppRepoRefreshAllButton } from "./AppRepoRefreshAllButton";

export interface IAppRepoListProps {
  errors: {
    create?: Error;
    delete?: Error;
    fetch?: Error;
    update?: Error;
    validate?: Error;
  };
  repos: IAppRepository[];
  repoSecrets: ISecret[];
  fetchRepos: (namespace: string) => void;
  deleteRepo: (name: string, namespace: string) => Promise<boolean>;
  resyncRepo: (name: string, namespace: string) => void;
  resyncAllRepos: (repos: IAppRepositoryKey[]) => void;
  install: (
    name: string,
    namespace: string,
    url: string,
    authHeader: string,
    customCA: string,
    syncJobPodTemplate: string,
    registrySecrets: string[],
  ) => Promise<boolean>;
  update: (
    name: string,
    namespace: string,
    url: string,
    authHeader: string,
    customCA: string,
    syncJobPodTemplate: string,
    registrySecrets: string[],
  ) => Promise<boolean>;
  validating: boolean;
  validate: (url: string, authHeader: string, customCA: string) => Promise<any>;
  cluster: string;
  namespace: string;
  kubeappsCluster: string;
  kubeappsNamespace: string;
  displayReposPerNamespaceMsg: boolean;
  isFetching: boolean;
  imagePullSecrets: ISecret[];
  fetchImagePullSecrets: (namespace: string) => void;
  createDockerRegistrySecret: (
    name: string,
    user: string,
    password: string,
    email: string,
    server: string,
    namespace: string,
  ) => Promise<boolean>;
}

const RequiredRBACRoles: { [s: string]: IRBACRole[] } = {
  delete: [
    {
      apiGroup: "kubeapps.com",
      resource: "apprepositories",
      verbs: ["delete"],
    },
  ],
  update: [
    {
      apiGroup: "kubeapps.com",
      resource: "apprepositories",
      verbs: ["get, update"],
    },
  ],
  fetch: [
    {
      apiGroup: "kubeapps.com",
      resource: "apprepositories",
      verbs: ["list"],
    },
  ],
};

class AppRepoList extends React.Component<IAppRepoListProps> {
  public componentDidMount() {
    this.props.fetchRepos(this.props.namespace);
    this.props.fetchImagePullSecrets(this.props.namespace);
  }

  public componentDidUpdate(prevProps: IAppRepoListProps) {
    const {
      cluster,
      errors: { fetch },
      fetchRepos,
      fetchImagePullSecrets,
      namespace,
    } = this.props;
    // refetch if namespace changes or if error removed due to location change
    if (
      prevProps.namespace !== namespace ||
      prevProps.cluster !== cluster ||
      (prevProps.errors.fetch && !fetch)
    ) {
      fetchRepos(namespace);
      fetchImagePullSecrets(namespace);
    }
  }

  public render() {
    const {
      errors,
      repos,
      install,
      update,
      cluster,
      namespace,
      kubeappsCluster,
      kubeappsNamespace,
      displayReposPerNamespaceMsg,
      isFetching,
      deleteRepo,
      resyncRepo,
      resyncAllRepos,
      validate,
      repoSecrets,
      validating,
      imagePullSecrets,
      fetchImagePullSecrets,
      createDockerRegistrySecret,
    } = this.props;
    const renderNamespace = namespace === definedNamespaces.all;

    // We do not currently support app repositories on additional clusters.
    if (cluster !== kubeappsCluster) {
      return (
        <MessageAlert header="AppRepositories can be created on the default cluster only">
          <div>
            <p className="margin-v-normal">
              Kubeapps' multi-cluster support currently enables creation of custom app repositories
              on the cluster on which Kubeapps is installed.
            </p>
            <p className="margin-v-normal">
              You cannot currently create an app repository on an additional cluster.
              {kubeappsCluster && (
                <>
                  You can create an app repository with charts available for installation across
                  clusters and namespaces on the{" "}
                  <Link to={url.app.config.apprepositories(kubeappsCluster, definedNamespaces.all)}>
                    cluster on which Kubeapps is installed
                  </Link>
                  , if you have the appropriate authorization.
                </>
              )}
            </p>
          </div>
        </MessageAlert>
      );
    }

    return (
      <div className="app-repo-list">
        <h1>App Repositories</h1>
        {errors.fetch && this.renderError("fetch")}
        {errors.delete && this.renderError("delete")}
        {errors.update && this.renderError("update")}
        {!errors.fetch && (
          <>
            <LoadingWrapper loaded={!isFetching}>
              <table>
                <thead>
                  <tr>
                    <th>Repo</th>
                    {renderNamespace && <th>Namespace</th>}
                    <th>URL</th>
                    <th>Actions</th>
                  </tr>
                </thead>
                <tbody>
                  {repos.map(repo => (
                    <AppRepoListItem
                      key={repo.metadata.uid}
                      deleteRepo={deleteRepo}
                      resyncRepo={resyncRepo}
                      repo={repo}
                      renderNamespace={renderNamespace}
                      cluster={cluster}
                      namespace={namespace}
                      kubeappsNamespace={kubeappsNamespace}
                      errors={errors}
                      validating={validating}
                      validate={validate}
                      secret={repoSecrets.find(secret =>
                        secret.metadata.ownerReferences?.some(
                          ownerRef => ownerRef.name === repo.metadata.name,
                        ),
                      )}
                      update={update}
                      imagePullSecrets={imagePullSecrets}
                      fetchImagePullSecrets={fetchImagePullSecrets}
                      createDockerRegistrySecret={createDockerRegistrySecret}
                    />
                  ))}
                </tbody>
              </table>
            </LoadingWrapper>
            <AppRepoAddButton
              errors={errors}
              onSubmit={install}
              validate={validate}
              namespace={namespace}
              kubeappsNamespace={kubeappsNamespace}
              validating={validating}
              primary={true}
              imagePullSecrets={imagePullSecrets}
              fetchImagePullSecrets={fetchImagePullSecrets}
              createDockerRegistrySecret={createDockerRegistrySecret}
            />
            <AppRepoRefreshAllButton resyncAllRepos={resyncAllRepos} repos={repos} />
          </>
        )}
        {displayReposPerNamespaceMsg && (
          <MessageAlert header="Looking for other app repositories?">
            <div>
              <p className="margin-v-normal">
                Administrators can view App Repositories across all namespaces using the Namespace
                selector at the top and clicking on "All Namespaces".
              </p>
              <p className="margin-v-normal">
                Kubeapps now enables you to create App Repositories in your own namespace that will
                be available in your own namespace and, in the future, optionally available in other
                namespaces to which you have access. You can read more information in the{" "}
                <a href="https://github.com/kubeapps/kubeapps/blob/master/docs/user/private-app-repository.md">
                  Private App Repository docs
                </a>
                .
              </p>
            </div>
          </MessageAlert>
        )}
      </div>
    );
  }

  private renderError(action: string) {
    return (
      <ErrorSelector
        error={this.props.errors[action]}
        defaultRequiredRBACRoles={RequiredRBACRoles}
        action={action}
        namespace={this.props.namespace}
        resource="App Repositories"
      />
    );
  }
}

export default AppRepoList;
