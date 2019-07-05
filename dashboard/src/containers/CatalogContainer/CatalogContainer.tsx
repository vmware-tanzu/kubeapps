import * as qs from "qs";
import { connect } from "react-redux";
import { RouteComponentProps } from "react-router";
import { Action } from "redux";
import { ThunkDispatch } from "redux-thunk";
import actions from "../../actions";
import Catalog from "../../components/Catalog";
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

function mapDispatchToProps(dispatch: ThunkDispatch<IStoreState, null, Action>) {
  return {
    fetchCharts: (repo: string) => dispatch(actions.charts.fetchCharts(repo)),
    pushSearchFilter: (filter: string) => dispatch(actions.shared.pushSearchFilter(filter)),
  };
}

export default connect(
  mapStateToProps,
  mapDispatchToProps,
)(Catalog);
