import * as React from "react";
import { Save, Trash2 } from "react-feather";
import { Redirect } from "react-router";

import ConfirmDialog from "../ConfirmDialog";

interface IFunctionControlsProps {
  enableSaveButton: boolean;
  deleteFunction: () => Promise<void>;
  updateFunction: () => void;
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
          modalIsOpen={this.state.modalIsOpen}
          closeModal={this.closeModal}
        />
        {this.state.redirectToFunctionsList && <Redirect to="/functions" />}
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
    await this.props.deleteFunction();
    this.setState({
      modalIsOpen: false,
      redirectToFunctionsList: true,
    });
  };
}

export default FunctionControls;
