import * as React from "react";
import * as Modal from "react-modal";
import LoadingWrapper from "../LoadingWrapper";
import "./ConfirmDialog.css";

interface IConfirmDialogProps {
  modalIsOpen: boolean;
  loading: boolean;
  extraElem?: JSX.Element;
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
          className="centered-modal"
          isOpen={this.props.modalIsOpen}
          onRequestClose={this.closeModal}
          contentLabel="Modal"
        >
          {this.state.error && (
            <div className="padding-big margin-b-big bg-action">{this.state.error}</div>
          )}
          {this.props.loading === true ? (
            <div className="row confirm-dialog-loading-info">
              <div className="col-8 loading-legend">Loading, please wait</div>
              <div className="col-4">
                <LoadingWrapper />
              </div>
            </div>
          ) : (
            <div>
              <div className="margin-b-normal"> Are you sure you want to delete this? </div>
              {this.props.extraElem}
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
          )}
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
