import * as React from "react";
import * as Modal from "react-modal";

import { ForbiddenError, IRBACRole, NotFoundError } from "../../shared/types";
import { NotFoundErrorAlert, PermissionsErrorAlert, UnexpectedErrorAlert } from "../ErrorAlert";

interface IAddBindingButtonProps {
  error?: Error;
  bindingName: string;
  instanceRefName: string;
  namespace: string;
  addBinding: (bindingName: string, instanceName: string, namespace: string) => Promise<boolean>;
  onAddBinding: () => void;
}

interface IAddBindingButtonState {
  modalIsOpen: boolean;
  // deployment options
  bindingName: string;
  instanceRefName: string;
}

const RequiredRBACRoles: IRBACRole[] = [
  {
    apiGroup: "servicecatalog.k8s.io",
    resource: "servicebindings",
    verbs: ["create"],
  },
];

export class AddBindingButton extends React.Component<
  IAddBindingButtonProps,
  IAddBindingButtonState
> {
  public state = {
    error: undefined,
    modalIsOpen: false,
    ...this.props,
  };

  public render() {
    const { modalIsOpen, bindingName, instanceRefName } = this.state;
    return (
      <div className="AddBindingButton">
        <button className="button button-primary" onClick={this.openModal}>
          Add Binding
        </button>
        <Modal isOpen={modalIsOpen} onRequestClose={this.closeModal}>
          {this.props.error && <div className="margin-b-big">{this.renderError()}</div>}
          <div className="bind-form">
            <h1>Add Binding</h1>
            <label htmlFor="binding-name">
              <span>Name:</span>
              <input
                type="text"
                id="binding-name"
                value={bindingName}
                onChange={this.handleNameChange}
              />
            </label>
            <br />
            <label htmlFor="instance-ref-name">
              <span>Instance Name:</span>
              <input
                type="text"
                id="instance-ref-name"
                value={instanceRefName}
                onChange={this.handleInstanceNameChange}
              />
            </label>
            <button className="button button-primary" onClick={this.bind}>
              Create Binding
            </button>
            <button className="button" onClick={this.closeModal}>
              Cancel
            </button>
          </div>
        </Modal>
      </div>
    );
  }

  private closeModal = () => this.setState({ modalIsOpen: false });
  private openModal = () => this.setState({ modalIsOpen: true });
  private handleNameChange = (e: React.ChangeEvent<HTMLInputElement>) =>
    this.setState({ bindingName: e.target.value });
  private handleInstanceNameChange = (e: React.ChangeEvent<HTMLInputElement>) =>
    this.setState({ instanceRefName: e.target.value });
  private bind = async () => {
    const added = await this.props.addBinding(
      this.state.bindingName,
      this.state.instanceRefName,
      this.props.namespace,
    );
    if (added) {
      this.closeModal();
      this.props.onAddBinding();
    }
  };

  private renderError() {
    const { error, namespace } = this.props;
    const { bindingName } = this.state;
    switch (error && error.constructor) {
      case ForbiddenError:
        return (
          <PermissionsErrorAlert
            namespace={namespace}
            roles={RequiredRBACRoles}
            action={`create Service Binding "${bindingName}"`}
          />
        );
      case NotFoundError:
        return <NotFoundErrorAlert resource={`Namespace "${namespace}"`} />;
      default:
        return <UnexpectedErrorAlert />;
    }
  }
}
