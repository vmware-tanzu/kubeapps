import * as React from 'react';
import { Link } from 'react-router-dom';

import { Chart } from '../store/types';
import ChartIcon from './ChartIcon';
import './ChartListItem.css';

interface Props {
  chart: Chart;
}

class ChartListItem extends React.Component<Props> {
  render() {
    const { chart } = this.props;
    const latestAppVersion = chart.relationships.latestChartVersion.data.app_version;
    return (
      <div className="ChartListItem padding-normal margin-big elevation-5">
        <Link to={`/charts/` + chart.id}>
          <ChartIcon icon={chart.attributes.icon} />
          <div className="ChartListName__details">
            <h6>{chart.id}</h6>
            {latestAppVersion &&
              <span>v{latestAppVersion}</span>
            }
          </div>
        </Link>
      </div>
    );
  }
}

export default ChartListItem;
