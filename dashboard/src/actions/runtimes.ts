import { Dispatch } from "redux";
import { createAction, getReturnOfExpression } from "typesafe-actions";

import KubelessConfig from "../shared/KubelessConfig";
import { IRuntime, IStoreState } from "../shared/types";

export const requestRuntimes = createAction("REQUEST_RUNTIMES");
export const receiveRuntimes = createAction("RECEIVE_RUNTIMES", (runtimes: IRuntime[]) => ({
  runtimes,
  type: "RECEIVE_RUNTIMES",
}));
const allActions = [requestRuntimes, receiveRuntimes].map(getReturnOfExpression);
export type RuntimesAction = typeof allActions[number];

export function fetchRuntimes() {
  return async (dispatch: Dispatch<IStoreState>) => {
    dispatch(requestRuntimes());
    const runtimeList = await KubelessConfig.getRuntimes();
    dispatch(receiveRuntimes(runtimeList));
    return runtimeList;
  };
}
