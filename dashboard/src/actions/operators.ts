import { ThunkAction } from "redux-thunk";
import { ActionType, createAction } from "typesafe-actions";

import { Operators } from "../shared/Operators";
import { IStoreState } from "../shared/types";

export const checkingOLM = createAction("CHECKING_OLM");
export const OLMInstalled = createAction("OLM_INSTALLED");
export const OLMNotInstalled = createAction("OLM_NOT_INSTALLED");

const actions = [checkingOLM, OLMInstalled, OLMNotInstalled];

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
