import configureMockStore from "redux-mock-store";
import thunk from "redux-thunk";
import { getType } from "typesafe-actions";
import { axios } from "../shared/AxiosInstance";

import actions from ".";
import { NotFoundError } from "../shared/types";

const mockStore = configureMockStore([thunk]);

let axiosGetMock = jest.fn();
let store: any;
let response: any;

beforeEach(() => {
  store = mockStore();
  axiosGetMock.mockImplementation(() => {
    return {
      status: 200,
      data: {
        data: response,
      },
    };
  });
  axios.get = axiosGetMock;
});

afterEach(() => {
  jest.resetAllMocks();
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
    expect(axiosGetMock.mock.calls[0][0]).toBe("api/chartsvc/v1/charts/foo");
  });

  it("returns a 404 error", async () => {
    const expectedActions = [
      { type: getType(actions.charts.requestCharts) },
      { type: getType(actions.charts.errorChart), payload: new NotFoundError("not found") },
    ];
    axiosGetMock = jest.fn(() => {
      return {
        ok: false,
        status: 404,
        data: "not found",
      };
    });
    axios.get = axiosGetMock;
    await store.dispatch(actions.charts.fetchCharts("foo"));
    expect(store.getActions()).toEqual(expectedActions);
  });

  it("returns a generic error", async () => {
    const expectedActions = [
      { type: getType(actions.charts.requestCharts) },
      { type: getType(actions.charts.errorChart), payload: new Error("something went wrong") },
    ];
    axiosGetMock = jest.fn(() => {
      return {
        ok: false,
        status: 500,
        data: "something went wrong",
      };
    });
    axios.get = axiosGetMock;
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
    expect(axiosGetMock.mock.calls[0][0]).toBe("api/chartsvc/v1/charts/foo/versions");
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
    expect(axiosGetMock.mock.calls[0][0]).toBe("api/chartsvc/v1/charts/foo/versions/1.0.0");
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
    expect(axiosGetMock.mock.calls[0][0]).toBe("api/chartsvc/v1/charts/foo/versions");
  });

  it("returns a not found error", async () => {
    response = [{ id: "foo", attributes: { version: "1.0.0" } }];
    const expectedActions = [
      { type: getType(actions.charts.requestCharts) },
      { type: getType(actions.charts.errorChart), payload: new NotFoundError("not found") },
    ];
    axiosGetMock = jest.fn(() => {
      return {
        ok: false,
        status: 404,
        data: "not found",
      };
    });
    axios.get = axiosGetMock;
    await store.dispatch(actions.charts.fetchChartVersionsAndSelectVersion("foo", "1.0.0"));
    expect(store.getActions()).toEqual(expectedActions);
    expect(axiosGetMock.mock.calls[0][0]).toBe("api/chartsvc/v1/charts/foo/versions");
  });
});
