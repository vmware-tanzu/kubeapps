import * as React from "react";

import { IAppRepository } from "shared/types";
import "./AppRepo.css";

interface IAppRepoRefreshAllButtonProps {
  resyncAllRepos: (names: string[]) => void;
  repos: IAppRepository[];
  kubeappsNamespace: string;
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
      const repoNames = this.props.repos.map(repo => {
        return repo.metadata.name;
      });
      this.props.resyncAllRepos(repoNames);
    }
  };
}
