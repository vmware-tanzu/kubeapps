import { connect } from "react-redux";
import { Dispatch } from "redux";

import actions from "../../actions";
import AppView from "../../components/AppView";
import { IStoreState } from "../../shared/types";

interface IRouteProps {
  match: {
    params: {
      namespace: string;
      releaseName: string;
    };
  };
}

function mapStateToProps({ apps }: IStoreState, { match: { params } }: IRouteProps) {
  return {
    app: apps.selected,
    namespace: params.namespace,
    releaseName: params.releaseName,
  };
}

function mapDispatchToProps(dispatch: Dispatch<IStoreState>) {
  return {
    getApp: (r: string) => dispatch(actions.apps.getApp(r)),
  };
}

export default connect(mapStateToProps, mapDispatchToProps)(AppView);
