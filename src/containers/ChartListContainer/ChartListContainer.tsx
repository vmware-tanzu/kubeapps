import { connect } from 'react-redux';
import { Dispatch } from 'redux';

import actions from '../../actions';
import ChartList from '../../components/ChartList';
import { StoreState } from '../../shared/types';

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
  return {
    fetchCharts: (repo: string) => dispatch(actions.charts.fetchCharts(repo))
  };
}

export default connect(mapStateToProps, mapDispatchToProps)(ChartList);
