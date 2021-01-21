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

  const chartItem2 = {
    id: "bar",
    attributes: {
      name: "bar",
      description: "",
      category: "",
      repo: { name: "bar", namespace: "chart-namespace" },
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

  it("requestCharts (without page) sets records", () => {
    const state = chartsReducer(undefined, {
      type: getType(actions.charts.requestCharts) as any,
    });
    expect(state).toEqual({
      ...initialState,
      isFetching: true,
      records: initialState.records.set(1, false),
    });
  });

  it("requestCharts (with page) sets records", () => {
    const state = chartsReducer(undefined, {
      type: getType(actions.charts.requestCharts) as any,
      payload: 2,
    });
    expect(state).toEqual({
      ...initialState,
      isFetching: true,
      records: initialState.records.set(2, false),
    });
  });

  it("single receiveCharts (first page) should be returned", () => {
    const state = chartsReducer(undefined, {
      type: getType(actions.charts.receiveCharts) as any,
      payload: { items: [chartItem], page: 1, totalPages: 3 } as IReceiveChartsActionPayload,
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

  it("single receiveCharts (middle page) having visited the previous ones should be ignored", () => {
    const state = chartsReducer(
      { ...initialState, page: 2, records: initialState.records.set(1, true).set(2, false) },
      {
        type: getType(actions.charts.receiveCharts) as any,
        payload: { items: [chartItem], page: 2, totalPages: 3 } as IReceiveChartsActionPayload,
      },
    );
    expect(state).toEqual({
      ...initialState,
      isFetching: false,
      hasFinishedFetching: false,
      items: [chartItem],
      page: 3,
      records: initialState.records
        .set(1, true)
        .set(2, true)
        .set(3, false),
    });
  });

  it("single receiveCharts (middle page) not visiting the previous ones should be ignored", () => {
    const state = chartsReducer(
      { ...initialState, page: 2, records: initialState.records.set(1, false).set(2, false) },
      {
        type: getType(actions.charts.receiveCharts) as any,
        payload: { items: [chartItem], page: 2, totalPages: 3 } as IReceiveChartsActionPayload,
      },
    );
    expect(state).toEqual({
      ...initialState,
      isFetching: false,
      hasFinishedFetching: false,
      items: [chartItem],
      page: 3,
      records: initialState.records.set(1, false).set(2, false),
    });
  });

  it("single receiveCharts (last page) not incrementing page", () => {
    const state = chartsReducer(
      {
        ...initialState,
        page: 3,
        records: initialState.records
          .set(1, true)
          .set(2, true)
          .set(3, false),
      },
      {
        type: getType(actions.charts.receiveCharts) as any,
        payload: { items: [chartItem], page: 3, totalPages: 3 } as IReceiveChartsActionPayload,
      },
    );
    expect(state).toEqual({
      ...initialState,
      isFetching: false,
      hasFinishedFetching: true,
      items: [chartItem],
      page: 3,
      records: initialState.records
        .set(1, true)
        .set(2, true)
        .set(3, true),
    });
  });

  it("two receiveCharts should add items (no dups)", () => {
    const state1 = chartsReducer(undefined, {
      type: getType(actions.charts.receiveCharts) as any,
      payload: { items: [chartItem], page: 1, totalPages: 2 } as IReceiveChartsActionPayload,
    });
    const state2 = chartsReducer(state1, {
      type: getType(actions.charts.receiveCharts) as any,
      payload: { items: [chartItem2], page: 2, totalPages: 2 } as IReceiveChartsActionPayload,
    });
    expect(state2).toEqual({
      ...initialState,
      isFetching: false,
      hasFinishedFetching: true,
      items: [chartItem, chartItem2],
      page: 2,
      records: initialState.records.set(1, true).set(2, true),
    });
    expect(state2.items.length).toBe(2);
  });

  it("two receiveCharts should add items (remove dups)", () => {
    const state1 = chartsReducer(undefined, {
      type: getType(actions.charts.receiveCharts) as any,
      payload: { items: [chartItem], page: 1, totalPages: 2 } as IReceiveChartsActionPayload,
    });
    const state2 = chartsReducer(state1, {
      type: getType(actions.charts.receiveCharts) as any,
      payload: { items: [chartItem], page: 2, totalPages: 2 } as IReceiveChartsActionPayload,
    });
    expect(state2).toEqual({
      ...initialState,
      isFetching: false,
      hasFinishedFetching: true,
      items: [chartItem],
      page: 2,
      records: initialState.records.set(1, true).set(2, true),
    });
    expect(state2.items.length).toBe(1);
  });

  it("requestCharts and receiveCharts with multiple pages", () => {
    const stateReq1 = chartsReducer(initialState, {
      type: getType(actions.charts.requestCharts) as any,
      payload: 1,
    });
    expect(stateReq1).toEqual({
      ...initialState,
      isFetching: true,
      hasFinishedFetching: false,
      items: [],
      page: 1,
      records: initialState.records.set(1, true),
    });
    const stateRec1 = chartsReducer(stateReq1, {
      type: getType(actions.charts.receiveCharts) as any,
      payload: { items: [chartItem], page: 1, totalPages: 3 } as IReceiveChartsActionPayload,
    });
    expect(stateRec1).toEqual({
      ...initialState,
      isFetching: false,
      items: [chartItem],
      hasFinishedFetching: false,
      page: 2,
      records: initialState.records.set(1, true).set(2, false),
    });
    const stateReq2 = chartsReducer(stateRec1, {
      type: getType(actions.charts.requestCharts) as any,
      payload: 2,
    });
    expect(stateReq2).toEqual({
      ...initialState,
      isFetching: true,
      hasFinishedFetching: false,
      items: [chartItem],
      page: 2,
      records: initialState.records
        .set(1, true)
        .set(2, true)
        .set(3, false),
    });
    const stateRec2 = chartsReducer(stateReq2, {
      type: getType(actions.charts.receiveCharts) as any,
      payload: { items: [chartItem2], page: 2, totalPages: 3 } as IReceiveChartsActionPayload,
    });
    expect(stateRec2).toEqual({
      ...initialState,
      isFetching: false,
      hasFinishedFetching: false,
      page: 3,
      items: [chartItem, chartItem2],
      records: initialState.records
        .set(1, true)
        .set(2, true)
        .set(3, false),
    });
    const stateReq3 = chartsReducer(stateRec2, {
      type: getType(actions.charts.requestCharts) as any,
      payload: 3,
    });
    expect(stateReq3).toEqual({
      ...initialState,
      isFetching: true,
      hasFinishedFetching: false,
      items: [chartItem, chartItem2],
      page: 3,
      records: initialState.records
        .set(1, true)
        .set(2, true)
        .set(3, false),
    });
    const stateRec3 = chartsReducer(stateReq3, {
      type: getType(actions.charts.receiveCharts) as any,
      payload: { items: [], page: 3, totalPages: 3 } as IReceiveChartsActionPayload,
    });
    expect(stateRec3).toEqual({
      ...initialState,
      isFetching: false,
      hasFinishedFetching: true,
      page: 3,
      items: [chartItem, chartItem2],
      records: initialState.records
        .set(1, true)
        .set(2, true)
        .set(3, true),
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

  it("resetRequestCharts resets to the initial", () => {
    const state = chartsReducer(undefined, {
      type: getType(actions.charts.resetRequestCharts) as any,
    });
    expect(state).toEqual({
      ...initialState,
    });
  });

  it("errorChart resets to the initial state", () => {
    const state = chartsReducer(undefined, {
      type: getType(actions.charts.errorChart) as any,
    });
    expect(state).toEqual({
      ...initialState,
    });
  });
});
