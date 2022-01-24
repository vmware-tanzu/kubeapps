// Copyright 2018-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

import { connect } from "react-redux";
import { Action } from "redux";
import { ThunkDispatch } from "redux-thunk";
import { IStoreState } from "shared/types";
import actions from "../../actions";
import LoginForm from "../../components/LoginForm";

function mapStateToProps({
  auth: { authenticated, authenticating, authenticationError },
  config: { authProxyEnabled, oauthLoginURI, appVersion, authProxySkipLoginPage },
  clusters,
}: IStoreState) {
  return {
    authenticated,
    authenticating,
    authenticationError,
    cluster: clusters.currentCluster,
    oauthLoginURI: authProxyEnabled ? oauthLoginURI : "",
    authProxySkipLoginPage,
    appVersion,
  };
}

function mapDispatchToProps(dispatch: ThunkDispatch<IStoreState, null, Action>) {
  return {
    authenticate: (cluster: string, token: string) =>
      dispatch(actions.auth.authenticate(cluster, token, false)),
    checkCookieAuthentication: (cluster: string) =>
      dispatch(actions.auth.checkCookieAuthentication(cluster)),
  };
}

export default connect(mapStateToProps, mapDispatchToProps)(LoginForm);
