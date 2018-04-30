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
        <Link to={`/apps/ns/${release.namespace}/${release.name}`}>
          <CardIcon icon={iconSrc} />
          <CardContent>
            <div className="ChartListItem__content">
              <div className="ChartListItem__content__title type-big">{release.name}</div>
              <div className="ChartListItem__content__info">
                <div className="ChartListItem__content__info_version type-small padding-t-tiny type-color-light-blue">
                  {(metadata && metadata.appVersion) || "-"}
                </div>
                <div
                  className={`ChartListItem__content__info_repo ${
                    release.namespace
                  } type-small type-color-white padding-t-tiny padding-h-normal`}
                >
                  {release.namespace}
                </div>
              </div>
            </div>
          </CardContent>
        </Link>
      </Card>
    );
  }
}

export default AppListItem;
