import { getType } from "typesafe-actions";

import actions from "../actions";
import { RuntimesAction } from "../actions/runtimes";
import { IRuntime } from "../shared/types";

export interface IRuntimeState {
  isFetching: boolean;
  items: IRuntime[];
}

const initialState: IRuntimeState = {
  isFetching: false,
  items: [],
};

const runtimesReducer = (
  state: IRuntimeState = initialState,
  action: RuntimesAction,
): IRuntimeState => {
  switch (action.type) {
    case getType(actions.runtimes.receiveRuntimes):
      const { runtimes } = action;
      return { ...state, isFetching: false, items: runtimes };
    case getType(actions.runtimes.requestRuntimes):
      return { ...state, isFetching: true };
    default:
      return state;
  }
};

export default runtimesReducer;
