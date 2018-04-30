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
export const errorFunctions = createAction(
  "ERROR_FUNCTIONS",
  (err: Error, op: "create" | "update" | "fetch" | "delete") => ({
    err,
    op,
    type: "ERROR_FUNCTIONS",
  }),
);
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
  requestRuntimes,
  receiveRuntimes,
  errorFunctions,
  selectFunction,
  setPodName,
].map(getReturnOfExpression);
export type FunctionsAction = typeof allActions[number];

export function fetchFunctions(ns?: string) {
  return async (dispatch: Dispatch<IStoreState>) => {
    if (ns && ns === "_all") {
      ns = undefined;
    }
    dispatch(requestFunctions());
    try {
      const functionList = await Function.list(ns);
      dispatch(receiveFunctions(functionList.items));
    } catch (e) {
      dispatch(errorFunctions(e, "fetch"));
    }
  };
}

export function getFunction(name: string, namespace: string) {
  return async (dispatch: Dispatch<IStoreState>) => {
    dispatch(requestFunctions());
    try {
      const f = await Function.get(name, namespace);
      dispatch(selectFunction(f));
    } catch (e) {
      dispatch(errorFunctions(e, "fetch"));
    }
  };
}

export function createFunction(name: string, namespace: string, spec: IFunction["spec"]) {
  return async (dispatch: Dispatch<IStoreState>) => {
    try {
      await Function.create(name, namespace, spec);
      return true;
    } catch (e) {
      dispatch(errorFunctions(e, "create"));
      return false;
    }
  };
}

export function deleteFunction(name: string, namespace: string) {
  return async (dispatch: Dispatch<IStoreState>) => {
    try {
      await Function.delete(name, namespace);
      return true;
    } catch (e) {
      dispatch(errorFunctions(e, "delete"));
      return false;
    }
  };
}

export function updateFunction(name: string, namespace: string, newFn: IFunction) {
  return async (dispatch: Dispatch<IStoreState>) => {
    try {
      const f = await Function.update(name, namespace, newFn);
      dispatch(selectFunction(f));
    } catch (e) {
      dispatch(errorFunctions(e, "update"));
    }
  };
}

export function getPodName(fn: IFunction) {
  return async (dispatch: Dispatch<IStoreState>) => {
    try {
      const name = await Function.getPodName(fn);
      if (name) {
        dispatch(setPodName(name));
      }
    } catch (e) {
      dispatch(errorFunctions(e, "fetch"));
    }
  };
}

export function fetchRuntimes() {
  return async (dispatch: Dispatch<IStoreState>) => {
    dispatch(requestRuntimes());
    try {
      const runtimeList = await KubelessConfig.getRuntimes();
      dispatch(receiveRuntimes(runtimeList));
      return runtimeList;
    } catch (e) {
      dispatch(errorFunctions(e, "fetch"));
    }
  };
}
