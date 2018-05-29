import * as React from "react";
import { Redirect } from "react-router";

import { IApp } from "../../shared/types";
import ConfirmDialog from "../ConfirmDialog";

interface IAppControlsProps {
  app: IApp;
  deleteApp: () => Promise<boolean>;
}

interface IAppControlsState {
  migrate: boolean;
  modalIsOpen: boolean;
  redirectToAppList: boolean;
  upgrade: boolean;
  deleting: boolean;
}

class AppControls extends React.Component<IAppControlsProps, IAppControlsState> {
  public state: IAppControlsState = {
    deleting: false,
    migrate: false,
    modalIsOpen: false,
    redirectToAppList: false,
    upgrade: false,
  };

  public render() {
    const { name, namespace } = this.props.app.data;
    if (this.props.app.hr && this.props.app.hr.metadata) {
      return (
        <div className="AppControls">
          <button className="button" onClick={this.handleUpgradeClick}>
            Upgrade
          </button>
          {this.state.upgrade && <Redirect push={true} to={`/apps/ns/${namespace}/edit/${name}`} />}
          <button className="button button-danger" onClick={this.openModel}>
            Delete
          </button>
          <ConfirmDialog
            onConfirm={this.handleDeleteClick}
            modalIsOpen={this.state.modalIsOpen}
            loading={this.state.deleting}
            closeModal={this.closeModal}
          />
          {this.state.redirectToAppList && <Redirect to={`/apps/ns/${namespace}`} />}
        </div>
      );
    }
    return (
      <div className="AppControls">
        <button className="button" onClick={this.handleMigrateClick}>
          Import App
        </button>
        {this.state.migrate && (
          <Redirect push={true} to={`/apps/ns/${namespace}/migrate/${name}`} />
        )}
      </div>
    );
  }

  public handleUpgradeClick = () => {
    this.setState({ upgrade: true });
  };

  public handleMigrateClick = () => {
    this.setState({ migrate: true });
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
    this.setState({ deleting: true });
    const deleted = await this.props.deleteApp();
    const s: Partial<IAppControlsState> = { modalIsOpen: false };
    if (deleted) {
      s.redirectToAppList = true;
    }
    this.setState(s as IAppControlsState);
  };
}

export default AppControls;
