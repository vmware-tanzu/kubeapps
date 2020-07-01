import * as React from "react";
import helmIcon from "../../icons/helm.svg";
import placeholder from "../../placeholder.png";
import { IAppOverview } from "../../shared/types";
import * as url from "../../shared/url";
import InfoCard from "../InfoCard";
import "./AppListItem.css";

export interface IAppListItemProps {
  app: IAppOverview;
  cluster: string;
}

class AppListItem extends React.Component<IAppListItemProps> {
  public render() {
    const { app, cluster } = this.props;
    const icon = app.icon ? app.icon : placeholder;
    const banner =
      app.updateInfo && !app.updateInfo.error && !app.updateInfo.upToDate
        ? "Update available"
        : undefined;
    return (
      <InfoCard
        key={app.releaseName}
        link={url.app.apps.get(cluster, app.namespace, app.releaseName)}
        title={app.releaseName}
        icon={icon}
        info={`${app.chart} v${app.version || "-"}`}
        banner={banner}
        tag1Content={app.namespace}
        tag2Content={app.status.toLocaleLowerCase()}
        tag2Class={app.status.toLocaleLowerCase()}
        subIcon={helmIcon}
      />
    );
  }
}

export default AppListItem;
