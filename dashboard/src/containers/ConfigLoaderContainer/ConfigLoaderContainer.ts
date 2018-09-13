import { connect } from "react-redux";
import { Dispatch } from "redux";

import actions from "../../actions";
import LoadingWrapper from "../../components/LoadingWrapper";
import { IStoreState } from "../../shared/types";

interface IMapProps {
  loaded?: boolean;
}

function mapStateToProps({ config }: IStoreState): IMapProps {
  return {
    loaded: config.loaded,
  };
}

function mapDispatchToProps(dispatch: Dispatch<IStoreState>) {
  dispatch(actions.config.getConfig());
}

export default connect(mapStateToProps, mapDispatchToProps)(LoadingWrapper);
