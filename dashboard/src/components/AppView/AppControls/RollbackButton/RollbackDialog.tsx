import * as React from "react";
import LoadingWrapper from "../../../LoadingWrapper";

import "./RollbackDialog.css";

interface IRollbackDialogProps {
  loading: boolean;
  currentRevision: number;
  onConfirm: (revision: number) => Promise<any>;
  closeModal: () => Promise<any>;
}

interface IRollbackDialogState {
  targetRevision: number;
}

class RollbackDialog extends React.Component<IRollbackDialogProps, IRollbackDialogState> {
  public state: IRollbackDialogState = {
    targetRevision: this.props.currentRevision - 1,
  };

  public render() {
    const options = [];
    // Use as options the number of versions without the latest
    for (let i = this.props.currentRevision - 1; i > 0; i--) {
      options.push(i);
    }
    return this.props.loading === true ? (
      <div className="row confirm-dialog-loading-info">
        <div className="col-8 loading-legend">Loading, please wait</div>
        <div className="col-4">
          <LoadingWrapper />
        </div>
      </div>
    ) : (
      <div>
        <label>
          Select the revision to which you want to rollback (current: {this.props.currentRevision})
        </label>
        <div>
          <select className="margin-t-normal" onChange={this.selectRevision}>
            {options.map(o => (
              <option key={o} value={o}>
                {o}
              </option>
            ))}
          </select>
        </div>
        <div className="margin-t-normal button-row">
          <button className="button" type="button" onClick={this.props.closeModal}>
            Cancel
          </button>
          <button
            className="button button-primary button-danger"
            type="submit"
            onClick={this.onConfirm}
          >
            Rollback
          </button>
        </div>
      </div>
    );
  }

  private selectRevision = (e: React.FormEvent<HTMLSelectElement>) => {
    this.setState({ targetRevision: Number(e.currentTarget.value) });
  };

  private onConfirm = () => {
    this.props.onConfirm(this.state.targetRevision);
  };
}

export default RollbackDialog;
