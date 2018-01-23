import * as React from "react";

import { IChartVersion } from "../../shared/types";

interface IChartVersionsListProps {
  versions: IChartVersion[];
}

class ChartVersionsList extends React.Component<IChartVersionsListProps> {
  public render() {
    const items = this.props.versions.slice(0, 5).map(v => (
      <li key={v.id}>
        {v.attributes.version} - {this.formatDate(v.attributes.created)}
      </li>
    ));
    return (
      <div className="ChartVersionsList">
        <ul className="remove-style padding-l-reset margin-b-reset">{items}</ul>
        <span className="type-small">Show all...</span>
      </div>
    );
  }

  public formatDate(dateStr: string) {
    const d = new Date(dateStr);
    return d.toLocaleDateString("en-US", { month: "short", day: "numeric", year: "numeric" });
  }
}

export default ChartVersionsList;
