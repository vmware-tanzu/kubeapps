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
    function: functions.selected,
    name: params.name,
    namespace: params.namespace,
  };
}

function mapDispatchToProps(
  dispatch: Dispatch<IStoreState>,
  { match: { params: { name, namespace } } }: IRouteProps,
) {
  return {
    deleteFunction: () => dispatch(actions.functions.deleteFunction(name, namespace)),
    getFunction: () => dispatch(actions.functions.getFunction(name, namespace)),
    updateFunction: (fn: IFunction) =>
      dispatch(actions.functions.updateFunction(name, namespace, fn)),
  };
}

export default connect(mapStateToProps, mapDispatchToProps)(FunctionView);
