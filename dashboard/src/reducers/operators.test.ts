import { getType } from "typesafe-actions";
import actions from "../actions";

import operatorReducer from "./operators";
import { IOperatorsState } from "./operators";

describe("catalogReducer", () => {
  let initialState: IOperatorsState;

  beforeEach(() => {
    initialState = {
      isFetching: false,
      isOLMInstalled: false,
      operators: [],
    };
  });

  describe("operators", () => {
    const actionTypes = {
      checkingOLM: getType(actions.operators.checkingOLM),
      OLMInstalled: getType(actions.operators.OLMInstalled),
      OLMNotInstalled: getType(actions.operators.OLMNotInstalled),
    };

    describe("reducer actions", () => {
      it("sets isFetching when checking if the OLM is installed", () => {
        expect(
          operatorReducer(undefined, {
            type: actionTypes.checkingOLM as any,
          }),
        ).toEqual({ ...initialState, isFetching: true });
      });

      it("unsets isFetching and mark OLM as installed", () => {
        const state = operatorReducer(undefined, {
          type: actionTypes.checkingOLM as any,
        });
        expect(state).toEqual({ ...initialState, isFetching: true });
        expect(
          operatorReducer(undefined, {
            type: actionTypes.OLMInstalled as any,
          }),
        ).toEqual({ ...initialState, isOLMInstalled: true });
      });

      it("unsets isFetching and mark OLM as not installed", () => {
        const state = operatorReducer(undefined, {
          type: actionTypes.checkingOLM as any,
        });
        expect(state).toEqual({ ...initialState, isFetching: true });
        expect(
          operatorReducer(undefined, {
            type: actionTypes.OLMNotInstalled as any,
          }),
        ).toEqual({ ...initialState, isOLMInstalled: false });
      });
    });
  });

  // TODO(andresmgot): getOperators tests
});
