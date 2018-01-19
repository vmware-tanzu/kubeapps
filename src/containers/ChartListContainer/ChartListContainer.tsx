import { connect } from "react-redux";
import { Dispatch } from "redux";

import actions from "../../actions";
import ChartList from "../../components/ChartList";
import { IStoreState } from "../../shared/types";

interface IRouteProps {
  match: {
    params: {
      repo: string;
    };
  };
}

function mapStateToProps({ charts }: IStoreState, { match: { params } }: IRouteProps) {
  return {
    charts,
    repo: params.repo,
  };
}

function mapDispatchToProps(dispatch: Dispatch<IStoreState>) {
  return {
    fetchCharts: (repo: string) => dispatch(actions.charts.fetchCharts(repo)),
  };
}

export default connect(mapStateToProps, mapDispatchToProps)(ChartList);
