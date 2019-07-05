import { connect } from "react-redux";
import { ThunkDispatch } from "redux-thunk";

import actions from "../../actions";
import { ConfigAction } from "../../actions/config";

import ConfigLoader from "../../components/ConfigLoader";
import { IStoreState } from "../../shared/types";

function mapStateToProps({ config }: IStoreState) {
  return {
    loaded: config.loaded,
    error: config.error,
  };
}
function mapDispatchToProps(dispatch: ThunkDispatch<IStoreState, null, ConfigAction>) {
  return {
    getConfig: () => dispatch(actions.config.getConfig()),
  };
}

export default connect(
  mapStateToProps,
  mapDispatchToProps,
)(ConfigLoader);
