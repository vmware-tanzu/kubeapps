import { connect } from "react-redux";
import { Dispatch } from "redux";

import actions from "../../actions";
import FunctionList from "../../components/FunctionList";
import { IStoreState } from "../../shared/types";

interface IRouteProps {
  match: {
    params: {
      namespace: string;
    };
  };
}

function mapStateToProps(
  { functions: { items } }: IStoreState,
  { match: { params } }: IRouteProps,
) {
  return {
    functions: items,
    namespace: params.namespace,
  };
}

function mapDispatchToProps(dispatch: Dispatch<IStoreState>) {
  return {
    fetchFunctions: (namespace: string) => dispatch(actions.functions.fetchFunctions(namespace)),
  };
}

export default connect(mapStateToProps, mapDispatchToProps)(FunctionList);
