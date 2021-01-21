import { getType } from "typesafe-actions";
import actions from "../actions";

import { IChart, IChartCategory, IChartState, IReceiveChartsActionPayload } from "../shared/types";
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
      hasFinishedFetching: false,
      items: [],
      categories: [],
      selected: {
        versions: [],
      },
      deployed: {},
      page: 1,
      size: 100,
      records: new Map<number, boolean>().set(1, false),
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
      payload: 1,
    });
    expect(state).toEqual({
      ...initialState,
      isFetching: true,
    });
  });

  it("single requestCharts with query", () => {
    const state = chartsReducer(undefined, {
      type: getType(actions.charts.requestCharts) as any,
      payload: 1,
    });
    expect(state).toEqual({
      ...initialState,
      isFetching: true,
    });
  });

  it("single receiveCharts no finished", () => {
    const state = chartsReducer(undefined, {
      type: getType(actions.charts.receiveCharts) as any,
      payload: { items: [chartItem], page: 1, totalPages: 2 } as IReceiveChartsActionPayload,
    });
    expect(state).toEqual({
      ...initialState,
      isFetching: false,
      hasFinishedFetching: false,
      items: [chartItem],
      page: 2,
      records: new Map<number, boolean>().set(1, true).set(2, false),
    });
  });

  it("single receiveCharts finished", () => {
    const state = chartsReducer(undefined, {
      type: getType(actions.charts.receiveCharts) as any,
      payload: { items: [chartItem], page: 1, totalPages: 1 } as IReceiveChartsActionPayload,
    });
    expect(state).toEqual({
      ...initialState,
      isFetching: false,
      hasFinishedFetching: true,
      items: [chartItem],
      page: 1,
      records: new Map<number, boolean>().set(1, true).set(2, false),
    });
  });

  it("two receiveCharts should add items", () => {
    const state1 = chartsReducer(undefined, {
      type: getType(actions.charts.receiveCharts) as any,
      payload: { items: [chartItem], page: 1, totalPages: 2 } as IReceiveChartsActionPayload,
    });
    const state2 = chartsReducer(state1, {
      type: getType(actions.charts.receiveCharts) as any,
      payload: { items: [], page: 2, totalPages: 2 } as IReceiveChartsActionPayload,
    });
    expect(state2).toEqual({
      ...initialState,
      isFetching: false,
      hasFinishedFetching: true,
      items: [chartItem],
      page: 2,
      records: new Map<number, boolean>()
        .set(1, true)
        .set(2, true)
        .set(3, false),
    });
  });

  it("two receiveCharts and then errorChart", () => {
    const state1 = chartsReducer(undefined, {
      type: getType(actions.charts.receiveCharts) as any,
      payload: { items: [chartItem], page: 1, totalPages: 1 } as IReceiveChartsActionPayload,
    });
    const state2 = chartsReducer(state1, {
      type: getType(actions.charts.receiveCharts) as any,
      payload: { items: [], page: 2, totalPages: 2 } as IReceiveChartsActionPayload,
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
