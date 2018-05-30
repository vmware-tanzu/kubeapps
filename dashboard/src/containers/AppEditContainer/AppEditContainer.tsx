import { connect } from "react-redux";
import { push } from "react-router-redux";
import { Dispatch } from "redux";

import actions from "../../actions";
import AppEdit from "../../components/AppEdit";
import { IChartVersion, IStoreState } from "../../shared/types";

interface IRouteProps {
  match: {
    params: {
      namespace: string;
      releaseName: string;
    };
  };
}

function mapStateToProps(
  { apps, catalog, charts }: IStoreState,
  { match: { params } }: IRouteProps,
) {
  return {
    app: apps.selected,
    bindings: catalog.bindings,
    error: apps.error,
    namespace: params.namespace,
    releaseName: params.releaseName,
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
    getApp: (releaseName: string, ns: string) => dispatch(actions.apps.getApp(releaseName, ns)),
    getBindings: (ns: string) => dispatch(actions.catalog.getBindings(ns)),
    getChartValues: (id: string, version: string) =>
      dispatch(actions.charts.getChartValues(id, version)),
    getChartVersion: (id: string, version: string) =>
      dispatch(actions.charts.getChartVersion(id, version)),
    push: (location: string) => dispatch(push(location)),
  };
}

export default connect(mapStateToProps, mapDispatchToProps)(AppEdit);
