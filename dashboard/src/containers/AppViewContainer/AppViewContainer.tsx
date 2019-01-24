import { connect } from "react-redux";
import { Action } from "redux";
import { ThunkDispatch } from "redux-thunk";

import actions from "../../actions";
import AppView from "../../components/AppView";
import { IChartUpdate, IResource, IStoreState } from "../../shared/types";

interface IRouteProps {
  match: {
    params: {
      namespace: string;
      releaseName: string;
    };
  };
}

function mapStateToProps({ apps, kube, charts }: IStoreState, { match: { params } }: IRouteProps) {
  let update: IChartUpdate | undefined;
  if (
    apps.selected &&
    apps.selected.chart &&
    apps.selected.chart.metadata &&
    apps.selected.chart.metadata.name &&
    charts.updates[apps.selected.chart.metadata.name]
  ) {
    update = charts.updates[apps.selected.chart.metadata.name];
  }
  return {
    app: apps.selected,
    deleteError: apps.deleteError,
    resources: kube.items,
    error: apps.error,
    namespace: params.namespace,
    releaseName: params.releaseName,
    update,
  };
}

function mapDispatchToProps(dispatch: ThunkDispatch<IStoreState, null, Action>) {
  return {
    deleteApp: (releaseName: string, ns: string, purge: boolean) =>
      dispatch(actions.apps.deleteApp(releaseName, ns, purge)),
    getApp: (releaseName: string, ns: string) => dispatch(actions.apps.getApp(releaseName, ns)),
    getChartUpdates: (name: string, version: string, appVersion: string) =>
      dispatch(actions.charts.getChartUpdates(name, version, appVersion)),
    // TODO: remove once WebSockets are moved to Redux store (#882)
    receiveResource: (payload: { key: string; resource: IResource }) =>
      dispatch(actions.kube.receiveResource(payload)),
  };
}

export default connect(
  mapStateToProps,
  mapDispatchToProps,
)(AppView);
