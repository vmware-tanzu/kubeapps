import { connect } from "react-redux";
import { Action } from "redux";
import { ThunkDispatch } from "redux-thunk";

import actions from "../../actions";

import ServiceClassList from "../../components/ServiceClassList";
import { IStoreState } from "../../shared/types";

interface IRouteProps {
  match: {
    params: {
      brokerName: string;
      className: string;
    };
  };
}

function mapStateToProps({ catalog }: IStoreState, props: IRouteProps) {
  const { classes, errors } = catalog;

  return {
    classes,
    error: errors.fetch,
  };
}

function mapDispatchToProps(dispatch: ThunkDispatch<IStoreState, null, Action>) {
  return {
    getClasses: async () => dispatch(actions.catalog.getClasses()),
  };
}

export default connect(
  mapStateToProps,
  mapDispatchToProps,
)(ServiceClassList);
