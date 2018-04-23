import { connect } from "react-redux";
import { push } from "react-router-redux";
import { Dispatch } from "redux";

import actions from "../../actions";
import FunctionList from "../../components/FunctionList";
import { IFunction, IStoreState } from "../../shared/types";

function mapStateToProps({ functions, namespace }: IStoreState) {
  return {
    functions: functions.items,
    namespace: namespace.current,
    runtimes: functions.runtimes,
  };
}

function mapDispatchToProps(dispatch: Dispatch<IStoreState>) {
  return {
    deployFunction: (name: string, namespace: string, spec: IFunction["spec"]) =>
      dispatch(actions.functions.createFunction(name, namespace, spec)),
    fetchFunctions: (ns: string) => dispatch(actions.functions.fetchFunctions(ns)),
    fetchRuntimes: () => dispatch(actions.functions.fetchRuntimes()),
    navigateToFunction: (name: string, namespace: string) =>
      dispatch(push(`/functions/ns/${namespace}/${name}`)),
  };
}

export default connect(mapStateToProps, mapDispatchToProps)(FunctionList);
