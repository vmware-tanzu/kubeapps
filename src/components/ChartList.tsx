import * as React from 'react';
import { Dispatch } from 'redux';

import './ChartList.css';
import { ChartState, StoreState } from '../store/types';
import ChartListItem from './ChartListItem';

interface Props {
  charts: ChartState;
  fetchCharts: () => (dispatch: Dispatch<StoreState>) => Promise<{}>;
}

class ChartList extends React.Component<Props> {
  componentDidMount() {
    this.props.fetchCharts();
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
