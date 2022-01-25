// Copyright 2018-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { connect } from "react-redux";
import { ThunkDispatch } from "redux-thunk";
import { IStoreState } from "shared/types";
import actions from "../../actions";
import { ConfigAction } from "../../actions/config";
import ConfigLoader from "../../components/ConfigLoader";

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

export default connect(mapStateToProps, mapDispatchToProps)(ConfigLoader);
