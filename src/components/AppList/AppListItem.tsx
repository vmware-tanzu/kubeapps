import * as React from "react";
import { Link } from "react-router-dom";

import placeholder from "../../placeholder.png";
import { IApp } from "../../shared/types";
import Card, { CardContent, CardIcon } from "../Card";
import "../ChartList/ChartListItem.css";

interface IAppListItemProps {
  app: IApp;
}

class AppListItem extends React.Component<IAppListItemProps> {
  public render() {
    const { app } = this.props;
    const release = app.data;
    const icon = app.chart && app.chart.attributes.icon;
    const iconSrc = icon ? `/api/chartsvc/${icon}` : placeholder;
    const metadata = release.chart && release.chart.metadata;

    return (
      <Card key={release.name} responsive={true} className="AppListItem">
        <Link to={`/apps/${release.namespace}/${release.name}`}>
          <CardIcon icon={iconSrc} />
          <CardContent>
            <div className="ChartListItem__content">
              <h3 className="ChartListItem__content__title">{release.name}</h3>
              <div className="ChartListItem__content__info text-r">
                <p className="margin-reset type-color-light-blue">
                  {(metadata && metadata.appVersion) || "-"}
                </p>
                <span
                  className={`ChartListItem__content__repo padding-tiny
                  padding-h-normal type-small margin-t-small`}
                >
                  {release.namespace}
                </span>
              </div>
            </div>
          </CardContent>
        </Link>
      </Card>
    );
  }
}

export default AppListItem;
