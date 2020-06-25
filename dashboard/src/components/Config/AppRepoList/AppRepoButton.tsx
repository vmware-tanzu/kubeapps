import * as React from "react";
import Modal from "react-modal";
import { Redirect } from "react-router";

import { IAppRepository, IRBACRole, ISecret } from "../../../shared/types";
import ErrorSelector from "../../ErrorAlert/ErrorSelector";
import "./AppRepo.css";
import { AppRepoForm } from "./AppRepoForm";

const RequiredRBACRoles: IRBACRole[] = [
  {
    apiGroup: "kubeapps.com",
    resource: "apprepositories",
    verbs: ["create"],
  },
  {
    apiGroup: "",
    resource: "secrets",
    verbs: ["create"],
  },
];

interface IAppRepoAddButtonProps {
  errors: {
    create?: Error;
    delete?: Error;
    fetch?: Error;
    update?: Error;
    validate?: Error;
  };
  onSubmit: (
    name: string,
    namespace: string,
    url: string,
    authHeader: string,
    customCA: string,
    syncJobPodTemplate: string,
    registrySecrets: string[],
  ) => Promise<boolean>;
  validate: (url: string, authHeader: string, customCA: string) => Promise<any>;
  validating: boolean;
  redirectTo?: string;
  namespace: string;
  kubeappsNamespace: string;
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
  text?: string;
  primary?: boolean;
  repo?: IAppRepository;
  secret?: ISecret;
}
interface IAppRepoAddButtonState {
  lastSubmittedName: string;
  modalIsOpen: boolean;
}

export class AppRepoAddButton extends React.Component<
  IAppRepoAddButtonProps,
  IAppRepoAddButtonState
> {
  public state = {
    lastSubmittedName: "",
    modalIsOpen: false,
  };

  public render() {
    const { redirectTo, text, primary, namespace, kubeappsNamespace } = this.props;
    return (
      <React.Fragment>
        <button className={`button ${primary ? "button-primary" : ""}`} onClick={this.openModal}>
          {text || "Add App Repository"}
        </button>
        <Modal
          isOpen={this.state.modalIsOpen}
          onRequestClose={this.closeModal}
          contentLabel="Modal"
        >
          {this.props.errors.create && (
            <ErrorSelector
              error={this.props.errors.create}
              defaultRequiredRBACRoles={{ create: RequiredRBACRoles }}
              action="create"
              namespace={this.props.namespace}
              resource={`App Repository ${this.state.lastSubmittedName}`}
            />
          )}
          <AppRepoForm
            onSubmit={this.onSubmit}
            validate={this.props.validate}
            onAfterInstall={this.closeModal}
            validating={this.props.validating}
            validationError={this.props.errors.validate}
            repo={this.props.repo}
            secret={this.props.secret}
            imagePullSecrets={this.props.imagePullSecrets}
            namespace={namespace}
            kubeappsNamespace={kubeappsNamespace}
            fetchImagePullSecrets={this.props.fetchImagePullSecrets}
            createDockerRegistrySecret={this.props.createDockerRegistrySecret}
          />
        </Modal>
        {redirectTo && <Redirect to={redirectTo} />}
      </React.Fragment>
    );
  }

  private closeModal = async () => this.setState({ modalIsOpen: false });
  private onSubmit = (
    name: string,
    url: string,
    authHeader: string,
    customCA: string,
    syncJobPodTemplate: string,
    registrySecrets: string[],
  ) => {
    // Store last submitted name to show it in an error if needed
    this.setState({ lastSubmittedName: name });
    return this.props.onSubmit(
      name,
      this.props.namespace,
      url,
      authHeader,
      customCA,
      syncJobPodTemplate,
      registrySecrets,
    );
  };
  private openModal = async () => this.setState({ modalIsOpen: true });
}
