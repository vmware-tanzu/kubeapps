import * as React from 'react';
import { Chart, ChartVersion } from '../shared/types';
import ChartDeployButton from './ChartDeployButton';
import { RouterAction } from 'react-router-redux';

interface Props {
  chartID: string;
  getChart: (id: string) => Promise<{}>;
  deployChart: (chart: Chart, releaseName: string, namespace: string) => Promise<{}>;
  push: (location: string) => RouterAction;
  isFetching: boolean;
  chart: Chart;
  version: ChartVersion;
}

class ChartView extends React.Component<Props> {
  componentDidMount() {
    const { chartID, getChart } = this.props;
    getChart(chartID);
  }

  render() {
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
