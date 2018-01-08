import * as React from 'react';
import { Link } from 'react-router-dom';

import { Chart } from '../store/types';
import './ChartListItem.css';

const placeholder = require('../placeholder.png');

interface Props {
  chart: Chart;
}

class ChartListItem extends React.Component<Props> {
  render() {
    const { chart } = this.props;
    return (
      <div className="ChartListItem padding-normal margin-big elevation-5">
        <Link to={`/charts/`+chart.id}>
          <div className="ChartListItem__icon">
            <img className="ChartListItem__icon" src={this.chartIconSrc()} />
          </div>
          <div className="ChartListName__details">
            <h6>{chart.id}</h6>
          </div>
        </Link>
      </div>
    );
  }

  chartIconSrc() {
    const icon = this.props.chart.attributes.icon;
    if (icon.length > 0) {
      return `/api/chartsvc/${icon}`;
    } else {
      return placeholder;
    }
  }
}

export default ChartListItem;
