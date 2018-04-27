import * as React from "react";

import { ForbiddenError, IAppRepository, IRBACRole } from "../../../shared/types";
import { PermissionsErrorAlert, UnexpectedErrorAlert } from "../../ErrorAlert";
import { AppRepoAddButton } from "./AppRepoButton";
import { AppRepoListItem } from "./AppRepoListItem";

export interface IAppRepoListProps {
  errors: {
    create?: Error;
    delete?: Error;
    fetch?: Error;
    update?: Error;
  };
  repos: IAppRepository[];
  fetchRepos: () => Promise<any>;
  deleteRepo: (name: string) => Promise<any>;
  resyncRepo: (name: string) => Promise<any>;
  install: (name: string, url: string, authHeader: string) => Promise<boolean>;
}

const RequiredRBACRoles: { [s: string]: IRBACRole[] } = {
  delete: [
    {
      apiGroup: "kubeapps.com",
      resource: "apprepositories",
      verbs: ["delete"],
    },
  ],
  refresh: [
    {
      apiGroup: "kubeapps.com",
      resource: "apprepositories",
      verbs: ["get, update"],
    },
  ],
  view: [
    {
      apiGroup: "kubeapps.com",
      resource: "apprepositories",
      verbs: ["list"],
    },
  ],
};

export class AppRepoList extends React.Component<IAppRepoListProps> {
  public componentDidMount() {
    this.props.fetchRepos();
  }

  public componentWillReceiveProps(nextProps: IAppRepoListProps) {
    const { errors: { fetch }, fetchRepos } = this.props;
    // refetch if error removed due to location change
    if (fetch && !nextProps.errors.fetch) {
      fetchRepos();
    }
  }

  public render() {
    const { errors, repos, install, deleteRepo, resyncRepo } = this.props;
    return (
      <div className="app-repo-list">
        <h1>App Repositories</h1>
        {errors.fetch && this.renderError(errors.fetch)}
        {errors.delete && this.renderError(errors.delete, "delete")}
        {errors.update && this.renderError(errors.update, "refresh")}
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
        <AppRepoAddButton error={errors.create} install={install} />
      </div>
    );
  }

  private renderError(error: Error, action: string = "view") {
    switch (error.constructor) {
      case ForbiddenError:
        return (
          <PermissionsErrorAlert
            namespace="kubeapps"
            roles={RequiredRBACRoles[action]}
            action={`${action} App Repositories`}
          />
        );
      default:
        return <UnexpectedErrorAlert />;
    }
  }
}
