import * as React from "react";
import { Save, Trash2 } from "react-feather";
import { Redirect } from "react-router";

import ConfirmDialog from "../ConfirmDialog";

interface IFunctionControlsProps {
  enableSaveButton: boolean;
  deleteFunction: () => Promise<boolean>;
  updateFunction: () => void;
  namespace: string;
}

interface IFunctionControlsState {
  modalIsOpen: boolean;
  redirectToFunctionsList: boolean;
}

class FunctionControls extends React.Component<IFunctionControlsProps, IFunctionControlsState> {
  public state: IFunctionControlsState = {
    modalIsOpen: false,
    redirectToFunctionsList: false,
  };

  public render() {
    const { namespace } = this.props;
    return (
      <div className="FunctionControls">
        <button
          className="button"
          onClick={this.props.updateFunction}
          disabled={!this.props.enableSaveButton}
        >
          <Save className="icon" /> Save
        </button>
        <button className="button button-danger" onClick={this.openModel}>
          <Trash2 className="icon" /> Delete
        </button>
        <ConfirmDialog
          onConfirm={this.handleDeleteClick}
          loading={false}
          modalIsOpen={this.state.modalIsOpen}
          closeModal={this.closeModal}
        />
        {this.state.redirectToFunctionsList && <Redirect to={`/functions/ns/${namespace}`} />}
      </div>
    );
  }

  public openModel = () => {
    this.setState({
      modalIsOpen: true,
    });
  };

  public closeModal = async () => {
    this.setState({
      modalIsOpen: false,
    });
  };

  public handleDeleteClick = async () => {
    const deleted = await this.props.deleteFunction();
    const s: Partial<IFunctionControlsState> = { modalIsOpen: false };
    if (deleted) {
      s.redirectToFunctionsList = true;
    }
    this.setState(s as IFunctionControlsState);
  };
}

export default FunctionControls;
