import qs from "qs";
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
    filter: qs.parse(location.search, { ignoreQueryPrefix: true }),
    repos: params.repo,
    csvs: operators.csvs,
    cluster: params.cluster,
    namespace: params.namespace,
    kubeappsNamespace: config.kubeappsNamespace,
  };
}

function mapDispatchToProps(dispatch: ThunkDispatch<IStoreState, null, Action>) {
  return {
    fetchCharts: (
      cluster: string,
      namespace: string,
      repos: string,
      page: number,
      size: number,
      query?: string,
    ) => dispatch(actions.charts.fetchCharts(cluster, namespace, repos, page, size, query)),
    fetchChartCategories: (cluster: string, namespace: string) =>
      dispatch(actions.charts.fetchChartCategories(cluster, namespace)),
    fetchRepos: (namespace: string, listGlobal?: boolean) =>
      dispatch(actions.repos.fetchRepos(namespace, listGlobal)),
    resetRequestCharts: () => dispatch(actions.charts.resetRequestCharts()),
    pushSearchFilter: (filter: string) => dispatch(actions.shared.pushSearchFilter(filter)),
    getCSVs: (cluster: string, namespace: string) =>
      dispatch(actions.operators.getCSVs(cluster, namespace)),
  };
}

export default connect(mapStateToProps, mapDispatchToProps)(Catalog);
