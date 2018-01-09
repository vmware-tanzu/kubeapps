import { connect } from 'react-redux';
import { bindActionCreators, Dispatch } from 'redux';

import * as actions from '../actions';
import ChartList from '../components/ChartList';
import { StoreState } from '../store/types';

interface RouteProps {
  match: {
    params: {
      repo: string;
    }
  };
}

function mapStateToProps({ charts }: StoreState, { match: { params } }: RouteProps) {
  return {
    charts,
    repo: params.repo
  };
}

function mapDispatchToProps(dispatch: Dispatch<StoreState>) {
  return bindActionCreators(
    {
      fetchCharts: actions.fetchCharts
    },
    dispatch);
}

export default connect(mapStateToProps, mapDispatchToProps)(ChartList);
