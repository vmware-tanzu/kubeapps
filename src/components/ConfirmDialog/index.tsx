import * as React from "react";
import * as Modal from "react-modal";

interface IConfirmDialogProps {
  modalIsOpen: boolean;
  onConfirm: () => Promise<any>;
  closeModal: () => Promise<any>;
}

interface IConfirmDialogState {
  error?: string;
  modalIsOpen: boolean;
}

class ConfirmDialog extends React.Component<IConfirmDialogProps, IConfirmDialogState> {
  public state: IConfirmDialogState = {
    error: undefined,
    modalIsOpen: this.props.modalIsOpen,
  };

  public render() {
    return (
      <div className="ConfirmDialog">
        <Modal
          style={{
            content: {
              bottom: "auto",
              left: "50%",
              marginRight: "-50%",
              right: "auto",
              top: "50%",
              transform: "translate(-50%, -50%)",
            },
          }}
          isOpen={this.props.modalIsOpen}
          onRequestClose={this.closeModal}
          contentLabel="Modal"
        >
          {this.state.error && (
            <div className="container padding-v-small bg-action">{this.state.error}</div>
          )}
          <div>Are you sure you want to delete this?</div>
          <div>
            <button className="button" onClick={this.props.closeModal}>
              Cancel
            </button>
            <button
              className="button button-primary button-danger"
              type="submit"
              onClick={this.props.onConfirm}
            >
              Delete
            </button>
          </div>
        </Modal>
      </div>
    );
  }

  public openModel = () => {
    this.setState({
      modalIsOpen: true,
    });
  };

  public closeModal = () => {
    this.setState({
      modalIsOpen: false,
    });
  };
}

export default ConfirmDialog;
