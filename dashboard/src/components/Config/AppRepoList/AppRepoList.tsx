import * as React from "react";

import { IAppRepository, IRBACRole } from "../../../shared/types";
import ErrorSelector from "../../ErrorAlert/ErrorSelector";
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
  fetchRepos: () => void;
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
  kubeappsNamespace: string;
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
    this.props.fetchRepos();
  }

  public componentWillReceiveProps(nextProps: IAppRepoListProps) {
    const {
      errors: { fetch },
      fetchRepos,
    } = this.props;
    // refetch if error removed due to location change
    if (fetch && !nextProps.errors.fetch) {
      fetchRepos();
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
      kubeappsNamespace,
    } = this.props;
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
              />
            ))}
          </tbody>
        </table>
        <AppRepoAddButton
          error={errors.create}
          install={install}
          kubeappsNamespace={kubeappsNamespace}
        />
        <AppRepoRefreshAllButton
          resyncAllRepos={resyncAllRepos}
          repos={repos}
          kubeappsNamespace={kubeappsNamespace}
        />
      </div>
    );
  }

  private renderError(action: string) {
    return (
      <ErrorSelector
        error={this.props.errors[action]}
        defaultRequiredRBACRoles={RequiredRBACRoles}
        action={action}
        namespace={this.props.kubeappsNamespace}
        resource="App Repositories"
      />
    );
  }
}

export default AppRepoList;
