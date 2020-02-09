import * as React from "react";

import { definedNamespaces } from "../../../shared/Namespace";
import { IAppRepository, IRBACRole } from "../../../shared/types";
import { ErrorSelector, MessageAlert } from "../../ErrorAlert";
import { AppRepoAddButton } from "./AppRepoButton";
import { AppRepoListItem } from "./AppRepoListItem";
import { AppRepoRefreshAllButton } from "./AppRepoRefreshAllButton";

export interface IAppRepoListProps {
  errors: {
    create?: Error;
    delete?: Error;
    fetch?: Error;
    update?: Error;
  };
  repos: IAppRepository[];
  fetchRepos: (namespace: string) => void;
  deleteRepo: (name: string) => Promise<boolean>;
  resyncRepo: (name: string) => void;
  resyncAllRepos: (names: string[]) => void;
  install: (
    name: string,
    url: string,
    authHeader: string,
    customCA: string,
    syncJobPodTemplate: string,
  ) => Promise<boolean>;
  namespace: string;
  displayReposPerNamespaceMsg: boolean;
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
  }

  public componentDidUpdate(prevProps: IAppRepoListProps) {
    const {
      errors: { fetch },
      fetchRepos,
      namespace,
    } = this.props;
    // refetch if namespace changes or if error removed due to location change
    if (prevProps.namespace !== namespace || (prevProps.errors.fetch && !fetch)) {
      fetchRepos(namespace);
    }
  }

  public render() {
    const {
      errors,
      repos,
      install,
      deleteRepo,
      resyncRepo,
      resyncAllRepos,
      namespace,
      displayReposPerNamespaceMsg,
    } = this.props;
    const renderNamespace = namespace === definedNamespaces.all;
    return (
      <div className="app-repo-list">
        <h1>App Repositories</h1>
        {errors.fetch && this.renderError("fetch")}
        {errors.delete && this.renderError("delete")}
        {errors.update && this.renderError("update")}
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
              />
            ))}
          </tbody>
        </table>
        <AppRepoAddButton error={errors.create} install={install} namespace={namespace} />
        <AppRepoRefreshAllButton
          resyncAllRepos={resyncAllRepos}
          repos={repos}
          namespace={namespace}
        />
        {displayReposPerNamespaceMsg && (
          <MessageAlert header="Looking for other app repositories?">
            <div>
              <p className="margin-v-normal">
                You can view App Repositories across all namespaces by selecting "All Namespaces"
                above, if you have permission to view App Repositories cluster-wide.
              </p>
              <p className="margin-v-normal">
                Kubeapps now enables you to create App Repositories in your own namespace that will
                be available in your own namespace and, in the future, optionally available in other
                namespaces to which you have access. You can read more information in the{" "}
                <a href="https://github.com/kubeapps/kubeapps/blob/master/docs/user/private-app-repository.md#per-namespace-app-repositories">
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
