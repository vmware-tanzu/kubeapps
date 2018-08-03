import * as React from "react";
import * as Modal from "react-modal";

interface IConfirmDialogProps {
  modalIsOpen: boolean;
  loading: boolean;
  onConfirm: () => Promise<any>;
  closeModal: () => Promise<any>;
}

interface IConfirmDialogState {
  modalIsOpen: boolean;
}

class ConfirmDialog extends React.Component<IConfirmDialogProps, IConfirmDialogState> {
  public state: IConfirmDialogState = {
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
          ariaHideApp={false}
          isOpen={this.props.modalIsOpen}
          onRequestClose={this.closeModal}
          contentLabel="Modal"
        >
          {this.props.loading === true ? (
            <div> Loading ... </div>
          ) : (
            <div>
              <div> Are you sure you want to delete this? </div>
              <button id="cancel" className="button" onClick={this.props.closeModal}>
                Cancel
              </button>
              <button
                id="delete"
                className="button button-primary button-danger"
                type="submit"
                onClick={this.props.onConfirm}
              >
                Delete
              </button>
            </div>
          )}
        </Modal>
      </div>
    );
  }

  public closeModal = () => {
    this.setState({
      modalIsOpen: false,
    });
  };
}

export default ConfirmDialog;
