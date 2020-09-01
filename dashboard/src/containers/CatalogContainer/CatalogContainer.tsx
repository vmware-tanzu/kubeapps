import * as qs from "qs";
import { connect } from "react-redux";
import { RouteComponentProps } from "react-router";
import { Action } from "redux";
import { ThunkDispatch } from "redux-thunk";
import actions from "../../actions";
import Catalog from "../../components/Catalog";
import { IStoreState } from "../../shared/types";

function mapStateToProps(
  { charts, operators, config }: IStoreState,
  {
    match: { params },
    location,
  }: RouteComponentProps<{ cluster: string; namespace: string; repo: string }>,
) {
  return {
    charts,
    filter: qs.parse(location.search, { ignoreQueryPrefix: true }).q || "",
    repo: params.repo,
    csvs: operators.csvs,
    cluster: params.cluster,
    namespace: params.namespace,
    kubeappsNamespace: config.kubeappsNamespace,
    featureFlags: config.featureFlags,
    UI: config.featureFlags.ui,
  };
}

function mapDispatchToProps(dispatch: ThunkDispatch<IStoreState, null, Action>) {
  return {
    fetchCharts: (namespace: string, repo: string) =>
      dispatch(actions.charts.fetchCharts(namespace, repo)),
    pushSearchFilter: (filter: string) => dispatch(actions.shared.pushSearchFilter(filter)),
    getCSVs: (cluster: string, namespace: string) =>
      dispatch(actions.operators.getCSVs(cluster, namespace)),
  };
}

export default connect(mapStateToProps, mapDispatchToProps)(Catalog);
