import { connect } from "react-redux";
import { Action } from "redux";
import { ThunkDispatch } from "redux-thunk";

import actions from "../../actions";
import LoginForm from "../../components/LoginForm";
import { IStoreState } from "../../shared/types";

function mapStateToProps({
  auth: { authenticated, authenticating, authenticationError, checkingOIDCToken },
}: IStoreState) {
  return {
    authenticated,
    authenticating,
    checkingOIDCToken,
    authenticationError,
  };
}

function mapDispatchToProps(dispatch: ThunkDispatch<IStoreState, null, Action>) {
  return {
    authenticate: (token: string) => dispatch(actions.auth.authenticate(token)),
    tryToAutoAuthenticate: () => dispatch(actions.auth.tryToAutoAuthenticate()),
  };
}

export default connect(
  mapStateToProps,
  mapDispatchToProps,
)(LoginForm);
