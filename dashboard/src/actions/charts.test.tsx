import configureMockStore from "redux-mock-store";
import thunk from "redux-thunk";
import { getType } from "typesafe-actions";
import { axiosWithAuth } from "../shared/AxiosInstance";

import actions from ".";
import { NotFoundError } from "../shared/types";

const mockStore = configureMockStore([thunk]);

let axiosGetMock = jest.fn();
let store: any;
let response: any;

const namespace = "kubeapps-namespace";

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
  axiosWithAuth.get = axiosGetMock;
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
    await store.dispatch(actions.charts.fetchCharts(namespace, "foo"));
    expect(store.getActions()).toEqual(expectedActions);
    expect(axiosGetMock.mock.calls[0][0]).toBe(`api/assetsvc/v1/ns/${namespace}/charts/foo`);
  });

  it("returns a 404 error", async () => {
    const expectedActions = [
      { type: getType(actions.charts.requestCharts) },
      {
        type: getType(actions.charts.errorChart),
        payload: new NotFoundError("could not find chart"),
      },
    ];
    axiosGetMock = jest.fn(() => {
      throw new Error("could not find chart");
    });
    axiosWithAuth.get = axiosGetMock;
    await store.dispatch(actions.charts.fetchCharts(namespace, "foo"));
    expect(store.getActions()).toEqual(expectedActions);
  });

  it("returns a generic error", async () => {
    const expectedActions = [
      { type: getType(actions.charts.requestCharts) },
      { type: getType(actions.charts.errorChart), payload: new Error("something went wrong") },
    ];
    axiosGetMock = jest.fn(() => {
      throw new Error("something went wrong");
    });
    axiosWithAuth.get = axiosGetMock;
    await store.dispatch(actions.charts.fetchCharts(namespace, "foo"));
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
    await store.dispatch(actions.charts.fetchChartVersions("chart-namespace", "foo"));
    expect(store.getActions()).toEqual(expectedActions);
    expect(axiosGetMock.mock.calls[0][0]).toBe(
      `api/assetsvc/v1/ns/chart-namespace/charts/foo/versions`,
    );
  });
});

describe("getChartVersion", () => {
  it("gets a chart version", async () => {
    response = { id: "foo" };
    const expectedActions = [
      { type: getType(actions.charts.requestCharts) },
      {
        type: getType(actions.charts.selectChartVersion),
        payload: { chartVersion: response, schema: { data: response }, values: { data: response } },
      },
    ];
    await store.dispatch(actions.charts.getChartVersion(namespace, "foo", "1.0.0"));
    expect(store.getActions()).toEqual(expectedActions);
    expect(axiosGetMock.mock.calls[0][0]).toBe(
      `api/assetsvc/v1/ns/${namespace}/charts/foo/versions/1.0.0`,
    );
  });

  it("gets a chart version with tag", async () => {
    response = { id: "foo" };
    const expectedActions = [
      { type: getType(actions.charts.requestCharts) },
      {
        type: getType(actions.charts.selectChartVersion),
        payload: { chartVersion: response, schema: { data: response }, values: { data: response } },
      },
    ];
    await store.dispatch(actions.charts.getChartVersion(namespace, "foo", "1.0.0-alpha+1.2.3-beta2"));
    expect(store.getActions()).toEqual(expectedActions);
    expect(axiosGetMock.mock.calls[0][0]).toBe(
      `api/assetsvc/v1/ns/${namespace}/charts/foo/versions/1.0.0-alpha%2B1.2.3-beta2`,
    );
  });

  it("gets a chart version with values and schema", async () => {
    // Call to get the chart version
    axiosGetMock.mockImplementationOnce(() => {
      return {
        status: 200,
        data: { data: { id: "foo" } },
      };
    });
    // Call to get the chart values
    axiosGetMock.mockImplementationOnce(() => {
      return {
        status: 200,
        data: "foo: bar",
      };
    });
    // Call to get the chart schema
    axiosGetMock.mockImplementationOnce(() => {
      return {
        status: 200,
        data: { properties: "foo" },
      };
    });
    const expectedActions = [
      { type: getType(actions.charts.requestCharts) },
      {
        type: getType(actions.charts.selectChartVersion),
        payload: { chartVersion: { id: "foo" }, values: "foo: bar", schema: { properties: "foo" } },
      },
    ];
    await store.dispatch(actions.charts.getChartVersion(namespace, "foo", "1.0.0"));
    expect(store.getActions()).toEqual(expectedActions);
  });

  it("returns an empty schema if not found", async () => {
    // Call to get the chart version
    axiosGetMock.mockImplementationOnce(() => {
      return {
        status: 200,
        data: {
          data: { id: "foo" },
        },
      };
    });
    // Call to get the chart values
    axiosGetMock.mockImplementationOnce(() => {
      throw new NotFoundError();
    });
    // Call to get the chart schema
    axiosGetMock.mockImplementationOnce(() => {
      throw new NotFoundError();
    });
    const expectedActions = [
      { type: getType(actions.charts.requestCharts) },
      {
        type: getType(actions.charts.selectChartVersion),
        payload: { chartVersion: { id: "foo" }, values: "", schema: {} },
      },
    ];
    await store.dispatch(actions.charts.getChartVersion(namespace, "foo", "1.0.0"));
    expect(store.getActions()).toEqual(expectedActions);
  });

  it("dispatches an error if it's unexpected", async () => {
    // Call to get the chart version
    axiosGetMock.mockImplementationOnce(() => {
      return {
        status: 200,
        data: {
          data: { id: "foo" },
        },
      };
    });
    // Call to get the chart values
    axiosGetMock.mockImplementationOnce(() => {
      throw new Error("Boom!");
    });
    // Call to get the chart schema
    axiosGetMock.mockImplementationOnce(() => {
      throw new Error("Boom!");
    });

    const expectedActions = [
      { type: getType(actions.charts.requestCharts) },
      { type: getType(actions.charts.errorChart), payload: new Error("Boom!") },
    ];
    await store.dispatch(actions.charts.getChartVersion(namespace, "foo", "1.0.0"));
    expect(store.getActions()).toEqual(expectedActions);
  });
});

