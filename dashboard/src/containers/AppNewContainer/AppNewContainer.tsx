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
    UI: config.featureFlags.ui,
  };
}

function mapDispatchToProps(dispatch: ThunkDispatch<IStoreState, null, Action>) {
  return {
    deployChart: (
      targetCluster: string,
      targetNamespace: string,
      version: IChartVersion,
      chartNamespace: string,
      releaseName: string,
      values?: string,
      schema?: JSONSchema4,
    ) =>
      dispatch(
        actions.apps.deployChart(
          targetCluster,
          targetNamespace,
          version,
          chartNamespace,
          releaseName,
          values,
          schema,
        ),
      ),
    fetchChartVersions: (namespace: string, id: string) =>
      dispatch(actions.charts.fetchChartVersions(namespace, id)),
    getChartVersion: (namespace: string, id: string, version: string) =>
      dispatch(actions.charts.getChartVersion(namespace, id, version)),
    push: (location: string) => dispatch(push(location)),
  };
}

export default connect(mapStateToProps, mapDispatchToProps)(DeploymentForm);
