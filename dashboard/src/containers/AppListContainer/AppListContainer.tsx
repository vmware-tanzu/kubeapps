import { connect } from "react-redux";
import { Dispatch } from "redux";

import actions from "../../actions";
import AppList from "../../components/AppList";
import { IStoreState } from "../../shared/types";

function mapStateToProps({ apps, namespace }: IStoreState) {
  return {
    apps,
    namespace: namespace.current,
  };
}

function mapDispatchToProps(dispatch: Dispatch<IStoreState>) {
  return {
    fetchApps: (ns: string) => dispatch(actions.apps.fetchApps(ns)),
  };
}

export default connect(mapStateToProps, mapDispatchToProps)(AppList);
