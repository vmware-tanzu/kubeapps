import { getType } from "typesafe-actions";
import actions from "../actions";

import { IChartState } from "../shared/types";
import chartsReducer from "./charts";

describe("chartReducer", () => {
  let initialState: IChartState;

  beforeEach(() => {
    initialState = {
      isFetching: false,
      items: [],
      selected: {
        versions: [],
      },
      deployed: {},
    };
  });
  const error = new Error("Boom");
  it("unsets an error when changing namespace", () => {
    const state = chartsReducer(undefined, {
      type: getType(actions.charts.errorChart) as any,
      payload: error,
    });
    expect(state).toEqual({
      ...initialState,
      isFetching: false,
      selected: {
        ...initialState.selected,
        error,
      },
    });

    expect(
      chartsReducer(undefined, {
        type: getType(actions.namespace.setNamespace) as any,
      }),
    ).toEqual({ ...initialState });
  });

  it("sets the initial state when changing namespace", () => {
    expect(
      chartsReducer(
        {
          ...initialState,
          isFetching: true,
          selected: {
            ...initialState.selected,
            error,
          },
        },
        {
          type: getType(actions.namespace.setNamespace) as any,
        },
      ),
    ).toEqual({ ...initialState });
  });
});
