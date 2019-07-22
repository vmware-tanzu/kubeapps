import * as React from "react";

import "./AppRepo.css";

interface IAppRepoRefreshAllButtonProps {
  error?: Error;
  resyncAllRepos: () => void;
  kubeappsNamespace: string;
}

export class AppRepoRefreshAllButton extends React.Component<IAppRepoRefreshAllButtonProps> {
  public render() {
    return (
      <span>
        <button
          className="button button-primary margin-l-big"
          onClick={this.handleResyncAllClick()}
          title="Refresh All App Repositories"
        >
          Refresh All
        </button>
      </span>
    );
  }

  private handleResyncAllClick() {
    return () => {
      this.props.resyncAllRepos();
    };
  }
}
