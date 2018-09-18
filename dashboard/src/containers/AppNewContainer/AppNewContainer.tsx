import { connect } from "react-redux";
import { push } from "react-router-redux";
import { ThunkDispatch } from "redux-thunk";

import actions from "../../actions";
import { ServiceCatalogAction } from "../../actions/catalog";
import { ChartsAction } from "../../actions/charts";
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
  { apps, catalog, charts, config, namespace }: IStoreState,
  { match: { params } }: IRouteProps,
) {
  return {
    bindingsWithSecrets: catalog.bindingsWithSecrets,
    chartID: `${params.repo}/${params.id}`,
    chartVersion: params.version,
    error: apps.error,
    kubeappsNamespace: config.namespace,
    namespace: namespace.current,
    selected: charts.selected,
  };
}

function mapDispatchToProps(
  dispatch: ThunkDispatch<IStoreState, null, ChartsAction & ServiceCatalogAction>,
) {
  return {
    deployChart: (
      version: IChartVersion,
      releaseName: string,
      namespace: string,
      values?: string,
    ) => dispatch(actions.apps.deployChart(version, releaseName, namespace, values)),
    fetchChartVersions: (id: string) => dispatch(actions.charts.fetchChartVersions(id)),
    getBindings: (ns: string) => dispatch(actions.catalog.getBindings(ns)),
    getChartValues: (id: string, version: string) =>
      dispatch(actions.charts.getChartValues(id, version)),
    getChartVersion: (id: string, version: string) =>
      dispatch(actions.charts.getChartVersion(id, version)),
    push: (location: string) => dispatch(push(location) as any),
  };
}

export default connect(mapStateToProps, mapDispatchToProps)(DeploymentForm);
