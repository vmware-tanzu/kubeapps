import * as React from "react";
import { ArrowUpCircle } from "react-feather";
import { Redirect } from "react-router";

interface IUpgradeButtonProps {
  updateVersion?: string;
  upgradeURL: string;
}

interface IUpgradeButtonState {
  upgrade: boolean;
}

class UpgradeButton extends React.Component<IUpgradeButtonProps, IUpgradeButtonState> {
  public state: IUpgradeButtonState = {
    upgrade: false,
  };

  public render() {
    const { updateVersion, upgradeURL } = this.props;
    let upgradeButton = (
      <button className="button" onClick={this.handleUpgradeClick}>
        Upgrade
      </button>
    );
    // If the app is outdated highlight the upgrade button
    if (updateVersion) {
      upgradeButton = (
        <div className="tooltip">
          <button className="button upgrade-button" onClick={this.handleUpgradeClick}>
            <span className="upgrade-text">Upgrade</span>
            <ArrowUpCircle color="white" size={25} fill="#82C341" className="notification" />
          </button>
          <span className="tooltiptext tooltip-top">New version ({updateVersion}) found!</span>
        </div>
      );
    }
    return (
      <React.Fragment>
        {upgradeButton}
        {this.state.upgrade && <Redirect push={true} to={upgradeURL} />}
      </React.Fragment>
    );
  }

  private handleUpgradeClick = () => {
    this.setState({ upgrade: true });
  };
}

export default UpgradeButton;
