import * as React from "react";
import { Link } from "react-router-dom";

import { IChart } from "../../shared/types";
import ChartIcon from "../ChartIcon";
import "./ChartListItem.css";

interface IChartListItemProps {
  chart: IChart;
}

class ChartListItem extends React.Component<IChartListItemProps> {
  public render() {
    const { chart } = this.props;
    const latestAppVersion = chart.relationships.latestChartVersion.data.app_version;
    return (
      <div className="ChartListItem padding-normal margin-big elevation-5">
        <Link to={`/charts/` + chart.id}>
          <ChartIcon icon={chart.attributes.icon} />
          <div className="ChartListItem__details">
            <h6>{chart.id}</h6>
            {latestAppVersion && <span>v{latestAppVersion}</span>}
          </div>
        </Link>
      </div>
    );
  }
}

export default ChartListItem;
