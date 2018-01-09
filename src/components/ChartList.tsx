import * as React from 'react';

import './ChartList.css';
import { ChartState, AsyncAction } from '../store/types';
import ChartListItem from './ChartListItem';

interface Props {
  charts: ChartState;
  repo: string;
  fetchCharts: (repo: string) => AsyncAction;
}

class ChartList extends React.Component<Props> {
  componentDidMount() {
    const { repo, fetchCharts } = this.props;
    fetchCharts(repo);
  }

  render() {
    let chartItems = this.props.charts.items.map(c => (<ChartListItem key={c.id} chart={c} />));
    return (
      <section className="ChartList">
        <header className="ChartList__header">
          <h1>Charts</h1>
          <hr />
        </header>
        <main className="text-c">
          <div className="ChartList__items">
            {chartItems}
          </div>
        </main>
      </section>
    );
  }
}

export default ChartList;
