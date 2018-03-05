import * as React from "react";

import { IAppRepository } from "../../../shared/types";
import { AppRepoAddButton } from "./AppRepoButton";
import { AppRepoListItem } from "./AppRepoListItem";

export interface IAppRepoListProps {
  repos: IAppRepository[];
  fetchRepos: () => Promise<any>;
  deleteRepo: (name: string) => Promise<any>;
  resyncRepo: (name: string) => Promise<any>;
  install: (name: string, url: string) => Promise<any>;
}

export class AppRepoList extends React.Component<IAppRepoListProps> {
  public componentDidMount() {
    this.props.fetchRepos();
  }

  public render() {
    const { repos, install, deleteRepo, resyncRepo } = this.props;
    return (
      <div className="app-repo-list">
        <h1>App Repositories</h1>
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
        <AppRepoAddButton install={install} />
      </div>
    );
  }
}
