import { getType } from "typesafe-actions";
import actions from "../actions";

import { IClusterServiceClass } from "shared/ClusterServiceClass";
import catalogReducer, { IServiceCatalogState } from "./catalog";

describe("catalogReducer", () => {
  let initialState: IServiceCatalogState;

  beforeEach(() => {
    initialState = {
      bindingsWithSecrets: { isFetching: false, list: [] },
      brokers: { isFetching: false, list: [] },
      classes: { isFetching: false, list: [] },
      errors: {},
      instances: { isFetching: false, list: [] },
      isChecking: true,
      isServiceCatalogInstalled: false,
      plans: { isFetching: false, list: [] },
    };
  });

  describe("classes", () => {
    const actionTypes = {
      requestClasses: getType(actions.catalog.requestClasses),
      receiveClasses: getType(actions.catalog.receiveClasses),
    };

    describe("reducer actions", () => {
      it("sets isFetching when requesting classes", () => {
        expect(
          catalogReducer(undefined, {
            type: actionTypes.requestClasses as any,
          }),
        ).toEqual({ ...initialState, classes: { isFetching: true, list: [] } });
      });

      it("restart isFetching and return the list of classes", () => {
        let state = catalogReducer(undefined, {
          type: actionTypes.requestClasses as any,
        });
        expect(state).toEqual({ ...initialState, classes: { isFetching: true, list: [] } });
        state = catalogReducer(state, {
          type: actionTypes.receiveClasses,
          payload: [{} as IClusterServiceClass],
        });
        expect(state).toEqual({ ...initialState, classes: { isFetching: false, list: [{}] } });
      });
    });
  });
});
