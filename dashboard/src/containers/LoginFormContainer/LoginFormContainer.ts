import { connect } from "react-redux";
import { Action } from "redux";
import { ThunkDispatch } from "redux-thunk";

import actions from "../../actions";
import LoginForm from "../../components/LoginForm";
import { IStoreState } from "../../shared/types";

function mapStateToProps({
  auth: { authenticated, authenticating, authenticationError },
  config: { authProxyEnabled, oauthLoginURI, featureFlags, appVersion },
  clusters,
}: IStoreState) {
  return {
    authenticated,
    authenticating,
    authenticationError,
    cluster: clusters.currentCluster,
    oauthLoginURI: authProxyEnabled ? oauthLoginURI : "",
    UI: featureFlags.ui,
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
