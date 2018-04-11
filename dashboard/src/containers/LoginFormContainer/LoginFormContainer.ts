import { connect } from "react-redux";
import { Dispatch } from "redux";

import actions from "../../actions";
import LoginForm from "../../components/LoginForm";
import { IStoreState } from "../../shared/types";

interface IRouteProps {
  match: {
    params: {
      name: string;
      namespace: string;
    };
  };
}

function mapStateToProps(
  { auth: { authenticated } }: IStoreState,
  { match: { params } }: IRouteProps,
) {
  return {
    authenticated,
  };
}

function mapDispatchToProps(
  dispatch: Dispatch<IStoreState>,
  { match: { params: { name, namespace } } }: IRouteProps,
) {
  return {
    authenticate: (token: string) => dispatch(actions.auth.authenticate(token)),
  };
}

export default connect(mapStateToProps, mapDispatchToProps)(LoginForm);
