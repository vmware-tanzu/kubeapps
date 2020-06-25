import * as React from "react";
import Modal from "react-modal";
import { ErrorSelector } from "../ErrorAlert";

interface INewNamespaceProps {
  modalIsOpen: boolean;
  namespace: string;
  onConfirm: () => any;
  closeModal: () => any;
  onChange: (e: React.FormEvent<HTMLInputElement>) => void;
  error?: Error;
}

interface INewNamespaceState {
  modalIsOpen: boolean;
  namespace: string;
}

class NewNamespace extends React.Component<INewNamespaceProps, INewNamespaceState> {
  public state: INewNamespaceState = {
    modalIsOpen: this.props.modalIsOpen,
    namespace: "",
  };

  public render() {
    return (
      <Modal
        className="centered-modal"
        isOpen={this.props.modalIsOpen}
        onRequestClose={this.props.closeModal}
        contentLabel="Modal"
      >
        {this.props.error && (
          <ErrorSelector
            error={this.props.error}
            resource={`namespace ${this.state.namespace}`}
            action="create"
          />
        )}

        <div>
          <div>
            <label htmlFor="new-ns">
              <div className="row">
                <div className="col-3 block">
                  <div className="centered">Namespace</div>
                </div>
                <div className="col-9 margin-t-big margin-l-big">
                  <input
                    id="new-ns"
                    onChange={this.props.onChange}
                    value={this.props.namespace}
                    type="text"
                  />
                  <span className="description">Introduce the name of the new namespace</span>
                </div>
              </div>
            </label>
          </div>
          <div className="margin-t-normal button-row">
            <button className="button" onClick={this.props.closeModal}>
              Cancel
            </button>
            <button className="button button-primary" type="submit" onClick={this.onConfirm}>
              Create
            </button>
          </div>
        </div>
      </Modal>
    );
  }

  private onConfirm = () => {
    this.setState({ namespace: this.props.namespace });
    this.props.onConfirm();
  };
}

export default NewNamespace;
