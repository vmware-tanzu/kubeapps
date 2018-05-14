import * as qs from "qs";
import { connect } from "react-redux";
import { RouteComponentProps } from "react-router";
import { push } from "react-router-redux";
import { Dispatch } from "redux";

import actions from "../../actions";
import ChartList from "../../components/ChartList";
import { IStoreState } from "../../shared/types";

function mapStateToProps(
  { charts }: IStoreState,
  { match: { params }, location }: RouteComponentProps<{ repo: string }>,
) {
  return {
    charts,
    filter: qs.parse(location.search, { ignoreQueryPrefix: true }).q || "",
    repo: params.repo,
  };
}

function mapDispatchToProps(dispatch: Dispatch<IStoreState>) {
  return {
    fetchCharts: (repo: string) => dispatch(actions.charts.fetchCharts(repo)),
    pushSearchFilter: (filter: string) => dispatch(push(`?q=${filter}`)),
  };
}

export default connect(mapStateToProps, mapDispatchToProps)(ChartList);
