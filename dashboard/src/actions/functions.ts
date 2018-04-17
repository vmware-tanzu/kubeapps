import { Dispatch } from "redux";
import { createAction, getReturnOfExpression } from "typesafe-actions";

import Function from "../shared/Function";
import KubelessConfig from "../shared/KubelessConfig";
import { IFunction, IRuntime, IStoreState } from "../shared/types";

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
export const requestRuntimes = createAction("REQUEST_RUNTIMES");
export const receiveRuntimes = createAction("RECEIVE_RUNTIMES", (runtimes: IRuntime[]) => ({
  runtimes,
  type: "RECEIVE_RUNTIMES",
}));
const allActions = [
  requestFunctions,
  receiveFunctions,
  selectFunction,
  setPodName,
  requestRuntimes,
  receiveRuntimes,
].map(getReturnOfExpression);
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

export function fetchRuntimes() {
  return async (dispatch: Dispatch<IStoreState>) => {
    dispatch(requestRuntimes());
    const runtimeList = await KubelessConfig.getRuntimes();
    dispatch(receiveRuntimes(runtimeList));
    return runtimeList;
  };
}
