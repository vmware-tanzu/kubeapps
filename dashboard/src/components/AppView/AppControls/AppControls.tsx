import { RouterAction } from "connected-react-router";
import * as React from "react";
import { Redirect } from "react-router";

import { IRelease } from "shared/types";
import RollbackButtonContainer from "../../../containers/RollbackButtonContainer";
import ConfirmDialog from "../../ConfirmDialog";
import LoadingWrapper from "../../LoadingWrapper";
import "./AppControls.css";
import UpgradeButton from "./UpgradeButton";

interface IAppControlsProps {
  app: IRelease;
  deleteApp: (purge: boolean) => Promise<boolean>;
  push: (location: string) => RouterAction;
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
    const { app, push } = this.props;
    const { name, namespace } = app;
    const deleted = app.info && app.info.deleted;
    if (!name || !namespace) {
      return <LoadingWrapper />;
    }
    return (
      <div className="AppControls">
        {/* If the app has been deleted hide the upgrade button */}
        {!deleted && (
          <React.Fragment>
            <UpgradeButton
              newVersion={app.updateInfo && !app.updateInfo.upToDate}
              releaseName={name}
              releaseNamespace={namespace}
              push={push}
            />
            {/* We only show the rollback button if there are versions to rollback */}
            {app.version > 1 && (
              <RollbackButtonContainer
                releaseName={name}
                namespace={namespace}
                revision={app.version}
              />
            )}
          </React.Fragment>
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
}

export default AppControls;
