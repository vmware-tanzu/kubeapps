import { connect } from "react-redux";
import { Dispatch } from "redux";

import actions from "../../actions";
import AppList from "../../components/AppList";
import { IStoreState } from "../../shared/types";

function mapStateToProps({ apps }: IStoreState) {
  return {
    apps,
  };
}

function mapDispatchToProps(dispatch: Dispatch<IStoreState>) {
  return {
    fetchApps: () => dispatch(actions.apps.fetchApps()),
  };
}

export default connect(mapStateToProps, mapDispatchToProps)(AppList);
