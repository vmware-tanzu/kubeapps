import * as React from "react";
import { ArrowUpCircle } from "react-feather";
import { Redirect } from "react-router";

import { IChartUpdate } from "shared/types";
import { hapi } from "../../shared/hapi/release";
import ConfirmDialog from "../ConfirmDialog";
import LoadingWrapper from "../LoadingWrapper";
import "./AppControls.css";

interface IAppControlsProps {
  app: hapi.release.Release;
  updates?: IChartUpdate[];
  deleteApp: (purge: boolean) => Promise<boolean>;
}

interface IAppControlsState {
  migrate: boolean;
  modalIsOpen: boolean;
  redirectToAppList: boolean;
  upgrade: boolean;
  deleting: boolean;
  purge: boolean;
}

class AppControls extends React.Component<IAppControlsProps, IAppControlsState> {
  public state: IAppControlsState = {
    deleting: false,
    migrate: false,
    modalIsOpen: false,
    purge: false,
    redirectToAppList: false,
    upgrade: false,
  };

  public render() {
    const { app } = this.props;
    const { name, namespace } = app;
    const deleted = app.info && app.info.deleted;
    if (!name || !namespace) {
      return <LoadingWrapper />;
    }
    return (
      <div className="AppControls">
        {/* If the app has been deleted hide the upgrade button */}
        {this.renderUpgradeButton()}
        {this.state.upgrade && (
          <Redirect push={true} to={`/apps/ns/${namespace}/upgrade/${name}`} />
        )}
        <button className="button button-danger" onClick={this.openModel}>
          {deleted ? "Purge" : "Delete"}
        </button>
        <ConfirmDialog
          onConfirm={this.handleDeleteClick}
          modalIsOpen={this.state.modalIsOpen}
          loading={this.state.deleting}
          closeModal={this.closeModal}
          extraElem={
            deleted ? (
              undefined
            ) : (
              <div className="margin-b-normal text-c">
                <label className="checkbox margin-r-big">
                  <input type="checkbox" onChange={this.togglePurge} />
                  <span>Purge release</span>
                </label>
              </div>
            )
          }
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
    // Purge the release if the application has been already deleted
    const alreadyDeleted = this.props.app.info && !!this.props.app.info.deleted;
    const deleted = await this.props.deleteApp(alreadyDeleted || this.state.purge);
    const s: Partial<IAppControlsState> = { modalIsOpen: false };
    if (deleted) {
      s.redirectToAppList = true;
    }
    this.setState(s as IAppControlsState);
  };

  private togglePurge = () => {
    this.setState({ purge: !this.state.purge });
  };

  private renderUpgradeButton = () => {
    const { app, updates } = this.props;
    const deleted = app.info && app.info.deleted;
    let upgradeButton = null;
    // If the app has been deleted hide the upgrade button
    if (!deleted) {
      upgradeButton = (
        <button className="button" onClick={this.handleUpgradeClick}>
          Upgrade
        </button>
      );
      // If the app is outdated highlight the upgrade button
      if (updates && updates.length > 0) {
        upgradeButton = (
          <div className="tooltip">
            <button className="button upgrade-button" onClick={this.handleUpgradeClick}>
              <span className="upgrade-text">Upgrade</span>
              <ArrowUpCircle color="white" size={25} fill="#82C341" className="notification" />
            </button>
            <span className="tooltiptext tooltip-top">
              New version(s) ({updates.map(u => u.latestVersion).join(", ")}) found!
            </span>
          </div>
        );
      }
    }
    return upgradeButton;
  };
}

export default AppControls;
