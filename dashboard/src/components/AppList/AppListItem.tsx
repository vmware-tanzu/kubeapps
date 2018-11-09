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

    return (
      <InfoCard
        key={app.releaseName}
        link={`/apps/ns/${app.namespace}/${app.releaseName}`}
        title={app.releaseName}
        icon={icon}
        info={`${app.chart} v${app.version || "-"}`}
        tag1Content={app.namespace}
        tag2Content={app.status.toLocaleLowerCase()}
        tag2Class={app.status.toLocaleLowerCase()}
      />
    );
  }
}

export default AppListItem;
