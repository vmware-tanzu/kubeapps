import * as React from "react";

import { IChartState } from "../../shared/types";
import { CardGrid } from "../Card";
import ChartListItem from "./ChartListItem";

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
    const chartItems = this.props.charts.items.map(c => <ChartListItem key={c.id} chart={c} />);
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
