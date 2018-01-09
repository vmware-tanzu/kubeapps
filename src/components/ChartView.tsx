import * as React from 'react';
import { AsyncAction, Chart, ChartVersion } from '../store/types';

interface Props {
  chartID: string;
  getChart: (id: string) => AsyncAction;
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
    const { isFetching, chart } = this.props;
    if (isFetching || !chart) {
      return <div>Loading</div>;
    }
    return <div>{chart.id}</div>;
  }
}

export default ChartView;
