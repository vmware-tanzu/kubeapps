import { connect } from "react-redux";
import { push } from "react-router-redux";
import { Dispatch } from "redux";

import actions from "../../actions";
import DeploymentForm from "../../components/DeploymentForm";
import { IChartVersion, IStoreState } from "../../shared/types";

interface IRouteProps {
  match: {
    params: {
      repo: string;
      id: string;
      version: string;
    };
  };
}

function mapStateToProps(
  { apps, catalog, charts, namespace }: IStoreState,
  { match: { params } }: IRouteProps,
) {
  return {
    bindings: catalog.bindings,
    chartID: `${params.repo}/${params.id}`,
    chartVersion: params.version,
    error: apps.error,
    namespace: namespace.current,
    selected: charts.selected,
  };
}

function mapDispatchToProps(dispatch: Dispatch<IStoreState>) {
  return {
    deployChart: (
      version: IChartVersion,
      releaseName: string,
      namespace: string,
      values?: string,
      resourceVersion?: string,
    ) =>
      dispatch(actions.apps.deployChart(version, releaseName, namespace, values, resourceVersion)),
    fetchChartVersions: (id: string) => dispatch(actions.charts.fetchChartVersions(id)),
    getBindings: (ns: string) => dispatch(actions.catalog.getBindings(ns)),
    getChartValues: (id: string, version: string) =>
      dispatch(actions.charts.getChartValues(id, version)),
    getChartVersion: (id: string, version: string) =>
      dispatch(actions.charts.getChartVersion(id, version)),
    push: (location: string) => dispatch(push(location)),
  };
}

export default connect(mapStateToProps, mapDispatchToProps)(DeploymentForm);
