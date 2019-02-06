import { push } from "connected-react-router";
import { connect } from "react-redux";
import { Action } from "redux";
import { ThunkDispatch } from "redux-thunk";

import actions from "../../actions";
import AppView from "../../components/AppView";
import { IResource, IStoreState } from "../../shared/types";

interface IRouteProps {
  match: {
    params: {
      namespace: string;
      releaseName: string;
    };
  };
}

function mapStateToProps({ apps, kube, charts }: IStoreState, { match: { params } }: IRouteProps) {
  return {
    app: apps.selected,
    deleteError: apps.deleteError,
    resources: kube.items,
    error: apps.error,
    namespace: params.namespace,
    releaseName: params.releaseName,
  };
}

function mapDispatchToProps(dispatch: ThunkDispatch<IStoreState, null, Action>) {
  return {
    deleteApp: (releaseName: string, ns: string, purge: boolean) =>
      dispatch(actions.apps.deleteApp(releaseName, ns, purge)),
    getAppWithUpdateInfo: (releaseName: string, ns: string) =>
      dispatch(actions.apps.getAppWithUpdateInfo(releaseName, ns)),
    // TODO: remove once WebSockets are moved to Redux store (#882)
    receiveResource: (payload: { key: string; resource: IResource }) =>
      dispatch(actions.kube.receiveResource(payload)),
    push: (location: string) => dispatch(push(location)),
  };
}

export default connect(
  mapStateToProps,
  mapDispatchToProps,
)(AppView);
