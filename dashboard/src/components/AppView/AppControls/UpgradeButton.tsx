import { RouterAction } from "connected-react-router";
import * as React from "react";
import { ArrowUpCircle } from "react-feather";
import { useSelector } from "react-redux";
import { IStoreState } from "shared/types";
import * as url from "../../../shared/url";

export interface IUpgradeButtonProps {
  newVersion?: boolean;
  releaseName: string;
  releaseNamespace: string;
  push: (location: string) => RouterAction;
}

const UpgradeButton: React.SFC<IUpgradeButtonProps> = props => {
  const { newVersion, push, releaseName, releaseNamespace } = props;
  const cluster = useSelector((state: IStoreState) => state.clusters.currentCluster);
  const onClick = () => push(url.app.apps.upgrade(cluster, releaseNamespace, releaseName));
  let upgradeButton = (
    <button className="button" onClick={onClick}>
      Upgrade
    </button>
  );
  // If the app is outdated highlight the upgrade button
  if (newVersion) {
    upgradeButton = (
      <button className="button upgrade-button" onClick={onClick}>
        <span className="upgrade-text">Upgrade</span>
        <ArrowUpCircle color="white" size={25} fill="#82C341" className="notification" />
      </button>
    );
  }
  return upgradeButton;
};

export default UpgradeButton;
