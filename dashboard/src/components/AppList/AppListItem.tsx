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
        <Link to={`/apps/ns/${release.namespace}/${release.name}`} title={release.name}>
          <CardIcon icon={iconSrc} />
          <CardContent>
            <div className="ChartListItem__content">
              <h3 className="ChartListItem__content__title type-big">{release.name}</h3>
              <div className="ChartListItem__content__info">
                <p className="ChartListItem__content__info_version margin-reset type-small padding-t-tiny type-color-light-blue">
                  {(metadata && metadata.appVersion) || "-"}
                </p>
                <span
                  className={`ChartListItem__content__info_repo ${
                    release.namespace
                  } type-small type-color-white padding-t-tiny padding-h-normal`}
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
