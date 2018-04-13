import { connect } from "react-redux";
import { push } from "react-router-redux";
import { Dispatch } from "redux";

import actions from "../../actions";
import FunctionList from "../../components/FunctionList";
import { IFunction, IStoreState } from "../../shared/types";

interface IRouteProps {
  match: {
    params: {
      namespace: string;
    };
  };
}

function mapStateToProps({ functions, runtimes }: IStoreState, { match: { params } }: IRouteProps) {
  return {
    functions: functions.items,
    namespace: params.namespace,
    runtimes: runtimes.items,
  };
}

function mapDispatchToProps(dispatch: Dispatch<IStoreState>) {
  return {
    deployFunction: (name: string, namespace: string, spec: IFunction["spec"]) =>
      dispatch(actions.functions.createFunction(name, namespace, spec)),
    fetchFunctions: (namespace: string) => dispatch(actions.functions.fetchFunctions(namespace)),
    fetchRuntimes: (namespace: string) => dispatch(actions.runtimes.fetchRuntimes()),
    navigateToFunction: (name: string, namespace: string) =>
      dispatch(push(`/functions/${namespace}/${name}`)),
  };
}

export default connect(mapStateToProps, mapDispatchToProps)(FunctionList);
