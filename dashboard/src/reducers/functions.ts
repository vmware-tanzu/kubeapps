import { getType } from "typesafe-actions";

import actions from "../actions";
import { FunctionsAction } from "../actions/functions";
import { IFunction } from "../shared/types";

export interface IFunctionState {
  isFetching: boolean;
  items: IFunction[];
  selected: {
    function?: IFunction;
    podName?: string;
  };
}

const initialState: IFunctionState = {
  isFetching: false,
  items: [],
  selected: {},
};

const functionsReducer = (
  state: IFunctionState = initialState,
  action: FunctionsAction,
): IFunctionState => {
  switch (action.type) {
    case getType(actions.functions.receiveFunctions):
      const { functions } = action;
      return { ...state, isFetching: false, items: functions };
    case getType(actions.functions.requestFunctions):
      return { ...state, isFetching: true };
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
    default:
      return state;
  }
};

export default functionsReducer;
