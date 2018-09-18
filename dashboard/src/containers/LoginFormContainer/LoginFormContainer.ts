import { connect } from "react-redux";
import { ThunkDispatch } from "redux-thunk";

import actions from "../../actions";
import { AuthAction } from "../../actions/auth";
import LoginForm from "../../components/LoginForm";
import { IStoreState } from "../../shared/types";

function mapStateToProps({
  auth: { authenticated, authenticating, authenticationError },
}: IStoreState) {
  return {
    authenticated,
    authenticating,
    authenticationError,
  };
}

function mapDispatchToProps(dispatch: ThunkDispatch<IStoreState, null, AuthAction>) {
  return {
    authenticate: (token: string) => dispatch(actions.auth.authenticate(token)),
  };
}

export default connect(mapStateToProps, mapDispatchToProps)(LoginForm);
