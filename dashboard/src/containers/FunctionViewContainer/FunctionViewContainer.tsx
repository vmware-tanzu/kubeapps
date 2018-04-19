import { connect } from "react-redux";
import { Dispatch } from "redux";

import actions from "../../actions";
import FunctionView from "../../components/FunctionView";
import { IFunction, IStoreState } from "../../shared/types";

interface IRouteProps {
  match: {
    params: {
      name: string;
      namespace: string;
    };
  };
}

function mapStateToProps({ functions }: IStoreState, { match: { params } }: IRouteProps) {
  return {
    function: functions.selected.function,
    name: params.name,
    namespace: params.namespace,
    podName: functions.selected.podName,
  };
}

function mapDispatchToProps(dispatch: Dispatch<IStoreState>) {
  return {
    deleteFunction: (n: string, ns: string) => dispatch(actions.functions.deleteFunction(n, ns)),
    getFunction: (n: string, ns: string) => dispatch(actions.functions.getFunction(n, ns)),
    getPodName: (fn: IFunction) => dispatch(actions.functions.getPodName(fn)),
    updateFunction: (fn: IFunction) =>
      dispatch(actions.functions.updateFunction(fn.metadata.name, fn.metadata.namespace, fn)),
  };
}

export default connect(mapStateToProps, mapDispatchToProps)(FunctionView);
