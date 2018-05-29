import { connect } from "react-redux";
import { Dispatch } from "redux";

import actions from "../../actions";
import AppView from "../../components/AppView";
import { IStoreState } from "../../shared/types";

interface IRouteProps {
  match: {
    params: {
      namespace: string;
      tillerReleaseName: string;
    };
  };
}

function mapStateToProps({ apps }: IStoreState, { match: { params } }: IRouteProps) {
  return {
    app: apps.selected,
    deleteError: apps.deleteError,
    error: apps.error,
    namespace: params.namespace,
    tillerReleaseName: params.tillerReleaseName,
  };
}

function mapDispatchToProps(dispatch: Dispatch<IStoreState>) {
  return {
    deleteApp: (tillerReleaseName: string, ns: string) =>
      dispatch(actions.apps.deleteApp(tillerReleaseName, ns)),
    getApp: (helmCRDReleaseName: string, tillerReleaseName: string, ns: string) =>
      dispatch(actions.apps.getApp(helmCRDReleaseName, tillerReleaseName, ns)),
  };
}

export default connect(mapStateToProps, mapDispatchToProps)(AppView);
