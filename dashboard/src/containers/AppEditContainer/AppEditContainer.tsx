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
      tillerReleaseName: string;
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
    selected: charts.selected,
    tillerReleaseName: params.tillerReleaseName,
  };
}

function mapDispatchToProps(dispatch: Dispatch<IStoreState>) {
  return {
    deployChart: (
      helmCRDReleaseName: string,
      version: IChartVersion,
      tillerReleaseName: string,
      namespace: string,
      values?: string,
      resourceVersion?: string,
    ) =>
      dispatch(
        actions.apps.deployChart(
          helmCRDReleaseName,
          version,
          tillerReleaseName,
          namespace,
          values,
          resourceVersion,
        ),
      ),
    fetchChartVersions: (id: string) => dispatch(actions.charts.fetchChartVersions(id)),
    getApp: (helmCRDReleaseName: string, tillerReleaseName: string, ns: string) =>
      dispatch(actions.apps.getApp(helmCRDReleaseName, tillerReleaseName, ns)),
    getBindings: (ns: string) => dispatch(actions.catalog.getBindings(ns)),
    getChartValues: (id: string, version: string) =>
      dispatch(actions.charts.getChartValues(id, version)),
    getChartVersion: (id: string, version: string) =>
      dispatch(actions.charts.getChartVersion(id, version)),
    push: (location: string) => dispatch(push(location)),
  };
}

export default connect(mapStateToProps, mapDispatchToProps)(AppEdit);
