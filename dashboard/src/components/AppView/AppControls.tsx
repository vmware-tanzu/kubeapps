import * as React from "react";
import { Redirect } from "react-router";

import { hapi } from "../../shared/hapi/release";
import ConfirmDialog from "../ConfirmDialog";

interface IAppControlsProps {
  app: hapi.release.Release;
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
    const { name, namespace } = this.props.app;
    if (!name || !namespace) {
      return <div> Loading </div>;
    }
    return (
      <div className="AppControls">
        {/* If the app has been deleted hide the upgrade button */}
        {this.props.app.info &&
          !this.props.app.info.deleted && (
            <button className="button" onClick={this.handleUpgradeClick}>
              Upgrade
            </button>
          )}
        {this.state.upgrade && (
          <Redirect push={true} to={`/apps/ns/${namespace}/upgrade/${name}`} />
        )}
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
