import { LOCATION_CHANGE, LocationChangeAction } from "connected-react-router";
import { getType } from "typesafe-actions";

import actions from "../actions";
import { FunctionsAction } from "../actions/functions";
import { IFunction, IRuntime } from "../shared/types";

export interface IFunctionState {
  isFetching: boolean;
  items: IFunction[];
  runtimes: IRuntime[];
  errors: {
    create?: Error;
    delete?: Error;
    fetch?: Error;
    update?: Error;
  };
  selected: {
    function?: IFunction;
    podName?: string;
  };
}

const initialState: IFunctionState = {
  errors: {},
  isFetching: false,
  items: [],
  runtimes: [],
  selected: {},
};

const functionsReducer = (
  state: IFunctionState = initialState,
  action: FunctionsAction | LocationChangeAction,
): IFunctionState => {
  switch (action.type) {
    case getType(actions.functions.receiveFunctions):
      const { functions } = action;
      return { ...state, isFetching: false, items: functions };
    case getType(actions.functions.requestFunctions):
      return { ...state, isFetching: true };
    case getType(actions.functions.errorFunctions):
      return {
        ...state,
        // don't reset the fetch error
        errors: { fetch: state.errors.fetch, [action.op]: action.err },
        isFetching: false,
      };
    case getType(actions.functions.selectFunction):
      const { f } = action;
      return {
        ...state,
        isFetching: false,
        selected: { ...state.selected, function: f, podName: undefined },
      };
    case getType(actions.functions.setPodName):
      const { name } = action;
      return { ...state, isFetching: false, selected: { ...state.selected, podName: name } };
    case getType(actions.functions.receiveRuntimes):
      const { runtimes } = action;
      return { ...state, isFetching: false, runtimes };
    case getType(actions.functions.requestRuntimes):
      return { ...state, isFetching: true };
    case LOCATION_CHANGE:
      return {
        ...state,
        errors: {},
        isFetching: false,
        selected: {},
      };
    default:
      return state;
  }
};

export default functionsReducer;
