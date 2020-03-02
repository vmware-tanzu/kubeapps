import * as React from "react";
import * as Modal from "react-modal";
import { Redirect } from "react-router";

import { IRBACRole } from "../../../shared/types";
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
  install: (
    name: string,
    namespace: string,
    url: string,
    authHeader: string,
    customCA: string,
    syncJobPodTemplate: string,
  ) => Promise<boolean>;
  validate: (url: string, authHeader: string, customCA: string) => Promise<any>;
  isFetching: boolean;
  redirectTo?: string;
  namespace: string;
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
    const { redirectTo } = this.props;
    return (
      <React.Fragment>
        <button className="button button-primary" onClick={this.openModal}>
          Add App Repository
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
            install={this.install}
            validate={this.props.validate}
            onAfterInstall={this.closeModal}
            isFetching={this.props.isFetching}
            validationError={this.props.errors.validate}
          />
        </Modal>
        {redirectTo && <Redirect to={redirectTo} />}
      </React.Fragment>
    );
  }

  private closeModal = async () => this.setState({ modalIsOpen: false });
  private install = (
    name: string,
    url: string,
    authHeader: string,
    customCA: string,
    syncJobPodTemplate: string,
  ) => {
    // Store last submitted name to show it in an error if needed
    this.setState({ lastSubmittedName: name });
    return this.props.install(
      name,
      this.props.namespace,
      url,
      authHeader,
      customCA,
      syncJobPodTemplate,
    );
  };
  private openModal = async () => this.setState({ modalIsOpen: true });
}
