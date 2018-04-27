import * as React from "react";
import { Link } from "react-router-dom";

import { IChartState } from "../../shared/types";
import { CardGrid } from "../Card";
import ChartListItem from "./ChartListItem";

import "./ChartList.css";

interface IChartListProps {
  charts: IChartState;
  repo: string;
  fetchCharts: (repo: string) => Promise<{}>;
}

class ChartList extends React.Component<IChartListProps> {
  public componentDidMount() {
    const { repo, fetchCharts } = this.props;
    fetchCharts(repo);
  }

  public render() {
    let chartItems;
    if (this.props.charts.items) {
      chartItems = this.props.charts.items.map(c => <ChartListItem key={c.id} chart={c} />);
    } else {
      chartItems = (
        <div className="ChartList__error_nocharts">
          No charts available. Manage your Helm chart repositories in Kubeapps by visiting the{" "}
          <Link to={`/config/repos`}>App repositories configuration</Link> page.
        </div>
      );
    }
    return (
      <section className="ChartList">
        <header className="ChartList__header">
          <h1>Charts</h1>
          <hr />
        </header>
        <CardGrid>{chartItems}</CardGrid>
      </section>
    );
  }
}

export default ChartList;
