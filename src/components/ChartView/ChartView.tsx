import * as React from "react";
import { RouterAction } from "react-router-redux";
import { IChart, IChartVersion } from "../../shared/types";
import ChartDeployButton from "./ChartDeployButton";

interface IChartViewProps {
  chartID: string;
  getChart: (id: string) => Promise<{}>;
  deployChart: (chart: IChart, releaseName: string, namespace: string) => Promise<{}>;
  push: (location: string) => RouterAction;
  isFetching: boolean;
  chart: IChart;
  version: IChartVersion;
}

class ChartView extends React.Component<IChartViewProps> {
  public componentDidMount() {
    const { chartID, getChart } = this.props;
    getChart(chartID);
  }

  public render() {
    const { isFetching, chart, deployChart, push } = this.props;
    if (isFetching || !chart) {
      return <div>Loading</div>;
    }
    return (
      <section className="ChartListView">
        <header className="ChartListView__header">
          <h1>{chart.id}</h1>
          <hr />
        </header>
        <main className="text-c">
          <ChartDeployButton push={push} chart={chart} deployChart={deployChart} />
        </main>
      </section>
    );
  }
}

export default ChartView;
