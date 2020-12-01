import { connect } from "react-redux";
import { Action } from "redux";
import { ThunkDispatch } from "redux-thunk";

import actions from "../../actions";
import ChartView from "../../components/ChartView";
import { IChartVersion, IStoreState } from "../../shared/types";

interface IRouteProps {
  match: {
    params: {
      cluster: string;
      namespace: string;
      repo: string;
      global: string;
      id: string;
      version?: string;
    };
  };
}

function mapStateToProps({ charts, config }: IStoreState, { match: { params } }: IRouteProps) {
  return {
    chartID: chartID(params),
    chartNamespace: params.global === "global" ? config.kubeappsNamespace : params.namespace,
    isFetching: charts.isFetching,
    cluster: params.cluster,
    namespace: params.namespace,
    selected: charts.selected,
    version: params.version,
    kubeappsNamespace: config.kubeappsNamespace,
  };
}

function mapDispatchToProps(
  dispatch: ThunkDispatch<IStoreState, null, Action>,
  { match: { params } }: IRouteProps,
) {
  return {
    fetchChartVersionsAndSelectVersion: (
      cluster: string,
      namespace: string,
      id: string,
      version?: string,
    ) =>
      dispatch(actions.charts.fetchChartVersionsAndSelectVersion(cluster, namespace, id, version)),
    getChartReadme: (cluster: string, namespace: string, version: string) =>
      dispatch(actions.charts.getChartReadme(cluster, namespace, chartID(params), version)),
    resetChartVersion: () => dispatch(actions.charts.resetChartVersion()),
    selectChartVersion: (version: IChartVersion) =>
      dispatch(actions.charts.selectChartVersion(version)),
  };
}

function chartID(params: IRouteProps["match"]["params"]) {
  return `${params.repo}/${params.id}`;
}

export default connect(mapStateToProps, mapDispatchToProps)(ChartView);
