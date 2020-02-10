import * as React from "react";

import { IAppRepository, IAppRepositoryKey } from "shared/types";
import "./AppRepo.css";

interface IAppRepoRefreshAllButtonProps {
  resyncAllRepos: (repos: IAppRepositoryKey[]) => void;
  repos: IAppRepository[];
}

export class AppRepoRefreshAllButton extends React.Component<IAppRepoRefreshAllButtonProps> {
  public render() {
    return (
      <button
        className="button button-primary margin-l-big"
        onClick={this.handleResyncAllClick}
        title="Refresh All App Repositories"
      >
        Refresh All
      </button>
    );
  }

  private handleResyncAllClick = async () => {
    if (this.props.repos) {
      const repos = this.props.repos.map(repo => {
        return {
          name: repo.metadata.name,
          namespace: repo.metadata.namespace,
        };
      });
      this.props.resyncAllRepos(repos);
    }
  };
}
