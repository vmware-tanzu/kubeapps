import configureMockStore from "redux-mock-store";
import thunk from "redux-thunk";
import { getType } from "typesafe-actions";

import actions from ".";
import { NotFoundError } from "../shared/types";

const mockStore = configureMockStore([thunk]);

let store: any;
const fetchOrig = window.fetch;
let fetchMock: jest.Mock;
let response: any;

beforeEach(() => {
  store = mockStore();
  fetchMock = jest.fn(() => {
    return {
      ok: true,
      json: () => {
        return { data: response };
      },
    };
  });
  window.fetch = fetchMock;
});

afterEach(() => {
  window.fetch = fetchOrig;
});

describe("fetchCharts", () => {
  it("fetches charts from a repo", async () => {
    response = [{ id: "foo" }];
    const expectedActions = [
      { type: getType(actions.charts.requestCharts) },
      { type: getType(actions.charts.receiveCharts), payload: response },
    ];
    await store.dispatch(actions.charts.fetchCharts("foo"));
    expect(store.getActions()).toEqual(expectedActions);
    expect(fetchMock.mock.calls[0][0]).toBe("api/chartsvc/v1/charts/foo");
  });

  it("returns a 404 error", async () => {
    const expectedActions = [
      { type: getType(actions.charts.requestCharts) },
      { type: getType(actions.charts.errorChart), payload: new NotFoundError("not found") },
    ];
    fetchMock = jest.fn(() => {
      return {
        ok: false,
        status: 404,
        json: () => {
          return { data: "not found" };
        },
      };
    });
    window.fetch = fetchMock;
    await store.dispatch(actions.charts.fetchCharts("foo"));
    expect(store.getActions()).toEqual(expectedActions);
  });

  it("returns a generic error", async () => {
    const expectedActions = [
      { type: getType(actions.charts.requestCharts) },
      { type: getType(actions.charts.errorChart), payload: new Error("something went wrong") },
    ];
    fetchMock = jest.fn(() => {
      return {
        ok: false,
        status: 500,
        json: () => {
          return { data: "something went wrong" };
        },
      };
    });
    window.fetch = fetchMock;
    await store.dispatch(actions.charts.fetchCharts("foo"));
    expect(store.getActions()).toEqual(expectedActions);
  });
});

describe("fetchChartVersions", () => {
  it("fetches chart versions", async () => {
    response = [{ id: "foo" }];
    const expectedActions = [
      { type: getType(actions.charts.requestCharts) },
      { type: getType(actions.charts.receiveChartVersions), payload: response },
    ];
    await store.dispatch(actions.charts.fetchChartVersions("foo"));
    expect(store.getActions()).toEqual(expectedActions);
    expect(fetchMock.mock.calls[0][0]).toBe("api/chartsvc/v1/charts/foo/versions");
  });
});

describe("getChartVersion", () => {
  it("gets a chart version", async () => {
    response = { id: "foo" };
    const expectedActions = [
      { type: getType(actions.charts.requestCharts) },
      { type: getType(actions.charts.selectChartVersion), payload: response },
    ];
    await store.dispatch(actions.charts.getChartVersion("foo", "1.0.0"));
    expect(store.getActions()).toEqual(expectedActions);
    expect(fetchMock.mock.calls[0][0]).toBe("api/chartsvc/v1/charts/foo/versions/1.0.0");
  });
});

describe("fetchChartVersionsAndSelectVersion", () => {
  it("fetches charts and select a version", async () => {
    response = [{ id: "foo", attributes: { version: "1.0.0" } }];
    const expectedActions = [
      { type: getType(actions.charts.requestCharts) },
      { type: getType(actions.charts.receiveChartVersions), payload: response },
      { type: getType(actions.charts.selectChartVersion), payload: response[0] },
    ];
    await store.dispatch(actions.charts.fetchChartVersionsAndSelectVersion("foo", "1.0.0"));
    expect(store.getActions()).toEqual(expectedActions);
    expect(fetchMock.mock.calls[0][0]).toBe("api/chartsvc/v1/charts/foo/versions");
  });

  it("returns a not found error", async () => {
    response = [{ id: "foo", attributes: { version: "1.0.0" } }];
    const expectedActions = [
      { type: getType(actions.charts.requestCharts) },
      { type: getType(actions.charts.errorChart), payload: new NotFoundError("not found") },
    ];
    fetchMock = jest.fn(() => {
      return {
        ok: false,
        status: 404,
        json: () => {
          return { data: "not found" };
        },
      };
    });
    window.fetch = fetchMock;
    await store.dispatch(actions.charts.fetchChartVersionsAndSelectVersion("foo", "1.0.0"));
    expect(store.getActions()).toEqual(expectedActions);
    expect(fetchMock.mock.calls[0][0]).toBe("api/chartsvc/v1/charts/foo/versions");
  });
});
