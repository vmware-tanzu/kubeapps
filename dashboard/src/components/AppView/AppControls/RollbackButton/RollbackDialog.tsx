import * as React from "react";
import * as Modal from "react-modal";
import LoadingWrapper from "../../../LoadingWrapper";

import "./RollbackDialog.css";

interface IRollbackDialogProps {
  modalIsOpen: boolean;
  loading: boolean;
  revision: number;
  onConfirm: (revision: number) => () => Promise<any>;
  closeModal: () => Promise<any>;
}

interface IRollbackDialogState {
  error?: string;
  modalIsOpen: boolean;
  revision: number;
}

class RollbackDialog extends React.Component<IRollbackDialogProps, IRollbackDialogState> {
  public state: IRollbackDialogState = {
    error: undefined,
    modalIsOpen: this.props.modalIsOpen,
    revision: this.props.revision - 1,
  };

  public render() {
    const options = [];
    // Use as options the number of versions unless the latest
    for (let i = this.props.revision - 1; i > 0; i--) {
      options.push(i);
    }
    return (
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
            <div className="margin-b-normal"> Are you sure you want to rollback this release? </div>
            <label>Select the revision to rollback (current: {this.props.revision})</label>
            <select className="margin-t-normal" onChange={this.selectRevision}>
              {options.map(o => (
                <option key={o} value={o}>
                  {o}
                </option>
              ))}
            </select>
            <div className="margin-t-normal button-row">
              <button className="button" onClick={this.props.closeModal}>
                Cancel
              </button>
              <button
                className="button button-primary button-danger"
                type="submit"
                onClick={this.props.onConfirm(this.state.revision)}
              >
                Rollback
              </button>
            </div>
          </div>
        )}
      </Modal>
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

  private selectRevision = (e: React.FormEvent<HTMLSelectElement>) => {
    this.setState({ revision: Number(e.currentTarget.value) });
  };
}

export default RollbackDialog;
