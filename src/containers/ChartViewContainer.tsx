import { connect } from 'react-redux';
import { Dispatch } from 'redux';

import actions from '../actions';
import ChartView from '../components/ChartView';
import { Chart, StoreState } from '../store/types';
import { push } from 'react-router-redux';

interface RouteProps {
  match: {
    params: {
      repo: string;
      id: string;
    }
  };
}

function mapStateToProps({ charts }: StoreState, { match: { params } }: RouteProps) {
  return {
    chart: charts.selectedChart,
    version: charts.selectedVersion,
    isFetching: charts.isFetching,
    chartID: `${params.repo}/${params.id}`,
  };
}

function mapDispatchToProps(dispatch: Dispatch<StoreState>) {
  return {
    getChart: (id: string) => dispatch(actions.charts.getChart(id)),
    deployChart: (chart: Chart, releaseName: string, namespace: string) =>
      dispatch(actions.charts.deployChart(chart, releaseName, namespace)),
    push: (location: string) => dispatch(push(location)),
  };
}

export default connect(mapStateToProps, mapDispatchToProps)(ChartView);
