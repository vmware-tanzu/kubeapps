import * as React from "react";
import * as Modal from "react-modal";

interface IAddBindingButtonProps {
  bindingName: string;
  instanceRefName: string;
  namespace: string;
  addBinding: (bindingName: string, instanceName: string, namespace: string) => Promise<any>;
}

interface IAddBindingButtonState {
  modalIsOpen: boolean;
  // deployment options
  bindingName: string;
  instanceRefName: string;
  namespace: string;
  error?: string;
}

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
    const { modalIsOpen, bindingName, instanceRefName, namespace } = this.state;
    return (
      <div className="AddBindingButton">
        <button className="button button-primary" onClick={this.openModal}>
          Add Binding
        </button>
        <Modal isOpen={modalIsOpen} onRequestClose={this.closeModal}>
          {this.state.error && (
            <div className="container padding-v-bigger bg-action">{this.state.error}</div>
          )}
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
            <br />
            <label htmlFor="namespace">
              <span>Namespace:</span>
              <input
                type="text"
                id="namespace"
                value={namespace}
                onChange={this.handleNamespaceChange}
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
  private handleNamespaceChange = (e: React.ChangeEvent<HTMLInputElement>) =>
    this.setState({ namespace: e.target.value });
  private bind = async () => {
    await this.props.addBinding(
      this.state.bindingName,
      this.state.instanceRefName,
      this.state.namespace,
    );
    this.closeModal();
  };
}
