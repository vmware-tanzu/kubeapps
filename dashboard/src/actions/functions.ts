import { Dispatch } from "redux";
import { createAction, getReturnOfExpression } from "typesafe-actions";

import Function from "../shared/Function";
import { IFunction, IStoreState } from "../shared/types";

export const requestFunctions = createAction("REQUEST_FUNCTIONS");
export const receiveFunctions = createAction("RECEIVE_FUNCTIONS", (functions: IFunction[]) => ({
  functions,
  type: "RECEIVE_FUNCTIONS",
}));
export const selectFunction = createAction("SELECT_FUNCTION", (f: IFunction) => ({
  f,
  type: "SELECT_FUNCTION",
}));
export const setPodName = createAction("SET_FUNCTION_POD_NAME", (name: string) => ({
  name,
  type: "SET_FUNCTION_POD_NAME",
}));
const allActions = [requestFunctions, receiveFunctions, selectFunction, setPodName].map(
  getReturnOfExpression,
);
export type FunctionsAction = typeof allActions[number];

export function fetchFunctions(namespace: string) {
  return async (dispatch: Dispatch<IStoreState>) => {
    dispatch(requestFunctions());
    const functionList = await Function.list();
    dispatch(receiveFunctions(functionList.items));
    return functionList;
  };
}

export function getFunction(name: string, namespace: string) {
  return async (dispatch: Dispatch<IStoreState>) => {
    dispatch(requestFunctions());
    const f = await Function.get(name, namespace);
    dispatch(selectFunction(f));
    return f;
  };
}

export function createFunction(name: string, namespace: string, spec: IFunction["spec"]) {
  return async (dispatch: Dispatch<IStoreState>) => {
    return Function.create(name, namespace, spec);
  };
}

export function deleteFunction(name: string, namespace: string) {
  return async (dispatch: Dispatch<IStoreState>) => {
    return Function.delete(name, namespace);
  };
}

export function updateFunction(name: string, namespace: string, newFn: IFunction) {
  return async (dispatch: Dispatch<IStoreState>) => {
    const f = await Function.update(name, namespace, newFn);
    dispatch(selectFunction(f));
    return f;
  };
}

export function getPodName(fn: IFunction) {
  return async (dispatch: Dispatch<IStoreState>) => {
    const name = await Function.getPodName(fn);
    if (name) {
      dispatch(setPodName(name));
    }
    return name;
  };
}
