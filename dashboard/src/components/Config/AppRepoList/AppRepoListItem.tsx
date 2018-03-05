import * as React from "react";
import { IAppRepository } from "shared/types";
import ConfirmDialog from "../../ConfirmDialog";

interface IAppRepoListItemProps {
  repo: IAppRepository;
  deleteRepo: (name: string) => Promise<any>;
  resyncRepo: (name: string) => Promise<any>;
}

interface IAppRepoListItemState {
  modalIsOpen: boolean;
}

export class AppRepoListItem extends React.Component<IAppRepoListItemProps, IAppRepoListItemState> {
  public state: IAppRepoListItemState = {
    modalIsOpen: false,
  };

  public render() {
    const { repo } = this.props;
    return (
      <tr key={repo.metadata.name}>
        <td>{repo.metadata.name}</td>
        <td>{repo.spec && repo.spec.url}</td>
        <td>
          <ConfirmDialog
            onConfirm={this.handleDeleteClick(repo.metadata.name)}
            modalIsOpen={this.state.modalIsOpen}
            closeModal={this.closeModal}
          />

          <button className="button button-secondary" onClick={this.openModel}>
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
  }

  public openModel = () => {
    this.setState({
      modalIsOpen: true,
    });
  };

  public closeModal = async () => {
    this.setState({
      modalIsOpen: false,
    });
  };

  private handleDeleteClick(repoName: string) {
    return async () => {
      this.props.deleteRepo(repoName);
      this.setState({ modalIsOpen: false });
    };
  }

  private handleResyncClick(repoName: string) {
    return () => {
      this.props.resyncRepo(repoName);
    };
  }
}
