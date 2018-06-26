import * as React from "react";
import { Link } from "react-router-dom";

import placeholder from "../../placeholder.png";
import { hapi } from "../../shared/hapi/release";
import Card, { CardContent, CardIcon } from "../Card";
import "../ChartList/ChartListItem.css";

interface IAppListItemProps {
  app: hapi.release.Release;
}

class AppListItem extends React.Component<IAppListItemProps> {
  public render() {
    const { app } = this.props;
    const metadata = app.chart && app.chart.metadata;
    const icon = metadata && metadata.icon;
    const iconSrc = icon ? icon : placeholder;

    return (
      <Card key={app.name} responsive={true} className="AppListItem">
        <Link to={`/apps/ns/${app.namespace}/${app.name}`} title={app.name}>
          <CardIcon icon={iconSrc} />
          <CardContent>
            <div className="ChartListItem__content">
              <h3 className="ChartListItem__content__title type-big">{app.name}</h3>
              <div className="ChartListItem__content__info">
                <p className="ChartListItem__content__info_version margin-reset type-small padding-t-tiny type-color-light-blue">
                  {(metadata && metadata.appVersion) || "-"}
                </p>
                <span
                  className={`ChartListItem__content__info_repo ${
                    app.namespace
                  } type-small type-color-white padding-t-tiny padding-h-normal`}
                >
                  {app.namespace}
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
