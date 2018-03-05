import * as React from "react";
import { Link } from "react-router-dom";

import placeholder from "../../placeholder.png";
import { IChart } from "../../shared/types";

import Card, { CardContent, CardIcon } from "../Card";

import "./ChartListItem.css";

interface IChartListItemProps {
  chart: IChart;
}

class ChartListItem extends React.Component<IChartListItemProps> {
  public render() {
    const { chart } = this.props;
    const { icon, name, repo } = chart.attributes;
    const iconSrc = icon ? `/api/chartsvc/${icon}` : placeholder;
    const latestAppVersion = chart.relationships.latestChartVersion.data.app_version;
    return (
      <Card key={`${repo}/${name}`} responsive={true} className="ChartListItem">
        <Link to={`/charts/` + chart.id}>
          <CardIcon icon={iconSrc} />
          <CardContent>
            <div className="ChartListItem__content">
              <h3 className="ChartListItem__content__title">{name}</h3>
              <div className="ChartListItem__content__info text-r">
                <p className="margin-reset type-color-light-blue">{latestAppVersion || "-"}</p>
                <span
                  className={`ChartListItem__content__repo ${repo.name} padding-tiny
                  padding-h-normal type-small margin-t-small`}
                >
                  {repo.name}
                </span>
              </div>
            </div>
          </CardContent>
        </Link>
      </Card>
    );
  }
}

export default ChartListItem;
