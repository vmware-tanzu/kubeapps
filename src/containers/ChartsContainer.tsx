import { connect } from 'react-redux';
import { bindActionCreators, Dispatch } from 'redux';

import * as actions from '../actions';
import ChartList from '../components/ChartList';
import { StoreState } from '../store/types';

function mapStateToProps({ charts }: StoreState) {
  return {
    charts
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
