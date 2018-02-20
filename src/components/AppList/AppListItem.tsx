import * as React from "react";
import { Link } from "react-router-dom";

import { hapi } from "../../shared/hapi/release";
import { IApp } from "../../shared/types";

import ChartIcon from "../ChartIcon";
import "./AppListItem.css";

interface IAppListItemProps {
  app: IApp;
}

class AppListItem extends React.Component<IAppListItemProps> {
  public render() {
    const { app } = this.props;
    let release: hapi.release.Release | undefined;
    release = app.data;
    let nameSpace: string | undefined;
    if (release && release.chart && release.chart.metadata) {
      nameSpace = `${release.namespace}`;
    }

    return (
      <div className="AppListItem padding-normal margin-big elevation-5">
        <Link to={`/apps/` + nameSpace + `/` + release.name}>
          <div className="AppListList__details">
            <ChartIcon icon={app.chart && app.chart.attributes.icon} />
            <h6>{release.name}</h6>
          </div>
        </Link>
      </div>
    );
  }
}

export default AppListItem;
