import { connect } from "react-redux";
import { RouteComponentProps } from "react-router";
import { Action } from "redux";
import { ThunkDispatch } from "redux-thunk";

import actions from "../../actions";
import OperatorList from "../../components/OperatorList";
import { IStoreState } from "../../shared/types";

function mapStateToProps(
  { operators, namespace }: IStoreState,
  { location }: RouteComponentProps<{}>,
) {
  return {
    namespace: namespace.current,
    isFetching: operators.isFetching,
    isOLMInstalled: operators.isOLMInstalled,
    operators: operators.operators,
    error: operators.errors.fetch,
  };
}

function mapDispatchToProps(dispatch: ThunkDispatch<IStoreState, null, Action>) {
  return {
    checkOLMInstalled: () => dispatch(actions.operators.checkOLMInstalled()),
    getOperators: (namespace: string) => dispatch(actions.operators.getOperators(namespace)),
  };
}

export default connect(mapStateToProps, mapDispatchToProps)(OperatorList);
