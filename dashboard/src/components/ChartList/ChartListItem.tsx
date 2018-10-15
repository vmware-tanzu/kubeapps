import * as React from "react";

import placeholder from "../../placeholder.png";
import { IChart } from "../../shared/types";
import { PreformattedCard } from "../Card";

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
      <PreformattedCard
        key={`${repo}/${name}`}
        title={name}
        link={`/charts/${chart.id}`}
        info={latestAppVersion || "-"}
        icon={iconSrc}
        tag1Content={repo.name}
        tag1Class={repo.name}
      />
    );
  }
}

export default ChartListItem;
