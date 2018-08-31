import * as React from "react";
import { Link } from "react-router-dom";

import placeholder from "../../placeholder.png";
import { IAppOverview } from "../../shared/types";
import Card, { CardContent, CardIcon } from "../Card";
import "../ChartList/ChartListItem.css";

interface IAppListItemProps {
  app: IAppOverview;
}

class AppListItem extends React.Component<IAppListItemProps> {
  public render() {
    const { app } = this.props;
    const icon = app.icon ? app.icon : placeholder;

    return (
      <Card key={app.releaseName} responsive={true} className="AppListItem">
        <Link to={`/apps/ns/${app.namespace}/${app.releaseName}`} title={app.releaseName}>
          <CardIcon icon={icon} />
          <CardContent>
            <div className="ChartListItem__content">
              <h3 className="ChartListItem__content__title type-big">{app.releaseName}</h3>
              <div className="ChartListItem__content__info">
                <p className="ChartListItem__content__info_version margin-reset type-small padding-t-tiny type-color-light-blue">
                  {app.version || "-"}
                </p>
                <div>
                  <span
                    className={
                      "ChartListItem__content__info_repo type-small type-color-white padding-t-tiny padding-h-normal"
                    }
                  >
                    {app.namespace}
                  </span>
                  <span
                    className={`ChartListItem__content__info_other ${
                      app.status
                    } type-small type-color-white padding-t-tiny padding-h-normal`}
                  >
                    {app.status.toLowerCase()}
                  </span>
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
