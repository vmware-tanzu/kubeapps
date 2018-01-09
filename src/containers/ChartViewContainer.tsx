import { connect } from 'react-redux';
import { bindActionCreators, Dispatch } from 'redux';

import * as actions from '../actions';
import ChartView from '../components/ChartView';
import { StoreState } from '../store/types';

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
  return bindActionCreators(
    {
      getChart: actions.getChart
    },
    dispatch);
}

export default connect(mapStateToProps, mapDispatchToProps)(ChartView);