describe("fetchChartVersionsAndSelectVersion", () => {
  it("fetches charts and select a version", async () => {
    response = [{ id: "foo", attributes: { version: "1.0.0" } }];
    const expectedActions = [
      { type: getType(actions.charts.requestCharts) },
      { type: getType(actions.charts.receiveChartVersions), payload: response },
      { type: getType(actions.charts.selectChartVersion), payload: { chartVersion: response[0] } },
    ];
    await store.dispatch(actions.charts.fetchChartVersionsAndSelectVersion("chart-namespace", "foo", "1.0.0"));
    expect(store.getActions()).toEqual(expectedActions);
    expect(axiosGetMock.mock.calls[0][0]).toBe(
      `api/assetsvc/v1/ns/chart-namespace/charts/foo/versions`,
    );
  });

  it("returns a not found error", async () => {
    response = [{ id: "foo", attributes: { version: "1.0.0" } }];
    const expectedActions = [
      { type: getType(actions.charts.requestCharts) },
      {
        type: getType(actions.charts.errorChart),
        payload: new NotFoundError("could not find chart"),
      },
    ];
    axiosGetMock = jest.fn(() => {
      throw new Error("could not find chart");
    });
    axiosWithAuth.get = axiosGetMock;
    await store.dispatch(actions.charts.fetchChartVersionsAndSelectVersion("chart-namespace", "foo", "1.0.0"));
    expect(store.getActions()).toEqual(expectedActions);
    expect(axiosGetMock.mock.calls[0][0]).toBe(
      `api/assetsvc/v1/ns/chart-namespace/charts/foo/versions`,
    );
  });
});

describe("getDeployedChartVersion", () => {
  it("should request a deployed chart", async () => {
    response = { id: "foo" };
    const expectedActions = [
      { type: getType(actions.charts.requestDeployedChartVersion) },
      {
        type: getType(actions.charts.receiveDeployedChartVersion),
        payload: { chartVersion: response, schema: { data: response }, values: { data: response } },
      },
    ];
    await store.dispatch(actions.charts.getDeployedChartVersion(namespace, "foo", "1.0.0"));
    expect(store.getActions()).toEqual(expectedActions);
  });
});
