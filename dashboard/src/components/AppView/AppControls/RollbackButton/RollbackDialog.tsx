import * as React from "react";
import LoadingWrapper from "../../../LoadingWrapper";

import "./RollbackDialog.css";

interface IRollbackDialogProps {
  loading: boolean;
  revision: number;
  onConfirm: (revision: number) => () => Promise<any>;
  closeModal: () => Promise<any>;
}

interface IRollbackDialogState {
  revision: number;
}

class RollbackDialog extends React.Component<IRollbackDialogProps, IRollbackDialogState> {
  public state: IRollbackDialogState = {
    revision: this.props.revision - 1,
  };

  public render() {
    const options = [];
    // Use as options the number of versions without the latest
    for (let i = this.props.revision - 1; i > 0; i--) {
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
          <button className="button" type="button" onClick={this.props.closeModal}>
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
    );
  }

  private selectRevision = (e: React.FormEvent<HTMLSelectElement>) => {
    this.setState({ revision: Number(e.currentTarget.value) });
  };
}

export default RollbackDialog;
