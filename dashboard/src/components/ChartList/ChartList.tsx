import * as React from "react";

import { Link } from "react-router-dom";
import { IChartState } from "../../shared/types";
import { NotFoundErrorAlert } from "../ErrorAlert";

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
    if (!this.props.charts.items) {
      return (
        <NotFoundErrorAlert
          resource={"Charts"}
          children={
            <div>
              Manage your Helm chart repositories in Kubeapps by visiting the{" "}
              <Link to={`/config/repos`}>App repositories configuration</Link> page.
            </div>
          }
        />
      );
    }

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
