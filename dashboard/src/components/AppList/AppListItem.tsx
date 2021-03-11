import helmIcon from "../../icons/helm.svg";
import placeholder from "../../placeholder.png";
import { IAppOverview } from "../../shared/types";
import * as url from "../../shared/url";
import InfoCard from "../InfoCard/InfoCard";

import Tooltip from "components/js/Tooltip";
import "./AppListItem.css";

export interface IAppListItemProps {
  app: IAppOverview;
  cluster: string;
}

function AppListItem(props: IAppListItemProps) {
  const { app, cluster } = props;
  const icon = app.icon ? app.icon : placeholder;
  const appStatus = app.status.toLocaleLowerCase();
  let tooltip = <></>;
  const updateAvailable = app.updateInfo && !app.updateInfo.error && !app.updateInfo.upToDate;
  if (app.updateInfo && updateAvailable) {
    if (app.updateInfo.appLatestVersion !== app.chartMetadata.appVersion) {
      tooltip = (
        <div className="color-icon-info">
          <Tooltip
            label="update-tooltip"
            id={`${app.releaseName}-update-tooltip`}
            icon="circle-arrow"
            position="top-left"
            iconProps={{ solid: true, size: "md", color: "blue" }}
          >
            New App Version: {app.updateInfo.appLatestVersion}
          </Tooltip>
        </div>
      );
    } else {
      tooltip = (
        <div className="color-icon-info">
          <Tooltip
            label="update-tooltip"
            id={`${app.releaseName}-update-tooltip`}
            icon="circle-arrow"
            position="top-left"
            iconProps={{ solid: true, size: "md" }}
          >
            New Chart Version: {app.updateInfo.chartLatestVersion}
          </Tooltip>
        </div>
      );
    }
  }
  return (
    <InfoCard
      key={`${app.namespace}/${app.releaseName}`}
      link={url.app.apps.get(cluster, app.namespace, app.releaseName)}
      title={app.releaseName}
      icon={icon}
      info={
        <div>
          <span>
            App: {app.chartMetadata.name}{" "}
            {app.chartMetadata.appVersion
              ? `v${app.chartMetadata.appVersion.replace(/^v/, "")}`
              : ""}
          </span>
          <br />
          <span>Chart: {app.chartMetadata.version}</span>
        </div>
      }
      description={app.chartMetadata.description}
      tag1Content={appStatus}
      tag1Class={appStatus === "deployed" ? "label-success" : "label-warning"}
      tooltip={tooltip}
      bgIcon={helmIcon}
    />
  );
}

export default AppListItem;
