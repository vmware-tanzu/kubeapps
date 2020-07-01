import * as React from "react";
import { Link } from "react-router-dom";

import { IAppRepository, ISecret } from "shared/types";
import * as url from "../../../shared/url";
import ConfirmDialog from "../../ConfirmDialog";
import { AppRepoAddButton } from "./AppRepoButton";

interface IAppRepoListItemProps {
  errors: {
    create?: Error;
    delete?: Error;
    fetch?: Error;
    update?: Error;
    validate?: Error;
  };
  update: (
    name: string,
    namespace: string,
    url: string,
    authHeader: string,
    customCA: string,
    syncJobPodTemplate: string,
    registrySecrets: string[],
  ) => Promise<boolean>;
  validating: boolean;
  validate: (url: string, authHeader: string, customCA: string) => Promise<any>;
  repo: IAppRepository;
  secret?: ISecret;
  renderNamespace: boolean;
  cluster: string;
  namespace: string;
  kubeappsNamespace: string;
  deleteRepo: (name: string, namespace: string) => Promise<boolean>;
  resyncRepo: (name: string, namespace: string) => void;
  imagePullSecrets: ISecret[];
  fetchImagePullSecrets: (namespace: string) => void;
  createDockerRegistrySecret: (
    name: string,
    user: string,
    password: string,
    email: string,
    server: string,
    namespace: string,
  ) => Promise<boolean>;
}

interface IAppRepoListItemState {
  modalIsOpen: boolean;
}

export class AppRepoListItem extends React.Component<IAppRepoListItemProps, IAppRepoListItemState> {
  public state: IAppRepoListItemState = {
    modalIsOpen: false,
  };

  public render() {
    const {
      cluster,
      namespace,
      renderNamespace,
      repo,
      errors,
      update,
      validate,
      validating,
      secret,
      imagePullSecrets,
      fetchImagePullSecrets,
      kubeappsNamespace,
      createDockerRegistrySecret,
    } = this.props;
    return (
      <tr key={repo.metadata.name}>
        <td>
          <Link to={url.app.repo(cluster, namespace, repo.metadata.name)}>
            {repo.metadata.name}
          </Link>
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

          <AppRepoAddButton
            errors={errors}
            onSubmit={update}
            validate={validate}
            namespace={namespace}
            kubeappsNamespace={kubeappsNamespace}
            validating={validating}
            text="Edit"
            repo={repo}
            secret={secret}
            imagePullSecrets={imagePullSecrets}
            fetchImagePullSecrets={fetchImagePullSecrets}
            createDockerRegistrySecret={createDockerRegistrySecret}
          />

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
