import { connect } from "react-redux";
import { RouteComponentProps } from "react-router";
import { Action } from "redux";
import { ThunkDispatch } from "redux-thunk";

import actions from "../../actions";
import OperatorList from "../../components/OperatorList";
import { IStoreState } from "../../shared/types";

function mapStateToProps({ operators }: IStoreState, { location }: RouteComponentProps<{}>) {
  return {
    isFetching: operators.isFetching,
    isOLMInstalled: operators.isOLMInstalled,
  };
}

function mapDispatchToProps(dispatch: ThunkDispatch<IStoreState, null, Action>) {
  return {
    checkOLMInstalled: () => dispatch(actions.operators.checkOLMInstalled()),
  };
}

export default connect(mapStateToProps, mapDispatchToProps)(OperatorList);
