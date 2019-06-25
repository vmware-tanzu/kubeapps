import * as React from "react";
import { Link } from "react-router-dom";

import { IChartVersion } from "../../shared/types";
import * as url from "../../shared/url";

interface IChartVersionsListProps {
  selected: IChartVersion;
  versions: IChartVersion[];
}

interface IChartVersionsListState {
  showAll: boolean;
}

class ChartVersionsList extends React.Component<IChartVersionsListProps, IChartVersionsListState> {
  public state: IChartVersionsListState = {
    showAll: false,
  };

  public render() {
    const versions = this.state.showAll ? this.props.versions : this.props.versions.slice(0, 5);
    const items = versions.map(v => {
      const selectedClass =
        this.props.selected.attributes.version === v.attributes.version
          ? "type-bold type-color-action"
          : "";
      return (
        <li key={v.id}>
          <Link className={selectedClass} to={url.app.charts.version(v)}>
            {v.attributes.version} - {this.formatDate(v.attributes.created)}
          </Link>
        </li>
      );
    });
    return (
      <div className="ChartVersionsList">
        <ul className="remove-style padding-l-reset margin-b-reset">{items}</ul>
        {!this.state.showAll && this.props.versions.length > 5 && (
          <a className="type-small" onClick={this.handleShowAll}>
            Show all...
          </a>
        )}
      </div>
    );
  }

  public formatDate(dateStr: string) {
    const d = new Date(dateStr);
    return d.toLocaleDateString(undefined, { month: "short", day: "numeric", year: "numeric" });
  }

  public handleShowAll = () => {
    this.setState({
      showAll: true,
    });
  };
}

export default ChartVersionsList;
