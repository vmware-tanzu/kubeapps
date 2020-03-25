import * as React from "react";
import { Link } from "react-router-dom";

import { IAppRepository } from "shared/types";
import ConfirmDialog from "../../ConfirmDialog";

interface IAppRepoListItemProps {
  repo: IAppRepository;
  renderNamespace: boolean;
  namespace: string;
  deleteRepo: (name: string, namespace: string) => Promise<boolean>;
  resyncRepo: (name: string, namespace: string) => void;
}

interface IAppRepoListItemState {
  modalIsOpen: boolean;
}

export class AppRepoListItem extends React.Component<IAppRepoListItemProps, IAppRepoListItemState> {
  public state: IAppRepoListItemState = {
    modalIsOpen: false,
  };

  public render() {
    const { namespace, renderNamespace, repo } = this.props;
    return (
      <tr key={repo.metadata.name}>
        <td>
          <Link to={`/catalog/ns/${namespace}/${repo.metadata.name}`}>{repo.metadata.name}</Link>
        </td>
        {renderNamespace && <td>{repo.metadata.namespace}</td>}
        <td>{repo.spec && repo.spec.url}</td>
        <td>
          <ConfirmDialog
            onConfirm={this.handleDeleteClick(repo.metadata.name, repo.metadata.namespace)}
            modalIsOpen={this.state.modalIsOpen}
            loading={false}
            closeModal={this.closeModal}
          />

          <button className="button button-secondary" onClick={this.openModel}>
            Delete
          </button>

          <button
            className="button button-secondary"
            onClick={this.handleResyncClick(repo.metadata.name, repo.metadata.namespace)}
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

  private handleDeleteClick(repoName: string, namespace: string) {
    return async () => {
      this.props.deleteRepo(repoName, namespace);
      this.setState({ modalIsOpen: false });
    };
  }

  private handleResyncClick(repoName: string, namespace: string) {
    return () => {
      this.props.resyncRepo(repoName, namespace);
    };
  }
}
