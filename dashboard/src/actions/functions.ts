import { Dispatch } from "redux";
import { createAction, getReturnOfExpression } from "typesafe-actions";

import Function from "../shared/Function";
import { IFunction, IStoreState } from "../shared/types";

export const requestFunctions = createAction("REQUEST_FUNCTIONS");
export const receiveFunctions = createAction("RECEIVE_FUNCTIONS", (functions: IFunction[]) => ({
  functions,
  type: "RECEIVE_FUNCTIONS",
}));
const allActions = [requestFunctions, receiveFunctions].map(getReturnOfExpression);
export type FunctionsAction = typeof allActions[number];

export function fetchFunctions(namespace: string) {
  return async (dispatch: Dispatch<IStoreState>) => {
    dispatch(requestFunctions());
    const functionList = await Function.list();
    dispatch(receiveFunctions(functionList.items));
    return functionList;
  };
}
