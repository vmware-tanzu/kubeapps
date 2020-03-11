import { ThunkAction } from "redux-thunk";
import { ActionType, createAction } from "typesafe-actions";

import { Operators } from "../shared/Operators";
import { IPackageManifest, IStoreState } from "../shared/types";

export const checkingOLM = createAction("CHECKING_OLM");
export const OLMInstalled = createAction("OLM_INSTALLED");
export const OLMNotInstalled = createAction("OLM_NOT_INSTALLED");

export const requestOperators = createAction("REQUEST_OPERATORS");
export const receiveOperators = createAction("RECEIVE_OPERATORS", resolve => {
  return (operators: IPackageManifest[]) => resolve(operators);
});

export const requestOperator = createAction("REQUEST_OPERATOR");
export const receiveOperator = createAction("RECEIVE_OPERATOR", resolve => {
  return (operator: IPackageManifest) => resolve(operator);
});

export const errorOperators = createAction("ERROR_OPERATORS", resolve => {
  return (err: Error) => resolve(err);
});

const actions = [
  checkingOLM,
  OLMInstalled,
  OLMNotInstalled,
  requestOperators,
  receiveOperators,
  errorOperators,
  requestOperator,
  receiveOperator,
];

export type OperatorAction = ActionType<typeof actions[number]>;

export function checkOLMInstalled(): ThunkAction<
  Promise<boolean>,
  IStoreState,
  null,
  OperatorAction
> {
  return async dispatch => {
    dispatch(checkingOLM());
    const installed = await Operators.isOLMInstalled();
    installed ? dispatch(OLMInstalled()) : dispatch(OLMNotInstalled());
    return installed;
  };
}

export function getOperators(
  namespace: string,
): ThunkAction<Promise<void>, IStoreState, null, OperatorAction> {
  return async dispatch => {
    dispatch(requestOperators());
    try {
      const operators = await Operators.getOperators(namespace);
      const sortedOp = operators.sort((o1, o2) => (o1.metadata.name > o2.metadata.name ? 1 : -1));
      dispatch(receiveOperators(sortedOp));
    } catch (e) {
      dispatch(errorOperators(e));
    }
  };
}

export function getOperator(
  namespace: string,
  operatorName: string,
): ThunkAction<Promise<void>, IStoreState, null, OperatorAction> {
  return async dispatch => {
    dispatch(requestOperator());
    try {
      const operator = await Operators.getOperator(namespace, operatorName);
      dispatch(receiveOperator(operator));
    } catch (e) {
      dispatch(errorOperators(e));
    }
  };
}
