import { connect } from "react-redux";
import { push } from "react-router-redux";
import { Dispatch } from "redux";

import actions from "../../actions";
import FunctionList from "../../components/FunctionList";
import { IFunction, IStoreState } from "../../shared/types";

function mapStateToProps({ functions: { items } }: IStoreState) {
  return {
    functions: items,
  };
}

function mapDispatchToProps(dispatch: Dispatch<IStoreState>) {
  return {
    deployFunction: (name: string, namespace: string, spec: IFunction["spec"]) =>
      dispatch(actions.functions.createFunction(name, namespace, spec)),
    fetchFunctions: () => dispatch(actions.functions.fetchFunctions()),
    navigateToFunction: (name: string, namespace: string) =>
      dispatch(push(`/functions/ns/${namespace}/${name}`)),
  };
}

export default connect(mapStateToProps, mapDispatchToProps)(FunctionList);
