import * as React from "react";

import placeholder from "../../placeholder.png";
import { IAppOverview } from "../../shared/types";
import InfoCard from "../InfoCard";
import "./AppListItem.css";

interface IAppListItemProps {
  app: IAppOverview;
}

class AppListItem extends React.Component<IAppListItemProps> {
  public render() {
    const { app } = this.props;
    const icon = app.icon ? app.icon : placeholder;
    let banner;
    let bannerColor;
    if (app.updateInfo && !app.updateInfo.error && !app.updateInfo.upToDate) {
      if (app.chartMetadata.appVersion !== app.updateInfo.appLatestVersion) {
        // We assume that if there is a new chart version and the app version changes
        // this means that there is a new app version available
        // We cannot compare app versions since they don't follow the semver standard
        banner = `New app version ${app.updateInfo.appLatestVersion} available`;
        bannerColor = "blue";
      } else {
        // New chart version
        banner = `Chart v${app.updateInfo.chartLatestVersion} available`;
        bannerColor = "green";
      }
    }
    return (
      <InfoCard
        key={app.releaseName}
        link={`/apps/ns/${app.namespace}/${app.releaseName}`}
        title={app.releaseName}
        icon={icon}
        info={`${app.chart} v${app.version || "-"}`}
        banner={banner}
        bannerColor={bannerColor}
        tag1Content={app.namespace}
        tag2Content={app.status.toLocaleLowerCase()}
        tag2Class={app.status.toLocaleLowerCase()}
      />
    );
  }
}

export default AppListItem;
