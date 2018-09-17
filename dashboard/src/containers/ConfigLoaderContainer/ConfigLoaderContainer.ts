import { connect } from "react-redux";
import { Dispatch } from "redux";

import actions from "../../actions";
import LoadingWrapper, { ILoadingWrapperProps } from "../../components/LoadingWrapper";
import { IStoreState } from "../../shared/types";

function mapStateToProps({ config }: IStoreState): ILoadingWrapperProps {
  return {
    loaded: config.loaded,
  };
}

function mapDispatchToProps(dispatch: Dispatch<IStoreState>) {
  dispatch(actions.config.getConfig());
}

export default connect(mapStateToProps, mapDispatchToProps)(LoadingWrapper);
