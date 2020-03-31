import { connect } from "react-redux";
import { Action } from "redux";
import { ThunkDispatch } from "redux-thunk";

import actions from "../../actions";
import ChartView from "../../components/ChartView";
import { IChartVersion, IStoreState } from "../../shared/types";

interface IRouteProps {
  match: {
    params: {
      namespace: string;
      repo: string;
      id: string;
      version?: string;
    };
  };
}

function mapStateToProps({ charts, namespace }: IStoreState, { match: { params } }: IRouteProps) {
  return {
    chartID: chartID(params),
    chartNamespace: params.namespace,
    isFetching: charts.isFetching,
    namespace: namespace.current,
    selected: charts.selected,
    version: params.version,
  };
}

function mapDispatchToProps(
  dispatch: ThunkDispatch<IStoreState, null, Action>,
  { match: { params } }: IRouteProps,
) {
  return {
    fetchChartVersionsAndSelectVersion: (namespace: string, id: string, version?: string) =>
      dispatch(actions.charts.fetchChartVersionsAndSelectVersion(namespace, id, version)),
    getChartReadme: (namespace: string, version: string) =>
      dispatch(actions.charts.getChartReadme(namespace, chartID(params), version)),
    resetChartVersion: () => dispatch(actions.charts.resetChartVersion()),
    selectChartVersion: (version: IChartVersion) =>
      dispatch(actions.charts.selectChartVersion(version)),
  };
}

function chartID(params: IRouteProps["match"]["params"]) {
  return `${params.repo}/${params.id}`;
}

export default connect(mapStateToProps, mapDispatchToProps)(ChartView);
