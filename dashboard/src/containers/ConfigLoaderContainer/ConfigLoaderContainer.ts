import { connect } from "react-redux";
import { ThunkDispatch } from "redux-thunk";

import actions from "../../actions";
import { ConfigAction } from "../../actions/config";

import LoadingWrapper, { ILoadingWrapperProps } from "../../components/LoadingWrapper";
import { IStoreState } from "../../shared/types";

function mapStateToProps({ config }: IStoreState): ILoadingWrapperProps {
  return {
    loaded: config.loaded,
  };
}
function mapDispatchToProps(dispatch: ThunkDispatch<IStoreState, null, ConfigAction>) {
  dispatch(actions.config.getConfig());
}

export default connect(
  mapStateToProps,
  mapDispatchToProps,
)(LoadingWrapper);
