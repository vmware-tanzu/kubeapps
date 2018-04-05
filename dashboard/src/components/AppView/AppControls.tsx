import * as React from "react";
import { Redirect } from "react-router";

import { IApp } from "../../shared/types";
import ConfirmDialog from "../ConfirmDialog";

interface IAppControlsProps {
  app: IApp;
  deleteApp: () => Promise<void>;
}

interface IAppControlsState {
  modalIsOpen: boolean;
  redirectToAppList: boolean;
  upgrade: boolean;
}

class AppControls extends React.Component<IAppControlsProps, IAppControlsState> {
  public state: IAppControlsState = {
    modalIsOpen: false,
    redirectToAppList: false,
    upgrade: false,
  };

  public render() {
    const { name, namespace } = this.props.app.data;
    return (
      <div className="AppControls">
        <button className="button" onClick={this.handleUpgradeClick}>
          Upgrade
        </button>
        {this.state.upgrade && <Redirect to={`/apps/edit/${namespace}/${name}`} />}
        <button className="button button-danger" onClick={this.openModel}>
          Delete
        </button>
        <ConfirmDialog
          onConfirm={this.handleDeleteClick}
          modalIsOpen={this.state.modalIsOpen}
          closeModal={this.closeModal}
        />
        {this.state.redirectToAppList && <Redirect to="/" />}
      </div>
    );
  }

  public handleUpgradeClick = () => {
    this.setState({ upgrade: true });
  };

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
    await this.props.deleteApp();
    this.setState({
      modalIsOpen: false,
      redirectToAppList: true,
    });
  };
}

export default AppControls;
