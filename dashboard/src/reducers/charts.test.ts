import { getType } from "typesafe-actions";
import actions from "../actions";

import { IChart, IChartCategory, IChartState } from "../shared/types";
import chartsReducer from "./charts";

describe("chartReducer", () => {
  let initialState: IChartState;
  const chartItem = {
    id: "foo",
    attributes: {
      name: "foo",
      description: "",
      category: "",
      repo: { name: "foo", namespace: "chart-namespace" },
    },
    relationships: { latestChartVersion: { data: { app_version: "v1.0.0" } } },
  } as IChart;
  const chartCategoryItem = {
    name: "foo",
    count: 1,
  } as IChartCategory;

  beforeEach(() => {
    initialState = {
      isFetching: false,
      items: [],
      categories: [],
      selected: {
        versions: [],
      },
      deployed: {},
      page: 1,
      size: 0,
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
        type: getType(actions.namespace.setNamespaceState) as any,
      }),
    ).toEqual({ ...initialState });
  });

  it("requestChartsCategories", () => {
    const state = chartsReducer(undefined, {
      type: getType(actions.charts.requestChartsCategories) as any,
    });
    expect(state).toEqual({
      ...initialState,
      isFetching: true,
    });
  });

  it("receiveChartCategories", () => {
    const state = chartsReducer(undefined, {
      type: getType(actions.charts.receiveChartCategories) as any,
      payload: [chartCategoryItem],
    });
    expect(state).toEqual({
      ...initialState,
      isFetching: false,
      categories: [chartCategoryItem],
    });
  });

  it("errorChartCatetories", () => {
    const state = chartsReducer(undefined, {
      type: getType(actions.charts.errorChartCatetories) as any,
    });
    expect(state).toEqual({
      ...initialState,
      isFetching: false,
      categories: [],
    });
  });

  it("single requestCharts without query", () => {
    const state = chartsReducer(undefined, {
      type: getType(actions.charts.requestCharts) as any,
    });
    expect(state).toEqual({
      ...initialState,
      isFetching: true,
    });
  });

  it("single requestCharts with query", () => {
    const state = chartsReducer(undefined, {
      type: getType(actions.charts.requestCharts) as any,
      payload: "query",
    });
    expect(state).toEqual({
      ...initialState,
      isFetching: true,
    });
  });

  it("single receiveCharts without query", () => {
    const state = chartsReducer(undefined, {
      type: getType(actions.charts.receiveCharts) as any,
      payload: [chartItem],
    });
    expect(state).toEqual({
      ...initialState,
      isFetching: false,
      items: [chartItem],
    });
  });

  it("single receiveCharts with query", () => {
    const state = chartsReducer(undefined, {
      type: getType(actions.charts.receiveCharts) as any,
      payload: [chartItem],
    });
    expect(state).toEqual({
      ...initialState,
      isFetching: false,
      items: [chartItem],
    });
  });

  it("two mixed receiveCharts should override items", () => {
    const state1 = chartsReducer(undefined, {
      type: getType(actions.charts.receiveCharts) as any,
      payload: [chartItem],
    });
    const state2 = chartsReducer(state1, {
      type: getType(actions.charts.receiveCharts) as any,
      payload: [],
    });
    expect(state2).toEqual({
      ...initialState,
      isFetching: false,
      items: [],
    });
  });

  it("two mixed receiveCharts and then errorChart", () => {
    const state1 = chartsReducer(undefined, {
      type: getType(actions.charts.receiveCharts) as any,
      payload: [chartItem],
    });
    const state2 = chartsReducer(state1, {
      type: getType(actions.charts.receiveCharts) as any,
    });
    const state3 = chartsReducer(state2, {
      type: getType(actions.charts.errorChart) as any,
    });
    expect(state3).toEqual({
      ...initialState,
      isFetching: false,
      items: [],
    });
  });
});
