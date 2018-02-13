import * as React from "react";

import { IAppRepository } from "../../shared/types";
import { AppRepoAddButton } from "./AppRepoButton";

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
    const { repos, install } = this.props;
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
            {repos.map(repo => {
              return (
                <tr key={repo.metadata.name}>
                  <td>{repo.metadata.name}</td>
                  <td>{repo.spec && repo.spec.url}</td>
                  <td>
                    <button
                      className="button button-secondary"
                      onClick={this.handleDeleteClick(repo.metadata.name)}
                    >
                      Delete
                    </button>
                    <button
                      className="button button-secondary"
                      onClick={this.handleResyncClick(repo.metadata.name)}
                    >
                      Refresh
                    </button>
                  </td>
                </tr>
              );
            })}
          </tbody>
        </table>
        <AppRepoAddButton install={install} />
      </div>
    );
  }

  private handleDeleteClick = (repoName: string) => () => this.props.deleteRepo(repoName);
  private handleResyncClick = (repoName: string) => () => this.props.resyncRepo(repoName);
}
