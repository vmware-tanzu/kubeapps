import { push } from "connected-react-router";
import { connect } from "react-redux";
import { Action } from "redux";
import { ThunkDispatch } from "redux-thunk";

import { JSONSchema4 } from "json-schema";
import actions from "../../actions";
import DeploymentForm from "../../components/DeploymentForm";
import { IChartVersion, IStoreState } from "../../shared/types";

interface IRouteProps {
  match: {
    params: {
      cluster: string;
      namespace: string;
      repo: string;
      global: string;
      id: string;
      version: string;
    };
  };
}

function mapStateToProps(
  { apps, charts, config }: IStoreState,
  { match: { params } }: IRouteProps,
) {
  return {
    chartID: `${params.repo}/${params.id}`,
    chartNamespace: params.global === "global" ? config.namespace : params.namespace,
    cluster: params.cluster,
    chartVersion: params.version,
    error: apps.error,
    kubeappsNamespace: config.namespace,
    namespace: params.namespace,
    selected: charts.selected,
    chartsIsFetching: charts.isFetching,
  };
}

function mapDispatchToProps(dispatch: ThunkDispatch<IStoreState, null, Action>) {
  return {
    deployChart: (
      version: IChartVersion,
      chartNamespace: string,
      namespace: string,
      releaseName: string,
      values?: string,
      schema?: JSONSchema4,
    ) =>
      dispatch(
        actions.apps.deployChart(version, chartNamespace, namespace, releaseName, values, schema),
      ),
    fetchChartVersions: (namespace: string, id: string) =>
      dispatch(actions.charts.fetchChartVersions(namespace, id)),
    getChartVersion: (namespace: string, id: string, version: string) =>
      dispatch(actions.charts.getChartVersion(namespace, id, version)),
    push: (location: string) => dispatch(push(location)),
    getNamespace: (ns: string) => dispatch(actions.namespace.getNamespace(ns)),
  };
}

export default connect(mapStateToProps, mapDispatchToProps)(DeploymentForm);
